package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetMetrics(t *testing.T) {
	m1 := GetMetrics()
	m2 := GetMetrics()

	if m1 != m2 {
		t.Error("GetMetrics should return the same singleton instance")
	}
}

func TestMetrics_IncrementRequests(t *testing.T) {
	m := &Metrics{
		OperationCounts: make(map[string]*uint64),
		ErrorCounts:     make(map[string]*uint64),
		StartTime:       time.Now(),
		MinLatency:      ^uint64(0),
	}

	m.IncrementRequests()
	m.IncrementRequests()
	m.IncrementRequests()

	if m.TotalRequests != 3 {
		t.Errorf("Expected TotalRequests = 3, got %d", m.TotalRequests)
	}
}

func TestMetrics_IncrementSuccess(t *testing.T) {
	m := &Metrics{
		OperationCounts: make(map[string]*uint64),
		ErrorCounts:     make(map[string]*uint64),
		StartTime:       time.Now(),
		MinLatency:      ^uint64(0),
	}

	m.IncrementSuccess()
	m.IncrementSuccess()

	if m.SuccessfulRequest != 2 {
		t.Errorf("Expected SuccessfulRequest = 2, got %d", m.SuccessfulRequest)
	}
}

func TestMetrics_IncrementFailed(t *testing.T) {
	m := &Metrics{
		OperationCounts: make(map[string]*uint64),
		ErrorCounts:     make(map[string]*uint64),
		StartTime:       time.Now(),
		MinLatency:      ^uint64(0),
	}

	m.IncrementFailed()

	if m.FailedRequests != 1 {
		t.Errorf("Expected FailedRequests = 1, got %d", m.FailedRequests)
	}
}

func TestMetrics_IncrementRateLimitHits(t *testing.T) {
	m := &Metrics{
		OperationCounts: make(map[string]*uint64),
		ErrorCounts:     make(map[string]*uint64),
		StartTime:       time.Now(),
		MinLatency:      ^uint64(0),
	}

	m.IncrementRateLimitHits()
	m.IncrementRateLimitHits()

	if m.RateLimitHits != 2 {
		t.Errorf("Expected RateLimitHits = 2, got %d", m.RateLimitHits)
	}
}

func TestMetrics_IncrementOperation(t *testing.T) {
	m := &Metrics{
		OperationCounts: make(map[string]*uint64),
		ErrorCounts:     make(map[string]*uint64),
		StartTime:       time.Now(),
		MinLatency:      ^uint64(0),
	}

	m.IncrementOperation("createCluster")
	m.IncrementOperation("createCluster")
	m.IncrementOperation("deleteCluster")

	if m.OperationCounts["createCluster"] == nil {
		t.Error("Expected createCluster operation to be tracked")
	} else if *m.OperationCounts["createCluster"] != 2 {
		t.Errorf("Expected createCluster count = 2, got %d", *m.OperationCounts["createCluster"])
	}

	if m.OperationCounts["deleteCluster"] == nil {
		t.Error("Expected deleteCluster operation to be tracked")
	} else if *m.OperationCounts["deleteCluster"] != 1 {
		t.Errorf("Expected deleteCluster count = 1, got %d", *m.OperationCounts["deleteCluster"])
	}
}

func TestMetrics_IncrementError(t *testing.T) {
	m := &Metrics{
		OperationCounts: make(map[string]*uint64),
		ErrorCounts:     make(map[string]*uint64),
		StartTime:       time.Now(),
		MinLatency:      ^uint64(0),
	}

	m.IncrementError("validation_error")
	m.IncrementError("validation_error")
	m.IncrementError("auth_error")

	if m.ErrorCounts["validation_error"] == nil {
		t.Error("Expected validation_error to be tracked")
	} else if *m.ErrorCounts["validation_error"] != 2 {
		t.Errorf("Expected validation_error count = 2, got %d", *m.ErrorCounts["validation_error"])
	}

	if m.ErrorCounts["auth_error"] == nil {
		t.Error("Expected auth_error to be tracked")
	} else if *m.ErrorCounts["auth_error"] != 1 {
		t.Errorf("Expected auth_error count = 1, got %d", *m.ErrorCounts["auth_error"])
	}
}

func TestMetrics_RecordLatency(t *testing.T) {
	m := &Metrics{
		OperationCounts: make(map[string]*uint64),
		ErrorCounts:     make(map[string]*uint64),
		StartTime:       time.Now(),
		MinLatency:      ^uint64(0),
	}

	m.RecordLatency(100) // 100ms
	m.RecordLatency(200) // 200ms
	m.RecordLatency(50)  // 50ms

	if m.TotalLatency != 350 {
		t.Errorf("Expected TotalLatency = 350, got %d", m.TotalLatency)
	}

	if m.RequestCount != 3 {
		t.Errorf("Expected RequestCount = 3, got %d", m.RequestCount)
	}

	if m.MaxLatency != 200 {
		t.Errorf("Expected MaxLatency = 200, got %d", m.MaxLatency)
	}

	if m.MinLatency != 50 {
		t.Errorf("Expected MinLatency = 50, got %d", m.MinLatency)
	}
}

