package fsmloader

import (
    "context"
    "sync"

    "github.com/looplab/fsm"
    "go.uber.org/zap"

    "github.com/balaji-balu/margo-hello-world/internal/edgenode"
    "github.com/balaji-balu/margo-hello-world/pkg/deployment"
)

type EdgeNodeFSM struct {
    ID          string
    FSM         *fsm.FSM
    Containers  map[string]*ContainerFSM
    mu          sync.RWMutex
    logger      *zap.Logger
}

func NewEdgeNodeFSM(ctx context.Context, 
    id string, 
    en *edgenode.EdgeNode,
    logger *zap.Logger) *EdgeNodeFSM {
    enfsm := &EdgeNodeFSM{
        ID:         id,
        Containers: make(map[string]*ContainerFSM),
        logger:     logger,
    }

    enfsm.FSM = fsm.NewFSM(
        "node_idle",
        fsm.Events{
            {Name: "start_deployment", Src: []string{"node_idle"}, Dst: "node_deploying"},
            {Name: "deployment_success", Src: []string{"node_deploying"}, Dst: "node_idle"},
            {Name: "deployment_failed", Src: []string{"node_deploying"}, Dst: "node_error"},
			{Name: "reset", Src: []string{"failed", "completed"}, Dst: "idle"},
		},
        fsm.Callbacks{
			"enter_deploying": func(ctx context.Context, evt *fsm.Event) {
                en.ReportStatus("app", deployment.StatusRunning, "Deployment started", 
                evt.Dst)
                enfsm.logger.Info("Deployment started", zap.String("state", evt.Dst))
            },
            "enter_completed": func(ctx context.Context, evt *fsm.Event) {
                en.ReportStatus("app", deployment.StatusSuccess, 
                    "Deployment completed successfully",
                    evt.Dst)
                enfsm.logger.Info("Deployment successful", zap.String("state", evt.Dst))
            },
            "enter_failed": func(ctx context.Context, evt *fsm.Event) {
                en.ReportStatus("app", deployment.StatusFailed, "Deployment failed", 
                    evt.Dst)
                enfsm.logger.Warn("Deployment failed", zap.String("state", evt.Dst))
            },			
            // "enter_node_deploying": func(ctx context.Context, e2 *fsm.Event) {
            //     e.logger.Info("Starting deployment on node", zap.String("node_id", e.ID))
            // },
        },
    )

    return enfsm
}

// StartDeployment spins up container FSMs for each requested container.
func (e *EdgeNodeFSM) StartDeployment(ctx context.Context, containers []string) {
    if err := e.FSM.Event(ctx, "start_deployment"); err != nil {
        e.logger.Error("Cannot start deployment", zap.Error(err))
        return
    }

    e.logger.Info("Creating container FSMs", zap.Int("count", len(containers)))
    for _, cid := range containers {
        cfsm := NewContainerFSM(cid, e.logger)
        e.mu.Lock()
        e.Containers[cid] = cfsm
        e.mu.Unlock()
        cfsm.Start(ctx)
    }

    // Aggregate container FSMs
    go e.monitorContainers(ctx)
}

func (e *EdgeNodeFSM) monitorContainers(ctx context.Context) {
    for {
        allDone := true
        e.mu.RLock()
        for id, cfsm := range e.Containers {
            state := cfsm.FSM.Current()
            if state == "failed" {
                e.logger.Warn("Container failed", zap.String("container_id", id))
                _ = e.FSM.Event(ctx, "deployment_failed")
                e.mu.RUnlock()
                return
            }
            if state == "running" || state == "created" {
                allDone = false
            }
        }
        e.mu.RUnlock()

        if allDone {
            e.logger.Info("All containers completed successfully")
            _ = e.FSM.Event(ctx, "deployment_success")
            return
        }
    }
}
