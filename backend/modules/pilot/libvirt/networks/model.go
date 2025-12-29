package networks

import (
	"github.com/google/uuid"
)

// Network represents a libvirt network
type Network struct {
	HypervisorID uuid.UUID `json:"hypervisorId"`
	UUID         string    `json:"uuid"`
	Name         string    `json:"name"`
	Bridge       string    `json:"bridge"`
	Active       bool      `json:"active"`
	Persistent   bool      `json:"persistent"`
	Autostart    bool      `json:"autostart"`
}

// NetworkFilter contains filter options
type NetworkFilter struct {
	Search *string `json:"search,omitempty"`
	Active *bool   `json:"active,omitempty"`
}
