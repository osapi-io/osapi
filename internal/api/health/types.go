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
	"time"
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
	GetConsumerStats(ctx context.Context) (*ConsumerMetrics, error)
	GetJobStats(ctx context.Context) (*JobMetrics, error)
	GetAgentStats(ctx context.Context) (*AgentMetrics, error)
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

// ConsumerMetrics holds JetStream consumer statistics.
type ConsumerMetrics struct {
	Total int
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
	Total int
	Ready int
}

// ClosureMetricsProvider implements MetricsProvider using function closures.
type ClosureMetricsProvider struct {
	NATSInfoFn      func(ctx context.Context) (*NATSMetrics, error)
	StreamInfoFn    func(ctx context.Context) ([]StreamMetrics, error)
	KVInfoFn        func(ctx context.Context) ([]KVMetrics, error)
	ConsumerStatsFn func(ctx context.Context) (*ConsumerMetrics, error)
	JobStatsFn      func(ctx context.Context) (*JobMetrics, error)
	AgentStatsFn    func(ctx context.Context) (*AgentMetrics, error)
}

// GetNATSInfo delegates to the NATSInfoFn closure.
func (p *ClosureMetricsProvider) GetNATSInfo(
	ctx context.Context,
) (*NATSMetrics, error) {
	return p.NATSInfoFn(ctx)
}

// GetStreamInfo delegates to the StreamInfoFn closure.
func (p *ClosureMetricsProvider) GetStreamInfo(
	ctx context.Context,
) ([]StreamMetrics, error) {
	return p.StreamInfoFn(ctx)
}

// GetKVInfo delegates to the KVInfoFn closure.
func (p *ClosureMetricsProvider) GetKVInfo(
	ctx context.Context,
) ([]KVMetrics, error) {
	return p.KVInfoFn(ctx)
}

// GetConsumerStats delegates to the ConsumerStatsFn closure.
func (p *ClosureMetricsProvider) GetConsumerStats(
	ctx context.Context,
) (*ConsumerMetrics, error) {
	return p.ConsumerStatsFn(ctx)
}

// GetJobStats delegates to the JobStatsFn closure.
func (p *ClosureMetricsProvider) GetJobStats(
	ctx context.Context,
) (*JobMetrics, error) {
	return p.JobStatsFn(ctx)
}

// GetAgentStats delegates to the AgentStatsFn closure.
func (p *ClosureMetricsProvider) GetAgentStats(
	ctx context.Context,
) (*AgentMetrics, error) {
	return p.AgentStatsFn(ctx)
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
	logger  *slog.Logger
}
