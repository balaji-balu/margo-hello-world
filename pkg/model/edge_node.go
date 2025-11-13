package model

import (
	"time"
)
type EdgeNode struct {
	NodeID   string            `json:"node_id"`
	SiteID   string            `json:"site_id"`
	Runtime  string            `json:"runtime"`
	Region   string            `json:"region"`
	LastSeen time.Time         `json:"last_seen"`
	CPUFree  float64           `json:"cpu_free"`
	Alive    bool              `json:"alive"`
	Labels   map[string]string `json:"labels,omitempty"`
}