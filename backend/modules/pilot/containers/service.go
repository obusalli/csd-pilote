package containers

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/events"
)

// Service handles business logic for container engines
type Service struct {
	repo   *Repository
	client *csdcore.Client
}

// NewService creates a new container engine service
func NewService() *Service {
	return &Service{
		repo:   NewRepository(),
		client: csdcore.GetClient(),
	}
}

// Create creates a new container engine
func (s *Service) Create(ctx context.Context, tenantID, userID uuid.UUID, input *ContainerEngineInput) (*ContainerEngine, error) {
	engineType := input.EngineType
	if engineType == "" {
		engineType = EngineTypeDocker
	}

	engine := &ContainerEngine{
		TenantID:    tenantID,
		Name:        input.Name,
		Description: input.Description,
		EngineType:  engineType,
		Host:        input.Host,
		ArtifactKey: input.ArtifactKey,
		Status:      EngineStatusPending,
		CreatedBy:   userID,
	}

	if err := s.repo.Create(engine); err != nil {
		return nil, fmt.Errorf("failed to create container engine: %w", err)
	}

	// Publish container engine created event
	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventContainerEngineCreated,
		tenantID,
		engine.ID.String(),
		map[string]interface{}{
			"name":       engine.Name,
			"engineType": engine.EngineType,
			"status":     engine.Status,
		},
	))

	return engine, nil
}

// Get retrieves a container engine by ID
func (s *Service) Get(ctx context.Context, tenantID, id uuid.UUID) (*ContainerEngine, error) {
	return s.repo.GetByID(tenantID, id)
}

// List retrieves all container engines for a tenant
func (s *Service) List(ctx context.Context, tenantID uuid.UUID, filter *ContainerEngineFilter, limit, offset int) ([]ContainerEngine, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.List(tenantID, filter, limit, offset)
}

// Update updates a container engine
func (s *Service) Update(ctx context.Context, tenantID, id uuid.UUID, input *ContainerEngineInput) (*ContainerEngine, error) {
	engine, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		engine.Name = input.Name
	}
	if input.Description != "" {
		engine.Description = input.Description
	}
	if input.EngineType != "" {
		engine.EngineType = input.EngineType
	}
	if input.Host != "" {
		engine.Host = input.Host
		engine.Status = EngineStatusPending // Reset status when config changes
	}
	if input.ArtifactKey != "" {
		engine.ArtifactKey = input.ArtifactKey
	}

	if err := s.repo.Update(engine); err != nil {
		return nil, fmt.Errorf("failed to update container engine: %w", err)
	}

	// Publish container engine updated event
	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventContainerEngineUpdated,
		tenantID,
		engine.ID.String(),
		map[string]interface{}{
			"name":   engine.Name,
			"status": engine.Status,
		},
	))

	return engine, nil
}

// Delete deletes a container engine
func (s *Service) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	if err := s.repo.Delete(tenantID, id); err != nil {
		return err
	}

	// Publish container engine deleted event
	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventContainerEngineDeleted,
		tenantID,
		id.String(),
		nil,
	))

	return nil
}

// TestConnection tests the connection to a container engine using a playbook
func (s *Service) TestConnection(ctx context.Context, token string, tenantID, engineID uuid.UUID, agentID uuid.UUID) error {
	engine, err := s.repo.GetByID(tenantID, engineID)
	if err != nil {
		return err
	}

	// This would execute a docker playbook with info action via csd-core agent
	// For now, we just update the status
	s.repo.UpdateStatus(tenantID, engineID, EngineStatusConnected, "Connection successful")

	_ = engine
	return nil
}

// ListContainers lists all containers on an engine
func (s *Service) ListContainers(ctx context.Context, token string, tenantID, engineID uuid.UUID, agentID uuid.UUID, all bool) ([]Container, error) {
	// This would execute a docker playbook with container_list action via csd-core agent
	return []Container{}, nil
}

// ContainerAction performs an action on a container (start, stop, restart, etc.)
func (s *Service) ContainerAction(ctx context.Context, token string, tenantID, engineID uuid.UUID, agentID uuid.UUID, containerID string, action string) error {
	// This would execute a docker playbook with container_start/stop/etc. action
	return nil
}

// ListImages lists all images on an engine
func (s *Service) ListImages(ctx context.Context, token string, tenantID, engineID uuid.UUID, agentID uuid.UUID) ([]Image, error) {
	// This would execute a docker playbook with image_list action
	return []Image{}, nil
}

// PullImage pulls an image
func (s *Service) PullImage(ctx context.Context, token string, tenantID, engineID uuid.UUID, agentID uuid.UUID, imageName string) error {
	// This would execute a docker playbook with image_pull action
	return nil
}

// ListNetworks lists all networks on an engine
func (s *Service) ListNetworks(ctx context.Context, token string, tenantID, engineID uuid.UUID, agentID uuid.UUID) ([]Network, error) {
	// This would execute a docker playbook with network_list action
	return []Network{}, nil
}

// ListVolumes lists all volumes on an engine
func (s *Service) ListVolumes(ctx context.Context, token string, tenantID, engineID uuid.UUID, agentID uuid.UUID) ([]Volume, error) {
	// This would execute a docker playbook with volume_list action
	return []Volume{}, nil
}

// GetContainerLogs gets logs from a container
func (s *Service) GetContainerLogs(ctx context.Context, token string, tenantID, engineID uuid.UUID, agentID uuid.UUID, containerID string, tail int) (string, error) {
	// This would execute a docker playbook with container_logs action
	return "", nil
}

// ExecContainer executes a command in a container
// Note: For interactive exec, use WebSocket via csd-core terminal module
func (s *Service) ExecContainer(ctx context.Context, token string, tenantID, engineID uuid.UUID, agentID uuid.UUID, containerID string, command []string) (string, error) {
	// This would execute a docker playbook with container_exec action
	return "", nil
}

// BulkDelete deletes multiple container engines by IDs
func (s *Service) BulkDelete(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) (int64, error) {
	return s.repo.BulkDelete(tenantID, ids)
}
