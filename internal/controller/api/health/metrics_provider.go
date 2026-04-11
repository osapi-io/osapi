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

import "context"

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

// GetObjectStoreInfo delegates to the ObjectStoreInfoFn closure.
func (p *ClosureMetricsProvider) GetObjectStoreInfo(
	ctx context.Context,
) ([]ObjectStoreMetrics, error) {
	return p.ObjectStoreInfoFn(ctx)
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

// GetComponentRegistry delegates to the ComponentRegistryFn closure.
// Returns nil, nil when the closure is not configured.
func (p *ClosureMetricsProvider) GetComponentRegistry(
	ctx context.Context,
) ([]ComponentEntry, error) {
	if p.ComponentRegistryFn == nil {
		return nil, nil
	}
	return p.ComponentRegistryFn(ctx)
}
