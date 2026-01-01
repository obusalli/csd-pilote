package hypervisors

import (
	"testing"

	"github.com/google/uuid"
)

func TestHypervisorFilter(t *testing.T) {
	t.Run("empty filter", func(t *testing.T) {
		filter := &HypervisorFilter{}
		if filter.Search != nil {
			t.Error("Expected Search to be nil")
		}
		if filter.Status != nil {
			t.Error("Expected Status to be nil")
		}
	})

	t.Run("filter with search", func(t *testing.T) {
		search := "kvm"
		filter := &HypervisorFilter{Search: &search}
		if filter.Search == nil || *filter.Search != "kvm" {
			t.Error("Expected Search to be 'kvm'")
		}
	})

	t.Run("filter with driver", func(t *testing.T) {
		driver := LibvirtDriverQEMU
		filter := &HypervisorFilter{Driver: &driver}
		if filter.Driver == nil || *filter.Driver != LibvirtDriverQEMU {
			t.Error("Expected Driver to be 'qemu'")
		}
	})
}

func TestHypervisorModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		hv := Hypervisor{}
		if hv.TableName() != "hypervisors" {
			t.Errorf("Expected table name 'hypervisors', got '%s'", hv.TableName())
		}
	})

	t.Run("hypervisor status values", func(t *testing.T) {
		validStatuses := []HypervisorStatus{
			HypervisorStatusPending,
			HypervisorStatusDeploying,
			HypervisorStatusConnected,
			HypervisorStatusDisconnected,
			HypervisorStatusError,
		}
		for _, status := range validStatuses {
			if status == "" {
				t.Error("Status should not be empty")
			}
		}
	})

	t.Run("hypervisor mode values", func(t *testing.T) {
		validModes := []HypervisorMode{
			HypervisorModeConnect,
			HypervisorModeDeploy,
		}
		for _, mode := range validModes {
			if mode == "" {
				t.Error("Mode should not be empty")
			}
		}
	})

	t.Run("libvirt driver values", func(t *testing.T) {
		validDrivers := []LibvirtDriver{
			LibvirtDriverQEMU,
			LibvirtDriverXen,
			LibvirtDriverLXC,
		}
		for _, driver := range validDrivers {
			if driver == "" {
				t.Error("Driver should not be empty")
			}
		}
	})
}

func TestHypervisorInput(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		agentID := uuid.New()
		input := HypervisorInput{
			Name:        "test-hypervisor",
			Description: "Test hypervisor",
			AgentID:     agentID.String(),
			URI:         "qemu:///system",
		}
		if input.Name != "test-hypervisor" {
			t.Error("Expected name to be 'test-hypervisor'")
		}
		if input.URI != "qemu:///system" {
			t.Error("Expected URI to be 'qemu:///system'")
		}
	})
}
