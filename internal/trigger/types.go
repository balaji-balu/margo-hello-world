package trigger

import (
	"context"
)
type TriggerType string

const (
	PollTriggerType   TriggerType = "poll"
	EventTriggerType  TriggerType = "event"
	GitOpsTriggerType TriggerType = "gitops"
)

// Trigger defines the base interface.
type Trigger interface {
	Start(ctx context.Context) error
	Stop()
}
