package deployments

import (
	"time"

	"github.com/google/uuid"
)

// Deployment represents a Kubernetes deployment
type Deployment struct {
	ClusterID         uuid.UUID         `json:"clusterId"`
	Namespace         string            `json:"namespace"`
	Name              string            `json:"name"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"readyReplicas"`
	AvailableReplicas int32             `json:"availableReplicas"`
	UpdatedReplicas   int32             `json:"updatedReplicas"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Images            []string          `json:"images"`
	CreatedAt         time.Time         `json:"createdAt"`
}

// DeploymentFilter contains filter options
type DeploymentFilter struct {
	Search *string `json:"search,omitempty"`
}

// CreateDeploymentInput contains input for creating a deployment
type CreateDeploymentInput struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Image     string            `json:"image"`
	Replicas  int32             `json:"replicas"`
	Labels    map[string]string `json:"labels,omitempty"`
	Ports     []ContainerPort   `json:"ports,omitempty"`
	Env       []EnvVar          `json:"env,omitempty"`
}

// ContainerPort represents a container port
type ContainerPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"`
}

// EnvVar represents an environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ScaleDeploymentInput contains input for scaling a deployment
type ScaleDeploymentInput struct {
	Replicas int32 `json:"replicas"`
}
