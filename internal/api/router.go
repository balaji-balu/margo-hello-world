package api

import (
    "github.com/gin-gonic/gin"
    "github.com/balaji/hello/ent"

    //"github.com/balaji/hello/internal/api"
    //. "github.com/balaji/hello/internal/api/handlers"   

    "github.com/balaji/hello/internal/api/handlers"
    "github.com/balaji/hello/internal/api/middleware"

)

func NewRouter(client *ent.Client) *gin.Engine {
    r := gin.New()
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    r.Use(middleware.CORSMiddleware())

    api := r.Group("/api/v1")
    {
        api.POST("/status", handlers.DeploymentStatusHandler)
        api.GET("/healthz", handlers.HealthzHandler)
        api.GET("/apps", func(c *gin.Context) { handlers.ListApps(c, client) })
        api.POST("/apps", func(c *gin.Context) { handlers.CreateApp(c, client) })
        api.GET("/apps/:id", func(c *gin.Context) { handlers.GetApp(c, client) })
        api.POST("/deployments", func(c *gin.Context) { handlers.CreateDeployment(c, client) })
        api.GET("/deployments/:id/status", func(c *gin.Context) { handlers.GetDeploymentStatus(c, client) })
    }

    return r
}
