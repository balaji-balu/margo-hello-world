package orchestrator

import (
    "log"
    "context"
    "net"
    "time"

    //. "github.com/balaji-balu/margo-hello-world/internal/config"
    
)

type SyncMode int

const (
    PushPreferred SyncMode = iota
    AdaptivePull
    OfflineDeterministic
)

func networkStable() bool {
    conn, err := net.DialTimeout("tcp", "github.com:443", 2*time.Second)
    if err != nil {
        return false
    }
    _ = conn.Close()
    return true
}


func (lo *LocalOrchestrator) DetectMode() SyncMode {
    if networkStable() {
        return PushPreferred
    }
    if time.Since(lo.Journal.LastSuccess).Hours() > 2 {
        return OfflineDeterministic
    }
    return AdaptivePull
}

func (lo *LocalOrchestrator) StartModeLoop(ctx context.Context) {
    var currentMode SyncMode
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    currentMode = -1
    lo.logger.Info("Starting mode detection loop...")
    for {
        select {
        case <-ctx.Done():
            log.Println("🌀 Mode detection loop exiting...")
            return
        case <-ticker.C:
            mode := lo.DetectMode()
            //log.Printf("Current mode: %v", mode)
            if mode != currentMode {
                log.Printf("🔄 Mode change detected: %v → %v", currentMode, mode)
                currentMode = mode
                lo.triggerFSM(ctx, mode)
            }
        }
    }
}

func (lo *LocalOrchestrator) triggerFSM(ctx context.Context, mode SyncMode) {
    switch mode {
    case PushPreferred:
        lo.logger.Info("Pushing preferred, enabling push mode")
        if err :=  lo.FSM.Event(ctx, "enable_push"); err != nil {
            log.Println("❌ Error:", err)
            return
        }
    case AdaptivePull:
        lo.logger.Info("Adaptive pull, enabling pull mode")
        if err := lo.FSM.Event(ctx, "enable_pull"); err != nil {
            log.Println("❌ Error:", err)
            return
        }
    case OfflineDeterministic:
        lo.logger.Info("Offline deterministic, enabling offline mode")
        if err :=  lo.FSM.Event(ctx, "enable_offline"); err != nil {
            log.Println("❌ Error:", err)
            return
        }
    }
}



