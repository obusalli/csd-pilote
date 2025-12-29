package networks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"csd-pilote/backend/modules/pilot/hypervisors"
	csdcore "csd-pilote/backend/modules/platform/csd-core"
)

// Service handles network operations via csd-core playbooks
type Service struct {
	hypervisorSvc *hypervisors.Service
	coreClient    *csdcore.Client
}

// NewService creates a new network service
func NewService() *Service {
	return &Service{
		hypervisorSvc: hypervisors.NewService(),
		coreClient:    csdcore.GetClient(),
	}
}

// List returns all networks for a hypervisor
func (s *Service) List(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, filter *NetworkFilter) ([]Network, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "list-networks", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawNetworks []rawNetwork
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawNetworks); err != nil {
		return nil, fmt.Errorf("failed to parse networks: %w", err)
	}

	networks := make([]Network, 0, len(rawNetworks))
	for _, net := range rawNetworks {
		// Apply filters
		if filter != nil {
			if filter.Search != nil && *filter.Search != "" {
				if !strings.Contains(strings.ToLower(net.Name), strings.ToLower(*filter.Search)) {
					continue
				}
			}
			if filter.Active != nil {
				if net.Active != *filter.Active {
					continue
				}
			}
		}

		networks = append(networks, s.toNetwork(hypervisorID, &net))
	}

	return networks, nil
}

// Get returns a specific network
func (s *Service) Get(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, name string) (*Network, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "get-network", map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get network: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawNet rawNetwork
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawNet); err != nil {
		return nil, fmt.Errorf("failed to parse network: %w", err)
	}

	result := s.toNetwork(hypervisorID, &rawNet)
	return &result, nil
}

// Start starts a network
func (s *Service) Start(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, name string) (*Network, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "start-network", map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start network: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, hypervisorID, name)
}

// Stop stops a network
func (s *Service) Stop(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, name string) (*Network, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "stop-network", map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to stop network: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, hypervisorID, name)
}

// SetAutostart sets the autostart flag for a network
func (s *Service) SetAutostart(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, name string, autostart bool) (*Network, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "set-network-autostart", map[string]interface{}{
		"name":      name,
		"autostart": autostart,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set autostart: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, hypervisorID, name)
}

// Delete deletes a network
func (s *Service) Delete(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, name string) error {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "delete-network", map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return fmt.Errorf("failed to delete network: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return fmt.Errorf("task failed: %s", execution.Error)
	}

	return nil
}

type rawNetwork struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	Bridge     string `json:"bridge"`
	Active     bool   `json:"active"`
	Persistent bool   `json:"persistent"`
	Autostart  bool   `json:"autostart"`
}

func (s *Service) toNetwork(hypervisorID uuid.UUID, net *rawNetwork) Network {
	return Network{
		HypervisorID: hypervisorID,
		UUID:         net.UUID,
		Name:         net.Name,
		Bridge:       net.Bridge,
		Active:       net.Active,
		Persistent:   net.Persistent,
		Autostart:    net.Autostart,
	}
}
