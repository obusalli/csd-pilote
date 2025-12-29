package clusters

import (
	"time"

	"github.com/google/uuid"
)

// ClusterStatus represents the status of a Kubernetes cluster
type ClusterStatus string

const (
	ClusterStatusPending      ClusterStatus = "PENDING"
	ClusterStatusDeploying    ClusterStatus = "DEPLOYING"
	ClusterStatusConnected    ClusterStatus = "CONNECTED"
	ClusterStatusDisconnected ClusterStatus = "DISCONNECTED"
	ClusterStatusError        ClusterStatus = "ERROR"
)

// ClusterMode represents how the cluster was added
type ClusterMode string

const (
	ClusterModeConnect ClusterMode = "CONNECT" // Connect to existing cluster
	ClusterModeDeploy  ClusterMode = "DEPLOY"  // Deploy new cluster via agents
)

// KubernetesDistribution represents the Kubernetes distribution
type KubernetesDistribution string

const (
	// Deployable distributions (can be installed via agents)
	K8sDistroK3s      KubernetesDistribution = "K3S"      // Lightweight, single binary
	K8sDistroRKE2     KubernetesDistribution = "RKE2"     // Rancher Kubernetes Engine 2
	K8sDistroKubeadm  KubernetesDistribution = "KUBEADM"  // Official Kubernetes installer
	K8sDistroK0s      KubernetesDistribution = "K0S"      // Zero friction Kubernetes
	K8sDistroMicroK8s KubernetesDistribution = "MICROK8S" // Canonical's lightweight K8s

	// Connect-only distributions (existing/managed clusters)
	K8sDistroEKS       KubernetesDistribution = "EKS"       // Amazon Elastic Kubernetes Service
	K8sDistroGKE       KubernetesDistribution = "GKE"       // Google Kubernetes Engine
	K8sDistroAKS       KubernetesDistribution = "AKS"       // Azure Kubernetes Service
	K8sDistroOpenShift KubernetesDistribution = "OPENSHIFT" // Red Hat OpenShift
	K8sDistroRancher   KubernetesDistribution = "RANCHER"   // Rancher managed
	K8sDistroOther     KubernetesDistribution = "OTHER"     // Unknown/other distribution
)

// Cluster represents a Kubernetes cluster configuration
type Cluster struct {
	ID            uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID     `json:"tenantId" gorm:"type:uuid;not null;index"`
	Name          string        `json:"name" gorm:"not null"`
	Description   string        `json:"description"`
	Mode          ClusterMode   `json:"mode" gorm:"default:'CONNECT'"`       // CONNECT or DEPLOY
	Distribution  KubernetesDistribution `json:"distribution"`               // K3S, RKE2, KUBEADM, etc.
	Version       string        `json:"version"`                             // Kubernetes version
	AgentID       uuid.UUID     `json:"agentId" gorm:"type:uuid"`            // Agent for CONNECT mode (executes K8s commands)
	ArtifactKey   string        `json:"artifactKey"`                         // Reference to kubeconfig artifact in csd-core
	ApiServerURL  string        `json:"apiServerUrl"`                        // API server URL
	Status        ClusterStatus `json:"status" gorm:"default:'PENDING'"`
	StatusMessage string        `json:"statusMessage"`
	LastCheckedAt *time.Time    `json:"lastCheckedAt"`
	CreatedAt     time.Time     `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt     time.Time     `json:"updatedAt" gorm:"autoUpdateTime"`
	CreatedBy     uuid.UUID     `json:"createdBy" gorm:"type:uuid"`

	// Relations
	Nodes []ClusterNode `json:"nodes,omitempty" gorm:"foreignKey:ClusterID"`
}

// TableName returns the table name for GORM
func (Cluster) TableName() string {
	return "clusters"
}

// NodeRole represents the role of a node in a cluster
type NodeRole string

const (
	NodeRoleMaster NodeRole = "MASTER" // Control plane node
	NodeRoleWorker NodeRole = "WORKER" // Worker node
)

// ClusterNode represents a node in a deployed cluster
type ClusterNode struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ClusterID uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index"`
	AgentID   uuid.UUID `json:"agentId" gorm:"type:uuid;not null"` // csd-core agent on this node
	Role      NodeRole  `json:"role" gorm:"not null"`              // MASTER or WORKER
	Hostname  string    `json:"hostname"`
	IP        string    `json:"ip"`
	Status    string    `json:"status" gorm:"default:'PENDING'"` // PENDING, DEPLOYING, READY, ERROR
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (ClusterNode) TableName() string {
	return "cluster_nodes"
}

// ClusterInput represents input for connecting to an existing cluster
type ClusterInput struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	AgentID      string                 `json:"agentId"`
	ArtifactKey  string                 `json:"artifactKey"`
	Distribution KubernetesDistribution `json:"distribution"` // Optional: K3S, RKE2, EKS, GKE, AKS, etc.
}

// DeployClusterInput represents input for deploying a new cluster
type DeployClusterInput struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Distribution KubernetesDistribution `json:"distribution"` // K3S, RKE2, etc.
	Version      string                 `json:"version"`      // Optional: specific version
	MasterNodes  []string               `json:"masterNodes"`  // Agent IDs for master nodes
	WorkerNodes  []string               `json:"workerNodes"`  // Agent IDs for worker nodes
}

// ClusterFilter represents filter options for listing clusters
type ClusterFilter struct {
	Search       *string                 `json:"search"`
	Status       *ClusterStatus          `json:"status"`
	Mode         *ClusterMode            `json:"mode"`
	Distribution *KubernetesDistribution `json:"distribution"`
}
