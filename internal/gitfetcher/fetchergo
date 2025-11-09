// internal/gitfetcher/fetcher.go
package gitfetcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type GitFetcher struct {
	RepoURL  string
	Branch   string
	LocalDir string
}

func (g *GitFetcher) CloneOrPull() error {
	if _, err := os.Stat(filepath.Join(g.LocalDir, ".git")); os.IsNotExist(err) {
		log.Printf("🌀 Cloning %s branch=%s", g.RepoURL, g.Branch)
		_, err := git.PlainClone(g.LocalDir, false, &git.CloneOptions{
			URL:           g.RepoURL,
			ReferenceName: plumbing.NewBranchReferenceName(g.Branch),
			SingleBranch:  true,
			Depth:         1,
		})
		return err
	}

	repo, err := git.PlainOpen(g.LocalDir)
	if err != nil {
		return fmt.Errorf("open repo failed: %w", err)
	}

	wt, _ := repo.Worktree()
	err = wt.Pull(&git.PullOptions{RemoteName: "origin"})
	if err == git.NoErrAlreadyUpToDate {
		log.Println("✅ Repo up-to-date")
		return nil
	}
	return err
}
