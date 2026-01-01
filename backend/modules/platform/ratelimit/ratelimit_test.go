package ratelimit

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRateLimiter_Allow(t *testing.T) {
	config := &Config{
		DefaultLimit: LimitConfig{
			MaxRequests: 5,
			Window:      time.Second,
			Burst:       2,
		},
		Limits:          make(map[string]LimitConfig),
		CleanupInterval: time.Hour, // Long interval for testing
	}

	rl := NewRateLimiter(config)
	tenantID := uuid.New()
	userID := uuid.New()

	t.Run("allows requests within limit", func(t *testing.T) {
		// Should allow MaxRequests + Burst requests
		for i := 0; i < 7; i++ {
			if !rl.Allow(tenantID, userID, "test-op") {
				t.Errorf("Request %d should be allowed", i+1)
			}
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		// Next request should be blocked
		if rl.Allow(tenantID, userID, "test-op") {
			t.Error("Request should be blocked after limit exceeded")
		}
	})

	t.Run("different operations have separate limits", func(t *testing.T) {
		newTenant := uuid.New()
		// First operation uses up tokens
		for i := 0; i < 7; i++ {
			rl.Allow(newTenant, userID, "op1")
		}
		// Second operation should still work
		if !rl.Allow(newTenant, userID, "op2") {
			t.Error("Different operation should have separate limit")
		}
	})

	t.Run("different tenants have separate limits", func(t *testing.T) {
		tenant1 := uuid.New()
		tenant2 := uuid.New()

		// Use up tenant1's tokens
		for i := 0; i < 7; i++ {
			rl.Allow(tenant1, userID, "test")
		}

		// tenant2 should still work
		if !rl.Allow(tenant2, userID, "test") {
			t.Error("Different tenant should have separate limit")
		}
	})
}

func TestRateLimiter_Reset(t *testing.T) {
	config := &Config{
		DefaultLimit: LimitConfig{
			MaxRequests: 5,
			Window:      time.Second,
			Burst:       0,
		},
		Limits:          make(map[string]LimitConfig),
		CleanupInterval: time.Hour,
	}

	rl := NewRateLimiter(config)
	tenantID := uuid.New()
	userID := uuid.New()

	// Use up all tokens
	for i := 0; i < 5; i++ {
		rl.Allow(tenantID, userID, "test")
	}

	// Should be blocked
	if rl.Allow(tenantID, userID, "test") {
		t.Error("Should be blocked before reset")
	}

	// Reset
	rl.Reset(tenantID)

	// Should work again
	if !rl.Allow(tenantID, userID, "test") {
		t.Error("Should be allowed after reset")
	}
}

func TestRateLimiter_OperationSpecificLimits(t *testing.T) {
	config := &Config{
		DefaultLimit: LimitConfig{
			MaxRequests: 10,
			Window:      time.Second,
			Burst:       0,
		},
		Limits: map[string]LimitConfig{
			"restricted": {
				MaxRequests: 2,
				Window:      time.Second,
				Burst:       0,
			},
		},
		CleanupInterval: time.Hour,
	}

	rl := NewRateLimiter(config)
	tenantID := uuid.New()
	userID := uuid.New()

	// Restricted operation should only allow 2
	for i := 0; i < 2; i++ {
		if !rl.Allow(tenantID, userID, "restricted") {
			t.Errorf("Restricted request %d should be allowed", i+1)
		}
	}
	if rl.Allow(tenantID, userID, "restricted") {
		t.Error("Restricted request 3 should be blocked")
	}

	// Default operation should allow more
	for i := 0; i < 10; i++ {
		if !rl.Allow(tenantID, userID, "default-op") {
			t.Errorf("Default request %d should be allowed", i+1)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	t.Run("has default limit", func(t *testing.T) {
		if config.DefaultLimit.MaxRequests <= 0 {
			t.Error("Default limit should have positive MaxRequests")
		}
		if config.DefaultLimit.Window <= 0 {
			t.Error("Default limit should have positive Window")
		}
	})

	t.Run("has operation-specific limits", func(t *testing.T) {
		expectedOps := []string{
			"clusters",
			"hypervisors",
			"containerEngines",
			"deployCluster",
			"bulkDeleteClusters",
		}
		for _, op := range expectedOps {
			if _, ok := config.Limits[op]; !ok {
				t.Errorf("Expected limit for operation %s", op)
			}
		}
	})

	t.Run("deploy operations are restrictive", func(t *testing.T) {
		deployLimit := config.Limits["deployCluster"]
		defaultLimit := config.DefaultLimit
		if deployLimit.MaxRequests >= defaultLimit.MaxRequests {
			t.Error("Deploy should be more restrictive than default")
		}
	})

	t.Run("bulk delete operations are very restrictive", func(t *testing.T) {
		bulkLimit := config.Limits["bulkDeleteClusters"]
		if bulkLimit.MaxRequests > 5 {
			t.Error("Bulk delete should be very restrictive")
		}
		if bulkLimit.Burst > 0 {
			t.Error("Bulk delete should have no burst")
		}
	})
}

func TestRateLimitError(t *testing.T) {
	err := &RateLimitError{
		Operation:   "test-op",
		RetryAfter:  time.Minute,
		MaxRequests: 10,
		Window:      time.Minute,
	}

	if err.Error() == "" {
		t.Error("Error message should not be empty")
	}
	if err.Operation != "test-op" {
		t.Error("Operation should be preserved")
	}
}

func TestMakeKey(t *testing.T) {
	config := DefaultConfig()
	rl := NewRateLimiter(config)

	tenantID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	userID := uuid.New()

	key := rl.makeKey(tenantID, userID, "test-op")

	// Key should contain tenant ID and operation
	if key == "" {
		t.Error("Key should not be empty")
	}
	// Key format is tenantID:operation (userID is not included for tenant-level limiting)
	expected := tenantID.String() + ":test-op"
	if key != expected {
		t.Errorf("Key = %q, want %q", key, expected)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b     float64
		expected float64
	}{
		{1.0, 2.0, 1.0},
		{2.0, 1.0, 1.0},
		{1.0, 1.0, 1.0},
		{-1.0, 1.0, -1.0},
		{0.5, 0.3, 0.3},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%f, %f) = %f, want %f", tt.a, tt.b, result, tt.expected)
		}
	}
}
