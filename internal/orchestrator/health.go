package orchestrator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	//"github.com/balaji-balu/margo-hello-world/ent"
	//"github.com/balaji-balu/margo-hello-world/internal/natsbroker"
	//"github.com/balaji-balu/margo-hello-world/internal/orchestrator"
	"github.com/balaji-balu/margo-hello-world/pkg/model"
)

var (
	nodeCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "node_count",
		Help: "The total number of nodes",
	})
)

func (lo *LocalOrchestrator) MonitorHealthandStatusFromEN(coUrl string) {

	// Subscribe to health
	go func() {
		log.Println("lo with siteid:", fmt.Sprintf("health.%s.*", lo.Config.Site))
		subHealth := fmt.Sprintf("health.%s.*", lo.Config.Site)
		err := lo.nc.Subscribe2(subHealth, func(h model.HealthMsg) {
			log.Printf("[LO] health from %s runtime=%s", h.NodeID, h.Runtime)
			err := lo.CreateEdgeNode(lo.RootCtx, h)
			if err != nil {
				log.Printf("[LO] error saving the Node:", err)
			}
			//nodeCount.Set(float64(len(orchestrator.GetAllNodes(db))))
			//if fsm.GetState() == shared.Discovering  {
			//	fsm.Transition(shared.Running)
			//}
		})
		if err != nil {
			log.Println("subscribe error:", err)
		} else {
			log.Println("subscribed to", subHealth)
		}
		lo.nc.Flush()
		log.Println("subscription ready for", subHealth)

	}()

	go func() {
		subStatus := fmt.Sprintf("status.%s.*", lo.Config.Site)
		err := lo.nc.Subscribe4(subStatus, func(s model.DeploymentStatus) {
			//log.Printf("[LO] status %s from %s: success=%v, msg=%s",
			//	s.DeploymentID, s.NodeID, s.Status, s.Message)
			log.Println("[LO] status %s: state=%v",
				s.DeploymentID, s.Status.State,
			)
	
			lo.UpdateDeploymentRecord(lo.RootCtx, s)
			
			//forward to CO
			forwardToCO(coUrl, s)

		})
		if err != nil {
			log.Fatal("[LO] failed to subscribe to status updates:", err)
		}

	}()

}

func forwardToCO(baseurl string, report model.DeploymentStatus) {
	url := fmt.Sprintf("%s/deployments/%s/status", baseurl, report.DeploymentID)
	payload, err := json.Marshal(report)
	if err != nil {
		log.Println("[LO] failed to marshal report:", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Println("[LO] failed to send report to CO:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[LO] CO returned status %d: %s", resp.StatusCode, string(body))
		return
	}

	log.Printf("[LO] ✅ Report forwarded to CO successfully (deployment_id=%s)", 
					report.DeploymentID)
}
