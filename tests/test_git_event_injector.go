package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// GitDesiredStateChanged models a GitOps-style event
type GitDesiredStateChanged struct {
	Event     string `json:"event"`
	Repo      string `json:"repo"`
	Branch    string `json:"branch"`
	CommitID  string `json:"commit_id"`
	Site      string `json:"site"`
	Timestamp string `json:"timestamp"`
}

func main() {
	url := "nats://localhost:4222"
	subject := "git.desiredstate.changed"

	// Connect to NATS
	nc, err := nats.Connect(url)
	if err != nil {
		log.Fatalf("❌ Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("✅ Connected to", url)

	// Build test event
	event := GitDesiredStateChanged{
		Event:     "git.desiredstate.changed",
		Repo:      "https://github.com/balaji-balu/margo-hello-world",
		Branch:    "main",
		CommitID:  "abc123fakecommitid",
		Site:      "f95d34b2-8019-4590-a3ff-ff1e15ecc5d5", //"tiruvannamalai",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	data, _ := json.MarshalIndent(event, "", "  ")

	// Publish
	if err := nc.Publish(subject, data); err != nil {
		log.Fatalf("❌ Failed to publish: %v", err)
	}

	log.Printf("📤 Published [%s]: %s\n", subject, string(data))

	// Flush & wait
	nc.Flush()
	if err := nc.LastError(); err != nil {
		log.Fatalf("❌ NATS error: %v", err)
	}

	log.Println("✅ Event sent successfully")
}
