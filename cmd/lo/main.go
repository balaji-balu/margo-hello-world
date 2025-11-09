// file: cmd/lo/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	//"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	//bolt "go.etcd.io/bbolt"

	cfffg "github.com/balaji-balu/margo-hello-world/internal/config"
	"github.com/balaji-balu/margo-hello-world/internal/fsmloader"
	"github.com/balaji-balu/margo-hello-world/internal/natsbroker"
	"github.com/balaji-balu/margo-hello-world/internal/orchestrator"
)

var (
	//db       *bolt.DB
	nc *natsbroker.Broker
)

func init() {
	if err := godotenv.Load("./.env"); err != nil {
		log.Println("No .env file found, reading from system environment")
	}
}

func main() {
	log.Println("Hello from lo.....")

	// ------------------------------------------------------------
	// 1️⃣ Context + Logger setup
	// ------------------------------------------------------------
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configPath := flag.String("config", "./configs/lo1.yaml", "path to config file")
	flag.Parse()

	cfg, err := cfffg.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ error loading config: %v", err)
	}

	log.Println("cfg loaded. init zap")
    logger, err := zap.NewProduction()
    if err != nil {
        // If Zap fails, fall back to standard log or panic
        log.Fatalf("can't initialize zap logger: %v", err)
    }
    defer logger.Sync() // Ensure all buffered logs are written
	zap.RedirectStdLog(logger)

	// Init DB
	dbPath := "./data/lo/nodes.db"
	if p := os.Getenv("BOLT_PATH"); p != "" {
		dbPath = p
	}
	db := orchestrator.InitDB(dbPath)
	defer db.Close()

	log.Println("boltz db inited")
	//reset all existing as not alive
	//TBD: also do reset once the node health is timeout
	for _, n := range orchestrator.GetAllNodes(db) {
		n.Alive = false
		orchestrator.SaveNode(db, n)
	}

	natsURL := cfg.NATS.URL
	log.Println("Connecting to", natsURL)

	nc, err = natsbroker.New(natsURL)
	//nc, err = nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("nats connect: %v", err)
	}
	//defer nc.Drain()
	log.Println("connected to", natsURL)

	runHealth(nc, cfg.CO.URL, cfg.Server.Site, db)

	log.Println("🚀 Starting adaptive mode manager...")

	// ------------------------------------------------------------
	// 2️⃣ Setup orchestrator + FSM loader
	// ------------------------------------------------------------
	lo := orchestrator.New(ctx, cfg, db, nc, logger)
	loader := fsmloader.NewLoader(ctx, logger, lo)
	lo.FSM = loader.FSM // ensure both share same FSM instance

	// ------------------------------------------------------------
	// 3️⃣ Setup Gin router
	// ------------------------------------------------------------
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/register", lo.RegisterRequest)
	r.POST("/deployment_status", lo.DeployStatus)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// ------------------------------------------------------------
	// 4️⃣ Start orchestrator and HTTP server
	// ------------------------------------------------------------
	//go lo.StartModeLoop(ctx)
	go lo.StartNetworkMonitor(ctx)

	go func() {
		logger.Info("🌐 HTTP server started on :", zap.Int("Port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server crashed", zap.Error(err))
		}
	}()

	// ------------------------------------------------------------
	// 5️⃣ Handle shutdown signals
	// ------------------------------------------------------------
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("🛑 Shutdown signal received...")
	cancel() // broadcast cancel to all goroutines

	// ------------------------------------------------------------
	// 6️⃣ Gracefully stop HTTP server
	// ------------------------------------------------------------
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("HTTP server shutdown gracefully")
	}

	// ------------------------------------------------------------
	// 7️⃣ Final cleanup
	// ------------------------------------------------------------
	time.Sleep(500 * time.Millisecond)
	logger.Info("🧹 All systems stopped. Goodbye!")
}
