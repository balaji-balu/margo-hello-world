package orchestrator

import (
	"context"

	"github.com/balaji-balu/margo-hello-world/internal/gitobserver"
	"github.com/balaji-balu/margo-hello-world/internal/natsbroker"
)

func Run(ctx context.Context, cfg *Config) error {
    natsClient, err := natsbroker.New(cfg.NATS.URL)
    if err != nil {
        return err
    }

    observer := gitobserver.New(cfg.Git.Repo, cfg.Git.Branch, cfg.SiteFilter)
    observer.OnChange(func(event gitobserver.GitEvent) {
        _ = natsClient.Publish("git.desiredstate.changed", event)
    })

    return observer.Start(ctx)
}
