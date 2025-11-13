package main

import (
	"os"
	"context"
	"fmt"
	"log"
	"net/http"
	//"flag"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/joho/godotenv"

	//"github.com/balaji-balu/margo-hello-world/internal/config"
	"github.com/balaji-balu/margo-hello-world/internal/edgenode"
	//"github.com/balaji-balu/margo-hello-world/internal/fsmloader"
	"github.com/balaji-balu/margo-hello-world/internal/natsbroker"
)

var localOrchestratorURL string // LO endpoint

func init() {
	err := godotenv.Load("./.env") // relative path to project root
	if err != nil {
		log.Println("No .env file found, reading from system environment")
	}
}

func main() {
	//configPath := flag.String("config", "./configs/edge1.yaml", "config file path")
	//flag.Parse()

	// // Load configuration
	// cfg, err := config.LoadConfig(*configPath)
	// if err != nil {
	// 	log.Fatalf("❌ Error loading config: %v", err)
	// }
	// log.Println("✅ Loaded config: ", cfg.Server)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8082"
    }
    siteID := os.Getenv("SITE_ID")
	if siteID == "" {
		siteID = "f95d34b2-8019-4590-a3ff-ff1e15ecc5d5"
	}
    nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "edge1-containerd"
	}
	runtime := os.Getenv("RUNTIME")
	if runtime == "" {
		runtime = "containerd"
	} 
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	log.Printf("siteid:%s nodeid:%s port:%s nats url:%s runtime:%s", 
		siteID, nodeID, port, natsURL, runtime)

	// Initialize NATS
	log.Printf("📡 Connecting to NATS at %s", natsURL)
	nc, err := natsbroker.New(natsURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to NATS: %v", err)
	}

	// Initialize Zap logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()
	zap.RedirectStdLog(logger)

	ctx := context.Background()

	// Initialize Edge Node + FSM
	en := edgenode.NewEdgeNode(ctx, siteID, nodeID, runtime, nc, logger)
	//nodeFSM := fsmloader.NewEdgeNodeFSM(ctx, "edge-node-1", en, logger)

	// Start background tasks
	en.Start()

	// Create Gin engine
	router := gin.Default()

	// --- Health API ---
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"node":    nodeID,
			"runtime": runtime,
			//"region":  cfg.Server.Region,
		})
	})

	// --- Deploy API ---
	// router.POST("/deploy", func(c *gin.Context) {
	// 	handleDeployGin(en, c, nodeFSM)
	// })

	addr := fmt.Sprintf(":%s", port)
	log.Printf("🚀 Edge Node Server started on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
