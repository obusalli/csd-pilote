package namespaces

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

// Service handles namespace operations via csd-core playbooks
type Service struct {
	clusterSvc *clusters.Service
	coreClient *csdcore.Client
}

// NewService creates a new namespace service
func NewService() *Service {
	return &Service{
		clusterSvc: clusters.NewService(),
		coreClient: csdcore.GetClient(),
	}
}

// List returns all namespaces for a cluster
func (s *Service) List(ctx context.Context, token string, tenantID, clusterID uuid.UUID, filter *NamespaceFilter) ([]Namespace, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	// Execute kubernetes task via csd-core
	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "list-namespaces", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	// Parse output
	var rawNamespaces []struct {
		Name        string            `json:"name"`
		Status      string            `json:"status"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		CreatedAt   string            `json:"createdAt"`
	}

	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawNamespaces); err != nil {
		return nil, fmt.Errorf("failed to parse namespaces: %w", err)
	}

	namespaces := make([]Namespace, 0, len(rawNamespaces))
	for _, ns := range rawNamespaces {
		// Apply filter
		if filter != nil && filter.Search != nil && *filter.Search != "" {
			if !strings.Contains(strings.ToLower(ns.Name), strings.ToLower(*filter.Search)) {
				continue
			}
		}

		createdAt, _ := time.Parse(time.RFC3339, ns.CreatedAt)
		namespaces = append(namespaces, Namespace{
			ClusterID:   clusterID,
			Name:        ns.Name,
			Status:      ns.Status,
			Labels:      ns.Labels,
			Annotations: ns.Annotations,
			CreatedAt:   createdAt,
		})
	}

	return namespaces, nil
}

// Get returns a specific namespace
func (s *Service) Get(ctx context.Context, token string, tenantID, clusterID uuid.UUID, name string) (*Namespace, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "get-namespace", map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawNs struct {
		Name        string            `json:"name"`
		Status      string            `json:"status"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		CreatedAt   string            `json:"createdAt"`
	}

	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawNs); err != nil {
		return nil, fmt.Errorf("failed to parse namespace: %w", err)
	}

	createdAt, _ := time.Parse(time.RFC3339, rawNs.CreatedAt)
	return &Namespace{
		ClusterID:   clusterID,
		Name:        rawNs.Name,
		Status:      rawNs.Status,
		Labels:      rawNs.Labels,
		Annotations: rawNs.Annotations,
		CreatedAt:   createdAt,
	}, nil
}

// Create creates a new namespace
func (s *Service) Create(ctx context.Context, token string, tenantID, clusterID uuid.UUID, name string, labels map[string]string) (*Namespace, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "create-namespace", map[string]interface{}{
		"name":   name,
		"labels": labels,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, clusterID, name)
}

// Delete deletes a namespace
func (s *Service) Delete(ctx context.Context, token string, tenantID, clusterID uuid.UUID, name string) error {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "delete-namespace", map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return fmt.Errorf("task failed: %s", execution.Error)
	}

	return nil
}
