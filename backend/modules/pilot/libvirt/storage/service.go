package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"csd-pilote/backend/modules/pilot/hypervisors"
	csdcore "csd-pilote/backend/modules/platform/csd-core"
)

// Service handles storage operations via csd-core playbooks
type Service struct {
	hypervisorSvc *hypervisors.Service
	coreClient    *csdcore.Client
}

// NewService creates a new storage service
func NewService() *Service {
	return &Service{
		hypervisorSvc: hypervisors.NewService(),
		coreClient:    csdcore.GetClient(),
	}
}

// ListPools returns all storage pools for a hypervisor
func (s *Service) ListPools(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, filter *StoragePoolFilter) ([]StoragePool, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "list-storage-pools", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list storage pools: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawPools []rawStoragePool
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawPools); err != nil {
		return nil, fmt.Errorf("failed to parse storage pools: %w", err)
	}

	pools := make([]StoragePool, 0, len(rawPools))
	for _, pool := range rawPools {
		// Apply filters
		if filter != nil {
			if filter.Search != nil && *filter.Search != "" {
				if !strings.Contains(strings.ToLower(pool.Name), strings.ToLower(*filter.Search)) {
					continue
				}
			}
			if filter.Active != nil {
				if pool.Active != *filter.Active {
					continue
				}
			}
		}

		pools = append(pools, s.toStoragePool(hypervisorID, &pool))
	}

	return pools, nil
}

// GetPool returns a specific storage pool
func (s *Service) GetPool(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, name string) (*StoragePool, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "get-storage-pool", map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get storage pool: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawPool rawStoragePool
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawPool); err != nil {
		return nil, fmt.Errorf("failed to parse storage pool: %w", err)
	}

	result := s.toStoragePool(hypervisorID, &rawPool)
	return &result, nil
}

// StartPool starts a storage pool
func (s *Service) StartPool(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, name string) (*StoragePool, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "start-storage-pool", map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start storage pool: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.GetPool(ctx, token, tenantID, hypervisorID, name)
}

// StopPool stops a storage pool
func (s *Service) StopPool(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, name string) (*StoragePool, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "stop-storage-pool", map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to stop storage pool: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.GetPool(ctx, token, tenantID, hypervisorID, name)
}

// RefreshPool refreshes a storage pool
func (s *Service) RefreshPool(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, name string) (*StoragePool, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "refresh-storage-pool", map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to refresh storage pool: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.GetPool(ctx, token, tenantID, hypervisorID, name)
}

// ListVolumes returns all volumes in a storage pool
func (s *Service) ListVolumes(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, poolName string, filter *StorageVolumeFilter) ([]StorageVolume, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "list-storage-volumes", map[string]interface{}{
		"poolName": poolName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawVolumes []rawStorageVolume
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawVolumes); err != nil {
		return nil, fmt.Errorf("failed to parse volumes: %w", err)
	}

	volumes := make([]StorageVolume, 0, len(rawVolumes))
	for _, vol := range rawVolumes {
		// Apply filters
		if filter != nil {
			if filter.Search != nil && *filter.Search != "" {
				if !strings.Contains(strings.ToLower(vol.Name), strings.ToLower(*filter.Search)) {
					continue
				}
			}
		}

		volumes = append(volumes, s.toStorageVolume(hypervisorID, poolName, &vol))
	}

	return volumes, nil
}

// GetVolume returns a specific volume
func (s *Service) GetVolume(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, poolName, volumeName string) (*StorageVolume, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "get-storage-volume", map[string]interface{}{
		"poolName":   poolName,
		"volumeName": volumeName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawVol rawStorageVolume
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawVol); err != nil {
		return nil, fmt.Errorf("failed to parse volume: %w", err)
	}

	result := s.toStorageVolume(hypervisorID, poolName, &rawVol)
	return &result, nil
}

// CreateVolume creates a new volume
func (s *Service) CreateVolume(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, poolName string, input *CreateVolumeInput) (*StorageVolume, error) {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor not found: %w", err)
	}

	format := input.Format
	if format == "" {
		format = "qcow2"
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "create-storage-volume", map[string]interface{}{
		"poolName": poolName,
		"name":     input.Name,
		"capacity": input.Capacity,
		"format":   format,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.GetVolume(ctx, token, tenantID, hypervisorID, poolName, input.Name)
}

// DeleteVolume deletes a volume
func (s *Service) DeleteVolume(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, poolName, volumeName string) error {
	hv, err := s.hypervisorSvc.Get(ctx, tenantID, hypervisorID)
	if err != nil {
		return fmt.Errorf("hypervisor not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteLibvirtTask(ctx, token, hv.AgentID, hv.URI, hv.ArtifactKey, "delete-storage-volume", map[string]interface{}{
		"poolName":   poolName,
		"volumeName": volumeName,
	})
	if err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return fmt.Errorf("task failed: %s", execution.Error)
	}

	return nil
}

type rawStoragePool struct {
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	State        string `json:"state"`
	Capacity     uint64 `json:"capacity"`
	Allocation   uint64 `json:"allocation"`
	Available    uint64 `json:"available"`
	Active       bool   `json:"active"`
	Persistent   bool   `json:"persistent"`
	Autostart    bool   `json:"autostart"`
	VolumesCount int    `json:"volumesCount"`
}

type rawStorageVolume struct {
	Name       string `json:"name"`
	Key        string `json:"key"`
	Path       string `json:"path"`
	Type       string `json:"type"`
	Capacity   uint64 `json:"capacity"`
	Allocation uint64 `json:"allocation"`
}

func (s *Service) toStoragePool(hypervisorID uuid.UUID, pool *rawStoragePool) StoragePool {
	return StoragePool{
		HypervisorID: hypervisorID,
		UUID:         pool.UUID,
		Name:         pool.Name,
		State:        pool.State,
		Capacity:     pool.Capacity,
		Allocation:   pool.Allocation,
		Available:    pool.Available,
		Active:       pool.Active,
		Persistent:   pool.Persistent,
		Autostart:    pool.Autostart,
		VolumesCount: pool.VolumesCount,
	}
}

func (s *Service) toStorageVolume(hypervisorID uuid.UUID, poolName string, vol *rawStorageVolume) StorageVolume {
	return StorageVolume{
		HypervisorID: hypervisorID,
		PoolName:     poolName,
		Name:         vol.Name,
		Key:          vol.Key,
		Path:         vol.Path,
		Type:         vol.Type,
		Capacity:     vol.Capacity,
		Allocation:   vol.Allocation,
	}
}
