package hypervisors

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"csd-pilote/backend/modules/platform/config"
	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/events"
	"csd-pilote/backend/modules/platform/logger"
	"csd-pilote/backend/modules/platform/middleware"
	"csd-pilote/backend/modules/platform/pagination"
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

// Create creates a new hypervisor (CONNECT mode - connect to existing hypervisor)
func (s *Service) Create(ctx context.Context, tenantID, userID uuid.UUID, input *HypervisorInput) (*Hypervisor, error) {
	agentID, err := uuid.Parse(input.AgentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agentId: %w", err)
	}

	hypervisor := &Hypervisor{
		TenantID:    tenantID,
		Name:        input.Name,
		Description: input.Description,
		Mode:        HypervisorModeConnect,
		AgentID:     agentID,
		URI:         input.URI,
		ArtifactKey: input.ArtifactKey,
		Status:      HypervisorStatusPending,
		CreatedBy:   userID,
	}

	if err := s.repo.Create(hypervisor); err != nil {
		return nil, fmt.Errorf("failed to create hypervisor: %w", err)
	}

	// Publish hypervisor created event
	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventHypervisorCreated,
		tenantID,
		hypervisor.ID.String(),
		map[string]interface{}{
			"name":   hypervisor.Name,
			"mode":   hypervisor.Mode,
			"status": hypervisor.Status,
		},
	))

	return hypervisor, nil
}

// Get retrieves a hypervisor by ID
func (s *Service) Get(ctx context.Context, tenantID, id uuid.UUID) (*Hypervisor, error) {
	return s.repo.GetByID(tenantID, id)
}

// List retrieves all hypervisors for a tenant
func (s *Service) List(ctx context.Context, tenantID uuid.UUID, filter *HypervisorFilter, limit, offset int) ([]Hypervisor, int64, error) {
	p := pagination.Normalize(limit, offset)
	return s.repo.List(tenantID, filter, p.Limit, p.Offset)
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
	if input.AgentID != "" {
		agentID, err := uuid.Parse(input.AgentID)
		if err != nil {
			return nil, fmt.Errorf("invalid agentId: %w", err)
		}
		hypervisor.AgentID = agentID
		hypervisor.Status = HypervisorStatusPending // Reset status when agent changes
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

	// Publish hypervisor updated event
	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventHypervisorUpdated,
		tenantID,
		hypervisor.ID.String(),
		map[string]interface{}{
			"name":   hypervisor.Name,
			"status": hypervisor.Status,
		},
	))

	return hypervisor, nil
}

// Delete deletes a hypervisor
func (s *Service) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	if err := s.repo.Delete(tenantID, id); err != nil {
		return err
	}

	// Publish hypervisor deleted event
	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventHypervisorDeleted,
		tenantID,
		id.String(),
		nil,
	))

	return nil
}

// TestConnection tests the connection to a hypervisor using a playbook
func (s *Service) TestConnection(ctx context.Context, token string, tenantID, hypervisorID uuid.UUID, agentID uuid.UUID) error {
	hypervisor, err := s.repo.GetByID(tenantID, hypervisorID)
	if err != nil {
		return err
	}

	// Execute a libvirt playbook with node_info action to test connection
	_, err = s.client.ExecuteLibvirtTask(ctx, token, hypervisor.AgentID, hypervisor.URI, hypervisor.ArtifactKey, "node-info", nil)
	if err != nil {
		s.repo.UpdateStatus(tenantID, hypervisorID, HypervisorStatusDisconnected, err.Error())
		return err
	}

	s.repo.UpdateStatus(tenantID, hypervisorID, HypervisorStatusConnected, "Connection successful")
	return nil
}

// Deploy deploys Libvirt on an agent
func (s *Service) Deploy(ctx context.Context, tenantID, userID uuid.UUID, input *DeployHypervisorInput) (*Hypervisor, error) {
	token, _ := middleware.GetTokenFromContext(ctx)

	agentID, err := uuid.Parse(input.AgentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agentId: %w", err)
	}

	// Validate agent can deploy this driver
	capability := "libvirt-deploy-" + string(input.Driver)
	if err := s.client.ValidateAgentCapability(ctx, token, agentID, capability); err != nil {
		return nil, fmt.Errorf("agent cannot deploy %s: %w", input.Driver, err)
	}

	// Create the hypervisor record
	hypervisor := &Hypervisor{
		TenantID:    tenantID,
		Name:        input.Name,
		Description: input.Description,
		Mode:        HypervisorModeDeploy,
		Driver:      input.Driver,
		AgentID:     agentID,
		URI:         "qemu:///system", // Local connection after deployment
		Status:      HypervisorStatusDeploying,
		CreatedBy:   userID,
	}

	if err := s.repo.Create(hypervisor); err != nil {
		return nil, fmt.Errorf("failed to create hypervisor: %w", err)
	}

	// Start async deployment (in background)
	go s.runDeployment(hypervisor.ID, tenantID, input)

	return hypervisor, nil
}

