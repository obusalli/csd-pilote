package namespaces

import (
	"time"

	"github.com/google/uuid"
)

// Namespace represents a Kubernetes namespace
type Namespace struct {
	ClusterID   uuid.UUID         `json:"clusterId"`
	Name        string            `json:"name"`
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
}

// NamespaceFilter contains filter options
type NamespaceFilter struct {
	Search *string `json:"search,omitempty"`
}
