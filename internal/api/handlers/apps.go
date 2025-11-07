package handlers

import (
	"context"
	"net/http"

	"github.com/balaji-balu/margo-hello-world/ent"
	"github.com/gin-gonic/gin"

	"github.com/balaji-balu/margo-hello-world/pkg/application"
)

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

func CreateApp(c *gin.Context, client *ent.Client) {
	//var app ent.ApplicationDesc
	appDesc, err := application.ParseFromFile("./tests/app3.yaml")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = Persist(context.Background(), client, appDesc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, appDesc)
}
