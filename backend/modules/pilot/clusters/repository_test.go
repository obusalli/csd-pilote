package clusters

import (
	"testing"

	"github.com/google/uuid"
)

func TestClusterFilter(t *testing.T) {
	t.Run("empty filter", func(t *testing.T) {
		filter := &ClusterFilter{}
		if filter.Search != nil {
			t.Error("Expected Search to be nil")
		}
		if filter.Status != nil {
			t.Error("Expected Status to be nil")
		}
	})

	t.Run("filter with search", func(t *testing.T) {
		search := "test"
		filter := &ClusterFilter{Search: &search}
		if filter.Search == nil || *filter.Search != "test" {
			t.Error("Expected Search to be 'test'")
		}
	})

	t.Run("filter with status", func(t *testing.T) {
		status := ClusterStatusConnected
		filter := &ClusterFilter{Status: &status}
		if filter.Status == nil || *filter.Status != ClusterStatusConnected {
			t.Error("Expected Status to be CONNECTED")
		}
	})
}

func TestClusterModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		cluster := Cluster{}
		if cluster.TableName() != "clusters" {
			t.Errorf("Expected table name 'clusters', got '%s'", cluster.TableName())
		}
	})

	t.Run("cluster node table name", func(t *testing.T) {
		node := ClusterNode{}
		if node.TableName() != "cluster_nodes" {
			t.Errorf("Expected table name 'cluster_nodes', got '%s'", node.TableName())
		}
	})

	t.Run("cluster status values", func(t *testing.T) {
		validStatuses := []ClusterStatus{
			ClusterStatusPending,
			ClusterStatusDeploying,
			ClusterStatusConnected,
			ClusterStatusDisconnected,
			ClusterStatusError,
		}
		for _, status := range validStatuses {
			if status == "" {
				t.Error("Status should not be empty")
			}
		}
	})

	t.Run("cluster mode values", func(t *testing.T) {
		validModes := []ClusterMode{
			ClusterModeConnect,
			ClusterModeDeploy,
		}
		for _, mode := range validModes {
			if mode == "" {
				t.Error("Mode should not be empty")
			}
		}
	})

	t.Run("kubernetes distribution values", func(t *testing.T) {
		validDistros := []KubernetesDistribution{
			K8sDistroK3s,
			K8sDistroRKE2,
			K8sDistroKubeadm,
		}
		for _, distro := range validDistros {
			if distro == "" {
				t.Error("Distribution should not be empty")
			}
		}
	})
}

func TestClusterInput(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		agentID := uuid.New()
		input := ClusterInput{
			Name:         "test-cluster",
			Description:  "Test cluster",
			Distribution: K8sDistroK3s,
			AgentID:      agentID.String(),
		}
		if input.Name != "test-cluster" {
			t.Error("Expected name to be 'test-cluster'")
		}
		if input.Distribution != K8sDistroK3s {
			t.Error("Expected distribution to be K3S")
		}
	})
}

func TestClusterNodeRole(t *testing.T) {
	t.Run("valid roles", func(t *testing.T) {
		roles := []NodeRole{
			NodeRoleMaster,
			NodeRoleWorker,
		}
		for _, role := range roles {
			if role == "" {
				t.Error("Role should not be empty")
			}
		}
	})
}
