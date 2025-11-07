package orchestrator

import (
	"context"
	"log"
	//"net"
	"time"

	"go.uber.org/zap"
	//. "github.com/balaji-balu/margo-hello-world/internal/config"
)

var (
	currentMode SyncMode
	cancelFunc  context.CancelFunc
)

type SyncMode int

const (
	PushPreferred SyncMode = iota
	AdaptivePull
	OfflineDeterministic
)

func (lo *LocalOrchestrator) StartModeLoop(ctx context.Context) {

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
			mode := AdaptivePull //lo.DetectMode()
			//log.Printf("Current mode: %v", mode)
			if mode != currentMode {
				log.Printf("🔄 Mode change detected: %v → %v", currentMode, mode)
				currentMode = mode
				lo.triggerFSM(ctx, mode)
			}
		}
	}
}

func (lo *LocalOrchestrator) triggerFSM(ctx context.Context, newMode SyncMode) {
	if cancelFunc != nil {
		lo.logger.Info("Stopping current process for mode %s", zap.Int("mode", int(currentMode)))
		cancelFunc() // Graceful stop
		cancelFunc = nil
	}

	// Start new mode
	ctx, cancel := context.WithCancel(context.Background())
	cancelFunc = cancel
	currentMode = newMode
	lo.logger.Info("Starting new process for mode %s", zap.Int("mode", int(newMode)))

	switch newMode {
	case PushPreferred:
		lo.logger.Info("Pushing preferred, enabling push mode")
		if err := lo.FSM.Event(ctx, "enable_push"); err != nil {
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
		if err := lo.FSM.Event(ctx, "enable_offline"); err != nil {
			log.Println("❌ Error:", err)
			return
		}
	}
}
