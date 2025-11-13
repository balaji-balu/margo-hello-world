package model

type DeploymentStage string

const (
    StatePending    DeploymentStage = "pending"
    StateInstalling DeploymentStage = "installing"
    StateInstalled  DeploymentStage = "installed"
    StateFailed     DeploymentStage = "failed"
)

type DeploymentStatus struct {
    APIVersion   string                 `json:"apiVersion"`
    Kind         string                 `json:"kind"`
    DeploymentID string                 `json:"deploymentId"`
    Status       DeploymentState        `json:"status"`
    Components   []DeploymentComponent  `json:"components"`
}

type DeploymentState struct {
    State string        `json:"state"`
    Error StatusError   `json:"error"`
}

type DeploymentComponent struct {
    Name  string      `json:"name"`
    State string      `json:"state"`
    Error StatusError `json:"error"`
}

type StatusError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
