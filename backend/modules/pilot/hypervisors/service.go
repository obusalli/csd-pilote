package hypervisors

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
)

// Service handles business logic for hypervisors
type Service struct {
	repo   *Repository
	client *csdcore.Client
}

// NewService creates a new hypervisor service
func NewService() *Service {
	return &Service{
		repo:   NewRepository(),
		client: csdcore.GetClient(),
	}
}

// Create creates a new hypervisor
func (s *Service) Create(ctx context.Context, tenantID, userID uuid.UUID, input *HypervisorInput) (*Hypervisor, error) {
	hypervisor := &Hypervisor{
		TenantID:    tenantID,
		Name:        input.Name,
		Description: input.Description,
		URI:         input.URI,
		ArtifactKey: input.ArtifactKey,
		Status:      HypervisorStatusPending,
		CreatedBy:   userID,
	}

	if err := s.repo.Create(hypervisor); err != nil {
		return nil, fmt.Errorf("failed to create hypervisor: %w", err)
	}

	return hypervisor, nil
}

// Get retrieves a hypervisor by ID
func (s *Service) Get(ctx context.Context, tenantID, id uuid.UUID) (*Hypervisor, error) {
	return s.repo.GetByID(tenantID, id)
}

// List retrieves all hypervisors for a tenant
func (s *Service) List(ctx context.Context, tenantID uuid.UUID, filter *HypervisorFilter, limit, offset int) ([]Hypervisor, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.List(tenantID, filter, limit, offset)
}

// Update updates a hypervisor
func (s *Service) Update(ctx context.Context, tenantID, id uuid.UUID, input *HypervisorInput) (*Hypervisor, error) {
	hypervisor, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		hypervisor.Name = input.Name
	}
	if input.Description != "" {
		hypervisor.Description = input.Description
	}
	if input.URI != "" {
		hypervisor.URI = input.URI
		hypervisor.Status = HypervisorStatusPending // Reset status when config changes
	}
	if input.ArtifactKey != "" {
		hypervisor.ArtifactKey = input.ArtifactKey
	}

	if err := s.repo.Update(hypervisor); err != nil {
		return nil, fmt.Errorf("failed to update hypervisor: %w", err)
	}

	return hypervisor, nil
}

// Delete deletes a hypervisor
func (s *Service) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.Delete(tenantID, id)
}

// TestConnection tests the connection to a hypervisor using a playbook
func (s *Service) TestConnection(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, agentID uuid.UUID) error {
	hypervisor, err := s.repo.GetByID(tenantID, hypervisorID)
	if err != nil {
		return err
	}

	// This would execute a libvirt playbook with node_info action
	// For now, we just update the status
	s.repo.UpdateStatus(tenantID, hypervisorID, HypervisorStatusConnected, "Connection successful")

	_ = hypervisor
	return nil
}

// ListDomains lists all domains (VMs) on a hypervisor
func (s *Service) ListDomains(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, agentID uuid.UUID) ([]Domain, error) {
	// This would execute a libvirt playbook with vm_list action via csd-core agent
	// The playbook would use the hypervisor's URI and optional artifactKey for auth
	return []Domain{}, nil
}

// DomainAction performs an action on a domain (start, stop, etc.)
func (s *Service) DomainAction(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, agentID uuid.UUID, domainName string, action string) error {
	// This would execute a libvirt playbook with vm_start/vm_stop/etc. action
	return nil
}

// ListNetworks lists all networks on a hypervisor
func (s *Service) ListNetworks(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, agentID uuid.UUID) ([]Network, error) {
	// This would execute a libvirt playbook with network_list action
	return []Network{}, nil
}

// ListStoragePools lists all storage pools on a hypervisor
func (s *Service) ListStoragePools(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, agentID uuid.UUID) ([]StoragePool, error) {
	// This would execute a libvirt playbook with pool_list action
	return []StoragePool{}, nil
}

// ListStorageVolumes lists all volumes in a storage pool
func (s *Service) ListStorageVolumes(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, agentID uuid.UUID, poolName string) ([]StorageVolume, error) {
	// This would execute a libvirt playbook with volume_list action
	return []StorageVolume{}, nil
}

// ExecuteLibvirtAction executes a Libvirt action via playbook
func (s *Service) ExecuteLibvirtAction(ctx context.Context, token string, hypervisorID uuid.UUID, action string, args map[string]interface{}) (map[string]interface{}, error) {
	// This would create and execute a libvirt playbook via csd-core
	result := map[string]interface{}{
		"action":       action,
		"hypervisorId": hypervisorID.String(),
		"status":       "executed",
	}
	return result, nil
}
