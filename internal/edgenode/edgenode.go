package edgenode 

import (
	"bytes"
	"encoding/json"
	"fmt"
	//"log"
	"net/http"
	"time"

	"github.com/balaji-balu/margo-hello-world/pkg/deployment"
)

type EdgeNode struct {
	localOrchestratorURL string
}

func NewEdgeNode(localOrchestratorURL string) *EdgeNode {
	return &EdgeNode{
		localOrchestratorURL: localOrchestratorURL,
	}
}

func (edgeNode *EdgeNode) ReportStatus(
		app string, 
		status deployment.DeploymentStatus, 
		msg string, 
		currentState string) {
	report := deployment.DeploymentReport{
		NodeID:  "edge-node-1",
		AppName: app,
		Status:  status,
		Message: msg,
		State:   currentState,
		Timestamp:    time.Now().Format(time.RFC3339),
	}
	body, _ := json.Marshal(report)
	url := fmt.Sprintf("%s/deployment_status", edgeNode.localOrchestratorURL)
	http.Post(url, "application/json", bytes.NewReader(body))
}