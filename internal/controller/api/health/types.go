// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package health

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/retr0h/osapi/internal/controller/api/health/gen"
)

// Checker checks the health of a dependency.
type Checker interface {
	CheckHealth(ctx context.Context) error
}

// MetricsProvider retrieves system metrics for the status endpoint.
type MetricsProvider interface {
	GetNATSInfo(ctx context.Context) (*NATSMetrics, error)
	GetStreamInfo(ctx context.Context) ([]StreamMetrics, error)
	GetKVInfo(ctx context.Context) ([]KVMetrics, error)
	GetObjectStoreInfo(ctx context.Context) ([]ObjectStoreMetrics, error)
	GetJobStats(ctx context.Context) (*JobMetrics, error)
	GetAgentStats(ctx context.Context) (*AgentMetrics, error)
	GetComponentRegistry(ctx context.Context) ([]ComponentEntry, error)
}

// NATSMetrics holds NATS connection information.
type NATSMetrics struct {
	URL     string
	Version string
}

// StreamMetrics holds JetStream stream statistics.
type StreamMetrics struct {
	Name      string
	Messages  uint64
	Bytes     uint64
	Consumers int
}

// KVMetrics holds KV bucket statistics.
type KVMetrics struct {
	Name  string
	Keys  int
	Bytes uint64
}

// ObjectStoreMetrics holds Object Store bucket statistics.
type ObjectStoreMetrics struct {
	Name string
	Size uint64
}

// ConsumerMetrics holds JetStream consumer statistics.
type ConsumerMetrics struct {
	Total     int
	Consumers []ConsumerDetail
}

// ConsumerDetail holds per-consumer information.
type ConsumerDetail struct {
	Name        string
	Pending     uint64
	AckPending  int
	Redelivered int
}

// JobMetrics holds job queue statistics.
type JobMetrics struct {
	Total       int
	Unprocessed int
	Processing  int
	Completed   int
	Failed      int
	DLQ         int
}

// AgentMetrics holds agent fleet statistics.
type AgentMetrics struct {
	Total  int
	Ready  int
	Agents []AgentDetail
}

// AgentDetail holds per-agent registration info.
type AgentDetail struct {
	Hostname   string
	Labels     string
	Registered string
}

// ComponentEntry holds unified component registration details for the registry.
type ComponentEntry struct {
	Type          string
	Hostname      string
	Status        string
	Conditions    []string
	Age           string
	CPUPercent    float64
	MemBytes      int64
	SubComponents map[string]SubComponentInfo
}

// ClosureMetricsProvider implements MetricsProvider using function closures.
type ClosureMetricsProvider struct {
	NATSInfoFn          func(ctx context.Context) (*NATSMetrics, error)
	StreamInfoFn        func(ctx context.Context) ([]StreamMetrics, error)
	KVInfoFn            func(ctx context.Context) ([]KVMetrics, error)
	ObjectStoreInfoFn   func(ctx context.Context) ([]ObjectStoreMetrics, error)
	JobStatsFn          func(ctx context.Context) (*JobMetrics, error)
	AgentStatsFn        func(ctx context.Context) (*AgentMetrics, error)
	ComponentRegistryFn func(ctx context.Context) ([]ComponentEntry, error)
}

// SubComponentInfo holds the status and optional address of a sub-component.
type SubComponentInfo struct {
	Status  string
	Address string // Empty means no network endpoint.
}

// Health implementation of the Health APIs operations.
type Health struct {
	// Checker performs dependency health checks.
	Checker Checker
	// StartTime records when the server started.
	StartTime time.Time
	// Version is the application version string.
	Version string
	// Metrics provides system metrics (optional, can be nil).
	Metrics MetricsProvider
	// SubComponents reports the status of internal services.
	SubComponents map[string]SubComponentInfo
	logger        *slog.Logger

	// metricsCache holds the last populateMetrics result to avoid
	// querying NATS on every poll cycle.
	metricsCache   *cachedMetrics
	metricsCacheMu sync.RWMutex
}

// cachedMetrics holds a snapshot of metrics with a timestamp for TTL.
type cachedMetrics struct {
	resp      gen.StatusResponse
	fetchedAt time.Time
}

// metricsCacheTTL is how long cached metrics are served before refreshing.
const metricsCacheTTL = 15 * time.Second
