package handlers

import (
	"context"
	"net/http"
	"log"


	"github.com/gin-gonic/gin"
	"github.com/goccy/go-yaml"

	"github.com/balaji-balu/margo-hello-world/ent"
	"github.com/balaji-balu/margo-hello-world/ent/applicationdesc"	
	"github.com/balaji-balu/margo-hello-world/pkg/application"
	"github.com/balaji-balu/margo-hello-world/internal/gitfetcher"
)
type CreateAppRequest struct {
	AppName string `json:"app_name" binding:"required"`
	RepoURL string `json:"repo_url" binding:"required"`
}

func ListApps(c *gin.Context, client *ent.Client) {
	apps, err := client.ApplicationDesc.Query().All(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"apps": apps})
}

func GetApp(c *gin.Context, client *ent.Client) {
	id := c.Param("id")
	app, err := client.ApplicationDesc.Get(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}
	c.JSON(http.StatusOK, app)
}

func CreateApp(c *gin.Context, client *ent.Client, fetcher *gitfetcher.GitFetcher) {

	// verify already added, if yes, reject
	//post content : app name and app repo url
	// git private repo read margo.yaml file
	// then parse the yaml.
	// persist the application contents 
	var req CreateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appName := req.AppName // or from query, form, etc.
	apps, err := client.ApplicationDesc.
		Query().
		Where(
			applicationdesc.NameContainsFold(appName), // case-insensitive partial match
		).
		All(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("apps matched:", apps)

	if len(apps) > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "duplicate apps found"})
		return
	}

	fetcher.RepoURL = req.RepoURL
	content, err := fetcher.FetchAppResource(req.AppName, "margo.yaml")
	if err != nil {
		log.Printf("❌ failed to fetch resource: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("📄 margo.yaml contents:\n", string(content))

	//var app ent.ApplicationDesc
	var appDesc application.ApplicationDescription
	if err := yaml.Unmarshal([]byte(content), &appDesc); err != nil {
		log.Printf("❌ failed to unmarshall resource: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}	

	// appDesc, err := application.ParseFromFile("./tests/app3.yaml")
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	err = Persist(context.Background(), client, &appDesc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, appDesc)
}
