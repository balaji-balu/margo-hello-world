package model

type HealthMsg struct {
	NodeID     string  `json:"node_id"`
	SiteID     string  `json:"site_id"`
	CPUPercent float64 `json:"cpu_percent"`
	MemMB      float64 `json:"mem_mb"`
	Timestamp  int64   `json:"timestamp"`
	Runtime    string  `json:"runtime"`
	Region     string  `json:"region"`
}
