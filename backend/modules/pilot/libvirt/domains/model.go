package domains

import (
	"github.com/google/uuid"
)

// DomainState represents the state of a domain
type DomainState string

const (
	DomainStateRunning     DomainState = "RUNNING"
	DomainStatePaused      DomainState = "PAUSED"
	DomainStateShutoff     DomainState = "SHUTOFF"
	DomainStateCrashed     DomainState = "CRASHED"
	DomainStateSuspended   DomainState = "SUSPENDED"
	DomainStateUnknown     DomainState = "UNKNOWN"
)

// Domain represents a libvirt domain (VM)
type Domain struct {
	HypervisorID uuid.UUID `json:"hypervisorId"`
	ID           int       `json:"id"`
	UUID         string    `json:"uuid"`
	Name         string    `json:"name"`
	State        DomainState `json:"state"`
	MaxMemory    uint64    `json:"maxMemory"`   // KB
	Memory       uint64    `json:"memory"`      // KB
	VCPUs        int       `json:"vcpus"`
	CPUTime      uint64    `json:"cpuTime"`     // nanoseconds
	Autostart    bool      `json:"autostart"`
	Persistent   bool      `json:"persistent"`
}

// DomainFilter contains filter options
type DomainFilter struct {
	Search *string      `json:"search,omitempty"`
	State  *DomainState `json:"state,omitempty"`
}

// CreateDomainInput contains input for creating a domain
type CreateDomainInput struct {
	Name     string `json:"name"`
	Memory   uint64 `json:"memory"`    // MB
	VCPUs    int    `json:"vcpus"`
	DiskSize uint64 `json:"diskSize"`  // GB
	DiskPool string `json:"diskPool"`
	Network  string `json:"network"`
	ISOPath  string `json:"isoPath,omitempty"`
}

// DomainStats contains domain statistics
type DomainStats struct {
	CPUTime    uint64  `json:"cpuTime"`
	CPUPercent float64 `json:"cpuPercent"`
	Memory     uint64  `json:"memory"`
	MaxMemory  uint64  `json:"maxMemory"`
}
