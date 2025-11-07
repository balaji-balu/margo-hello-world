package gitobserver

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitEvent struct {
	Site       string
	CommitHash string
	EventType  string
	Timestamp  time.Time
}

type DeploymentChange struct {
	DeploymentID string
	FilePath     string
	YAMLContent  string
}

type RepoWatcher struct {
	RepoURL   string
	Branch    string
	Interval  time.Duration
	OnChange  func(commit string, deployments []DeploymentChange)
	Token     string
	LocalPath string
	stopCh    chan struct{}
}

func New(repoURL, branch string, interval time.Duration) *RepoWatcher {
	return &RepoWatcher{
		RepoURL:  repoURL,
		Branch:   branch,
		Interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (r *RepoWatcher) Stop() {
	close(r.stopCh)
}

func (r *RepoWatcher) Start(siteID string) error {
	var repo *git.Repository
	var err error

	log.Println("[GitObserver] Starting watcher for", r.Branch)

	if r.Token == "" {
		r.Token = os.Getenv("GITHUB_TOKEN")
	}

	r.LocalPath = "/tmp/repo-cache"
	if _, statErr := os.Stat(r.LocalPath); os.IsNotExist(statErr) {
		log.Println("[GitObserver] Cloning fresh repo...")
		cloneOpts := &git.CloneOptions{
			URL:           r.RepoURL,
			ReferenceName: plumbing.NewBranchReferenceName(r.Branch),
			Depth:         1,
		}
		if r.Token != "" {
			cloneOpts.Auth = &http.BasicAuth{
				Username: "git",
				Password: r.Token,
			}
		}
		repo, err = git.PlainClone(r.LocalPath, false, cloneOpts)
	} else {
		log.Println("[GitObserver] Opening existing repo...")
		repo, err = git.PlainOpen(r.LocalPath)
	}

	if err != nil {
		return fmt.Errorf("git open/clone failed: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("worktree failed: %w", err)
	}

	log.Println("[GitObserver] Checking out branch:", r.Branch)

	// Initialize last commit
	head, _ := repo.Head()
	lastCommit := head.Hash().String()

	siteDir := fmt.Sprintf("%s/", siteID)
	log.Println("[GitObserver] Watching site:", siteDir)

	for {
		select {
		case <-time.After(r.Interval):
			pullOpts := &git.PullOptions{
				RemoteName: "origin",
				Force:      true,
			}
			if r.Token != "" {
				pullOpts.Auth = &http.BasicAuth{
					Username: "git",
					Password: r.Token,
				}
			}

			if err := wt.Pull(pullOpts); err != nil && err != git.NoErrAlreadyUpToDate {
				log.Println("[GitObserver] Pull error:", err)
				continue
			}

			ref, _ := repo.Head()
			commit := ref.Hash().String()

			if commit != lastCommit {
				changedFiles := getChangedFiles(repo, lastCommit, commit)
				lastCommit = commit // move early to avoid duplicate triggers

				var siteFiles []string
				for _, f := range changedFiles {
					if siteDir == "" || strings.HasPrefix(f, siteDir) {
						siteFiles = append(siteFiles, f)
					}
				}

				var deployments []DeploymentChange
				for _, f := range siteFiles {
					if strings.HasSuffix(f, "desiredstate.yaml") {
						log.Println("[xxxxxxxxxxxxxx] Deployment detected:", f)
						deploymentID := ExtractDeploymentID(f)
						buf, err := FetchYamlFile(repo, f)
						if err != nil {
							log.Println("[GitObserver] Error reading file:", err)
							continue
						}
						deployments = append(deployments, DeploymentChange{
							DeploymentID: deploymentID,
							FilePath:     f,
							YAMLContent:  buf.String(),
						})
					}
				}

				if len(deployments) > 0 && r.OnChange != nil {
					r.OnChange(commit, deployments)
				}
			}

		case <-r.stopCh:
			log.Printf("[GitObserver] Stopping watcher for %s", r.Branch)
			return nil
		}
	}
}

func getChangedFiles(repo *git.Repository, oldCommit, newCommit string) []string {
	if oldCommit == "" {
		return []string{}
	}
	oldC, _ := repo.CommitObject(plumbing.NewHash(oldCommit))
	newC, _ := repo.CommitObject(plumbing.NewHash(newCommit))
	patch, _ := oldC.Patch(newC)

	var files []string
	for _, stat := range patch.Stats() {
		files = append(files, stat.Name)
	}
	return files
}

func ExtractDeploymentID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return ""
}

func FetchYamlFile(repo *git.Repository, filePath string) (*strings.Builder, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	entry, err := tree.FindEntry(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found in tree: %s", filePath)
	}

	blob, err := repo.BlobObject(entry.Hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob: %w", err)
	}

	reader, err := blob.Reader()
	if err != nil {
		return nil, fmt.Errorf("failed to open blob reader: %w", err)
	}
	defer reader.Close()

	buf := new(strings.Builder)
	if _, err := io.Copy(buf, reader); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}
	return buf, nil
}
