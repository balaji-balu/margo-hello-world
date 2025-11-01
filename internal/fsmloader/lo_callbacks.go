package fsmloader

import (
	"context"
	"go.uber.org/zap"
	"github.com/looplab/fsm"
)

type Callbacks struct {
	ctx    context.Context
	logger *zap.Logger
}

func NewCallbacks(ctx context.Context, logger *zap.Logger) *Callbacks {
	return &Callbacks{ctx: ctx, logger: logger}
}

// Note the (ctx context.Context, e *fsm.Event) parameters now
func (cb *Callbacks) OnEnterState(ctx context.Context, e *fsm.Event) {
	cb.logger.Info("🔁 Entering state", zap.String("state", e.Dst))
	switch e.Dst {
	case "pending":
		cb.logger.Info("📦 Deployment request received")
	case "deploying":
		cb.logger.Info("🚀 Starting deployment process")
	case "monitoring":
		cb.logger.Info("🩺 Monitoring deployed nodes")
	case "paused":
		cb.logger.Warn("⚠️ Connection lost – entering paused mode")
	case "degraded":
		cb.logger.Warn("⚙️ Node degraded – partial operation")
	case "idle":
		cb.logger.Info("✅ Back to idle – ready for next deployment")
	case "failed":
		cb.logger.Error("❌ FSM entered failed state")
	}
}

func (cb *Callbacks) OnLeaveState(ctx context.Context, e *fsm.Event) {
	cb.logger.Debug("↩️ Leaving state", zap.String("state", e.Src))
}

func (cb *Callbacks) AfterEvent(ctx context.Context, e *fsm.Event) {
	cb.logger.Debug("📨 Event processed", zap.String("event", e.Event), zap.String("to", e.Dst))
}
