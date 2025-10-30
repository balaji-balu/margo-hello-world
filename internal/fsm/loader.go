package fsm

import (
	"context"
	"fmt"
	"os"

	"github.com/looplab/fsm"
	"gopkg.in/yaml.v3"
)

type Event struct {
	Name string   `yaml:"name"`
	Src  []string `yaml:"src"`
	Dst  string   `yaml:"dst"`
}

type FSMConfig struct {
	InitialState string            `yaml:"initial_state"`
	Events       []Event           `yaml:"events"`
	Callbacks    map[string]string `yaml:"callbacks"`
}

func LoadFSM(path string) (*fsm.FSM, error) {
	wd, err := os.Getwd()
	fmt.Println("inside loader", path, wd)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg FSMConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	fmt.Println("✅ Loaded FSM config:", cfg.InitialState, "events:", len(cfg.Events))

	// Build fsm.EventDesc list directly from YAML "events"
	events := []fsm.EventDesc{}
	for _, e := range cfg.Events {
		events = append(events, fsm.EventDesc{
			Name: e.Name,
			Src:  e.Src,
			Dst:  e.Dst,
		})
	}

	// Create callbacks
	callbacks := map[string]fsm.Callback{}
	for name, msg := range cfg.Callbacks {
		text := msg // capture value
		callbacks[name] = func(ctx context.Context, e *fsm.Event) {
			fmt.Printf("[FSM] %s: %s → %s | %s\n", e.Event, e.Src, e.Dst, text)
		}
	}

	// Initialize FSM with initial state
	m := fsm.NewFSM(cfg.InitialState, events, callbacks)

	fmt.Println("FSM started at state:", m.Current())
	return m, nil
}
