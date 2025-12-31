package clusters

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/events"
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

// Create creates a new cluster (CONNECT mode - connect to existing cluster)
func (s *Service) Create(ctx context.Context, tenantID, userID uuid.UUID, input *ClusterInput) (*Cluster, error) {
	agentID, err := uuid.Parse(input.AgentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agentId: %w", err)
	}

	cluster := &Cluster{
		TenantID:     tenantID,
		Name:         input.Name,
		Description:  input.Description,
		Mode:         ClusterModeConnect,
		Distribution: input.Distribution, // Optional: can be empty or set to known distribution
		AgentID:      agentID,
		ArtifactKey:  input.ArtifactKey,
		Status:       ClusterStatusPending,
		CreatedBy:    userID,
	}

	if err := s.repo.Create(cluster); err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	// Publish cluster created event
	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventClusterCreated,
		tenantID,
		cluster.ID.String(),
		map[string]interface{}{
			"name":   cluster.Name,
			"mode":   cluster.Mode,
			"status": cluster.Status,
		},
	))

	return cluster, nil
}

// Deploy deploys a new Kubernetes cluster on selected agents
func (s *Service) Deploy(ctx context.Context, tenantID, userID uuid.UUID, input *DeployClusterInput) (*Cluster, error) {
	token, _ := ctx.Value("token").(string)

	// Validate all agent IDs can deploy this distribution
	capability := "kubernetes-deploy-" + string(input.Distribution)
	allAgentIDs := append(input.MasterNodes, input.WorkerNodes...)

	for _, agentIDStr := range allAgentIDs {
		agentID, err := uuid.Parse(agentIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid agent ID %s: %w", agentIDStr, err)
		}
		if err := s.client.ValidateAgentCapability(ctx, token, agentID, capability); err != nil {
			return nil, fmt.Errorf("agent %s cannot deploy %s: %w", agentIDStr, input.Distribution, err)
		}
	}

	// Create the cluster record
	cluster := &Cluster{
		TenantID:     tenantID,
		Name:         input.Name,
		Description:  input.Description,
		Mode:         ClusterModeDeploy,
		Distribution: input.Distribution,
		Version:      input.Version,
		Status:       ClusterStatusDeploying,
		CreatedBy:    userID,
	}

	if err := s.repo.Create(cluster); err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	// Create node records
	nodes := make([]ClusterNode, 0, len(allAgentIDs))

	for _, agentIDStr := range input.MasterNodes {
		agentID, _ := uuid.Parse(agentIDStr)
		nodes = append(nodes, ClusterNode{
			ClusterID: cluster.ID,
			AgentID:   agentID,
			Role:      NodeRoleMaster,
			Status:    "PENDING",
		})
	}

	for _, agentIDStr := range input.WorkerNodes {
		agentID, _ := uuid.Parse(agentIDStr)
		nodes = append(nodes, ClusterNode{
			ClusterID: cluster.ID,
			AgentID:   agentID,
			Role:      NodeRoleWorker,
			Status:    "PENDING",
		})
	}

	if err := s.repo.CreateNodes(nodes); err != nil {
		return nil, fmt.Errorf("failed to create cluster nodes: %w", err)
	}

	// Start async deployment (in background)
	go s.runDeployment(cluster.ID, tenantID, input, nodes)

	// Return cluster with nodes
	cluster.Nodes = nodes
	return cluster, nil
}

