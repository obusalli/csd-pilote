package services

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

// Service handles Kubernetes service operations via csd-core playbooks
type Service struct {
	clusterSvc *clusters.Service
	coreClient *csdcore.Client
}

// NewService creates a new K8s service service
func NewService() *Service {
	return &Service{
		clusterSvc: clusters.NewService(),
		coreClient: csdcore.GetClient(),
	}
}

// List returns all services for a cluster/namespace
func (s *Service) List(ctx context.Context, token string, tenantID, clusterID uuid.UUID, namespace string, filter *ServiceFilter) ([]K8sService, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	params := map[string]interface{}{}
	if namespace != "" {
		params["namespace"] = namespace
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "list-services", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawServices []rawK8sService
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawServices); err != nil {
		return nil, fmt.Errorf("failed to parse services: %w", err)
	}

	services := make([]K8sService, 0, len(rawServices))
	for _, svc := range rawServices {
		// Apply filters
		if filter != nil {
			if filter.Search != nil && *filter.Search != "" {
				if !strings.Contains(strings.ToLower(svc.Name), strings.ToLower(*filter.Search)) {
					continue
				}
			}
			if filter.Type != nil && *filter.Type != "" {
				if svc.Type != *filter.Type {
					continue
				}
			}
		}

		services = append(services, s.toK8sService(clusterID, &svc))
	}

	return services, nil
}

// Get returns a specific service
func (s *Service) Get(ctx context.Context, token string, tenantID, clusterID uuid.UUID, namespace, name string) (*K8sService, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "get-service", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawSvc rawK8sService
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawSvc); err != nil {
		return nil, fmt.Errorf("failed to parse service: %w", err)
	}

	result := s.toK8sService(clusterID, &rawSvc)
	return &result, nil
}

// Create creates a new service
func (s *Service) Create(ctx context.Context, token string, tenantID, clusterID uuid.UUID, input *CreateServiceInput) (*K8sService, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "create-service", map[string]interface{}{
		"namespace": input.Namespace,
		"name":      input.Name,
		"type":      input.Type,
		"selector":  input.Selector,
		"ports":     input.Ports,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	return s.Get(ctx, token, tenantID, clusterID, input.Namespace, input.Name)
}

// Delete deletes a service
func (s *Service) Delete(ctx context.Context, token string, tenantID, clusterID uuid.UUID, namespace, name string) error {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "delete-service", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	})
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return fmt.Errorf("task failed: %s", execution.Error)
	}

	return nil
}

type rawServicePort struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort string `json:"targetPort"`
	NodePort   int32  `json:"nodePort,omitempty"`
}

type rawK8sService struct {
	Namespace   string            `json:"namespace"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	ClusterIP   string            `json:"clusterIP"`
	ExternalIP  string            `json:"externalIP"`
	Ports       []rawServicePort  `json:"ports"`
	Selector    map[string]string `json:"selector"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   string            `json:"createdAt"`
}

func (s *Service) toK8sService(clusterID uuid.UUID, svc *rawK8sService) K8sService {
	ports := make([]ServicePort, len(svc.Ports))
	for i, p := range svc.Ports {
		ports[i] = ServicePort{
			Name:       p.Name,
			Protocol:   p.Protocol,
			Port:       p.Port,
			TargetPort: p.TargetPort,
			NodePort:   p.NodePort,
		}
	}

	createdAt, _ := time.Parse(time.RFC3339, svc.CreatedAt)

	return K8sService{
		ClusterID:   clusterID,
		Namespace:   svc.Namespace,
		Name:        svc.Name,
		Type:        svc.Type,
		ClusterIP:   svc.ClusterIP,
		ExternalIP:  svc.ExternalIP,
		Ports:       ports,
		Selector:    svc.Selector,
		Labels:      svc.Labels,
		Annotations: svc.Annotations,
		CreatedAt:   createdAt,
	}
}
