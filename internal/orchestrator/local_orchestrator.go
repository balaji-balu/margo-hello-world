package orchestrator

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	//"encoding/json"

	"github.com/balaji-balu/margo-hello-world/ent"
	// _ "github.com/lib/pq"  // enables the 'postgres' driver
	// "entgo.io/ent/dialect"
	// "entgo.io/ent/dialect/sql"

	//"github.com/looplab/fsm"
	//fsmloader "github.com/balaji-balu/margo-hello-world/internal/fsm"
	//"github.com/balaji-balu/margo-hello-world/internal/fsmloader"
	. "github.com/balaji-balu/margo-hello-world/internal/config"
	"github.com/balaji-balu/margo-hello-world/internal/gitobserver"
	"github.com/balaji-balu/margo-hello-world/internal/natsbroker"
	"github.com/balaji-balu/margo-hello-world/pkg/deployment"
	"github.com/gin-gonic/gin"
	"github.com/looplab/fsm"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"net"
	"net/http"
	//"github.com/balaji-balu/margo-hello-world/pkg/deployment"
	//"fmt"
)

type EventType string

const (
	EventGitPolled      = "EventGitPolled"
	EventNetworkChange  = "EventNetworkChange"
	EventDeployComplete = "EventDeployComplete"
)

type Event struct {
	Name string
	Data interface{}
	Time time.Time
}

type GitPolledPayload struct {
	Commit      string
	Deployments []gitobserver.DeploymentChange
}

type NetworkChangePayload struct {
	OldMode string
	NewMode string
}

type LoConfig struct {
	Owner   string
	Repo    string
	Token   string
	Path    string
	NatsUrl string
	Site    string
}

// type Journal struct {
//     ETag        string
//     LastSuccess string
// }

