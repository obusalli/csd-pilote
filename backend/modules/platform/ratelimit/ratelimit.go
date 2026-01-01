package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"csd-pilote/backend/modules/platform/middleware"

	"github.com/google/uuid"
)

// RateLimiter represents a token bucket rate limiter
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*Bucket
	config   *Config
	cleanupT *time.Ticker
	stopCh   chan struct{}
	stopped  bool
}

// Bucket represents a token bucket for a specific key
type Bucket struct {
	tokens     float64
	lastRefill time.Time
}

// Config configures rate limits for different operations
type Config struct {
	// Default limit for unspecified operations
	DefaultLimit LimitConfig

	// Operation-specific limits
	Limits map[string]LimitConfig

	// Cleanup interval for expired buckets
	CleanupInterval time.Duration
}

// LimitConfig defines rate limit parameters
type LimitConfig struct {
	// Maximum number of requests per window
	MaxRequests int

	// Time window duration
	Window time.Duration

	// Burst capacity (additional tokens for spikes)
	Burst int
}

// DefaultConfig returns default rate limit configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultLimit: LimitConfig{
			MaxRequests: 60,
			Window:      time.Minute,
			Burst:       10,
		},
		Limits: map[string]LimitConfig{
			// ===== UNAUTHENTICATED REQUESTS =====
			// Stricter limits for requests without valid auth
			"__unauthenticated__": {
				MaxRequests: 20,
				Window:      time.Minute,
				Burst:       5,
			},
			// ===== QUERIES =====
			// List queries - moderate limits to prevent data dumps
			"clusters": {
				MaxRequests: 60,
				Window:      time.Minute,
				Burst:       10,
			},
			"hypervisors": {
				MaxRequests: 60,
				Window:      time.Minute,
				Burst:       10,
			},
			"containerEngines": {
				MaxRequests: 60,
				Window:      time.Minute,
				Burst:       10,
			},
			"pods": {
				MaxRequests: 120,
				Window:      time.Minute,
				Burst:       20,
			},
			"deployments": {
				MaxRequests: 120,
				Window:      time.Minute,
				Burst:       20,
			},
			"services": {
				MaxRequests: 120,
				Window:      time.Minute,
				Burst:       20,
			},
			"namespaces": {
				MaxRequests: 60,
				Window:      time.Minute,
				Burst:       10,
			},
			// Security queries
			"securityRules": {
				MaxRequests: 60,
				Window:      time.Minute,
				Burst:       10,
			},
			"securityProfiles": {
				MaxRequests: 60,
				Window:      time.Minute,
				Burst:       10,
			},
			"securityDeployments": {
				MaxRequests: 30,
				Window:      time.Minute,
				Burst:       5,
			},
			// Log queries - more restrictive (can be large)
			"podLogs": {
				MaxRequests: 30,
				Window:      time.Minute,
				Burst:       5,
			},
			"containerLogs": {
				MaxRequests: 30,
				Window:      time.Minute,
				Burst:       5,
			},

			// ===== MUTATIONS =====
			// Deploy operations - very restrictive
			"deployCluster": {
				MaxRequests: 5,
				Window:      time.Minute,
				Burst:       1,
			},
			"deployHypervisor": {
				MaxRequests: 5,
				Window:      time.Minute,
				Burst:       1,
			},
			"deploySecurityProfile": {
				MaxRequests: 10,
				Window:      time.Minute,
				Burst:       2,
			},
			"flushSecurityRules": {
				MaxRequests: 3,
				Window:      time.Minute,
				Burst:       0,
			},

			// Test connection operations
			"testClusterConnection": {
				MaxRequests: 10,
				Window:      time.Minute,
				Burst:       2,
			},
			"testHypervisorConnection": {
				MaxRequests: 10,
				Window:      time.Minute,
				Burst:       2,
			},
			"testContainerEngineConnection": {
				MaxRequests: 10,
				Window:      time.Minute,
				Burst:       2,
			},

			// Bulk delete operations - very restrictive
			"bulkDeleteClusters": {
				MaxRequests: 3,
				Window:      time.Minute,
				Burst:       0,
			},
			"bulkDeleteHypervisors": {
				MaxRequests: 3,
				Window:      time.Minute,
				Burst:       0,
			},
			"bulkDeleteContainerEngines": {
				MaxRequests: 3,
				Window:      time.Minute,
				Burst:       0,
			},
			"bulkDeleteSecurityRules": {
				MaxRequests: 3,
				Window:      time.Minute,
				Burst:       0,
			},

			// Create/Update operations - moderate limits
			"createCluster": {
				MaxRequests: 20,
				Window:      time.Minute,
				Burst:       5,
			},
			"createHypervisor": {
				MaxRequests: 20,
				Window:      time.Minute,
				Burst:       5,
			},
			"createContainerEngine": {
				MaxRequests: 20,
				Window:      time.Minute,
				Burst:       5,
			},
			"createSecurityRule": {
				MaxRequests: 30,
				Window:      time.Minute,
				Burst:       10,
			},
			"createSecurityProfile": {
				MaxRequests: 20,
				Window:      time.Minute,
				Burst:       5,
			},

			// Container/Pod actions - moderate
			"containerAction": {
				MaxRequests: 30,
				Window:      time.Minute,
				Burst:       10,
			},
			"pullImage": {
				MaxRequests: 10,
				Window:      time.Minute,
				Burst:       2,
			},
			"scaleDeployment": {
				MaxRequests: 20,
				Window:      time.Minute,
				Burst:       5,
			},
		},
		CleanupInterval: 5 * time.Minute,
	}
}

var globalLimiter *RateLimiter
var limiterOnce sync.Once