// runDeployment executes the cluster deployment in background
func (s *Service) runDeployment(clusterID, tenantID uuid.UUID, input *DeployClusterInput, nodes []ClusterNode) {
	ctx := context.Background()
	distribution := string(input.Distribution)

	// Get a system token for background operations
	// In production, this would use a service account token
	token := "" // Background tasks use internal auth

	var joinToken, joinURL, kubeconfig string
	var masterNodes, workerNodes []ClusterNode

	// Separate master and worker nodes
	for _, node := range nodes {
		if node.Role == NodeRoleMaster {
			masterNodes = append(masterNodes, node)
		} else {
			workerNodes = append(workerNodes, node)
		}
	}

	// Step 1: Deploy on first master node (init cluster)
	if len(masterNodes) > 0 {
		firstMaster := masterNodes[0]
		s.repo.UpdateNodeStatus(firstMaster.ID, "DEPLOYING", "Initializing cluster...")

		params := map[string]interface{}{
			"role":    "init-master",
			"version": input.Version,
		}

		execution, err := s.client.DeployKubernetesTask(ctx, token, firstMaster.AgentID, distribution, "install", params)
		if err != nil {
			s.handleDeploymentError(clusterID, tenantID, firstMaster.ID, "Failed to initialize cluster", err)
			return
		}

		if execution.Status != "SUCCESS" {
			s.handleDeploymentError(clusterID, tenantID, firstMaster.ID, "Cluster initialization failed", fmt.Errorf(execution.Error))
			return
		}

		// Extract join token and URL from output
		if output, ok := execution.Output.(map[string]interface{}); ok {
			joinToken, _ = output["joinToken"].(string)
			joinURL, _ = output["joinUrl"].(string)
			kubeconfig, _ = output["kubeconfig"].(string)
		}

		s.repo.UpdateNodeStatus(firstMaster.ID, "READY", "Master node initialized")
	}

	// Step 2: Deploy on additional master nodes
	for i := 1; i < len(masterNodes); i++ {
		node := masterNodes[i]
		s.repo.UpdateNodeStatus(node.ID, "DEPLOYING", "Joining cluster as master...")

		params := map[string]interface{}{
			"role":      "join-master",
			"joinToken": joinToken,
			"joinUrl":   joinURL,
			"version":   input.Version,
		}

		execution, err := s.client.DeployKubernetesTask(ctx, token, node.AgentID, distribution, "install", params)
		if err != nil {
			s.repo.UpdateNodeStatus(node.ID, "ERROR", err.Error())
			continue // Continue with other nodes
		}

		if execution.Status != "SUCCESS" {
			s.repo.UpdateNodeStatus(node.ID, "ERROR", execution.Error)
			continue
		}

		s.repo.UpdateNodeStatus(node.ID, "READY", "Master node joined")
	}

	// Step 3: Deploy on worker nodes
	for _, node := range workerNodes {
		s.repo.UpdateNodeStatus(node.ID, "DEPLOYING", "Joining cluster as worker...")

		params := map[string]interface{}{
			"role":      "join-worker",
			"joinToken": joinToken,
			"joinUrl":   joinURL,
			"version":   input.Version,
		}

		execution, err := s.client.DeployKubernetesTask(ctx, token, node.AgentID, distribution, "install", params)
		if err != nil {
			s.repo.UpdateNodeStatus(node.ID, "ERROR", err.Error())
			continue
		}

		if execution.Status != "SUCCESS" {
			s.repo.UpdateNodeStatus(node.ID, "ERROR", execution.Error)
			continue
		}

		s.repo.UpdateNodeStatus(node.ID, "READY", "Worker node joined")
	}

	// Step 4: Store kubeconfig as artifact
	if kubeconfig != "" {
		artifactKey := fmt.Sprintf("cluster-%s-kubeconfig", clusterID.String())
		err := s.client.CreateArtifact(ctx, token, tenantID, artifactKey, "kubeconfig", kubeconfig)
		if err != nil {
			s.repo.UpdateStatus(tenantID, clusterID, ClusterStatusError, "Failed to store kubeconfig: "+err.Error())
			return
		}

		// Update cluster with artifact key
		s.repo.UpdateClusterArtifact(clusterID, artifactKey)
	}

	// Step 5: Update cluster status to connected
	s.repo.UpdateStatus(tenantID, clusterID, ClusterStatusConnected, "Cluster deployed successfully")
}

// handleDeploymentError handles deployment errors
func (s *Service) handleDeploymentError(clusterID, tenantID, nodeID uuid.UUID, message string, err error) {
	fullMessage := fmt.Sprintf("%s: %v", message, err)
	s.repo.UpdateNodeStatus(nodeID, "ERROR", fullMessage)
	s.repo.UpdateStatus(tenantID, clusterID, ClusterStatusError, fullMessage)
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
	if input.AgentID != "" {
		agentID, err := uuid.Parse(input.AgentID)
		if err != nil {
			return nil, fmt.Errorf("invalid agentId: %w", err)
		}
		cluster.AgentID = agentID
		cluster.Status = ClusterStatusPending // Reset status when agent changes
	}
	if input.ArtifactKey != "" {
		cluster.ArtifactKey = input.ArtifactKey
		cluster.Status = ClusterStatusPending // Reset status when config changes
	}
	if input.Distribution != "" {
		cluster.Distribution = input.Distribution
	}

	if err := s.repo.Update(cluster); err != nil {
		return nil, fmt.Errorf("failed to update cluster: %w", err)
	}

	// Publish cluster updated event
	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventClusterUpdated,
		tenantID,
		cluster.ID.String(),
		map[string]interface{}{
			"name":   cluster.Name,
			"status": cluster.Status,
		},
	))

	return cluster, nil
}

// Delete deletes a cluster
func (s *Service) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	if err := s.repo.Delete(tenantID, id); err != nil {
		return err
	}

	// Publish cluster deleted event
	events.GetEventBus().PublishAsync(events.NewEvent(
		events.EventClusterDeleted,
		tenantID,
		id.String(),
		nil,
	))

	return nil
}

// TestConnection tests the connection to a cluster using a playbook
func (s *Service) TestConnection(ctx context.Context, token string, tenantID, clusterID uuid.UUID, agentID uuid.UUID) error {
	cluster, err := s.repo.GetByID(tenantID, clusterID)
	if err != nil {
		return err
	}

	// Execute a kubernetes playbook to test connection
	_, err = s.client.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "get-server-version", nil)
	if err != nil {
		s.repo.UpdateStatus(tenantID, clusterID, ClusterStatusDisconnected, err.Error())
		return err
	}

	s.repo.UpdateStatus(tenantID, clusterID, ClusterStatusConnected, "Connection successful")
	return nil
}

// BulkDelete deletes multiple clusters by IDs
func (s *Service) BulkDelete(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) (int64, error) {
	return s.repo.BulkDelete(tenantID, ids)
}