type LocalOrchestrator struct {
	Config  LoConfig
	Journal Journal
	Client  *ent.Client
	EOPort  string
	//Hosturls []string
	Hosts []string
	FSM   *fsm.FSM
	//machine *fsm.FSMWrapper
	//Callback CallbackHandler
	//fsm     *fsm.FSM
	//logger *zap.Logger
	rb *ResultBus
	//eventCh chan string
	RootCtx context.Context
	db      *bolt.DB
	nc      *natsbroker.Broker

	eventCh     chan Event
	logger      *zap.Logger
	currentMode string
	cancelFunc  context.CancelFunc // for stopping running process
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

func New(
	ctx context.Context,
	cfg *Config,
	db *bolt.DB,
	nc *natsbroker.Broker,
	logger *zap.Logger,
) *LocalOrchestrator {
	rb := NewResultBus()

	return &LocalOrchestrator{
		Config: LoConfig{
			//Owner: cfg..Owner,
			Repo:    cfg.Git.Repo,
			NatsUrl: cfg.NATS.URL,
			Token:   os.Getenv("GITHUB_TOKEN"),
			Site:    cfg.Server.Site,
		},
		logger: logger,
		rb:     rb,
		//ventChan: make(chan Event, 20),
		eventCh: make(chan Event, 20),
		RootCtx: ctx,
		db:      db,
		nc:      nc,
	}
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

func (lo *LocalOrchestrator) DeployToEdges(
	id string,
	dep deployment.ApplicationDeployment,
) {
	targetEdges := GetAllNodes(lo.db)
	log.Println("Deploying to edges:", targetEdges)

	var wg sync.WaitGroup

	for _, node := range targetEdges {
		wg.Add(1)
		go func(node EdgeNode) {
			defer wg.Done()
			lo.logger.Info("Deploying to node", zap.String("node_id", node.NodeID))

			// Extract repo + revision from spec
			repoVal := dep.Spec.DeploymentProfile.Components[0].Properties.Repository
			revision := dep.Spec.DeploymentProfile.Components[0].Properties.Revision

			req := deployment.DeployRequest{
				DeploymentID: dep.Metadata.Annotations.ID,
				GitRepoURL:   repoVal,
				Revision:     revision,
				//RuntimeFilter: dep.Spec.DeploymentProfile.Type,
				//Region:       node.Region,
			}
			req.ContainerImages = append(req.ContainerImages, repoVal)

			// ✅ create a record for this node in DB as "pending"
			rec := map[string]string{
				node.NodeID: "pending",
			}
			SaveDeploymentRecord(lo.db, req.DeploymentID, rec)

			log.Println("Deploying to node req:", req)
			// // ✅ prepare payload and subject for NATS publish
			// payload, err := json.Marshal(req)
			// if err != nil {
			//     lo.logger.Error("Failed to marshal deploy request", zap.Error(err))
			//     return
			// }

			subj := fmt.Sprintf("site.%s.deploy.%s", node.SiteID, node.NodeID)
			if err := lo.nc.Publish(subj, req); err != nil {
				lo.logger.Error("Failed to publish deploy message", zap.Error(err))
				return
			}

			lo.logger.Info("Deployment message published", zap.String("subject", subj))
		}(node)
	}

	go func() {
		wg.Wait()
		lo.rb.Publish("all", "done", nil)
	}()
}

func networkStable() bool {
	conn, err := net.DialTimeout("tcp", "github.com:443", 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func (lo *LocalOrchestrator) DetectMode() string {
	// if networkStable() {
	//     return "pushpreferred"
	// }
	// if time.Since(lo.Journal.LastSuccess).Hours() > 2 {
	//     return "offline"
	// }
	return "adaptive"
}

func (l *LocalOrchestrator) StartNetworkMonitor(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	currentMode := l.currentMode

	for {
		select {
		case <-ticker.C:
			newMode := l.DetectMode()
			if newMode != currentMode {
				payload := NetworkChangePayload{OldMode: currentMode, NewMode: newMode}
				l.TriggerEvent(ctx, EventNetworkChange, payload)
				currentMode = newMode
				l.currentMode = newMode
			}

		case <-ctx.Done():
			return
		}
	}
}

func (l *LocalOrchestrator) handleGitPolled(data GitPolledPayload) {
	l.logger.Info("Handling GitPolled",
		zap.String("commit", data.Commit),
		zap.Int("deployments", len(data.Deployments)),
	)

	for _, d := range data.Deployments {
		l.logger.Info("Deploying", zap.String("deployment_id", d.DeploymentID))
		var dep deployment.ApplicationDeployment
		if err := yaml.Unmarshal([]byte(d.YAMLContent), &dep); err != nil {
			l.logger.Error("Failed to unmarshal deployment YAML", zap.Error(err))
			continue
		}
		// Call your deployer here
		l.DeployToEdges(d.DeploymentID, dep)
	}
}

func (l *LocalOrchestrator) handleNetworkChange(ctx context.Context, cfg LoConfig, data NetworkChangePayload) {
	l.logger.Info("Network mode change detected",
		zap.String("old_mode", data.OldMode),
		zap.String("new_mode", data.NewMode),
	)

	// Cancel any existing process (pull/push/offline)
	if l.cancelFunc != nil {
		l.cancelFunc()
		l.logger.Info("Stopped previous mode process", zap.String("mode", data.OldMode))
	}

	// Start the new mode
	ctxNew, cancel := context.WithCancel(ctx)
	l.cancelFunc = cancel

	switch data.NewMode {
	case "adaptive":
		go l.StartPullMode(ctxNew, cfg)
	case "pushpreferred":
		go l.StartPushMode(ctxNew, cfg)
	case "offline":
		go l.StartOfflineMode(ctxNew, cfg)
	default:
		l.logger.Warn("Unknown mode", zap.String("mode", data.NewMode))
	}
}

func (l *LocalOrchestrator) StartEventDispatcher(ctx context.Context) {
	l.logger.Info("Starting FSM event dispatcher...")

	for {
		select {
		case <-ctx.Done():
			l.logger.Info("Event dispatcher stopped")
			return
		case ev := <-l.eventCh:
			l.logger.Info("Processing FSM event", zap.Any("event", ev))

			switch ev.Name {
			case EventGitPolled:
				if data, ok := ev.Data.(GitPolledPayload); ok {
					l.handleGitPolled(data)
				}
			case EventNetworkChange:
				if data, ok := ev.Data.(NetworkChangePayload); ok {
					l.handleNetworkChange(ctx, l.Config, data)
				}
			}

			if err := l.FSM.Event(ctx, ev.Name); err != nil {
				l.logger.Error("FSM event failed",
					zap.Any("event", ev),
					zap.String("state", l.FSM.Current()),
					zap.Error(err))
			}
		}
	}
}

/*
	func (l *LocalOrchestrator) StartEventDispatcher(ctx context.Context) {
	    l.logger.Info("Starting FSM event dispatcher...")
	    for {
	        select {
	        case <-ctx.Done():
	            l.logger.Info("Event dispatcher stopped")
	            return
	        case ev := <-l.eventCh:
	            l.logger.Info("Processing FSM event",
	                    zap.Any("event", ev),
	                    zap.String("state", l.FSM.Current()))
	            if err := l.FSM.Event(ctx, ev.Name); err != nil {
	                l.logger.Error("FSnM event failed",
	                    zap.Any("event", ev),
	                    zap.String("state", l.FSM.Current()),
	                    zap.Error(err))
	            } else {
	                l.logger.Info("FSM event completed", zap.Any("event", ev))
	            }
	        }
	    }
	}
*/
func (l *LocalOrchestrator) TriggerEvent(ctx context.Context, name string, data interface{}) {
	ev := Event{Name: name, Data: data, Time: time.Now()}

	select {
	case l.eventCh <- ev:
		l.logger.Info("Queued FSM event", zap.String("event", name))
	default:
		l.logger.Warn("Event queue full, dropping event", zap.String("event", name))
	}
}

/*
func (lo *LocalOrchestrator) pickTargetNodes(req deployment.DeployRequest) ([]EdgeNode, error) {

    // := DeployAPIRequest{

	// choose targets
	targets := []EdgeNode{}
	if len(req.TargetNodes) > 0 {
		// load nodes from db
		all := GetAllNodes(lo.db)
		for _, n := range all {
			for _, id := range req.TargetNodes {
				if n.NodeID == id {
					targets = append(targets, n);
					break
				}
			}
		}
	} else {
		all := GetAllNodes(lo.db)
		rt := req.RuntimeFilter
		if rt == "" && len(req.WasmImages) > 0 { rt = "wasm" }
		if rt == "" && len(req.ContainerImages) > 0 { rt = "containerd" }
		targets = PickNodes(all, rt, req.Region, 10)
	}

	if len(targets) == 0 {
		//http.Error(w, "no target nodes available", 400)
		log.Println("[LO] no target nodes available")
		return nil, fmt.Errorf("no target nodes available")
	}

    return targets, nil

}

*/
