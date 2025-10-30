package deployment


type DeploymentStatus string

type DeployRequest struct {
	AppName string `json:"app_name"`
	Image   string `json:"image"`
	Token   string `json:"token"`
	Revision string `json:"revision"`
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
	AppName string           `json:"app_name"`
	Status  DeploymentStatus `json:"status"`
	Message string           `json:"message,omitempty"`
}
