package validation

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"csd-pilote/backend/modules/platform/config"
)

// Limits defines validation limits
const (
	MaxNameLength        = 255
	MaxDescriptionLength = 2000
	MaxSearchLength      = 255
	MaxBulkIDs           = 100
	MaxPaginationLimit   = 100
	MaxTailLines         = 10000
	MaxReplicas          = 1000
	MaxArrayLength       = 1000
	MaxPortNumber        = 65535
	MinPortNumber        = 1
)

// Pre-compiled regex patterns for performance (avoid recompiling on each call)
var (
	portRangeRegex    = regexp.MustCompile(`^(\d+)(-(\d+))?$`)
	k8sNameRegex      = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	dockerImageRegex  = regexp.MustCompile(`^[a-z0-9]([a-z0-9._/-]*[a-z0-9])?(:[a-zA-Z0-9._-]+)?(@sha256:[a-f0-9]{64})?$`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Code    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors holds multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

func (e *ValidationErrors) Add(field, message, code string) {
	e.Errors = append(e.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// Validator provides validation methods
type Validator struct {
	errors *ValidationErrors
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		errors: &ValidationErrors{},
	}
}

// Errors returns all validation errors
func (v *Validator) Errors() *ValidationErrors {
	return v.errors
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return v.errors.HasErrors()
}

// FirstError returns the first error message or empty string
func (v *Validator) FirstError() string {
	if len(v.errors.Errors) > 0 {
		return v.errors.Errors[0].Message
	}
	return ""
}

// Required validates that a string is not empty
func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.errors.Add(field, field+" is required", "REQUIRED")
	}
	return v
}

// MaxLength validates string maximum length
func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if utf8.RuneCountInString(value) > max {
		v.errors.Add(field, fmt.Sprintf("%s must be at most %d characters", field, max), "MAX_LENGTH")
	}
	return v
}

// MinLength validates string minimum length
func (v *Validator) MinLength(field, value string, min int) *Validator {
	if utf8.RuneCountInString(value) < min {
		v.errors.Add(field, fmt.Sprintf("%s must be at least %d characters", field, min), "MIN_LENGTH")
	}
	return v
}

// UUID validates that a string is a valid UUID
func (v *Validator) UUID(field, value string) *Validator {
	if value == "" {
		return v
	}
	if _, err := uuid.Parse(value); err != nil {
		v.errors.Add(field, fmt.Sprintf("%s must be a valid UUID", field), "INVALID_UUID")
	}
	return v
}

// Enum validates that a value is in allowed set
func (v *Validator) Enum(field, value string, allowed []string) *Validator {
	if value == "" {
		return v
	}
	for _, a := range allowed {
		if value == a {
			return v
		}
	}
	v.errors.Add(field, fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowed, ", ")), "INVALID_ENUM")
	return v
}

// Range validates that a number is within range
func (v *Validator) Range(field string, value, min, max int) *Validator {
	if value < min || value > max {
		v.errors.Add(field, fmt.Sprintf("%s must be between %d and %d", field, min, max), "OUT_OF_RANGE")
	}
	return v
}

// Positive validates that a number is positive
func (v *Validator) Positive(field string, value int) *Validator {
	if value < 0 {
		v.errors.Add(field, fmt.Sprintf("%s must be positive", field), "NEGATIVE")
	}
	return v
}

// MaxItems validates array maximum length
func (v *Validator) MaxItems(field string, length, max int) *Validator {
	if length > max {
		v.errors.Add(field, fmt.Sprintf("%s must have at most %d items", field, max), "MAX_ITEMS")
	}
	return v
}

// MinItems validates array minimum length
func (v *Validator) MinItems(field string, length, min int) *Validator {
	if length < min {
		v.errors.Add(field, fmt.Sprintf("%s must have at least %d items", field, min), "MIN_ITEMS")
	}
	return v
}

// IP validates that a string is a valid IP address
func (v *Validator) IP(field, value string) *Validator {
	if value == "" {
		return v
	}
	if net.ParseIP(value) == nil {
		v.errors.Add(field, fmt.Sprintf("%s must be a valid IP address", field), "INVALID_IP")
	}
	return v
}

// CIDR validates that a string is a valid CIDR notation
func (v *Validator) CIDR(field, value string) *Validator {
	if value == "" {
		return v
	}
	if _, _, err := net.ParseCIDR(value); err != nil {
		v.errors.Add(field, fmt.Sprintf("%s must be a valid CIDR notation", field), "INVALID_CIDR")
	}
	return v
}

// Port validates that a number is a valid port
func (v *Validator) Port(field string, value int) *Validator {
	if value < MinPortNumber || value > MaxPortNumber {
		v.errors.Add(field, fmt.Sprintf("%s must be a valid port (1-65535)", field), "INVALID_PORT")
	}
	return v
}

// PortRange validates a port range string (e.g., "80", "80-443")
func (v *Validator) PortRange(field, value string) *Validator {
	if value == "" {
		return v
	}
	if !portRangeRegex.MatchString(value) {
		v.errors.Add(field, fmt.Sprintf("%s must be a valid port or port range", field), "INVALID_PORT_RANGE")
	}
	return v
}

