package containers

import (
	"testing"
)

func TestContainerEngineFilter(t *testing.T) {
	t.Run("empty filter", func(t *testing.T) {
		filter := &ContainerEngineFilter{}
		if filter.Search != nil {
			t.Error("Expected Search to be nil")
		}
		if filter.Status != nil {
			t.Error("Expected Status to be nil")
		}
	})

	t.Run("filter with search", func(t *testing.T) {
		search := "docker"
		filter := &ContainerEngineFilter{Search: &search}
		if filter.Search == nil || *filter.Search != "docker" {
			t.Error("Expected Search to be 'docker'")
		}
	})

	t.Run("filter with engine type", func(t *testing.T) {
		engineType := EngineTypeDocker
		filter := &ContainerEngineFilter{EngineType: &engineType}
		if filter.EngineType == nil || *filter.EngineType != EngineTypeDocker {
			t.Error("Expected EngineType to be DOCKER")
		}
	})
}

func TestContainerEngineModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		engine := ContainerEngine{}
		if engine.TableName() != "container_engines" {
			t.Errorf("Expected table name 'container_engines', got '%s'", engine.TableName())
		}
	})

	t.Run("engine type values", func(t *testing.T) {
		validTypes := []EngineType{
			EngineTypeDocker,
			EngineTypePodman,
		}
		for _, engineType := range validTypes {
			if engineType == "" {
				t.Error("Engine type should not be empty")
			}
		}
	})

	t.Run("engine status values", func(t *testing.T) {
		validStatuses := []EngineStatus{
			EngineStatusConnected,
			EngineStatusDisconnected,
			EngineStatusError,
			EngineStatusPending,
		}
		for _, status := range validStatuses {
			if status == "" {
				t.Error("Status should not be empty")
			}
		}
	})
}

func TestContainerEngineInput(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		input := ContainerEngineInput{
			Name:        "test-engine",
			Description: "Test engine",
			EngineType:  EngineTypeDocker,
			Host:        "unix:///var/run/docker.sock",
		}
		if input.Name != "test-engine" {
			t.Error("Expected name to be 'test-engine'")
		}
		if input.EngineType != EngineTypeDocker {
			t.Error("Expected engine type to be DOCKER")
		}
		if input.Host != "unix:///var/run/docker.sock" {
			t.Error("Expected host to be 'unix:///var/run/docker.sock'")
		}
	})
}

func TestContainerModel(t *testing.T) {
	t.Run("container port", func(t *testing.T) {
		port := ContainerPort{
			IP:          "0.0.0.0",
			PrivatePort: 80,
			PublicPort:  8080,
			Type:        "tcp",
		}
		if port.PrivatePort != 80 {
			t.Error("Expected private port to be 80")
		}
		if port.PublicPort != 8080 {
			t.Error("Expected public port to be 8080")
		}
	})

	t.Run("container mount", func(t *testing.T) {
		mount := ContainerMount{
			Type:        "bind",
			Source:      "/host/path",
			Destination: "/container/path",
			Mode:        "rw",
			RW:          true,
		}
		if mount.Type != "bind" {
			t.Error("Expected mount type to be 'bind'")
		}
		if !mount.RW {
			t.Error("Expected mount to be read-write")
		}
	})
}

func TestNetworkModel(t *testing.T) {
	t.Run("network IPAM", func(t *testing.T) {
		ipam := NetworkIPAM{
			Driver: "default",
			Configs: []IPAMConfig{
				{
					Subnet:  "172.17.0.0/16",
					Gateway: "172.17.0.1",
				},
			},
		}
		if ipam.Driver != "default" {
			t.Error("Expected IPAM driver to be 'default'")
		}
		if len(ipam.Configs) != 1 {
			t.Error("Expected 1 IPAM config")
		}
	})
}
