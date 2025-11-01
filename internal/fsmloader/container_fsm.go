package fsmloader

import (
    "context"
    "github.com/looplab/fsm"
    "go.uber.org/zap"
)

type ContainerFSM struct {
    ID     string
    FSM    *fsm.FSM
    logger *zap.Logger
}

func NewContainerFSM(id string, logger *zap.Logger) *ContainerFSM {
    c := &ContainerFSM{ID: id, logger: logger}

    c.FSM = fsm.NewFSM(
        "created",
        fsm.Events{
            {Name: "start_container", Src: []string{"created", "stopped"}, Dst: "running"},
            {Name: "stop_container", Src: []string{"running"}, Dst: "stopped"},
            {Name: "crash_detected", Src: []string{"running"}, Dst: "failed"},
            {Name: "restart_container", Src: []string{"failed", "stopped"}, Dst: "running"},
            {Name: "remove_container", Src: []string{"created", "stopped", "failed"}, Dst: "removed"},
        },
        fsm.Callbacks{
            "enter_state": func(ctx context.Context, e *fsm.Event) {
                c.logger.Info("Container FSM transition",
                    zap.String("container_id", c.ID),
                    zap.String("event", e.Event),
                    zap.String("src", e.Src),
                    zap.String("dst", e.Dst),
                )
            },
        },
    )

    return c
}

func (c *ContainerFSM) Start(ctx context.Context) {
    _ = c.FSM.Event(ctx, "start_container")
}

func (c *ContainerFSM) Crash(ctx context.Context) {
    _ = c.FSM.Event(ctx, "crash_detected")
}

func (c *ContainerFSM) Stop(ctx context.Context) {
    _ = c.FSM.Event(ctx, "stop_container")
}
