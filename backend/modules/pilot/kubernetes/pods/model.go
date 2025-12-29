package pods

import (
	"time"

	"github.com/google/uuid"
)

// Pod represents a Kubernetes pod
type Pod struct {
	ClusterID     uuid.UUID         `json:"clusterId"`
	Namespace     string            `json:"namespace"`
	Name          string            `json:"name"`
	Phase         string            `json:"phase"`
	Status        string            `json:"status"`
	Ready         string            `json:"ready"`
	Restarts      int32             `json:"restarts"`
	Age           string            `json:"age"`
	IP            string            `json:"ip"`
	Node          string            `json:"node"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Containers    []Container       `json:"containers"`
	CreatedAt     time.Time         `json:"createdAt"`
}

// Container represents a container in a pod
type Container struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restartCount"`
	State        string `json:"state"`
}

// PodFilter contains filter options
type PodFilter struct {
	Search *string `json:"search,omitempty"`
	Phase  *string `json:"phase,omitempty"`
}

// PodLogs contains pod logs
type PodLogs struct {
	PodName   string `json:"podName"`
	Container string `json:"container"`
	Logs      string `json:"logs"`
}
