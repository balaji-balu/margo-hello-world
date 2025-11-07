package handlers

import (
	"github.com/balaji-balu/margo-hello-world/pkg/deployment"
	"github.com/gin-gonic/gin"
	"log"
	//"io"
	//"net/http"
	//"encoding/json"
	"github.com/balaji-balu/margo-hello-world/internal/streammanager"
	"time"
)

func DeploymentStatusHandler(c *gin.Context, sm *streammanager.StreamManager) {
	log.Println("DeploymentStatusHandler called")

	var dr deployment.DeploymentReport
	if err := c.ShouldBindJSON(&dr); err != nil {
		log.Println("Error binding JSON:", err)
		return
	}
	// body, err := io.ReadAll(c.Request.Body)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	log.Println("Received deployment status:", dr)

	// var status deployment.DeploymentStatus
	// if err := json.Unmarshal(body, &status); err != nil {
	// 	log.Println("Failed to parse deployment status:", err)
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	log.Println("Deployment id:", dr.DeploymentID, sm)

	sm.Broadcast(dr.DeploymentID, streammanager.DeployEvent{
		DeploymentId: dr.DeploymentID,
		Timestamp:    time.Now().Format(time.RFC3339),
		SiteID:       "",
		Message:      dr.Message,
		Status:       string(dr.Status),
	})

}