// GetRateLimiter returns the global rate limiter singleton
func GetRateLimiter() *RateLimiter {
	limiterOnce.Do(func() {
		globalLimiter = NewRateLimiter(DefaultConfig())
	})
	return globalLimiter
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *Config) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*Bucket),
		config:  config,
		stopCh:  make(chan struct{}),
	}

	// Start cleanup goroutine
	rl.cleanupT = time.NewTicker(config.CleanupInterval)
	go rl.cleanup()

	return rl
}

// Allow checks if an operation is allowed for a given tenant/user
func (rl *RateLimiter) Allow(tenantID, userID uuid.UUID, operation string) bool {
	key := rl.makeKey(tenantID, userID, operation)
	limit := rl.getLimit(operation)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[key]
	now := time.Now()

	if !exists {
		// Create new bucket with full capacity
		bucket = &Bucket{
			tokens:     float64(limit.MaxRequests + limit.Burst),
			lastRefill: now,
		}
		rl.buckets[key] = bucket
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := float64(limit.MaxRequests) * (elapsed.Seconds() / limit.Window.Seconds())
	maxTokens := float64(limit.MaxRequests + limit.Burst)

	bucket.tokens = min(bucket.tokens+tokensToAdd, maxTokens)
	bucket.lastRefill = now

	// Check if we have at least 1 token
	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}

	return false
}

// getLimit returns the limit config for an operation
func (rl *RateLimiter) getLimit(operation string) LimitConfig {
	if limit, ok := rl.config.Limits[operation]; ok {
		return limit
	}
	return rl.config.DefaultLimit
}

// makeKey creates a unique key for a rate limit bucket
func (rl *RateLimiter) makeKey(tenantID, userID uuid.UUID, operation string) string {
	// Rate limit per tenant + operation (not per user to prevent abuse)
	return tenantID.String() + ":" + operation
}

// cleanup removes expired buckets periodically
func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.stopCh:
			return
		case <-rl.cleanupT.C:
			rl.mu.Lock()
			now := time.Now()
			for key, bucket := range rl.buckets {
				// Remove buckets that haven't been used for 10 minutes
				if now.Sub(bucket.lastRefill) > 10*time.Minute {
					delete(rl.buckets, key)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Stop stops the rate limiter cleanup goroutine
func (rl *RateLimiter) Stop() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.stopped {
		return
	}
	rl.stopped = true
	rl.cleanupT.Stop()
	close(rl.stopCh)
}

// Reset resets rate limits for a tenant (useful for testing)
func (rl *RateLimiter) Reset(tenantID uuid.UUID) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	prefix := tenantID.String() + ":"
	for key := range rl.buckets {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(rl.buckets, key)
		}
	}
}

// RateLimitError represents a rate limit exceeded error
type RateLimitError struct {
	Operation   string
	RetryAfter  time.Duration
	MaxRequests int
	Window      time.Duration
}

func (e *RateLimitError) Error() string {
	return "rate limit exceeded for operation: " + e.Operation
}

// CheckRateLimit checks rate limit and returns an error if exceeded
func CheckRateLimit(r *http.Request, operation string) error {
	limiter := GetRateLimiter()

	// Use typed context keys from middleware (not raw strings)
	tenantID, ok := middleware.GetTenantIDFromContext(r.Context())
	if !ok {
		// No tenant - apply IP-based rate limiting for unauthenticated requests
		clientIP := getClientIP(r)
		if !limiter.AllowByIP(clientIP, operation) {
			limit := limiter.getLimit("__unauthenticated__")
			return &RateLimitError{
				Operation:   operation,
				RetryAfter:  limit.Window,
				MaxRequests: limit.MaxRequests,
				Window:      limit.Window,
			}
		}
		return nil
	}

	// Get user ID for more granular limiting if needed
	var userID uuid.UUID
	if user, ok := middleware.GetUserFromContext(r.Context()); ok {
		userID = user.UserID
	}

	if !limiter.Allow(tenantID, userID, operation) {
		limit := limiter.getLimit(operation)
		return &RateLimitError{
			Operation:   operation,
			RetryAfter:  limit.Window,
			MaxRequests: limit.MaxRequests,
			Window:      limit.Window,
		}
	}

	return nil
}

// AllowByIP checks if an operation is allowed for a given IP address (unauthenticated requests)
func (rl *RateLimiter) AllowByIP(ip string, operation string) bool {
	key := "ip:" + ip + ":" + operation
	limit := rl.getLimit("__unauthenticated__")

	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[key]
	now := time.Now()

	if !exists {
		bucket = &Bucket{
			tokens:     float64(limit.MaxRequests + limit.Burst),
			lastRefill: now,
		}
		rl.buckets[key] = bucket
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := float64(limit.MaxRequests) * (elapsed.Seconds() / limit.Window.Seconds())
	maxTokens := float64(limit.MaxRequests + limit.Burst)

	bucket.tokens = min(bucket.tokens+tokensToAdd, maxTokens)
	bucket.lastRefill = now

	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}

	return false
}

// getClientIP extracts the client IP from the request
// Handles X-Forwarded-For and X-Real-IP headers for proxied requests
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain multiple IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client)
		if idx := indexByte(xff, ','); idx != -1 {
			return trimSpace(xff[:idx])
		}
		return trimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return trimSpace(xri)
	}

	// Fall back to RemoteAddr
	addr := r.RemoteAddr
	// Remove port if present
	if idx := lastIndexByte(addr, ':'); idx != -1 {
		// Check if it's IPv6 (has brackets)
		if bracketIdx := lastIndexByte(addr, ']'); bracketIdx != -1 && bracketIdx < idx {
			return addr[:idx]
		} else if indexByte(addr, ':') == idx {
			// Only one colon, so it's IPv4:port
			return addr[:idx]
		}
	}
	return addr
}

// Helper functions to avoid importing strings package
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
