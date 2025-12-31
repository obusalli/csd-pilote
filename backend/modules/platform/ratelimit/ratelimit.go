package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RateLimiter represents a token bucket rate limiter
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*Bucket
	config   *Config
	cleanupT *time.Ticker
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
			MaxRequests: 100,
			Window:      time.Minute,
			Burst:       10,
		},
		Limits: map[string]LimitConfig{
			// Deploy operations - more restrictive
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

			// Bulk delete operations
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
	for range rl.cleanupT.C {
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
	tenantID, ok := r.Context().Value("tenantId").(uuid.UUID)
	if !ok {
		// No tenant, no rate limiting
		return nil
	}
	userID, ok := r.Context().Value("userId").(uuid.UUID)
	if !ok {
		userID = uuid.Nil
	}

	limiter := GetRateLimiter()
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

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
