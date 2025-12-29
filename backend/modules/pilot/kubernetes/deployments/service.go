package deployments

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"csd-pilote/backend/modules/pilot/clusters"
	csdcore "csd-pilote/backend/modules/platform/csd-core"
)

// Service handles deployment operations via csd-core playbooks
type Service struct {
	clusterSvc *clusters.Service
	coreClient *csdcore.Client
}

// NewService creates a new deployment service
func NewService() *Service {
	return &Service{
		clusterSvc: clusters.NewService(),
		coreClient: csdcore.GetClient(),
	}
}

// List returns all deployments for a cluster/namespace
func (s *Service) List(ctx context.Context, token string, tenantID, clusterID uuid.UUID, namespace string, filter *DeploymentFilter) ([]Deployment, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	params := map[string]interface{}{}
	if namespace != "" {
		params["namespace"] = namespace
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "list-deployments", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawDeployments []rawDeployment
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawDeployments); err != nil {
		return nil, fmt.Errorf("failed to parse deployments: %w", err)
	}

	deployments := make([]Deployment, 0, len(rawDeployments))
	for _, dep := range rawDeployments {
		// Apply filter
		if filter != nil && filter.Search != nil && *filter.Search != "" {
			if !strings.Contains(strings.ToLower(dep.Name), strings.ToLower(*filter.Search)) {
				continue
			}
		}

		deployments = append(deployments, s.toDeployment(clusterID, &dep))
	}

	return deployments, nil
}

// Get returns a specific deployment
func (s *Service) Get(ctx context.Context, token string, tenantID, clusterID uuid.UUID, namespace, name string) (*Deployment, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "get-deployment", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawDep rawDeployment
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawDep); err != nil {
		return nil, fmt.Errorf("failed to parse deployment: %w", err)
	}

	result := s.toDeployment(clusterID, &rawDep)
	return &result, nil
}

// Create creates a new deployment
func (s *Service) Create(ctx context.Context, token string, tenantID, clusterID uuid.UUID, input *CreateDeploymentInput) (*Deployment, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	replicas := input.Replicas
	if replicas <= 0 {
		replicas = 1
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "create-deployment", map[string]interface{}{
		"namespace": input.Namespace,
		"name":      input.Name,
		"image":     input.Image,
		"replicas":  replicas,
		"labels":    input.Labels,
		"ports":     input.Ports,
		"env":       input.Env,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, clusterID, input.Namespace, input.Name)
}

// Scale scales a deployment
func (s *Service) Scale(ctx context.Context, token string, tenantID, clusterID uuid.UUID, namespace, name string, replicas int32) (*Deployment, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "scale-deployment", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
		"replicas":  replicas,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scale deployment: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, clusterID, namespace, name)
}

// Restart restarts a deployment
func (s *Service) Restart(ctx context.Context, token string, tenantID, clusterID uuid.UUID, namespace, name string) (*Deployment, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "restart-deployment", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to restart deployment: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, clusterID, namespace, name)
}

// Delete deletes a deployment
func (s *Service) Delete(ctx context.Context, token string, tenantID, clusterID uuid.UUID, namespace, name string) error {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "delete-deployment", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	})
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return fmt.Errorf("task failed: %s", execution.Error)
	}

	return nil
}

type rawDeployment struct {
	Namespace         string            `json:"namespace"`
	Name              string            `json:"name"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"readyReplicas"`
	AvailableReplicas int32             `json:"availableReplicas"`
	UpdatedReplicas   int32             `json:"updatedReplicas"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	Images            []string          `json:"images"`
	CreatedAt         string            `json:"createdAt"`
}

func (s *Service) toDeployment(clusterID uuid.UUID, dep *rawDeployment) Deployment {
	createdAt, _ := time.Parse(time.RFC3339, dep.CreatedAt)
	return Deployment{
		ClusterID:         clusterID,
		Namespace:         dep.Namespace,
		Name:              dep.Name,
		Replicas:          dep.Replicas,
		ReadyReplicas:     dep.ReadyReplicas,
		AvailableReplicas: dep.AvailableReplicas,
		UpdatedReplicas:   dep.UpdatedReplicas,
		Labels:            dep.Labels,
		Annotations:       dep.Annotations,
		Images:            dep.Images,
		CreatedAt:         createdAt,
	}
}
