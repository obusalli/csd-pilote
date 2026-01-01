package metrics

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds application metrics
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	TotalRequests     uint64
	SuccessfulRequest uint64
	FailedRequests    uint64

	// Rate limit metrics
	RateLimitHits uint64

	// Operation metrics by type
	OperationCounts map[string]*uint64

	// Latency tracking (in milliseconds)
	TotalLatency     uint64
	RequestCount     uint64
	MaxLatency       uint64
	MinLatency       uint64
	latencyInitOnce  sync.Once

	// Error tracking by type
	ErrorCounts map[string]*uint64

	// Start time for uptime calculation
	StartTime time.Time
}

var (
	globalMetrics *Metrics
	metricsOnce   sync.Once
)

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	metricsOnce.Do(func() {
		globalMetrics = &Metrics{
			OperationCounts: make(map[string]*uint64),
			ErrorCounts:     make(map[string]*uint64),
			StartTime:       time.Now(),
			MinLatency:      ^uint64(0), // Max uint64 value initially
		}
	})
	return globalMetrics
}

// IncrementRequests increments total request count
func (m *Metrics) IncrementRequests() {
	atomic.AddUint64(&m.TotalRequests, 1)
}

// IncrementSuccess increments successful request count
func (m *Metrics) IncrementSuccess() {
	atomic.AddUint64(&m.SuccessfulRequest, 1)
}

// IncrementFailed increments failed request count
func (m *Metrics) IncrementFailed() {
	atomic.AddUint64(&m.FailedRequests, 1)
}

// IncrementRateLimitHits increments rate limit hit count
func (m *Metrics) IncrementRateLimitHits() {
	atomic.AddUint64(&m.RateLimitHits, 1)
}

// IncrementOperation increments operation count by type
func (m *Metrics) IncrementOperation(operation string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.OperationCounts[operation] == nil {
		var count uint64
		m.OperationCounts[operation] = &count
	}
	atomic.AddUint64(m.OperationCounts[operation], 1)
}

// IncrementError increments error count by type
func (m *Metrics) IncrementError(errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ErrorCounts[errorType] == nil {
		var count uint64
		m.ErrorCounts[errorType] = &count
	}
	atomic.AddUint64(m.ErrorCounts[errorType], 1)
}

// RecordLatency records request latency in milliseconds
func (m *Metrics) RecordLatency(latencyMs uint64) {
	atomic.AddUint64(&m.TotalLatency, latencyMs)
	atomic.AddUint64(&m.RequestCount, 1)

	// Update max latency
	for {
		oldMax := atomic.LoadUint64(&m.MaxLatency)
		if latencyMs <= oldMax {
			break
		}
		if atomic.CompareAndSwapUint64(&m.MaxLatency, oldMax, latencyMs) {
			break
		}
	}

	// Update min latency
	for {
		oldMin := atomic.LoadUint64(&m.MinLatency)
		if latencyMs >= oldMin {
			break
		}
		if atomic.CompareAndSwapUint64(&m.MinLatency, oldMin, latencyMs) {
			break
		}
	}
}

// GetAverageLatency returns average latency in milliseconds
func (m *Metrics) GetAverageLatency() float64 {
	count := atomic.LoadUint64(&m.RequestCount)
	if count == 0 {
		return 0
	}
	total := atomic.LoadUint64(&m.TotalLatency)
	return float64(total) / float64(count)
}

// Snapshot represents a point-in-time snapshot of metrics
type Snapshot struct {
	Timestamp   time.Time          `json:"timestamp"`
	Uptime      string             `json:"uptime"`
	UptimeSecs  float64            `json:"uptime_seconds"`

	// Request metrics
	TotalRequests     uint64 `json:"total_requests"`
	SuccessfulRequest uint64 `json:"successful_requests"`
	FailedRequests    uint64 `json:"failed_requests"`
	SuccessRate       float64 `json:"success_rate"`

	// Rate limiting
	RateLimitHits uint64 `json:"rate_limit_hits"`

	// Latency
	AverageLatencyMs float64 `json:"average_latency_ms"`
	MaxLatencyMs     uint64  `json:"max_latency_ms"`
	MinLatencyMs     uint64  `json:"min_latency_ms"`

	// Operations
	OperationCounts map[string]uint64 `json:"operation_counts"`

	// Errors
	ErrorCounts map[string]uint64 `json:"error_counts"`

	// Runtime metrics
	Runtime RuntimeMetrics `json:"runtime"`
}

