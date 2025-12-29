package hypervisors

import (
	"time"

	"github.com/google/uuid"
)

// HypervisorStatus represents the status of a Libvirt hypervisor
type HypervisorStatus string

const (
	HypervisorStatusPending      HypervisorStatus = "PENDING"
	HypervisorStatusDeploying    HypervisorStatus = "DEPLOYING"
	HypervisorStatusConnected    HypervisorStatus = "CONNECTED"
	HypervisorStatusDisconnected HypervisorStatus = "DISCONNECTED"
	HypervisorStatusError        HypervisorStatus = "ERROR"
)

// HypervisorMode represents how the hypervisor was added
type HypervisorMode string

const (
	HypervisorModeConnect HypervisorMode = "CONNECT" // Connect to existing hypervisor
	HypervisorModeDeploy  HypervisorMode = "DEPLOY"  // Deploy libvirt on agent
)

// LibvirtDriver represents the virtualization driver
type LibvirtDriver string

const (
	LibvirtDriverQEMU LibvirtDriver = "QEMU" // QEMU/KVM
	LibvirtDriverXen  LibvirtDriver = "XEN"  // Xen hypervisor
	LibvirtDriverLXC  LibvirtDriver = "LXC"  // Linux Containers
)

// Hypervisor represents a Libvirt hypervisor configuration
type Hypervisor struct {
	ID            uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID        `json:"tenantId" gorm:"type:uuid;not null;index"`
	Name          string           `json:"name" gorm:"not null"`
	Description   string           `json:"description"`
	Mode          HypervisorMode   `json:"mode" gorm:"default:'CONNECT'"`       // CONNECT or DEPLOY
	Driver        LibvirtDriver    `json:"driver" gorm:"default:'QEMU'"`        // QEMU, XEN, LXC
	AgentID       uuid.UUID        `json:"agentId" gorm:"type:uuid;not null"`   // csd-core agent
	URI           string           `json:"uri"`                                  // e.g., qemu+ssh://user@host/system (for CONNECT)
	ArtifactKey   string           `json:"artifactKey"`                          // Reference to SSH key artifact (optional)
	Status        HypervisorStatus `json:"status" gorm:"default:'PENDING'"`
	StatusMessage string           `json:"statusMessage"`
	// Cached info from hypervisor
	Hostname       string     `json:"hostname"`
	LibvirtVersion string     `json:"libvirtVersion"`
	HypervisorType string     `json:"hypervisorType"` // QEMU, KVM, Xen, etc.
	CPUModel       string     `json:"cpuModel"`
	TotalMemoryMB  int64      `json:"totalMemoryMb"`
	TotalCPUs      int        `json:"totalCpus"`
	LastCheckedAt  *time.Time `json:"lastCheckedAt"`
	CreatedAt      time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
	CreatedBy      uuid.UUID  `json:"createdBy" gorm:"type:uuid"`
}

// TableName returns the table name for GORM
func (Hypervisor) TableName() string {
	return "hypervisors"
}

// HypervisorInput represents input for connecting to an existing hypervisor
type HypervisorInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AgentID     string `json:"agentId"`
	URI         string `json:"uri"`
	ArtifactKey string `json:"artifactKey"`
}

// DeployHypervisorInput represents input for deploying libvirt on an agent
type DeployHypervisorInput struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	AgentID     string        `json:"agentId"` // Agent where to deploy libvirt
	Driver      LibvirtDriver `json:"driver"`  // QEMU, XEN, LXC
}

// HypervisorFilter represents filter options for listing hypervisors
type HypervisorFilter struct {
	Search *string           `json:"search"`
	Status *HypervisorStatus `json:"status"`
	Mode   *HypervisorMode   `json:"mode"`
	Driver *LibvirtDriver    `json:"driver"`
}

// Domain represents a Libvirt domain (VM)
type Domain struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	State     string `json:"state"` // running, paused, shutoff, etc.
	MemoryKB  uint64 `json:"memoryKb"`
	MaxMemKB  uint64 `json:"maxMemKb"`
	VCPUs     uint   `json:"vcpus"`
	CPUTime   uint64 `json:"cpuTime"`
	Autostart bool   `json:"autostart"`
}

// Network represents a Libvirt network
type Network struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	Active     bool   `json:"active"`
	Persistent bool   `json:"persistent"`
	Autostart  bool   `json:"autostart"`
	Bridge     string `json:"bridge"`
}

// StoragePool represents a Libvirt storage pool
type StoragePool struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	State      string `json:"state"` // active, inactive
	Type       string `json:"type"`  // dir, lvm, nfs, etc.
	CapacityGB float64 `json:"capacityGb"`
	UsedGB     float64 `json:"usedGb"`
	AvailGB    float64 `json:"availGb"`
	Autostart  bool   `json:"autostart"`
}

// StorageVolume represents a Libvirt storage volume
type StorageVolume struct {
	Name       string  `json:"name"`
	Path       string  `json:"path"`
	Type       string  `json:"type"` // file, block, dir
	Format     string  `json:"format"` // qcow2, raw, etc.
	CapacityGB float64 `json:"capacityGb"`
	UsedGB     float64 `json:"usedGb"`
}
