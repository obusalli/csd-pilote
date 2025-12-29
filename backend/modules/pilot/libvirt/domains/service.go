package domains

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"csd-pilote/backend/modules/pilot/hypervisors"
	csdcore "csd-pilote/backend/modules/platform/csd-core"
)

// Service handles domain operations via csd-core playbooks
type Service struct {
	hypervisorSvc *hypervisors.Service
	coreClient    *csdcore.Client
}

// NewService creates a new domain service
func NewService() *Service {
	return &Service{
		hypervisorSvc: hypervisors.NewService(),
		coreClient:    csdcore.GetClient(),
	}
}

// List returns all domains for a hypervisor
func (s *Service) List(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, filter *DomainFilter) ([]Domain, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "list-domains", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawDomains []rawDomain
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawDomains); err != nil {
		return nil, fmt.Errorf("failed to parse domains: %w", err)
	}

	domains := make([]Domain, 0, len(rawDomains))
	for _, dom := range rawDomains {
		// Apply filters
		if filter != nil {
			if filter.Search != nil && *filter.Search != "" {
				if !strings.Contains(strings.ToLower(dom.Name), strings.ToLower(*filter.Search)) {
					continue
				}
			}
			if filter.State != nil && *filter.State != "" {
				if dom.State != string(*filter.State) {
					continue
				}
			}
		}

		domains = append(domains, s.toDomain(hypervisorID, &dom))
	}

	return domains, nil
}

// Get returns a specific domain
func (s *Service) Get(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, domainUUID string) (*Domain, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "get-domain", map[string]interface{}{
		"uuid": domainUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawDom rawDomain
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawDom); err != nil {
		return nil, fmt.Errorf("failed to parse domain: %w", err)
	}

	result := s.toDomain(hypervisorID, &rawDom)
	return &result, nil
}

// Start starts a domain
func (s *Service) Start(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, domainUUID string) (*Domain, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "start-domain", map[string]interface{}{
		"uuid": domainUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start domain: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, hypervisorID, domainUUID)
}

// Shutdown shuts down a domain gracefully
func (s *Service) Shutdown(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, domainUUID string) (*Domain, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "shutdown-domain", map[string]interface{}{
		"uuid": domainUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to shutdown domain: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, hypervisorID, domainUUID)
}

// ForceStop forces a domain to stop
func (s *Service) ForceStop(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, domainUUID string) (*Domain, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "destroy-domain", map[string]interface{}{
		"uuid": domainUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to force stop domain: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, hypervisorID, domainUUID)
}

// Reboot reboots a domain
func (s *Service) Reboot(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, domainUUID string) (*Domain, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "reboot-domain", map[string]interface{}{
		"uuid": domainUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to reboot domain: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, hypervisorID, domainUUID)
}

// Delete deletes a domain
func (s *Service) Delete(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, domainUUID string, deleteVolumes bool) error {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "delete-domain", map[string]interface{}{
		"uuid":          domainUUID,
		"deleteVolumes": deleteVolumes,
	})
	if err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return fmt.Errorf("task failed: %s", execution.Error)
	}

	return nil
}

// SetAutostart sets the autostart flag for a domain
func (s *Service) SetAutostart(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, domainUUID string, autostart bool) (*Domain, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "set-domain-autostart", map[string]interface{}{
		"uuid":      domainUUID,
		"autostart": autostart,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set autostart: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, hypervisorID, domainUUID)
}

type rawDomain struct {
	ID         int    `json:"id"`
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	State      string `json:"state"`
	MaxMemory  uint64 `json:"maxMemory"`
	Memory     uint64 `json:"memory"`
	VCPUs      int    `json:"vcpus"`
	CPUTime    uint64 `json:"cpuTime"`
	Autostart  bool   `json:"autostart"`
	Persistent bool   `json:"persistent"`
}

func (s *Service) toDomain(hypervisorID uuid.UUID, dom *rawDomain) Domain {
	return Domain{
		HypervisorID: hypervisorID,
		ID:           dom.ID,
		UUID:         dom.UUID,
		Name:         dom.Name,
		State:        DomainState(dom.State),
		MaxMemory:    dom.MaxMemory,
		Memory:       dom.Memory,
		VCPUs:        dom.VCPUs,
		CPUTime:      dom.CPUTime,
		Autostart:    dom.Autostart,
		Persistent:   dom.Persistent,
	}
}
