package api

import (
	"log"
	"os"
	"github.com/balaji-balu/margo-hello-world/ent"
	"github.com/balaji-balu/margo-hello-world/internal/api/handlers"
	"github.com/balaji-balu/margo-hello-world/internal/api/middleware"
	"github.com/balaji-balu/margo-hello-world/internal/config"
	"github.com/balaji-balu/margo-hello-world/internal/streammanager"
	"github.com/balaji-balu/margo-hello-world/internal/gitfetcher"
	"github.com/gin-gonic/gin"
)

func NewRouter(client *ent.Client, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	sm := streammanager.NewStreamManager()

	log.Println("App registry details:",
		cfg.Appregistry.Repo,
		cfg.Appregistry.Branch,
	)
	fetcher := gitfetcher.GitFetcher{
		RepoURL:  cfg.Appregistry.Repo, //"https://github.com/edge-orchestration-platform/app-registry",
		Branch:   cfg.Appregistry.Branch,
		LocalDir: "./cache/app-registry",
		Token: os.Getenv("GITHUB_TOKEN"),
	}

	api := r.Group("/api/v1")
	{
		api.POST("/deployments/:id/status", func(c *gin.Context) { handlers.DeploymentStatusHandler(c, sm) })
		api.GET("/healthz", handlers.HealthzHandler)
		api.GET("/apps", func(c *gin.Context) { handlers.ListApps(c, client) })
		api.POST("/apps", func(c *gin.Context) { handlers.CreateApp(c, client, &fetcher) })
		api.GET("/apps/:id", func(c *gin.Context) { handlers.GetApp(c, client) })
		api.POST("/deployments", func(c *gin.Context) { handlers.CreateDeployment(c, client, cfg) })
		api.GET("/deployments/:id/status", func(c *gin.Context) { handlers.GetDeploymentStatus(c, client) })
		api.GET("/deployments/:id/stream", func(c *gin.Context) { handlers.HandleStreamDeployment(c, sm) })
	}

	return r
}
