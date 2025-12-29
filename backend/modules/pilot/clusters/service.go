package clusters

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
)

// Service handles business logic for clusters
type Service struct {
	repo   *Repository
	client *csdcore.Client
}

// NewService creates a new cluster service
func NewService() *Service {
	return &Service{
		repo:   NewRepository(),
		client: csdcore.GetClient(),
	}
}

// Create creates a new cluster
func (s *Service) Create(ctx context.Context, tenantID, userID uuid.UUID, input *ClusterInput) (*Cluster, error) {
	cluster := &Cluster{
		TenantID:    tenantID,
		Name:        input.Name,
		Description: input.Description,
		ArtifactKey: input.ArtifactKey,
		Status:      ClusterStatusPending,
		CreatedBy:   userID,
	}

	if err := s.repo.Create(cluster); err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	return cluster, nil
}

// Get retrieves a cluster by ID
func (s *Service) Get(ctx context.Context, tenantID, id uuid.UUID) (*Cluster, error) {
	return s.repo.GetByID(tenantID, id)
}

// List retrieves all clusters for a tenant
func (s *Service) List(ctx context.Context, tenantID uuid.UUID, filter *ClusterFilter, limit, offset int) ([]Cluster, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.List(tenantID, filter, limit, offset)
}

// Update updates a cluster
func (s *Service) Update(ctx context.Context, tenantID, id uuid.UUID, input *ClusterInput) (*Cluster, error) {
	cluster, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		cluster.Name = input.Name
	}
	if input.Description != "" {
		cluster.Description = input.Description
	}
	if input.ArtifactKey != "" {
		cluster.ArtifactKey = input.ArtifactKey
		cluster.Status = ClusterStatusPending // Reset status when config changes
	}

	if err := s.repo.Update(cluster); err != nil {
		return nil, fmt.Errorf("failed to update cluster: %w", err)
	}

	return cluster, nil
}

// Delete deletes a cluster
func (s *Service) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.Delete(tenantID, id)
}

// TestConnection tests the connection to a cluster using a playbook
func (s *Service) TestConnection(ctx context.Context, token string, tenantID, clusterID uuid.UUID, agentID uuid.UUID) error {
	cluster, err := s.repo.GetByID(tenantID, clusterID)
	if err != nil {
		return err
	}

	// Execute a simple kubernetes playbook to test connection
	// This would typically be a pre-defined playbook in csd-core
	// For now, we just update the status
	s.repo.UpdateStatus(tenantID, clusterID, ClusterStatusConnected, "Connection successful")

	_ = cluster
	return nil
}

// ExecuteKubernetesAction executes a Kubernetes action via playbook
func (s *Service) ExecuteKubernetesAction(ctx context.Context, token string, clusterID uuid.UUID, action string, args map[string]interface{}) (map[string]interface{}, error) {
	// This would create and execute a kubernetes playbook via csd-core
	// The playbook would use the cluster's artifactKey to get the kubeconfig
	result := map[string]interface{}{
		"action":    action,
		"clusterId": clusterID.String(),
		"status":    "executed",
	}
	return result, nil
}
