package pods

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

// Service handles pod operations via csd-core playbooks
type Service struct {
	clusterSvc *clusters.Service
	coreClient *csdcore.Client
}

// NewService creates a new pod service
func NewService() *Service {
	return &Service{
		clusterSvc: clusters.NewService(),
		coreClient: csdcore.GetClient(),
	}
}

// List returns all pods for a cluster/namespace
func (s *Service) List(ctx context.Context, token string, tenantID, clusterID uuid.UUID, namespace string, filter *PodFilter) ([]Pod, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	params := map[string]interface{}{}
	if namespace != "" {
		params["namespace"] = namespace
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "list-pods", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawPods []rawPod
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawPods); err != nil {
		return nil, fmt.Errorf("failed to parse pods: %w", err)
	}

	pods := make([]Pod, 0, len(rawPods))
	for _, pod := range rawPods {
		// Apply filters
		if filter != nil {
			if filter.Search != nil && *filter.Search != "" {
				if !strings.Contains(strings.ToLower(pod.Name), strings.ToLower(*filter.Search)) {
					continue
				}
			}
			if filter.Phase != nil && *filter.Phase != "" {
				if pod.Phase != *filter.Phase {
					continue
				}
			}
		}

		pods = append(pods, s.toPod(clusterID, &pod))
	}

	return pods, nil
}

// Get returns a specific pod
func (s *Service) Get(ctx context.Context, token string, tenantID, clusterID uuid.UUID, namespace, name string) (*Pod, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "get-pod", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var rawP rawPod
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &rawP); err != nil {
		return nil, fmt.Errorf("failed to parse pod: %w", err)
	}

	result := s.toPod(clusterID, &rawP)
	return &result, nil
}

// GetLogs returns logs for a pod/container
func (s *Service) GetLogs(ctx context.Context, token string, tenantID, clusterID uuid.UUID, namespace, name, container string, tailLines int64, previous bool) (*PodLogs, error) {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	params := map[string]interface{}{
		"namespace": namespace,
		"name":      name,
		"previous":  previous,
	}
	if container != "" {
		params["container"] = container
	}
	if tailLines > 0 {
		params["tailLines"] = tailLines
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "get-pod-logs", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return nil, fmt.Errorf("task failed: %s", execution.Error)
	}

	var logsOutput struct {
		Logs string `json:"logs"`
	}
	outputBytes, err := json.Marshal(execution.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}
	if err := json.Unmarshal(outputBytes, &logsOutput); err != nil {
		return nil, fmt.Errorf("failed to parse logs: %w", err)
	}

	return &PodLogs{
		PodName:   name,
		Container: container,
		Logs:      logsOutput.Logs,
	}, nil
}

// Delete deletes a pod
func (s *Service) Delete(ctx context.Context, token string, tenantID, clusterID uuid.UUID, namespace, name string, gracePeriod int64) error {
	cluster, err := s.clusterSvc.Get(ctx, tenantID, clusterID)
	if err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	params := map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	}
	if gracePeriod >= 0 {
		params["gracePeriodSeconds"] = gracePeriod
	}

	execution, err := s.coreClient.ExecuteKubernetesTask(ctx, token, cluster.AgentID, cluster.ArtifactKey, "delete-pod", params)
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}

	if execution.Status != "SUCCESS" {
		return fmt.Errorf("task failed: %s", execution.Error)
	}

	return nil
}

type rawContainer struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restartCount"`
	State        string `json:"state"`
}

type rawPod struct {
	Namespace   string            `json:"namespace"`
	Name        string            `json:"name"`
	Phase       string            `json:"phase"`
	Status      string            `json:"status"`
	Ready       string            `json:"ready"`
	Restarts    int32             `json:"restarts"`
	Age         string            `json:"age"`
	IP          string            `json:"ip"`
	Node        string            `json:"node"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Containers  []rawContainer    `json:"containers"`
	CreatedAt   string            `json:"createdAt"`
}

func (s *Service) toPod(clusterID uuid.UUID, pod *rawPod) Pod {
	containers := make([]Container, len(pod.Containers))
	for i, c := range pod.Containers {
		containers[i] = Container{
			Name:         c.Name,
			Image:        c.Image,
			Ready:        c.Ready,
			RestartCount: c.RestartCount,
			State:        c.State,
		}
	}

	createdAt, _ := time.Parse(time.RFC3339, pod.CreatedAt)

	return Pod{
		ClusterID:   clusterID,
		Namespace:   pod.Namespace,
		Name:        pod.Name,
		Phase:       pod.Phase,
		Status:      pod.Status,
		Ready:       pod.Ready,
		Restarts:    pod.Restarts,
		Age:         pod.Age,
		IP:          pod.IP,
		Node:        pod.Node,
		Labels:      pod.Labels,
		Annotations: pod.Annotations,
		Containers:  containers,
		CreatedAt:   createdAt,
	}
}
