// file: cmd/lo/main.go
package main

import (
	"context"
	//"flag"
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
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	//bolt "go.etcd.io/bbolt"

	//cfffg "github.com/balaji-balu/margo-hello-world/internal/config"
	"github.com/balaji-balu/margo-hello-world/ent"
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

	//configPath := flag.String("config", "./configs/lo1.yaml", "path to config file")
	//flag.Parse()

	//cfg, err := cfffg.LoadConfig(*configPath)
	//if err != nil {
	//	log.Fatalf("❌ error loading config: %v", err)
	//}
    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }
	repo := os.Getenv("REPO")
	if repo == "" {
		repo = "https://github.com/edge-orchestration-platform/deployments.git"
	}	
	coURL := os.Getenv("CO_URL")
	if coURL == "" {
		coURL = "http://localhost:8080/api/v1"
	}
    siteID := os.Getenv("SITE_ID")
	if siteID == "" {
		siteID = "f95d34b2-8019-4590-a3ff-ff1e15ecc5d5"
	}
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	log.Printf("siteid:%s port:%s nats url:%s ", 
		siteID, port, natsURL, )

    logger, err := zap.NewProduction()
    if err != nil {
        // If Zap fails, fall back to standard log or panic
        log.Fatalf("can't initialize zap logger: %v", err)
    }
    defer logger.Sync() // Ensure all buffered logs are written
	zap.RedirectStdLog(logger)

	dsn := os.Getenv("DATABASE_URL")
	fmt.Println("[CO] connecting to postgres at", dsn)
	var drv *sql.Driver
	var err1 error
	for i := 1; i <= 10; i++ {
		drv, err1 = sql.Open(dialect.Postgres, dsn)
		log.Println("sql open with err:", err1)
		if err1 == nil {
			if err1 = drv.DB().Ping(); err1 == nil {
				fmt.Println("✅ Connected to Postgres")
				break
			}
		}
		fmt.Printf("⏳ Waiting for Postgres (attempt %d)...\n", i)
		time.Sleep(3 * time.Second)
	}
	if err1 != nil {
		log.Fatalf("❌ Failed to connect to Postgres after retries: %v", err)
	}

	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()	

	//reset all existing as not alive
	//TBD: also do reset once the node health is timeout
	// for _, n := range orchestrator.GetAllNodes(db) {
	// 	n.Alive = false
	// 	orchestrator.SaveNode(db, n)
	// }

	log.Println("Connecting to", natsURL)
	nc, err = natsbroker.New(natsURL)
	if err != nil {
		log.Fatalf("nats connect: %v", err)
	}
	log.Println("connected to", natsURL)


	// ------------------------------------------------------------
	// 2️⃣ Setup orchestrator + FSM loader
	// ------------------------------------------------------------
	lo := orchestrator.New(ctx, siteID, natsURL, repo, client, nc, logger)
	loader := fsmloader.NewLoader(ctx, logger, lo)
	lo.FSM = loader.FSM // ensure both share same FSM instance
	lo.MonitorHealthandStatusFromEN(coURL)

	log.Println("🚀 Starting adaptive mode manager...")

	// ------------------------------------------------------------
	// 3️⃣ Setup Gin router
	// ------------------------------------------------------------
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	//r.POST("/register", lo.RegisterRequest)
	//r.POST("/deployment_status", lo.DeployStatus)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
	}

	// ------------------------------------------------------------
	// 4️⃣ Start orchestrator and HTTP server
	// ------------------------------------------------------------
	//go lo.StartModeLoop(ctx)
	go lo.StartNetworkMonitor(ctx)

	go func() {
		logger.Info("🌐 HTTP server started on :", zap.String("Port", port))
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
