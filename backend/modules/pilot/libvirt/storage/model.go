package storage

import (
	"github.com/google/uuid"
)

// StoragePool represents a libvirt storage pool
type StoragePool struct {
	HypervisorID uuid.UUID `json:"hypervisorId"`
	UUID         string    `json:"uuid"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	State        string    `json:"state"`
	Capacity     uint64    `json:"capacity"`     // bytes
	Allocation   uint64    `json:"allocation"`   // bytes
	Available    uint64    `json:"available"`    // bytes
	Active       bool      `json:"active"`
	Persistent   bool      `json:"persistent"`
	Autostart    bool      `json:"autostart"`
	VolumesCount int       `json:"volumesCount"`
}

// StorageVolume represents a libvirt storage volume
type StorageVolume struct {
	HypervisorID uuid.UUID `json:"hypervisorId"`
	PoolName     string    `json:"poolName"`
	Name         string    `json:"name"`
	Key          string    `json:"key"`
	Path         string    `json:"path"`
	Type         string    `json:"type"`
	Capacity     uint64    `json:"capacity"`    // bytes
	Allocation   uint64    `json:"allocation"`  // bytes
}

// StoragePoolFilter contains filter options
type StoragePoolFilter struct {
	Search *string `json:"search,omitempty"`
	Active *bool   `json:"active,omitempty"`
}

// StorageVolumeFilter contains filter options
type StorageVolumeFilter struct {
	Search *string `json:"search,omitempty"`
}

// CreateVolumeInput contains input for creating a volume
type CreateVolumeInput struct {
	Name     string `json:"name"`
	Capacity uint64 `json:"capacity"`  // bytes
	Format   string `json:"format"`    // qcow2, raw, etc.
}
