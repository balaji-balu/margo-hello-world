package orchestrator

import (
	"encoding/json"
	//"log"
	"os"
	"time"
)

type Journal struct {
	ETag         string    `json:"etag"`
	LastSuccess  time.Time `json:"last_success"`
	LastCommitID string    `json:"last_commit_id"`
}

func (lo *LocalOrchestrator) PersistJournal() {
	data, _ := json.MarshalIndent(lo.Journal, "", "  ")
	_ = os.WriteFile("journal.json", data, 0644)
}

func (lo *LocalOrchestrator) LoadJournal() {
	data, err := os.ReadFile("journal.json")
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &lo.Journal)
}
