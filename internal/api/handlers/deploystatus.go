package handlers

import (
    "github.com/gin-gonic/gin"
	"github.com/balaji/hello/pkg/deployment"
	"log"
	//"io"
	//"net/http"
	//"encoding/json"	
)

func DeploymentStatusHandler(c *gin.Context) {
	log.Println("DeploymentStatusHandler called")

	var app deployment.DeploymentReport
	if err := c.ShouldBindJSON(&app); err != nil {
		log.Println("Error binding JSON:", err)
		return
	}
	// body, err := io.ReadAll(c.Request.Body)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	log.Println("Received deployment status:", app.AppName, app.Message, app.Status )

	// var status deployment.DeploymentStatus
	// if err := json.Unmarshal(body, &status); err != nil {
	// 	log.Println("Failed to parse deployment status:", err)
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	log.Println("Deployment status:", app.Status)
}