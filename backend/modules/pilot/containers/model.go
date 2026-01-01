package containers

import (
	"time"

	"github.com/google/uuid"
)

// EngineType represents the container engine type
type EngineType string

const (
	EngineTypeDocker EngineType = "DOCKER"
	EngineTypePodman EngineType = "PODMAN"
)

// EngineStatus represents the status of a container engine connection
type EngineStatus string

const (
	EngineStatusConnected    EngineStatus = "CONNECTED"
	EngineStatusDisconnected EngineStatus = "DISCONNECTED"
	EngineStatusError        EngineStatus = "ERROR"
	EngineStatusPending      EngineStatus = "PENDING"
)

// ContainerEngine represents a Docker/Podman engine configuration
type ContainerEngine struct {
	ID            uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID    `json:"tenantId" gorm:"type:uuid;not null;index:idx_engine_tenant;index:idx_engine_tenant_status"`
	Name          string       `json:"name" gorm:"not null"`
	Description   string       `json:"description"`
	EngineType    EngineType   `json:"engineType" gorm:"not null;default:'DOCKER'"`
	Host          string       `json:"host" gorm:"not null"` // unix:///var/run/docker.sock or tcp://host:port
	ArtifactKey   string       `json:"artifactKey"`          // Reference to TLS certs artifact (optional)
	Status        EngineStatus `json:"status" gorm:"default:'PENDING';index:idx_engine_tenant_status"`
	StatusMessage string       `json:"statusMessage"`
	// Cached info from engine
	EngineVersion  string     `json:"engineVersion"`
	APIVersion     string     `json:"apiVersion"`
	OSType         string     `json:"osType"`
	Architecture   string     `json:"architecture"`
	TotalMemoryMB  int64      `json:"totalMemoryMb"`
	TotalCPUs      int        `json:"totalCpus"`
	ContainerCount int        `json:"containerCount"`
	ImageCount     int        `json:"imageCount"`
	LastCheckedAt  *time.Time `json:"lastCheckedAt"`
	CreatedAt      time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
	CreatedBy      uuid.UUID  `json:"createdBy" gorm:"type:uuid"`
}

// TableName returns the table name for GORM
func (ContainerEngine) TableName() string {
	return "container_engines"
}

// ContainerEngineInput represents input for creating/updating a container engine
type ContainerEngineInput struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	EngineType  EngineType `json:"engineType"`
	Host        string     `json:"host"`
	ArtifactKey string     `json:"artifactKey"`
}

// ContainerEngineFilter represents filter options for listing container engines
type ContainerEngineFilter struct {
	Search     *string       `json:"search"`
	Status     *EngineStatus `json:"status"`
	EngineType *EngineType   `json:"engineType"`
}

// Container represents a running or stopped container
type Container struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	State      string            `json:"state"` // running, paused, exited, created
	Status     string            `json:"status"`
	Created    time.Time         `json:"created"`
	Ports      []ContainerPort   `json:"ports"`
	Labels     map[string]string `json:"labels"`
	Networks   []string          `json:"networks"`
	Mounts     []ContainerMount  `json:"mounts"`
	Command    string            `json:"command"`
	SizeRw     int64             `json:"sizeRw"`
	SizeRootFs int64             `json:"sizeRootFs"`
}

// ContainerPort represents a container port mapping
type ContainerPort struct {
	IP          string `json:"ip"`
	PrivatePort int    `json:"privatePort"`
	PublicPort  int    `json:"publicPort"`
	Type        string `json:"type"` // tcp, udp
}

// ContainerMount represents a container mount
type ContainerMount struct {
	Type        string `json:"type"` // bind, volume, tmpfs
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
}

// Image represents a container image
type Image struct {
	ID          string            `json:"id"`
	RepoTags    []string          `json:"repoTags"`
	RepoDigests []string          `json:"repoDigests"`
	Created     time.Time         `json:"created"`
	Size        int64             `json:"size"`
	Labels      map[string]string `json:"labels"`
}

// Network represents a container network
type Network struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	IPAM       NetworkIPAM       `json:"ipam"`
	Labels     map[string]string `json:"labels"`
}

// NetworkIPAM represents network IPAM configuration
type NetworkIPAM struct {
	Driver  string       `json:"driver"`
	Configs []IPAMConfig `json:"configs"`
}

// IPAMConfig represents IPAM configuration
type IPAMConfig struct {
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
}

// Volume represents a container volume
type Volume struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	CreatedAt  time.Time         `json:"createdAt"`
	Labels     map[string]string `json:"labels"`
	Scope      string            `json:"scope"`
}
