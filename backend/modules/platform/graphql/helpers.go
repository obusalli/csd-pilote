package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"csd-pilote/backend/modules/platform/middleware"
	"csd-pilote/backend/modules/platform/pagination"
	"csd-pilote/backend/modules/platform/validation"
)

// ========================================
// Context Extraction Helpers
// ========================================

// RequestContext holds common context values for GraphQL handlers
type RequestContext struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Token    string
}

// GetTenantContext extracts tenantID from context, writes error if not found
// Returns tenantID and true if successful, uuid.Nil and false otherwise
func GetTenantContext(ctx context.Context, w http.ResponseWriter) (uuid.UUID, bool) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w)
		return uuid.Nil, false
	}
	return tenantID, true
}

// GetRequestContext extracts tenant, user and token from context for mutations
// Returns RequestContext and true if successful, writes error and returns false otherwise
func GetRequestContext(ctx context.Context, w http.ResponseWriter) (*RequestContext, bool) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w)
		return nil, false
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		WriteUnauthorized(w)
		return nil, false
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	return &RequestContext{
		TenantID: tenantID,
		UserID:   user.UserID,
		Token:    token,
	}, true
}

// ========================================
// Input Extraction Helpers
// ========================================

// RequireInput extracts and validates the input map from variables
// Returns the input map and true if successful, nil and false otherwise
func RequireInput(variables map[string]interface{}, w http.ResponseWriter) (map[string]interface{}, bool) {
	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		WriteValidationError(w, "input is required")
		return nil, false
	}
	return inputRaw, true
}

// RequireFilter extracts the filter map from variables (optional)
// Returns nil if no filter provided (not an error)
func GetFilter(variables map[string]interface{}) map[string]interface{} {
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		return f
	}
	return nil
}

// ParseFilterSearch extracts and validates search from filter
// Returns empty string if not present
func ParseFilterSearch(filter map[string]interface{}) (string, error) {
	if filter == nil {
		return "", nil
	}
	if search, ok := filter["search"].(string); ok {
		if len(search) > validation.MaxSearchLength {
			return "", validation.NewValidationError("search term too long")
		}
		return search, nil
	}
	return "", nil
}

// ParsePagination extracts and validates pagination parameters
func ParsePagination(variables map[string]interface{}) (limit, offset int) {
	limit = pagination.DefaultLimit()
	offset = 0

	if l, ok := variables["limit"].(float64); ok {
		limit = int(l)
	}
	if o, ok := variables["offset"].(float64); ok {
		offset = int(o)
	}

	// Apply validation limits
	limit, offset, _ = validation.ValidatePagination(limit, offset)
	return limit, offset
}

// ParseUUID extracts and validates a UUID from variables
func ParseUUID(variables map[string]interface{}, key string) (uuid.UUID, error) {
	idStr, ok := variables[key].(string)
	if !ok {
		return uuid.Nil, validation.NewValidationError(key + " is required")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, validation.NewValidationError("invalid " + key)
	}

	return id, nil
}

// ParseBulkUUIDs extracts and validates a list of UUIDs for bulk operations
func ParseBulkUUIDs(variables map[string]interface{}, key string) ([]uuid.UUID, error) {
	idsRaw, ok := variables[key].([]interface{})
	if !ok || len(idsRaw) == 0 {
		return nil, validation.NewValidationError(key + " is required")
	}

	return validation.ValidateBulkIDs(idsRaw)
}

// ParseString extracts a string from variables
func ParseString(variables map[string]interface{}, key string) string {
	if v, ok := variables[key].(string); ok {
		return v
	}
	return ""
}

// ParseStringRequired extracts a required string from variables
func ParseStringRequired(variables map[string]interface{}, key string) (string, error) {
	v, ok := variables[key].(string)
	if !ok || v == "" {
		return "", validation.NewValidationError(key + " is required")
	}
	return v, nil
}

// ParseInt extracts an int from variables with default
func ParseInt(variables map[string]interface{}, key string, defaultVal int) int {
	if v, ok := variables[key].(float64); ok {
		return int(v)
	}
	return defaultVal
}

// ParseIntWithMax extracts an int from variables with max limit
func ParseIntWithMax(variables map[string]interface{}, key string, defaultVal, maxVal int) int {
	val := ParseInt(variables, key, defaultVal)
	if val > maxVal {
		return maxVal
	}
	if val < 0 {
		return defaultVal
	}
	return val
}

// ParseBool extracts a bool from variables with default
func ParseBool(variables map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := variables[key].(bool); ok {
		return v
	}
	return defaultVal
}

// WriteError writes a sanitized error response
func WriteError(w http.ResponseWriter, err error, context string) {
	safeMsg := validation.SafeErrorMessage(err, context)
	json.NewEncoder(w).Encode(NewErrorResponse(safeMsg))
}

// WriteValidationError writes a validation error response
func WriteValidationError(w http.ResponseWriter, message string) {
	json.NewEncoder(w).Encode(NewErrorResponse(message))
}

// WriteUnauthorized writes an unauthorized error response
func WriteUnauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(NewErrorResponseWithCode(
		string(validation.ErrCodeUnauthorized),
		validation.NewUnauthorizedError().Message,
	))
}

// WriteSuccess writes a successful data response
func WriteSuccess(w http.ResponseWriter, data map[string]interface{}) {
	json.NewEncoder(w).Encode(NewDataResponse(data))
}

// ValidateEnum validates that a value is in the allowed set
func ValidateEnum(value string, allowed []string, fieldName string) error {
	if value == "" {
		return nil // Empty is OK, will use default
	}
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return validation.NewValidationError(fieldName + " must be one of: " + joinStrings(allowed))
}

func joinStrings(s []string) string {
	return strings.Join(s, ", ")
}

// EnumValues for various types
var (
	ClusterStatusValues       = []string{"PENDING", "DEPLOYING", "CONNECTED", "DISCONNECTED", "ERROR"}
	ClusterModeValues         = []string{"CONNECT", "DEPLOY"}
	HypervisorStatusValues    = []string{"PENDING", "DEPLOYING", "CONNECTED", "DISCONNECTED", "ERROR"}
	HypervisorModeValues      = []string{"CONNECT", "DEPLOY"}
	LibvirtDriverValues       = []string{"qemu", "xen", "lxc"}
	ContainerEngineTypeValues   = []string{"DOCKER", "PODMAN"}
	ContainerEngineStatusValues = []string{"PENDING", "CONNECTED", "DISCONNECTED", "ERROR"}
	ContainerActionValues     = []string{"start", "stop", "restart", "pause", "unpause", "kill", "remove"}
	RuleChainValues           = []string{"INPUT", "OUTPUT", "FORWARD", "PREROUTING", "POSTROUTING"}
	RuleProtocolValues        = []string{"tcp", "udp", "icmp", "icmpv6", "all", "any"}
	RuleActionValues          = []string{"ACCEPT", "DROP", "REJECT", "LOG", "MASQUERADE", "SNAT", "DNAT", "RETURN", "JUMP"}
	DeploymentStatusValues    = []string{"PENDING", "RUNNING", "COMPLETED", "FAILED", "ROLLED_BACK"}
	TemplateCategoryValues    = []string{"BASIC", "WEBSERVER", "DATABASE", "MAIL", "DNS", "MONITORING", "SECURITY", "CUSTOM"}
	KubernetesDistroValues    = []string{"K3S", "RKE2", "KUBEADM", "K0S", "MICROK8S", "EKS", "GKE", "AKS", "OPENSHIFT", "RANCHER", "OTHER"}
)