// runDeployment executes the libvirt deployment in background
func (s *Service) runDeployment(hypervisorID, tenantID uuid.UUID, input *DeployHypervisorInput) {
	// Use timeout to prevent goroutine leaks
	timeout := 15 * time.Minute
	if cfg := config.GetConfig(); cfg != nil && cfg.Limits.HypervisorDeploymentTimeout > 0 {
		timeout = time.Duration(cfg.Limits.HypervisorDeploymentTimeout) * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	driver := string(input.Driver)

	logger.Info("[Hypervisor %s] Starting deployment: driver=%s", hypervisorID, driver)

	// Background tasks use internal auth
	token := ""

	agentID, err := uuid.Parse(input.AgentID)
	if err != nil {
		logger.Error("[Hypervisor %s] Invalid agent ID: %s", hypervisorID, err.Error())
		s.repo.UpdateStatus(tenantID, hypervisorID, HypervisorStatusError, "Invalid agent ID: "+err.Error())
		return
	}

	// Step 1: Install libvirt packages and configure driver
	logger.Info("[Hypervisor %s] Step 1: Installing libvirt packages for driver %s", hypervisorID, driver)
	params := map[string]interface{}{
		"driver": driver,
	}

	execution, err := s.client.DeployLibvirtTask(ctx, token, agentID, driver, "install", params)
	if err != nil {
		logger.Error("[Hypervisor %s] Failed to deploy libvirt: %s", hypervisorID, err.Error())
		s.repo.UpdateStatus(tenantID, hypervisorID, HypervisorStatusError, "Failed to deploy libvirt: "+err.Error())
		return
	}

	if execution.Status != "SUCCESS" {
		logger.Error("[Hypervisor %s] Libvirt deployment failed: %s", hypervisorID, execution.Error)
		s.repo.UpdateStatus(tenantID, hypervisorID, HypervisorStatusError, "Libvirt deployment failed: "+execution.Error)
		return
	}
	logger.Info("[Hypervisor %s] Libvirt packages installed successfully", hypervisorID)

	// Step 2: Start libvirtd service
	logger.Info("[Hypervisor %s] Step 2: Starting libvirtd service", hypervisorID)
	execution, err = s.client.DeployLibvirtTask(ctx, token, agentID, driver, "start", nil)
	if err != nil {
		logger.Error("[Hypervisor %s] Failed to start libvirtd: %s", hypervisorID, err.Error())
		s.repo.UpdateStatus(tenantID, hypervisorID, HypervisorStatusError, "Failed to start libvirtd: "+err.Error())
		return
	}

	if execution.Status != "SUCCESS" {
		logger.Error("[Hypervisor %s] Failed to start libvirtd: %s", hypervisorID, execution.Error)
		s.repo.UpdateStatus(tenantID, hypervisorID, HypervisorStatusError, "Failed to start libvirtd: "+execution.Error)
		return
	}
	logger.Info("[Hypervisor %s] Libvirtd service started successfully", hypervisorID)

	// Step 3: Verify connection works
	logger.Info("[Hypervisor %s] Step 3: Verifying libvirt connection", hypervisorID)
	execution, err = s.client.DeployLibvirtTask(ctx, token, agentID, driver, "verify", nil)
	if err != nil {
		logger.Error("[Hypervisor %s] Failed to verify libvirt: %s", hypervisorID, err.Error())
		s.repo.UpdateStatus(tenantID, hypervisorID, HypervisorStatusError, "Failed to verify libvirt: "+err.Error())
		return
	}

	if execution.Status != "SUCCESS" {
		logger.Error("[Hypervisor %s] Libvirt verification failed: %s", hypervisorID, execution.Error)
		s.repo.UpdateStatus(tenantID, hypervisorID, HypervisorStatusError, "Libvirt verification failed: "+execution.Error)
		return
	}

	// Step 4: Update hypervisor status to connected
	logger.Info("[Hypervisor %s] Deployment completed successfully", hypervisorID)
	s.repo.UpdateStatus(tenantID, hypervisorID, HypervisorStatusConnected, "Libvirt deployed successfully")
}

// BulkDelete deletes multiple hypervisors by IDs
func (s *Service) BulkDelete(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) (int64, error) {
	return s.repo.BulkDelete(tenantID, ids)
}
