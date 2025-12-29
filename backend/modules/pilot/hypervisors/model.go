package hypervisors

import (
	"time"

	"github.com/google/uuid"
)

// HypervisorStatus represents the status of a Libvirt hypervisor connection
type HypervisorStatus string

const (
	HypervisorStatusConnected    HypervisorStatus = "CONNECTED"
	HypervisorStatusDisconnected HypervisorStatus = "DISCONNECTED"
	HypervisorStatusError        HypervisorStatus = "ERROR"
	HypervisorStatusPending      HypervisorStatus = "PENDING"
)

// Hypervisor represents a Libvirt hypervisor configuration
type Hypervisor struct {
	ID            uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID        `json:"tenantId" gorm:"type:uuid;not null;index"`
	Name          string           `json:"name" gorm:"not null"`
	Description   string           `json:"description"`
	URI           string           `json:"uri" gorm:"not null"`                  // e.g., qemu+ssh://user@host/system
	ArtifactKey   string           `json:"artifactKey"`                          // Reference to SSH key artifact in csd-core (optional)
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

// HypervisorInput represents input for creating/updating a hypervisor
type HypervisorInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URI         string `json:"uri"`
	ArtifactKey string `json:"artifactKey"`
}

// HypervisorFilter represents filter options for listing hypervisors
type HypervisorFilter struct {
	Search *string           `json:"search"`
	Status *HypervisorStatus `json:"status"`
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
