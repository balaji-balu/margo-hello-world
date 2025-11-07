package deployment

type DeploymentStatus string

type DeployRequest struct {
	AppName string `json:"app_name"`
	Image   string `json:"image"`
	Token   string `json:"token"`
	//Revision string `json:"revision"`

	DeploymentID    string   `json:"deployment_id"`
	GitRepoURL      string   `json:"git_repo_url"`
	WasmImages      []string `json:"wasm_images,omitempty"`
	ContainerImages []string `json:"container_images,omitempty"`
	Revision        string   `json:"revision,omitempty"`
	//EdgeNodeIDs []string `json:"edge_node_ids,omitempty"`

}

const (
	StatusPending   DeploymentStatus = "pending"
	StatusStarted   DeploymentStatus = "started"
	StatusRunning   DeploymentStatus = "in-progress"
	StatusCompleted DeploymentStatus = "completed"
	StatusSuccess   DeploymentStatus = "success"
	StatusFailed    DeploymentStatus = "failed"
)

type DeploymentReport struct {
	DeploymentID string           `json:"deployment_id"`
	SiteID       string           `json:"site_id"`
	NodeID       string           `json:"node_id"`
	AppName      string           `json:"app_name"`
	Status       DeploymentStatus `json:"status"`
	Message      string           `json:"message,omitempty"`
	State        string           `json:"state"`
	Timestamp    string           `json:"time"`
}
