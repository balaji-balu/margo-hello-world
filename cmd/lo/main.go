// file: lo/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	cfffg "github.com/balaji-balu/margo-hello-world/internal/config"
	"github.com/balaji-balu/margo-hello-world/internal/orchestrator"
)

func init() {
	if err := godotenv.Load("./.env"); err != nil {
		log.Println("No .env file found, reading from system environment")
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	configPath := flag.String("config", "./configs/lo1.yaml", "path to config file")
	flag.Parse()

	cfg, err := cfffg.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Create Local Orchestrator (FSM logic)
	lo := orchestrator.NewLocalOrchestrator("./configs/fsm_lo.yaml")

	var wg sync.WaitGroup

	// ------------------------------
	// 🌐 Setup Gin HTTP Server
	// ------------------------------
	r := gin.Default()

	api := r.Group("")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		api.POST("/register", lo.RegisterRequest)
		api.POST("/deployment_status", lo.DeployStatus)
	}

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("🌐 HTTP server started on %s (Gin)", httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ HTTP server error: %v", err)
		}
	}()

	// ------------------------------
	// ⚙️ Start Orchestrator Loop
	// ------------------------------
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Println("🌀 FSM loop exiting...")
				return
			default:
				mode := lo.DetectMode()
				switch mode {
				case orchestrator.PushPreferred:
					lo.WaitForWebhook(ctx)
				case orchestrator.AdaptivePull:
					lo.Poll(ctx)
				case orchestrator.OfflineDeterministic:
					lo.ScanLocalInbox(ctx)
				default:
					log.Println("Unknown mode, defaulting to AdaptivePull")
					lo.Poll(ctx)
				}
			}

			lo.PersistJournal()
			time.Sleep(5 * time.Second)
		}
	}()

	// ------------------------------
	// 🧹 Graceful Shutdown
	// ------------------------------
	<-ctx.Done()
	log.Println("🛑 Shutdown signal received...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Error shutting down HTTP server: %v", err)
	}

	wg.Wait()
	log.Println("✅ Clean exit.")
}
