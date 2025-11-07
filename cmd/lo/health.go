package main

import (
	//"encoding/json"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/balaji-balu/margo-hello-world/pkg/deployment"

	//"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	bolt "go.etcd.io/bbolt"

	"github.com/balaji-balu/margo-hello-world/internal/natsbroker"
	"github.com/balaji-balu/margo-hello-world/internal/orchestrator"
	"github.com/balaji-balu/margo-hello-world/pkg"
)

var (
	nodeCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "node_count",
		Help: "The total number of nodes",
	})
)

func runHealth(nc *natsbroker.Broker, coUrl string, siteID string, db *bolt.DB) {

	// Subscribe to health
	go func() {
		log.Println("lo with siteid:", fmt.Sprintf("health.%s.*", siteID))
		subHealth := fmt.Sprintf("health.%s.*", siteID)
		err := nc.Subscribe2(subHealth, func(h pkg.HealthMsg) {
			// var h pkg.HealthMsg
			log.Println("health received")
			// if err := json.Unmarshal(msg.Data, &h); err != nil {
			// 	log.Println("unmarshal error:", err)
			// 	return
			// }
			en := orchestrator.EdgeNode{
				NodeID:   h.NodeID,
				SiteID:   h.SiteID,
				Runtime:  h.Runtime,
				Region:   h.Region,
				LastSeen: time.Now(),
				CPUFree:  100 - h.CPUPercent,
				Alive:    true,
			}
			orchestrator.SaveNode(db, en)
			nodeCount.Set(float64(len(orchestrator.GetAllNodes(db))))
			log.Printf("[LO] health from %s runtime=%s", h.NodeID, h.Runtime)
			//if fsm.GetState() == shared.Discovering  {
			//	fsm.Transition(shared.Running)
			//}
		})
		if err != nil {
			log.Println("subscribe error:", err)
		} else {
			log.Println("subscribed to", subHealth)
		}
		nc.Flush()
		log.Println("subscription ready for", subHealth)

	}()

	go func() {
		subStatus := fmt.Sprintf("status.%s.*", siteID)
		err := nc.Subscribe4(subStatus, func(s deployment.DeploymentReport) {
			//var s deployment.DeploymentReport
			// if err := json.Unmarshal(m.Data, &s); err != nil {
			// 	log.Println("[LO] invalid status message:", err)
			// 	return
			// }
			log.Printf("[LO] status %s from %s: success=%v, msg=%s",
				s.DeploymentID, s.NodeID, s.Status, s.Message)
			orchestrator.UpdateDeploymentRecord(db, s.DeploymentID, s.NodeID, s.Status, s.Message)

			//forward to CO
			forwardToCO(coUrl, s)

		})
		if err != nil {
			log.Fatal("[LO] failed to subscribe to status updates:", err)
		}

	}()

}

func forwardToCO(baseurl string, report deployment.DeploymentReport) {
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

	log.Printf("[LO] ✅ Report forwarded to CO successfully (deployment_id=%s)", report.DeploymentID)
}
