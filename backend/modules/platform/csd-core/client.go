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
