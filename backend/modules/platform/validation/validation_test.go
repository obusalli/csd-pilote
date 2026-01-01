package validation

import (
	"strings"
	"testing"
)

func TestValidator_Required(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"valid string", "hello", false},
		{"string with spaces", "hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Required("field", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("Required(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_MaxLength(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		max      int
		hasError bool
	}{
		{"within limit", "hello", 10, false},
		{"at limit", "hello", 5, false},
		{"exceeds limit", "hello world", 5, true},
		{"empty string", "", 5, false},
		{"unicode within limit", "héllo", 5, false},
		{"unicode exceeds limit", "héllo world", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.MaxLength("field", tt.value, tt.max)
			if v.HasErrors() != tt.hasError {
				t.Errorf("MaxLength(%q, %d) hasError = %v, want %v", tt.value, tt.max, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_MinLength(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		min      int
		hasError bool
	}{
		{"above minimum", "hello", 3, false},
		{"at minimum", "hello", 5, false},
		{"below minimum", "hi", 5, true},
		{"empty string", "", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.MinLength("field", tt.value, tt.min)
			if v.HasErrors() != tt.hasError {
				t.Errorf("MinLength(%q, %d) hasError = %v, want %v", tt.value, tt.min, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_UUID(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"valid UUID", "550e8400-e29b-41d4-a716-446655440000", false},
		{"invalid UUID", "not-a-uuid", true},
		{"empty string", "", false}, // Empty is allowed
		{"partial UUID", "550e8400-e29b-41d4", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.UUID("field", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("UUID(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_Enum(t *testing.T) {
	allowed := []string{"a", "b", "c"}
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"valid value", "a", false},
		{"another valid value", "c", false},
		{"invalid value", "d", true},
		{"empty string", "", false}, // Empty is allowed
		{"case sensitive", "A", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Enum("field", tt.value, allowed)
			if v.HasErrors() != tt.hasError {
				t.Errorf("Enum(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_Range(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		min      int
		max      int
		hasError bool
	}{
		{"within range", 5, 1, 10, false},
		{"at minimum", 1, 1, 10, false},
		{"at maximum", 10, 1, 10, false},
		{"below minimum", 0, 1, 10, true},
		{"above maximum", 11, 1, 10, true},
		{"negative value", -5, 1, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Range("field", tt.value, tt.min, tt.max)
			if v.HasErrors() != tt.hasError {
				t.Errorf("Range(%d, %d, %d) hasError = %v, want %v", tt.value, tt.min, tt.max, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_IP(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"valid IPv4", "192.168.1.1", false},
		{"valid IPv6", "::1", false},
		{"valid IPv6 full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"invalid IP", "not.an.ip", true},
		{"empty string", "", false}, // Empty is allowed
		{"partial IP", "192.168", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.IP("field", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("IP(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_CIDR(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"valid CIDR", "192.168.1.0/24", false},
		{"valid IPv6 CIDR", "2001:db8::/32", false},
		{"invalid CIDR", "192.168.1.0", true},
		{"empty string", "", false}, // Empty is allowed
		{"invalid prefix", "192.168.1.0/33", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.CIDR("field", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("CIDR(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_Port(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		hasError bool
	}{
		{"valid port", 80, false},
		{"minimum port", 1, false},
		{"maximum port", 65535, false},
		{"port zero", 0, true},
		{"negative port", -1, true},
		{"port too high", 65536, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Port("field", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("Port(%d) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_PortRange(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"single port", "80", false},
		{"port range", "80-443", false},
		{"empty string", "", false},
		{"invalid format", "80:443", true},
		{"invalid characters", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.PortRange("field", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("PortRange(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_SafeString(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"safe string", "hello world", false},
		{"empty string", "", false},
		{"script tag", "<script>alert('xss')</script>", true},
		{"javascript protocol", "javascript:alert('xss')", true},
		{"sql injection", "'; DROP TABLE users;--", true},
		{"sql select", "SELECT * FROM users", true},
		{"normal text with numbers", "user123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.SafeString("field", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("SafeString(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_KubernetesName(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"valid name", "my-app", false},
		{"valid with numbers", "app-v1", false},
		{"empty string", "", false},
		{"uppercase", "MyApp", true},
		{"starts with hyphen", "-myapp", true},
		{"ends with hyphen", "myapp-", true},
		{"too long", strings.Repeat("a", 64), true},
		{"max length", strings.Repeat("a", 63), false},
		{"special characters", "my_app", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.KubernetesName("field", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("KubernetesName(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_DockerImageName(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"simple name", "nginx", false},
		{"with tag", "nginx:latest", false},
		{"with registry", "docker.io/library/nginx", false},
		{"with sha256", "nginx@sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", false},
		{"empty string", "", false},
		{"invalid characters", "nginx:latest!", true},
		{"uppercase start", "Nginx", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.DockerImageName("field", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("DockerImageName(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_NftablesExpression(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"valid tcp rule", "tcp dport 22", false},
		{"valid ip rule", "ip saddr 192.168.1.0/24", false},
		{"empty string", "", false},
		{"shell injection backtick", "tcp `rm -rf /`", true},
		{"shell injection dollar", "tcp $(rm -rf /)", true},
		{"pipe injection", "tcp | rm -rf /", true},
		{"semicolon injection", "tcp; rm -rf /", true},
		{"invalid prefix", "invalid rule", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.NftablesExpression("field", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("NftablesExpression(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"valid name", "my-resource", false},
		{"empty name", "", true},
		{"too long", strings.Repeat("a", 256), true},
		{"with script", "<script>", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.value)
			hasError := err != nil
			if hasError != tt.hasError {
				t.Errorf("ValidateName(%q) hasError = %v, want %v", tt.value, hasError, tt.hasError)
			}
		})
	}
}

func TestValidatePagination(t *testing.T) {
	tests := []struct {
		name           string
		limit          int
		offset         int
		expectedLimit  int
		expectedOffset int
	}{
		{"default values", 0, 0, 20, 0},
		{"negative limit", -5, 0, 20, 0},
		{"negative offset", 10, -5, 10, 0},
		{"limit too high", 200, 0, MaxPaginationLimit, 0},
		{"valid values", 50, 100, 50, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, offset, _ := ValidatePagination(tt.limit, tt.offset)
			if limit != tt.expectedLimit {
				t.Errorf("ValidatePagination limit = %d, want %d", limit, tt.expectedLimit)
			}
			if offset != tt.expectedOffset {
				t.Errorf("ValidatePagination offset = %d, want %d", offset, tt.expectedOffset)
			}
		})
	}
}

func TestValidateBulkIDs(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		_, err := ValidateBulkIDs([]interface{}{})
		if err == nil {
			t.Error("Expected error for empty list")
		}
	})

	t.Run("valid UUIDs", func(t *testing.T) {
		ids := []interface{}{
			"550e8400-e29b-41d4-a716-446655440000",
			"550e8400-e29b-41d4-a716-446655440001",
		}
		result, err := ValidateBulkIDs(ids)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 IDs, got %d", len(result))
		}
	})

	t.Run("too many IDs", func(t *testing.T) {
		ids := make([]interface{}, MaxBulkIDs+1)
		for i := range ids {
			ids[i] = "550e8400-e29b-41d4-a716-446655440000"
		}
		_, err := ValidateBulkIDs(ids)
		if err == nil {
			t.Error("Expected error for too many IDs")
		}
	})

	t.Run("invalid UUIDs filtered", func(t *testing.T) {
		ids := []interface{}{
			"550e8400-e29b-41d4-a716-446655440000",
			"not-a-uuid",
			123, // wrong type
		}
		result, err := ValidateBulkIDs(ids)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("Expected 1 valid ID, got %d", len(result))
		}
	})
}

func TestValidator_Chaining(t *testing.T) {
	t.Run("multiple validations", func(t *testing.T) {
		v := NewValidator()
		v.Required("name", "test").
			MaxLength("name", "test", 10).
			MinLength("name", "test", 2)

		if v.HasErrors() {
			t.Error("Expected no errors for valid input")
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		v := NewValidator()
		v.Required("name", "").
			Required("email", "")

		if !v.HasErrors() {
			t.Error("Expected errors")
		}
		if len(v.Errors().Errors) != 2 {
			t.Errorf("Expected 2 errors, got %d", len(v.Errors().Errors))
		}
	})
}

func TestValidationError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		err := &ValidationError{
			Field:   "name",
			Message: "is required",
			Code:    "REQUIRED",
		}
		expected := "name: is required"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})
}

func TestValidationErrors(t *testing.T) {
	t.Run("empty errors", func(t *testing.T) {
		errs := &ValidationErrors{}
		if errs.Error() != "" {
			t.Error("Expected empty error message")
		}
		if errs.HasErrors() {
			t.Error("Expected no errors")
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		errs := &ValidationErrors{}
		errs.Add("field1", "error1", "CODE1")
		errs.Add("field2", "error2", "CODE2")

		if !errs.HasErrors() {
			t.Error("Expected errors")
		}
		if len(errs.Errors) != 2 {
			t.Errorf("Expected 2 errors, got %d", len(errs.Errors))
		}
	})
}
