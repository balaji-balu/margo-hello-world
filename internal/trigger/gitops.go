//go:build gitops
package trigger

import "fmt"

type GitOpsTrigger struct{}

func (g *GitOpsTrigger) Start() error {
    fmt.Println("GitOps mode enabled — external controller manages sync.")
    return nil
}

func (g *GitOpsTrigger) Stop() {
	log.Println("[GitOpsTrigger] Stopped GitOps sync.")
}