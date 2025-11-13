package edgenode

import (
	"context"
	"os"
	//"bytes"
	//"encoding/json"
	"os/exec"
	"fmt"
	"log"
	//"net/http"
	"time"
	"math/rand"

	"go.uber.org/zap"

	//"github.com/balaji-balu/margo-hello-world/internal/config"
	"github.com/balaji-balu/margo-hello-world/pkg/deployment"
	"github.com/balaji-balu/margo-hello-world/pkg/model"
	"github.com/balaji-balu/margo-hello-world/internal/ocifetch"
	"github.com/balaji-balu/margo-hello-world/internal/natsbroker"
)

type EdgeNode struct {
	SiteID, NodeID, Runtime, Region string
	//localOrchestratorURL string
	nc *natsbroker.Broker
	logger *zap.Logger
	ctx context.Context
}

func NewEdgeNode(ctx context.Context, 
		siteID, nodeID, runtime string, nc *natsbroker.Broker, logger *zap.Logger) *EdgeNode {
	return &EdgeNode{
		//localOrchestratorURL: localOrchestratorURL,
		ctx: ctx,
		nc : nc,
		logger: logger,
		SiteID: siteID,
		NodeID: nodeID,
		Runtime: runtime,
		//Region: cfg.Server.Region,
	}
}

func (en *EdgeNode) Start() {
	go en.startHeartbeat()
	go en.startDeployListener()
}

func (en *EdgeNode) startHeartbeat() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			msg := model.HealthMsg{
				NodeID:     en.NodeID,
				SiteID:     en.SiteID,
				CPUPercent: rand.Float64() * 20,
				MemMB:      50 + rand.Float64()*20,
				Timestamp:  time.Now().Unix(),
				Runtime:    en.Runtime,
				//Region:     en.Region,
			}
			//data, _ := json.Marshal(msg)
			subj := fmt.Sprintf("health.%s.%s", en.SiteID, en.NodeID)
			if err := en.nc.Publish(subj, msg); err != nil {
				log.Println("Heart publish failed:", err)
			} else {
				log.Println("Heart msg published", subj, msg)
				en.nc.Flush()
			}
			//log.Println("Heart msg published", id)
		}
	}()
}

// TBD: workload Update, delete
func (en *EdgeNode) startDeployListener() {
	subj := fmt.Sprintf("site.%s.deploy.%s", en.SiteID, en.NodeID)
	en.nc.Subscribe3(subj, func(req deployment.DeployRequest) {
		log.Printf("req:", req)
		log.Printf("[EN %s] deploy request received (%s)", en.NodeID, en.Runtime)
		success := true
		statusMsg := "Deployment successful"

		if en.Runtime == "wasm" {
			for _, w := range req.WasmImages {
				log.Println("[EN] Running wasm:", w)
				//err := deployWorkload(w, runtime)
				//if err != nil {
					//fsm.Transition(shared.Deploying)
					success = false
					statusMsg = fmt.Sprintf("WASM deploy failed for %s: %v", w)
					break
				//}
			}
		} else {
			for i, img := range req.ContainerImages {
				log.Println("[EN] Running container:", img)
				err := en.deployContainerd(req, fmt.Sprintf("component-%d", i))
				if err != nil {
					//fsm.Transition(shared.Deploying)
					success = false
					statusMsg = fmt.Sprintf("Container deploy failed for %s: %v", img, err)
					break
				}
			}
		}

		// Publish deployment status
		status := deployment.DeploymentReport{
			DeploymentID: req.DeploymentID,
			NodeID:       en.NodeID,
			SiteID:       en.SiteID,
			Status:       "completed",
			Message:      statusMsg,
			Timestamp:    time.Now().Format(time.RFC3339),
		}
		//data, _ := json.Marshal(status)
		statusSubj := fmt.Sprintf("status.%s.%s", en.SiteID, en.NodeID)
		if err := en.nc.Publish(statusSubj, status); err != nil {
			log.Println("[EN] failed to publish status:", err)
		} else {
			log.Printf("[EN] status published: %s (success=%v)", statusSubj, success)
		}
		en.nc.Flush()
		//fsm.Transition(shared.Running)
	})
}

// Create new oci based container
// TBD: if container is already running
//
func (en *EdgeNode) deployContainerd(
	req deployment.DeployRequest, compName string,
) error {
	deploymentID := req.DeploymentID
	log.Println("Deploying Containerd:", req.Image)

    // 1️⃣ PENDING
    en.UpdateStatus(deploymentID, string(model.StatePending), compName, nil)

    // 2️⃣ INSTALLING (OCI Fetch)
    en.UpdateStatus(deploymentID, string(model.StateInstalling), compName, nil)
    fetcher := ocifetch.Fetcher{
        Image: req.Image,
        Tag:   req.Revision,
        Token: os.Getenv("GITHUB_TOKEN"),
    }
    if err := fetcher.Fetch(en.ctx); err != nil {
        en.UpdateStatus(deploymentID, string(model.StateFailed), compName, err)
        //http.Error(w, fmt.Sprintf("OCI fetch failed: %v", err), 500)
        return err
    }

    // 3️⃣ Run container
    image := fmt.Sprintf("%s:%s", req.Image, req.Revision)
    cmd := exec.Command("docker", "run", "-d", "--name", req.Revision, image)
    out, err := cmd.CombinedOutput()
    if err != nil {
        en.UpdateStatus(deploymentID, string(model.StateFailed), compName, fmt.Errorf("docker run failed: %v, %s", err, out))
        //http.Error(w, string(out), 500)
        return err
    }
	
    // 4️⃣ INSTALLED (success)
    en.UpdateStatus(deploymentID, string(model.StateInstalled), compName, nil)
    //w.WriteHeader(http.StatusOK)
    //w.Write([]byte(fmt.Sprintf("✅ Deployment completed: %s", image)))	
	return nil
}

func (en *EdgeNode) UpdateStatus(
	deploymentID string, state string, compName string, err error,
) {
    ds := model.DeploymentStatus{
        DeploymentID: deploymentID,
        APIVersion:   "deployment.margo/v1",
        Kind:         "DeploymentStatus",
    }

    ds.Status.State = state
    if err != nil {
        ds.Status.Error = model.StatusError{
            Code:    "DEPLOYMENT_FAILED",
            Message: err.Error(),
        }
    }

    if compName != "" {
        ds.Components = []model.DeploymentComponent{
            {Name: compName, State: state, Error: ds.Status.Error},
        }
    }

	statusSubj := fmt.Sprintf("status.%s.%s", en.SiteID, en.NodeID)
    en.nc.Publish(statusSubj, ds)
}

