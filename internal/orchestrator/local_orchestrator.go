package orchestrator

import (
    "log"
    "os"
    "fmt"
    
    "github.com/balaji-balu/margo-hello-world/ent"
    // _ "github.com/lib/pq"  // enables the 'postgres' driver
    // "entgo.io/ent/dialect"
    // "entgo.io/ent/dialect/sql"

    //"github.com/looplab/fsm"
    fsmloader "github.com/balaji-balu/margo-hello-world/internal/fsm"
    "github.com/balaji-balu/margo-hello-world/pkg/deployment"
    "github.com/looplab/fsm"
    "go.uber.org/zap"
    "github.com/gin-gonic/gin"
    "net/http"
//"fmt"

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
    //Hosturls []string
    Hosts []string
    machine *fsm.FSM
    logger *zap.Logger
}

type RegisterRequest struct {
	EdgeURL string `json:"edge_url" binding:"required"`
}

func NewLocalOrchestrator(fsmPath string) *LocalOrchestrator {
    
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize zap: %v", err)
	}
        
    m, err := fsmloader.LoadFSM(fsmPath)
    if err != nil {
        return nil //, fmt.Errorf("failed to load LO FSM: %w", err)
    }

    log.Printf("LO FSM initialized at state: %s", m.Current())

    lo := &LocalOrchestrator{
        Config: Config{
            Owner: os.Getenv("GITHUB_OWNER"),
            Repo:  os.Getenv("GITHUB_REPO"),
            Token: os.Getenv("GITHUB_TOKEN"),
            Path: os.Getenv("GITHUB_PATH"),
        },
        machine: m,
        logger: logger,
    }

    log.Printf("LO FSM initialized at state: %s", lo.machine.Current())

    lo.LoadJournal()
    return lo
}

// RegisterRequest handles registration of an Edge Node.
func (lo *LocalOrchestrator) RegisterRequest(c *gin.Context) {
	var req RegisterRequest

	// Bind JSON payload
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("❌ Invalid register request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Add to host list
	log.Println("✅ Registering edge:", req.EdgeURL)
	lo.Hosts = append(lo.Hosts, req.EdgeURL)

	// Respond success
	c.JSON(http.StatusOK, gin.H{
		"message":  "edge registered successfully",
		"edge_url": req.EdgeURL,
	})
}

func (lo *LocalOrchestrator) DeployStatus(c *gin.Context) {
    var status deployment.DeploymentReport
    ctx := c.Request.Context()


    if err := c.ShouldBindJSON(&status); err != nil {
        log.Println("❌ Invalid deploy status request:", err)
        return
    }

    if status.Status == deployment.StatusSuccess {
        if err := lo.machine.Event(c, "edge_accepted"); err != nil {
            fmt.Println("❌ Error:", err)
            return
        }
    } else if status.Status == deployment.StatusStarted {
    } else if status.Status == deployment.StatusRunning {
    } else if status.Status == deployment.StatusCompleted {
        if err := lo.machine.Event(ctx, "edge_rejected"); err != nil {
            fmt.Println("❌ Error:", err)
            return
        }
    } else if status.Status == deployment.StatusFailed {
        log.Println("deploy failed status edge_rejected calling")
        if err := lo.machine.Event(ctx, "edge_rejected"); err != nil {
            fmt.Println("❌ Error:", err)
            return
        }
    } else {


    }


    log.Println("deploy app name", status.AppName) 
    log.Println("deploy status", status.Status) 
    log.Println("deploy message", status.Message)

}

// func (lo *LocalOrchestrator) ApplyDeployment(data []byte) {
//     log.Println("Delegating deployment to internal/api/handlers")
//     //handlers.CreateDeployment(data)
// }
