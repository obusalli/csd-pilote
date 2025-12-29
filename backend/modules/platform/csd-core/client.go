package csdcore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"csd-pilote/backend/modules/platform/config"
	"github.com/google/uuid"
)

// Client is a GraphQL client for csd-core
type Client struct {
	httpClient *http.Client
	baseURL    string
	endpoint   string
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
}

// UserInfo represents user information from csd-core
type UserInfo struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	TenantID      uuid.UUID `json:"tenantId"`
	Roles         []string  `json:"roles"`
	Permissions   []string  `json:"permissions"`
	IsActive      bool      `json:"isActive"`
	EmailVerified bool      `json:"emailVerified"`
}

var globalClient *Client

// NewClient creates a new csd-core client
func NewClient(cfg *config.CSDCoreConfig) *Client {
	client := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  cfg.URL,
		endpoint: cfg.GraphQLEndpoint,
	}
	globalClient = client
	return client
}

// GetClient returns the global client
func GetClient() *Client {
	return globalClient
}

// Execute executes a GraphQL query/mutation
func (c *Client) Execute(ctx context.Context, token string, query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	return c.ExecuteWithName(ctx, token, "", query, variables)
}

// ExecuteWithName executes a GraphQL query/mutation with an explicit operation name
func (c *Client) ExecuteWithName(ctx context.Context, token string, operationName string, query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	reqBody := GraphQLRequest{
		Query:         query,
		OperationName: operationName,
		Variables:     variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+c.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("empty response from server (status: %d)", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var graphqlResp GraphQLResponse
	if err := json.Unmarshal(body, &graphqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(graphqlResp.Errors) > 0 {
		return &graphqlResp, fmt.Errorf("graphql error: %s", graphqlResp.Errors[0].Message)
	}

	return &graphqlResp, nil
}

// ValidateToken validates a JWT token with csd-core and returns user info
func (c *Client) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	query := `
		query Me {
			me {
				id
				email
				firstName
				lastName
				isActive
				emailVerified
			}
		}
	`

	resp, err := c.Execute(ctx, token, query, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Me *UserInfo `json:"me"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	if result.Me == nil {
		return nil, fmt.Errorf("user not found")
	}

	return result.Me, nil
}

// CheckPermission checks if user has a specific permission
func (c *Client) CheckPermission(ctx context.Context, token string, permission string) (bool, error) {
	query := `
		query HasPermission($permission: String!) {
			hasPermission(permission: $permission)
		}
	`

	resp, err := c.ExecuteWithName(ctx, token, "HasPermission", query, map[string]interface{}{
		"permission": permission,
	})
	if err != nil {
		return false, err
	}

	var result struct {
		HasPermission bool `json:"hasPermission"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return false, fmt.Errorf("failed to parse permission response: %w", err)
	}

	return result.HasPermission, nil
}

// GetArtifactContent gets artifact content from csd-core by key
func (c *Client) GetArtifactContent(ctx context.Context, token string, key string) ([]byte, error) {
	query := `
		query GetArtifactByKey($key: String!) {
			artifactByKey(key: $key) {
				id
				content
			}
		}
	`

	resp, err := c.ExecuteWithName(ctx, token, "GetArtifactByKey", query, map[string]interface{}{
		"key": key,
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		ArtifactByKey *struct {
			ID      string `json:"id"`
			Content string `json:"content"`
		} `json:"artifactByKey"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse artifact: %w", err)
	}

	if result.ArtifactByKey == nil {
		return nil, fmt.Errorf("artifact not found: %s", key)
	}

	// Content is base64 encoded
	return []byte(result.ArtifactByKey.Content), nil
}

// ExecutePlaybook executes a playbook via csd-core
func (c *Client) ExecutePlaybook(ctx context.Context, token string, playbookID uuid.UUID, nodeIDs []uuid.UUID, vars map[string]interface{}) (*PlaybookExecution, error) {
	mutation := `
		mutation ExecutePlaybook($input: ExecutePlaybookInput!) {
			executePlaybook(input: $input) {
				id
				status
				startedAt
			}
		}
	`

	nodeIDStrings := make([]string, len(nodeIDs))
	for i, id := range nodeIDs {
		nodeIDStrings[i] = id.String()
	}

	resp, err := c.ExecuteWithName(ctx, token, "ExecutePlaybook", mutation, map[string]interface{}{
		"input": map[string]interface{}{
			"playbookId": playbookID.String(),
			"nodeIds":    nodeIDStrings,
			"vars":       vars,
		},
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		ExecutePlaybook *PlaybookExecution `json:"executePlaybook"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse execution: %w", err)
	}

	return result.ExecutePlaybook, nil
}

// PlaybookExecution represents a playbook execution result
type PlaybookExecution struct {
	ID        uuid.UUID `json:"id"`
	Status    string    `json:"status"`
	StartedAt string    `json:"startedAt"`
}

// TaskExecution represents a task execution
type TaskExecution struct {
	ID          uuid.UUID   `json:"id"`
	PlaybookID  uuid.UUID   `json:"playbookId"`
	AgentID     uuid.UUID   `json:"agentId"`
	TaskIndex   int         `json:"taskIndex"`
	Status      string      `json:"status"` // PENDING, RUNNING, SUCCESS, FAILED
	Output      interface{} `json:"output"`
	Error       string      `json:"error"`
	StartedAt   *string     `json:"startedAt"`
	CompletedAt *string     `json:"completedAt"`
}

// TaskInput defines a task to execute
type TaskInput struct {
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

// ExecuteTaskInput represents input for executing a single task
type ExecuteTaskInput struct {
	AgentID     uuid.UUID              `json:"agentId"`
	Task        TaskInput              `json:"task"`
	ArtifactKey string                 `json:"artifactKey,omitempty"` // For kubeconfig, SSH keys, etc.
	Vars        map[string]interface{} `json:"vars,omitempty"`
	Wait        bool                   `json:"wait"` // Wait for completion
	Timeout     int                    `json:"timeout,omitempty"` // Timeout in seconds
}

// ExecuteTask executes a single task on an agent via csd-core
func (c *Client) ExecuteTask(ctx context.Context, token string, input *ExecuteTaskInput) (*TaskExecution, error) {
	// Create a one-task playbook and execute it
	mutation := `
		mutation ExecuteTask($input: ExecuteTaskInput!) {
			executeTask(input: $input) {
				id
				playbookId
				agentId
				taskIndex
				status
				output
				error
				startedAt
				completedAt
			}
		}
	`

	vars := map[string]interface{}{
		"input": map[string]interface{}{
			"agentId": input.AgentID.String(),
			"task": map[string]interface{}{
				"type":   input.Task.Type,
				"name":   input.Task.Name,
				"config": input.Task.Config,
			},
			"artifactKey": input.ArtifactKey,
			"vars":        input.Vars,
			"wait":        input.Wait,
			"timeout":     input.Timeout,
		},
	}

	resp, err := c.ExecuteWithName(ctx, token, "ExecuteTask", mutation, vars)
	if err != nil {
		return nil, err
	}

	var result struct {
		ExecuteTask *TaskExecution `json:"executeTask"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse task execution: %w", err)
	}

	return result.ExecuteTask, nil
}

// GetTaskExecution gets the status of a task execution
func (c *Client) GetTaskExecution(ctx context.Context, token string, executionID uuid.UUID) (*TaskExecution, error) {
	query := `
		query GetTaskExecution($id: ID!) {
			taskExecution(id: $id) {
				id
				playbookId
				agentId
				taskIndex
				status
				output
				error
				startedAt
				completedAt
			}
		}
	`

	resp, err := c.ExecuteWithName(ctx, token, "GetTaskExecution", query, map[string]interface{}{
		"id": executionID.String(),
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		TaskExecution *TaskExecution `json:"taskExecution"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse task execution: %w", err)
	}

	return result.TaskExecution, nil
}

// ValidateAgentCapability checks if an agent supports a specific capability
func (c *Client) ValidateAgentCapability(ctx context.Context, token string, agentID uuid.UUID, capability string) error {
	agent, err := c.GetAgent(ctx, token, agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	if agent == nil {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	if agent.Status != "ONLINE" {
		return fmt.Errorf("agent is not online (status: %s)", agent.Status)
	}

	if !agent.HasCapability(capability) {
		return fmt.Errorf("agent %s does not support %s capability (available: %v)", agent.Name, capability, agent.Capabilities)
	}

	return nil
}

// ExecuteKubernetesTask executes a Kubernetes-specific task
func (c *Client) ExecuteKubernetesTask(ctx context.Context, token string, agentID uuid.UUID, kubeconfigKey string, action string, params map[string]interface{}) (*TaskExecution, error) {
	// Validate agent supports Kubernetes
	if err := c.ValidateAgentCapability(ctx, token, agentID, "kubernetes"); err != nil {
		return nil, err
	}

	config := map[string]interface{}{
		"action":        action,
		"kubeconfigKey": kubeconfigKey,
	}
	// Merge params into config
	for k, v := range params {
		config[k] = v
	}

	return c.ExecuteTask(ctx, token, &ExecuteTaskInput{
		AgentID: agentID,
		Task: TaskInput{
			Type:   "kubernetes",
			Name:   fmt.Sprintf("k8s-%s", action),
			Config: config,
		},
		ArtifactKey: kubeconfigKey,
		Wait:        true,
		Timeout:     30,
	})
}

// ExecuteLibvirtTask executes a Libvirt-specific task
func (c *Client) ExecuteLibvirtTask(ctx context.Context, token string, agentID uuid.UUID, uri string, sshKeyArtifact string, action string, params map[string]interface{}) (*TaskExecution, error) {
	// Validate agent supports Libvirt
	if err := c.ValidateAgentCapability(ctx, token, agentID, "libvirt"); err != nil {
		return nil, err
	}

	config := map[string]interface{}{
		"action": action,
		"uri":    uri,
	}
	// Merge params into config
	for k, v := range params {
		config[k] = v
	}

	return c.ExecuteTask(ctx, token, &ExecuteTaskInput{
		AgentID: agentID,
		Task: TaskInput{
			Type:   "libvirt",
			Name:   fmt.Sprintf("libvirt-%s", action),
			Config: config,
		},
		ArtifactKey: sshKeyArtifact,
		Wait:        true,
		Timeout:     30,
	})
}

// DeployKubernetesResult contains the result of a Kubernetes deployment
type DeployKubernetesResult struct {
	Success    bool   `json:"success"`
	JoinToken  string `json:"joinToken,omitempty"`
	JoinURL    string `json:"joinUrl,omitempty"`
	Kubeconfig string `json:"kubeconfig,omitempty"`
	Error      string `json:"error,omitempty"`
}

// DeployKubernetesTask deploys Kubernetes on an agent
func (c *Client) DeployKubernetesTask(ctx context.Context, token string, agentID uuid.UUID, distribution string, action string, params map[string]interface{}) (*TaskExecution, error) {
	// Validate agent supports this distribution deployment
	capability := "kubernetes-deploy-" + distribution
	if err := c.ValidateAgentCapability(ctx, token, agentID, capability); err != nil {
		return nil, err
	}

	config := map[string]interface{}{
		"distribution": distribution,
		"action":       action,
	}
	for k, v := range params {
		config[k] = v
	}

	return c.ExecuteTask(ctx, token, &ExecuteTaskInput{
		AgentID: agentID,
		Task: TaskInput{
			Type:   "kubernetes-deploy",
			Name:   fmt.Sprintf("k8s-deploy-%s-%s", distribution, action),
			Config: config,
		},
		Wait:    true,
		Timeout: 300, // 5 minutes for deployment tasks
	})
}

// DeployLibvirtTask deploys Libvirt on an agent
func (c *Client) DeployLibvirtTask(ctx context.Context, token string, agentID uuid.UUID, driver string, action string, params map[string]interface{}) (*TaskExecution, error) {
	// Validate agent supports this driver deployment
	capability := "libvirt-deploy-" + driver
	if err := c.ValidateAgentCapability(ctx, token, agentID, capability); err != nil {
		return nil, err
	}

	config := map[string]interface{}{
		"driver": driver,
		"action": action,
	}
	for k, v := range params {
		config[k] = v
	}

	return c.ExecuteTask(ctx, token, &ExecuteTaskInput{
		AgentID: agentID,
		Task: TaskInput{
			Type:   "libvirt-deploy",
			Name:   fmt.Sprintf("libvirt-deploy-%s-%s", driver, action),
			Config: config,
		},
		Wait:    true,
		Timeout: 300, // 5 minutes for deployment tasks
	})
}

// CreateArtifact creates a new artifact in csd-core
func (c *Client) CreateArtifact(ctx context.Context, token string, tenantID uuid.UUID, key, artifactType, content string) error {
	mutation := `
		mutation CreateArtifact($input: CreateArtifactInput!) {
			createArtifact(input: $input) {
				id
				key
			}
		}
	`

	_, err := c.ExecuteWithName(ctx, token, "CreateArtifact", mutation, map[string]interface{}{
		"input": map[string]interface{}{
			"key":      key,
			"type":     artifactType,
			"content":  content,
			"tenantId": tenantID.String(),
		},
	})
	return err
}

// Agent represents a csd-core agent
type Agent struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Hostname     string    `json:"hostname"`
	LastSeen     string    `json:"lastSeen"`
	Capabilities []string  `json:"capabilities"` // Supported task types: kubernetes, libvirt, shell, etc.
}

// GetAgent gets agent info from csd-core
func (c *Client) GetAgent(ctx context.Context, token string, agentID uuid.UUID) (*Agent, error) {
	query := `
		query GetAgent($id: ID!) {
			agent(id: $id) {
				id
				name
				status
				hostname
				lastSeen
				capabilities
			}
		}
	`

	resp, err := c.ExecuteWithName(ctx, token, "GetAgent", query, map[string]interface{}{
		"id": agentID.String(),
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		Agent *Agent `json:"agent"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse agent: %w", err)
	}

	return result.Agent, nil
}

// HasCapability checks if an agent has a specific capability
func (a *Agent) HasCapability(capability string) bool {
	for _, cap := range a.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// HasCapabilityPrefix checks if an agent has any capability starting with prefix
func (a *Agent) HasCapabilityPrefix(prefix string) bool {
	for _, cap := range a.Capabilities {
		if len(cap) >= len(prefix) && cap[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// ListAgentsByCapability lists agents that support a specific capability
func (c *Client) ListAgentsByCapability(ctx context.Context, token string, capability string) ([]Agent, error) {
	agents, err := c.ListAgents(ctx, token)
	if err != nil {
		return nil, err
	}

	filtered := make([]Agent, 0)
	for _, agent := range agents {
		if agent.HasCapability(capability) {
			filtered = append(filtered, agent)
		}
	}

	return filtered, nil
}

// ListAgentsByCapabilityPrefix lists agents that have any capability starting with prefix
func (c *Client) ListAgentsByCapabilityPrefix(ctx context.Context, token string, prefix string) ([]Agent, error) {
	agents, err := c.ListAgents(ctx, token)
	if err != nil {
		return nil, err
	}

	filtered := make([]Agent, 0)
	for _, agent := range agents {
		if agent.HasCapabilityPrefix(prefix) {
			filtered = append(filtered, agent)
		}
	}

	return filtered, nil
}

// GetAgentCapabilitiesByPrefix returns capabilities matching a prefix for an agent
func (a *Agent) GetCapabilitiesByPrefix(prefix string) []string {
	result := make([]string, 0)
	for _, cap := range a.Capabilities {
		if len(cap) >= len(prefix) && cap[:len(prefix)] == prefix {
			result = append(result, cap)
		}
	}
	return result
}

// ListAgents lists available agents
func (c *Client) ListAgents(ctx context.Context, token string) ([]Agent, error) {
	query := `
		query ListAgents {
			agents {
				id
				name
				status
				hostname
				lastSeen
				capabilities
			}
		}
	`

	resp, err := c.ExecuteWithName(ctx, token, "ListAgents", query, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Agents []Agent `json:"agents"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse agents: %w", err)
	}

	return result.Agents, nil
}

// ServiceRegistration represents the service registration info
type ServiceRegistration struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Version     string `json:"version"`
	BaseURL     string `json:"baseUrl"`
	CallbackURL string `json:"callbackUrl"`
	Description string `json:"description"`

	// Frontend integration (Module Federation)
	FrontendURL     string            `json:"frontendUrl,omitempty"`
	RemoteEntryPath string            `json:"remoteEntryPath,omitempty"`
	RoutePath       string            `json:"routePath,omitempty"`
	ExposedModules  map[string]string `json:"exposedModules,omitempty"`
}

// RegisterService registers this service with csd-core
func (c *Client) RegisterService(ctx context.Context, serviceToken string, reg *ServiceRegistration) error {
	regURL := c.baseURL + "/core/api/latest/services/register"

	jsonBody, err := json.Marshal(reg)
	if err != nil {
		return fmt.Errorf("failed to marshal registration: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", regURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+serviceToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
