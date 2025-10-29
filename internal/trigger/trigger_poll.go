//go:build poll
package trigger

import (
    "os"
    "log"
    "github.com/balaji/hello/pkg/config"
)

func New(cfg config.Config) Trigger {
    log.Println("Creating poll trigger with github token:", os.Getenv(cfg.Trigger.TokenEnv) )
    return &PollTrigger{
        RepoOwner: cfg.Trigger.RepoOwner,
        RepoName:  cfg.Trigger.RepoName,
        Path:      cfg.Trigger.YamlPath,
        Token:     os.Getenv(cfg.Trigger.TokenEnv),
        Interval:  cfg.Trigger.Interval,
    }
}
