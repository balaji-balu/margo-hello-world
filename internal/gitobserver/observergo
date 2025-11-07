package gitobserver

import (
	"context"
	"log"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type GitEvent struct {	
	Site        string
	CommitHash  string
	EventType   string
	Timestamp   time.Time
}

type GitObserver struct {
    repoURL     string
    branch      string
    siteFilter  string
    lastHash    string
    onChange    func(GitEvent)
}

func New(repo, branch, filter string) *GitObserver {
    return &GitObserver{repoURL: repo, branch: branch, siteFilter: filter}
}

func (g *GitObserver) OnChange(cb func(GitEvent)) {
    g.onChange = cb
}

func (g *GitObserver) Start(ctx context.Context) error {
    for {
        hash, err := g.getLatestCommit()
        if err != nil {
            log.Printf("git error: %v", err)
        }

        if hash != g.lastHash {
            event := GitEvent{
                Site:        g.siteFilter,
                CommitHash:  hash,
                EventType:   "file_changed",
                Timestamp:   time.Now(),
            }
            g.onChange(event)
            g.lastHash = hash
        }

        select {
        case <-time.After(20 * time.Second):
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (g *GitObserver) getLatestCommit() (string, error) {
    repo, err := git.PlainCloneContext(context.Background(), "/tmp/repo", true, &git.CloneOptions{
        URL:           g.repoURL,
        ReferenceName: plumbing.NewBranchReferenceName(g.branch),
        SingleBranch:  true,
        Depth:         1,
    })
    if err == git.ErrRepositoryAlreadyExists {
        repo, _ = git.PlainOpen("/tmp/repo")
        _ = repo.Fetch(&git.FetchOptions{RemoteName: "origin"})
    }
    ref, _ := repo.Head()
    return ref.Hash().String(), nil
}