// SafeString validates that a string doesn't contain dangerous characters
func (v *Validator) SafeString(field, value string) *Validator {
	if value == "" {
		return v
	}
	// Check for common injection patterns
	dangerous := []string{"<script", "javascript:", "onclick", "onerror", "--", ";--", "/*", "*/", "@@", "char(", "nchar(", "varchar(", "nvarchar(", "alter ", "begin ", "cast(", "create ", "cursor ", "declare ", "delete ", "drop ", "end ", "exec ", "execute ", "fetch ", "insert ", "kill ", "select ", "sys.", "sysobjects", "syscolumns", "table ", "update "}
	lowerValue := strings.ToLower(value)
	for _, d := range dangerous {
		if strings.Contains(lowerValue, d) {
			v.errors.Add(field, fmt.Sprintf("%s contains invalid characters", field), "UNSAFE_STRING")
			return v
		}
	}
	return v
}

// KubernetesName validates Kubernetes resource names (RFC 1123)
func (v *Validator) KubernetesName(field, value string) *Validator {
	if value == "" {
		return v
	}
	// RFC 1123 DNS label: lowercase alphanumeric, hyphens allowed (not at start/end), max 63 chars
	if len(value) > 63 || !k8sNameRegex.MatchString(value) {
		v.errors.Add(field, fmt.Sprintf("%s must be a valid Kubernetes name (lowercase alphanumeric, hyphens, max 63 chars)", field), "INVALID_K8S_NAME")
	}
	return v
}

// DockerImageName validates Docker image names
func (v *Validator) DockerImageName(field, value string) *Validator {
	if value == "" {
		return v
	}
	// Basic Docker image name validation
	// Allows: registry/repo:tag or repo:tag or repo
	if !dockerImageRegex.MatchString(value) {
		v.errors.Add(field, fmt.Sprintf("%s must be a valid Docker image name", field), "INVALID_DOCKER_IMAGE")
	}
	return v
}

// NftablesExpression validates nftables expression (basic safety check)
func (v *Validator) NftablesExpression(field, value string) *Validator {
	if value == "" {
		return v
	}
	// Disallow shell metacharacters and common injection patterns
	dangerous := []string{"`", "$", "$(", "${", "&&", "||", ";", "|", ">", "<", "\\n", "\\r", "\n", "\r"}
	for _, d := range dangerous {
		if strings.Contains(value, d) {
			v.errors.Add(field, fmt.Sprintf("%s contains invalid characters", field), "UNSAFE_NFTABLES_EXPR")
			return v
		}
	}
	// Basic nftables syntax validation - must start with allowed keywords
	validPrefixes := []string{"ip", "ip6", "tcp", "udp", "icmp", "icmpv6", "ct", "meta", "iif", "oif", "ether", "arp", "vlan", "fib", "rt", "accept", "drop", "reject", "jump", "goto", "return", "queue", "log", "counter", "limit", "masquerade", "snat", "dnat"}
	value = strings.TrimSpace(strings.ToLower(value))
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(value, prefix) {
			hasValidPrefix = true
			break
		}
	}
	if !hasValidPrefix && value != "" {
		v.errors.Add(field, fmt.Sprintf("%s must start with a valid nftables keyword", field), "INVALID_NFTABLES_EXPR")
	}
	return v
}

// ValidateName validates a name field with standard rules
func ValidateName(name string) error {
	v := NewValidator()
	v.Required("name", name).MaxLength("name", name, MaxNameLength).SafeString("name", name)
	if v.HasErrors() {
		return v.errors
	}
	return nil
}

// ValidateDescription validates a description field
func ValidateDescription(description string) error {
	if description == "" {
		return nil
	}
	v := NewValidator()
	v.MaxLength("description", description, MaxDescriptionLength)
	if v.HasErrors() {
		return v.errors
	}
	return nil
}

// ValidatePagination validates limit and offset parameters
func ValidatePagination(limit, offset int) (int, int, error) {
	defaultLimit := 20
	maxLimit := MaxPaginationLimit

	// Use config values if available
	if cfg := config.GetConfig(); cfg != nil {
		if cfg.Pagination.DefaultLimit > 0 {
			defaultLimit = cfg.Pagination.DefaultLimit
		}
		if cfg.Pagination.MaxLimit > 0 {
			maxLimit = cfg.Pagination.MaxLimit
		}
	}

	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset, nil
}

// ValidateBulkIDs validates a list of IDs for bulk operations
func ValidateBulkIDs(ids []interface{}) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("ids is required")
	}
	if len(ids) > MaxBulkIDs {
		return nil, fmt.Errorf("maximum %d IDs allowed per request", MaxBulkIDs)
	}

	result := make([]uuid.UUID, 0, len(ids))
	for _, idRaw := range ids {
		idStr, ok := idRaw.(string)
		if !ok {
			continue
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		result = append(result, id)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid IDs provided")
	}

	return result, nil
}