func TestMetrics_GetAverageLatency(t *testing.T) {
	m := &Metrics{
		OperationCounts: make(map[string]*uint64),
		ErrorCounts:     make(map[string]*uint64),
		StartTime:       time.Now(),
		MinLatency:      ^uint64(0),
	}

	// No requests yet
	if avg := m.GetAverageLatency(); avg != 0 {
		t.Errorf("Expected average latency = 0 with no requests, got %f", avg)
	}

	m.RecordLatency(100)
	m.RecordLatency(200)

	avg := m.GetAverageLatency()
	if avg != 150 {
		t.Errorf("Expected average latency = 150, got %f", avg)
	}
}

func TestMetrics_GetSnapshot(t *testing.T) {
	m := &Metrics{
		OperationCounts: make(map[string]*uint64),
		ErrorCounts:     make(map[string]*uint64),
		StartTime:       time.Now().Add(-time.Hour), // 1 hour ago
		MinLatency:      ^uint64(0),
	}

	m.IncrementRequests()
	m.IncrementRequests()
	m.IncrementSuccess()
	m.IncrementFailed()
	m.IncrementOperation("test")
	m.IncrementError("test_error")
	m.RecordLatency(100)

	snapshot := m.GetSnapshot()

	if snapshot.TotalRequests != 2 {
		t.Errorf("Expected TotalRequests = 2, got %d", snapshot.TotalRequests)
	}

	if snapshot.SuccessfulRequest != 1 {
		t.Errorf("Expected SuccessfulRequest = 1, got %d", snapshot.SuccessfulRequest)
	}

	if snapshot.FailedRequests != 1 {
		t.Errorf("Expected FailedRequests = 1, got %d", snapshot.FailedRequests)
	}

	if snapshot.SuccessRate != 50 {
		t.Errorf("Expected SuccessRate = 50%%, got %f%%", snapshot.SuccessRate)
	}

	if snapshot.OperationCounts["test"] != 1 {
		t.Error("Expected operation count for 'test' to be 1")
	}

	if snapshot.ErrorCounts["test_error"] != 1 {
		t.Error("Expected error count for 'test_error' to be 1")
	}

	if snapshot.AverageLatencyMs != 100 {
		t.Errorf("Expected AverageLatencyMs = 100, got %f", snapshot.AverageLatencyMs)
	}

	if snapshot.UptimeSecs < 3600 {
		t.Error("Expected uptime to be at least 1 hour")
	}

	if snapshot.Runtime.Goroutines <= 0 {
		t.Error("Expected positive number of goroutines")
	}

	if snapshot.Runtime.NumCPU <= 0 {
		t.Error("Expected positive number of CPUs")
	}

	if snapshot.Runtime.GoVersion == "" {
		t.Error("Expected Go version to be set")
	}
}

func TestMetrics_Reset(t *testing.T) {
	m := &Metrics{
		OperationCounts: make(map[string]*uint64),
		ErrorCounts:     make(map[string]*uint64),
		StartTime:       time.Now(),
		MinLatency:      ^uint64(0),
	}

	m.IncrementRequests()
	m.IncrementSuccess()
	m.IncrementFailed()
	m.IncrementRateLimitHits()
	m.IncrementOperation("test")
	m.IncrementError("error")
	m.RecordLatency(100)

	m.Reset()

	if m.TotalRequests != 0 {
		t.Error("TotalRequests should be 0 after reset")
	}
	if m.SuccessfulRequest != 0 {
		t.Error("SuccessfulRequest should be 0 after reset")
	}
	if m.FailedRequests != 0 {
		t.Error("FailedRequests should be 0 after reset")
	}
	if m.RateLimitHits != 0 {
		t.Error("RateLimitHits should be 0 after reset")
	}
	if len(m.OperationCounts) != 0 {
		t.Error("OperationCounts should be empty after reset")
	}
	if len(m.ErrorCounts) != 0 {
		t.Error("ErrorCounts should be empty after reset")
	}
	if m.TotalLatency != 0 {
		t.Error("TotalLatency should be 0 after reset")
	}
	if m.RequestCount != 0 {
		t.Error("RequestCount should be 0 after reset")
	}
	if m.MaxLatency != 0 {
		t.Error("MaxLatency should be 0 after reset")
	}
}

func TestMetricsHandler(t *testing.T) {
	// Reset global metrics for this test
	globalMetrics = &Metrics{
		OperationCounts: make(map[string]*uint64),
		ErrorCounts:     make(map[string]*uint64),
		StartTime:       time.Now(),
		MinLatency:      ^uint64(0),
	}

	handler := MetricsHandler()

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var snapshot Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snapshot); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if snapshot.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if snapshot.Runtime.GoVersion == "" {
		t.Error("Expected Go version in response")
	}
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	m := &Metrics{
		OperationCounts: make(map[string]*uint64),
		ErrorCounts:     make(map[string]*uint64),
		StartTime:       time.Now(),
		MinLatency:      ^uint64(0),
	}

	done := make(chan bool)

	// Simulate concurrent access
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				m.IncrementRequests()
				m.IncrementOperation("concurrent_op")
				m.RecordLatency(uint64(j))
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if m.TotalRequests != 1000 {
		t.Errorf("Expected TotalRequests = 1000 after concurrent access, got %d", m.TotalRequests)
	}

	if m.OperationCounts["concurrent_op"] == nil {
		t.Error("Expected concurrent_op to be tracked")
	} else if *m.OperationCounts["concurrent_op"] != 1000 {
		t.Errorf("Expected concurrent_op count = 1000, got %d", *m.OperationCounts["concurrent_op"])
	}

	if m.RequestCount != 1000 {
		t.Errorf("Expected RequestCount = 1000, got %d", m.RequestCount)
	}
}
