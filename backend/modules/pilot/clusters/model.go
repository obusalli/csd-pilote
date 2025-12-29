package clusters

import (
	"time"

	"github.com/google/uuid"
)

// ClusterStatus represents the status of a Kubernetes cluster connection
type ClusterStatus string

const (
	ClusterStatusConnected    ClusterStatus = "CONNECTED"
	ClusterStatusDisconnected ClusterStatus = "DISCONNECTED"
	ClusterStatusError        ClusterStatus = "ERROR"
	ClusterStatusPending      ClusterStatus = "PENDING"
)

// Cluster represents a Kubernetes cluster configuration
type Cluster struct {
	ID            uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID     `json:"tenantId" gorm:"type:uuid;not null;index"`
	Name          string        `json:"name" gorm:"not null"`
	Description   string        `json:"description"`
	ArtifactKey   string        `json:"artifactKey" gorm:"not null"` // Reference to kubeconfig artifact in csd-core
	ApiServerURL  string        `json:"apiServerUrl"`                // Cached from kubeconfig
	Status        ClusterStatus `json:"status" gorm:"default:'PENDING'"`
	StatusMessage string        `json:"statusMessage"`
	LastCheckedAt *time.Time    `json:"lastCheckedAt"`
	CreatedAt     time.Time     `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt     time.Time     `json:"updatedAt" gorm:"autoUpdateTime"`
	CreatedBy     uuid.UUID     `json:"createdBy" gorm:"type:uuid"`
}

// TableName returns the table name for GORM
func (Cluster) TableName() string {
	return "clusters"
}

// ClusterInput represents input for creating/updating a cluster
type ClusterInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ArtifactKey string `json:"artifactKey"`
}

// ClusterFilter represents filter options for listing clusters
type ClusterFilter struct {
	Search *string        `json:"search"`
	Status *ClusterStatus `json:"status"`
}
