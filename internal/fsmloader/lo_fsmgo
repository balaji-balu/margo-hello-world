package fsmloader

import (
	"fmt"
	"os"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type FSMConfig struct {
	Initial string `yaml:"initial_state"`
	Events  []struct {
		Name string   `yaml:"name"`
		Src  []string `yaml:"src"`
		Dst  string   `yaml:"dst"`
	} `yaml:"events"`
}

func LoadFSMConfig(path string, logger *zap.Logger, cb *Callbacks) (*fsm.FSM, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read FSM YAML: %w", err)
	}

	var cfg FSMConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal FSM YAML: %w", err)
	}

	logger.Info("✅ Loaded FSM configuration",
		zap.String("initial_state", cfg.Initial),
		zap.Int("events", len(cfg.Events)),
	)

	// Convert events to looplab/fsm transitions
	var transitions []fsm.EventDesc
	for _, e := range cfg.Events {
		transitions = append(transitions, fsm.EventDesc{
			Name: e.Name,
			Src:  e.Src,
			Dst:  e.Dst,
		})
	}

	// Initialize FSM
	machine := fsm.NewFSM(
		cfg.Initial,
		transitions,
		fsm.Callbacks{
			"enter_state": cb.OnEnterState,
			"leave_state": cb.OnLeaveState,
			"after_event": cb.AfterEvent,
		},
	)

	return machine, nil
}
