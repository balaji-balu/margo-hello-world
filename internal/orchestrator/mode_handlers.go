package orchestrator

import (
    "context"
    "log"
	"time"
    "go.uber.org/zap"
    //"errors"


    "github.com/balaji-balu/margo-hello-world/internal/gitobserver"
    "github.com/balaji-balu/margo-hello-world/internal/natsbroker"
    . "github.com/balaji-balu/margo-hello-world/internal/config"
)

// Handles PushPreferred mode
func (lo *LocalOrchestrator) StartPushMode(ctx context.Context, cfg LoConfig) error {
    log.Println("🚀 Starting Push Mode (NATS subscribe to desiredstate.changed)")
    log.Println("cfg.NATS.URL", cfg.NatsUrl)
    //log.Println("cfg.Server.Site", cfgSite)

    b, err := natsbroker.New(cfg.NatsUrl)
    if err != nil {
        return err
    }

    // Subscribe to global topic
    err = b.Subscribe("git.desiredstate.changed", func(ev gitobserver.GitEvent) {
        lo.logger.Info("Received push event", zap.Any("event", ev))

        if ev.Site != cfg.Site && ev.Site != "*" {
            lo.logger.Error("Site mismatch", 
                zap.String("expected", cfg.Site), 
                zap.String("received", ev.Site))
            return
        }
        log.Printf("[LO-%s] Received push event: %v", cfg.Site, ev)
        // Trigger FSM event
        // if err := lo.FSM.Event(ctx, "git_update_received"); err != nil {
        //     lo.logger.Error("Failed to trigger FSM event", zap.Error(err))
        // }
        log.Println("✅ FSM event triggered")
        // ✅ Queue FSM event safely
        go lo.TriggerEvent("git_update_received")
    })
    if err != nil {
        lo.logger.Error("Failed to subscribe to NATS", zap.Error(err))
        return err
    }

    // Block until canceled (Ctrl+C or mode switch)
    <-ctx.Done()
    b.Close()
    log.Println("🛑 Push Mode stopped gracefully")
    return nil
}

// Handles AdaptivePull mode
func (lo *LocalOrchestrator) start_pull_mode(ctx context.Context, cfg *Config) error {
    log.Println("📡 Starting Pull Mode (periodic git sync)")
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            log.Println("🛑 Pull Mode stopped gracefully")
            return nil
        case <-ticker.C:
            log.Println("🔁 Checking for new desired state...")
            // TODO: implement git pull logic
            lo.FSM.Event(ctx, "git_polled")
        }
    }
}

// Handles OfflineDeterministic mode
func (lo *LocalOrchestrator) start_offline_mode(ctx context.Context, cfg *Config) error {
    log.Println("📴 Starting Offline Mode (working from journal)")
    lo.FSM.Event(ctx, "offline_mode_start")
    <-ctx.Done()
    log.Println("🛑 Offline Mode stopped gracefully")
    return nil
}
