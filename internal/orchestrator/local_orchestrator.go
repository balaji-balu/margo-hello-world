package orchestrator

import (
    "context"
    "log"
    "os"
    "fmt"
    "sync"
    "time"
    
    "github.com/balaji-balu/margo-hello-world/ent"
    // _ "github.com/lib/pq"  // enables the 'postgres' driver
    // "entgo.io/ent/dialect"
    // "entgo.io/ent/dialect/sql"

    //"github.com/looplab/fsm"
    //fsmloader "github.com/balaji-balu/margo-hello-world/internal/fsm"
    //"github.com/balaji-balu/margo-hello-world/internal/fsmloader"
    . "github.com/balaji-balu/margo-hello-world/internal/config"
    "github.com/balaji-balu/margo-hello-world/pkg/deployment"
    "github.com/looplab/fsm"
    "go.uber.org/zap"
    "github.com/gin-gonic/gin"
    "net/http"
//"fmt"

)

type LoConfig struct {
    Owner string
    Repo  string
    Token string
    Path  string
    NatsUrl string
    Site string
}

// type Journal struct {
//     ETag        string
//     LastSuccess string
// }

type LocalOrchestrator struct {
    Config  LoConfig
    Journal Journal
    Client  *ent.Client
    EOPort string
    //Hosturls []string
    Hosts []string
    FSM *fsm.FSM
    //machine *fsm.FSMWrapper
    //Callback CallbackHandler
    //fsm     *fsm.FSM
    logger *zap.Logger
    rb *ResultBus
    eventCh chan string
    RootCtx context.Context
}

type RegisterRequest struct {
	EdgeURL string `json:"edge_url" binding:"required"`
}

func (lo *LocalOrchestrator) ResultBus() *ResultBus {
    return lo.rb
}

func (lo *LocalOrchestrator) Machine() *fsm.FSM {
    return lo.FSM
}

func New(ctx context.Context, cfg *Config, logger *zap.Logger) *LocalOrchestrator {
    rb := NewResultBus()

    return &LocalOrchestrator{
        Config: LoConfig{
            //Owner: cfg..Owner,
            Repo:  cfg.Git.Repo,
            NatsUrl: cfg.NATS.URL,
            Token : os.Getenv("GITHUB_TOKEN"),
            Site : cfg.Server.Site,
        },
        logger: logger,
        rb: rb,
        eventCh: make(chan string, 20),
        RootCtx: ctx,
    }
}

/*
func TobedeletedNewLocalOrchestrator(fsmPath string) *LocalOrchestrator {
    ctx := context.Background()
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize zap: %v", err)
	}
	defer logger.Sync()

    machine, err := fsmloader.LoadFSMConfig("./configs/lo_fsm_resilient.yaml", logger)
	if err != nil {
		logger.Fatal("Failed to initialize FSM", zap.Error(err))
	}

    lo := &LocalOrchestrator{
        Config: LoConfig{
            Owner: os.Getenv("GITHUB_OWNER"),
            Repo:  os.Getenv("GITHUB_REPO"),
            Token: os.Getenv("GITHUB_TOKEN"),
            Path: os.Getenv("GITHUB_PATH"),
        },
        machine: machine,
        logger: logger,
    }

    loader := fsmloader.NewLoader(ctx, logger)

	// Attach orchestrator callbacks
	log.Println("FSM initialized at state:", machine.Current())

    lo.LoadJournal()
    return lo
}
*/

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

    lo.logger.Info("Node registration", zap.String("node_id", req.EdgeURL))

    if lo.FSM.Current() == "waiting_for_nodes" {
        lo.FSM.Event(c.Request.Context(), "node_registered")
    }

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
        if err := lo.FSM.Event(ctx, "edge_accepted"); err != nil {
            fmt.Println("❌ Error:", err)
            return
        }
    } else if status.Status == deployment.StatusStarted {
    } else if status.Status == deployment.StatusRunning {
    } else if status.Status == deployment.StatusCompleted {
        // if err := lo.machine.Event(ctx, "edge_rejected"); err != nil {
        //     fmt.Println("❌ Error:", err)
        //     return
        // }
    } else if status.Status == deployment.StatusFailed {
        log.Println("deploy failed status edge_rejected calling")
        if err := lo.FSM.Event(ctx, "edge_rejected"); err != nil {
            fmt.Println("❌ Error:", err)
            return
        }
    } else {


    }

    log.Println("deploy app name", status.AppName) 
    log.Println("deploy status", status.Status) 
    log.Println("deploy message", status.Message)

}

func (lo *LocalOrchestrator) DeployToEdges() {
    edges := []string{"chennai-edge1", "tiruvannamalai-edge2"}
    var wg sync.WaitGroup

    for _, node := range edges {
        wg.Add(1)
        go func(node string) {
            defer wg.Done()
            lo.logger.Info("Deploying to node", zap.String("node", node))
            time.Sleep(2 * time.Second)

            // simulate outcome
            if node == "chennai-edge1" {
                lo.rb.Publish(node, "success", nil)
            } else {
                lo.rb.Publish(node, "fail", fmt.Errorf("network issue"))
            }
        }(node)
    }

    go func() {
        wg.Wait()
        lo.rb.Publish("all", "done", nil)
    }()
}


func (l *LocalOrchestrator) StartEventDispatcher(ctx context.Context) {
    l.logger.Info("Starting FSM event dispatcher...")
    for {
        select {
        case <-ctx.Done():
            l.logger.Info("Event dispatcher stopped")
            return
        case ev := <-l.eventCh:
            l.logger.Info("Processing FSM event", 
                    zap.String("event", ev),
                    zap.String("state", l.FSM.Current()))
            if err := l.FSM.Event(ctx, ev); err != nil {
                l.logger.Error("FSM event failed", 
                    zap.String("event", ev),
                    zap.String("state", l.FSM.Current()),
                    zap.Error(err))
            } else {
                l.logger.Info("FSM event completed", zap.String("event", ev))
            }
        }
    }
}

func (l *LocalOrchestrator) TriggerEvent(ev string) {
    select {
    case l.eventCh <- ev:
        l.logger.Info("Queued FSM event", zap.String("event", ev))
    default:
        l.logger.Warn("Event queue full, dropping event", zap.String("event", ev))
    }
}

/*
func (lo *LocalOrchestrator) TriggerFSM(event string) {
    if err := lo.machine.Event(context.Background(), event); err != nil {
        lo.logger.Warn("Event failed", zap.String("event", event), zap.Error(err))
    }
}
*/