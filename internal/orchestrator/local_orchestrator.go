package orchestrator

import (
    //"log"
    "os"
    
    //"github.com/balaji/hello/internal/api/handlers"
    "github.com/balaji/hello/ent"
    // _ "github.com/lib/pq"  // enables the 'postgres' driver
    // "entgo.io/ent/dialect"
    // "entgo.io/ent/dialect/sql"
)

type Config struct {
    Owner string
    Repo  string
    Token string
    Path  string
}

// type Journal struct {
//     ETag        string
//     LastSuccess string
// }

type LocalOrchestrator struct {
    Config  Config
    Journal Journal
    Client  *ent.Client
    EOPort string
}

func NewLocalOrchestrator() *LocalOrchestrator {
    lo := &LocalOrchestrator{
        Config: Config{
            Owner: os.Getenv("GITHUB_OWNER"),
            Repo:  os.Getenv("GITHUB_REPO"),
            Token: os.Getenv("GITHUB_TOKEN"),
            Path: os.Getenv("GITHUB_PATH"),
        },
    }
    lo.LoadJournal()
    return lo
}

// func (lo *LocalOrchestrator) ApplyDeployment(data []byte) {
//     log.Println("Delegating deployment to internal/api/handlers")
//     //handlers.CreateDeployment(data)
// }