// RuntimeMetrics contains Go runtime information
type RuntimeMetrics struct {
	Goroutines   int    `json:"goroutines"`
	HeapAllocMB  float64 `json:"heap_alloc_mb"`
	HeapSysMB    float64 `json:"heap_sys_mb"`
	StackInUseMB float64 `json:"stack_in_use_mb"`
	NumGC        uint32 `json:"num_gc"`
	GoVersion    string `json:"go_version"`
	NumCPU       int    `json:"num_cpu"`
}

// GetSnapshot returns a snapshot of current metrics
func (m *Metrics) GetSnapshot() Snapshot {
	uptime := time.Since(m.StartTime)
	totalReqs := atomic.LoadUint64(&m.TotalRequests)
	successReqs := atomic.LoadUint64(&m.SuccessfulRequest)

	var successRate float64
	if totalReqs > 0 {
		successRate = float64(successReqs) / float64(totalReqs) * 100
	}

	// Copy operation counts
	m.mu.RLock()
	opCounts := make(map[string]uint64, len(m.OperationCounts))
	for k, v := range m.OperationCounts {
		opCounts[k] = atomic.LoadUint64(v)
	}
	errCounts := make(map[string]uint64, len(m.ErrorCounts))
	for k, v := range m.ErrorCounts {
		errCounts[k] = atomic.LoadUint64(v)
	}
	m.mu.RUnlock()

	// Get runtime stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	minLatency := atomic.LoadUint64(&m.MinLatency)
	if minLatency == ^uint64(0) {
		minLatency = 0 // No requests yet
	}

	return Snapshot{
		Timestamp:         time.Now(),
		Uptime:            uptime.Round(time.Second).String(),
		UptimeSecs:        uptime.Seconds(),
		TotalRequests:     totalReqs,
		SuccessfulRequest: successReqs,
		FailedRequests:    atomic.LoadUint64(&m.FailedRequests),
		SuccessRate:       successRate,
		RateLimitHits:     atomic.LoadUint64(&m.RateLimitHits),
		AverageLatencyMs:  m.GetAverageLatency(),
		MaxLatencyMs:      atomic.LoadUint64(&m.MaxLatency),
		MinLatencyMs:      minLatency,
		OperationCounts:   opCounts,
		ErrorCounts:       errCounts,
		Runtime: RuntimeMetrics{
			Goroutines:   runtime.NumGoroutine(),
			HeapAllocMB:  float64(memStats.HeapAlloc) / 1024 / 1024,
			HeapSysMB:    float64(memStats.HeapSys) / 1024 / 1024,
			StackInUseMB: float64(memStats.StackInuse) / 1024 / 1024,
			NumGC:        memStats.NumGC,
			GoVersion:    runtime.Version(),
			NumCPU:       runtime.NumCPU(),
		},
	}
}

// Reset resets all metrics (useful for testing)
func (m *Metrics) Reset() {
	atomic.StoreUint64(&m.TotalRequests, 0)
	atomic.StoreUint64(&m.SuccessfulRequest, 0)
	atomic.StoreUint64(&m.FailedRequests, 0)
	atomic.StoreUint64(&m.RateLimitHits, 0)
	atomic.StoreUint64(&m.TotalLatency, 0)
	atomic.StoreUint64(&m.RequestCount, 0)
	atomic.StoreUint64(&m.MaxLatency, 0)
	atomic.StoreUint64(&m.MinLatency, ^uint64(0))

	m.mu.Lock()
	m.OperationCounts = make(map[string]*uint64)
	m.ErrorCounts = make(map[string]*uint64)
	m.StartTime = time.Now()
	m.mu.Unlock()
}

// MetricsHandler returns an HTTP handler for metrics endpoint
func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		snapshot := GetMetrics().GetSnapshot()
		json.NewEncoder(w).Encode(snapshot)
	}
}
