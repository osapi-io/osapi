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
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	api "github.com/osapi-io/osapi/internal/controller/endpoints/health"
)

// metricTimeout is the per-metric context deadline. Individual metric
// collectors that exceed this are logged and skipped so one slow NATS
// call doesn't stall the entire response.
const metricTimeout = 3 * time.Second

// Ready checks the controller's dependencies: nil when it can take
// traffic.
func (h *Health) Ready(
	ctx context.Context,
) error {
	return h.Checker.CheckHealth(ctx)
}

// Status is the health of each component, with the metrics of the system.
// Its status is degraded when a component is not healthy.
func (h *Health) Status(
	ctx context.Context,
) api.StatusResponse {
	return h.buildStatusResponse(ctx, h.Checker.CheckNATS(ctx), h.Checker.CheckKV(ctx))
}

// buildStatusResponse constructs the status response from component checks and metrics.
func (h *Health) buildStatusResponse(
	ctx context.Context,
	natsErr error,
	kvErr error,
) api.StatusResponse {
	uptime := time.Since(h.StartTime).Round(time.Second).String()

	overall := "ok"
	if natsErr != nil || kvErr != nil {
		overall = "degraded"
	}

	resp := api.StatusResponse{
		Status:  overall,
		Version: h.Version,
		Uptime:  uptime,
		Checks:  []api.Check{check("nats", natsErr), check("kv", kvErr)},
	}

	if h.Metrics != nil {
		h.populateMetricsWithCache(ctx, &resp)
	}

	return resp
}

// populateMetricsWithCache serves cached metrics if fresh enough,
// otherwise fetches new metrics and caches them.
func (h *Health) populateMetricsWithCache(
	ctx context.Context,
	resp *api.StatusResponse,
) {
	h.metricsCacheMu.RLock()
	cached := h.metricsCache
	h.metricsCacheMu.RUnlock()

	if cached != nil && time.Since(cached.fetchedAt) < metricsCacheTTL {
		copyMetricsFromCache(cached, resp)

		return
	}

	h.populateMetrics(ctx, resp)

	h.metricsCacheMu.Lock()
	h.metricsCache = &cachedMetrics{
		resp:      *resp,
		fetchedAt: time.Now(),
	}
	h.metricsCacheMu.Unlock()
}

// copyMetricsFromCache copies metric fields from a cached response.
// Component-level fields (Status, Uptime, Version) are always fresh.
func copyMetricsFromCache(
	cached *cachedMetrics,
	resp *api.StatusResponse,
) {
	resp.NATS = cached.resp.NATS
	resp.Streams = cached.resp.Streams
	resp.KVBuckets = cached.resp.KVBuckets
	resp.ObjectStores = cached.resp.ObjectStores
	resp.Jobs = cached.resp.Jobs
	resp.Agents = cached.resp.Agents
	resp.Consumers = cached.resp.Consumers
	resp.Registry = cached.resp.Registry
}

// check is the result of one check of the controller: ok, or error with
// why.
func check(
	name string,
	err error,
) api.Check {
	if err != nil {
		return api.Check{Name: name, Status: "error", Error: err.Error()}
	}

	return api.Check{Name: name, Status: "ok"}
}

// populateMetrics enriches the response with system metrics. All calls
// run concurrently with per-metric timeouts. If one fails, log and skip.
func (h *Health) populateMetrics(
	ctx context.Context,
	resp *api.StatusResponse,
) {
	var (
		mu               sync.Mutex
		wg               sync.WaitGroup
		natsInfo         *NATSMetrics
		streams          []StreamMetrics
		kvBuckets        []KVMetrics
		objectStores     []ObjectStoreMetrics
		jobStats         *JobMetrics
		agentStats       *AgentMetrics
		componentEntries []ComponentEntry
	)

	collect := func(
		fn func(ctx context.Context),
	) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metricCtx, cancel := context.WithTimeout(ctx, metricTimeout)
			defer cancel()
			fn(metricCtx)
		}()
	}

	collect(func(fnCtx context.Context) {
		info, err := h.Metrics.GetNATSInfo(fnCtx)
		if err != nil {
			h.logger.Warn("failed to get NATS info for status", slog.Any("error", err))
			return
		}
		mu.Lock()
		natsInfo = info
		mu.Unlock()
	})

	collect(func(fnCtx context.Context) {
		s, err := h.Metrics.GetStreamInfo(fnCtx)
		if err != nil {
			h.logger.Warn("failed to get stream info for status", slog.Any("error", err))
			return
		}
		mu.Lock()
		streams = s
		mu.Unlock()
	})

	collect(func(fnCtx context.Context) {
		b, err := h.Metrics.GetKVInfo(fnCtx)
		if err != nil {
			h.logger.Warn("failed to get KV info for status", slog.Any("error", err))
			return
		}
		mu.Lock()
		kvBuckets = b
		mu.Unlock()
	})

	collect(func(fnCtx context.Context) {
		o, err := h.Metrics.GetObjectStoreInfo(fnCtx)
		if err != nil {
			h.logger.Warn("failed to get Object Store info for status", slog.Any("error", err))
			return
		}
		mu.Lock()
		objectStores = o
		mu.Unlock()
	})

	collect(func(fnCtx context.Context) {
		j, err := h.Metrics.GetJobStats(fnCtx)
		if err != nil {
			h.logger.Warn("failed to get job stats for status", slog.Any("error", err))
			return
		}
		mu.Lock()
		jobStats = j
		mu.Unlock()
	})

	collect(func(fnCtx context.Context) {
		a, err := h.Metrics.GetAgentStats(fnCtx)
		if err != nil {
			h.logger.Warn("failed to get agent stats for status", slog.Any("error", err))
			return
		}
		mu.Lock()
		agentStats = a
		mu.Unlock()
	})

	// Consumer count is derived from stream info (Consumers field)
	// instead of enumerating individual consumers via ListConsumers.
	// This avoids N sequential NATS API calls per consumer.

	collect(func(fnCtx context.Context) {
		entries, err := h.Metrics.GetComponentRegistry(fnCtx)
		if err != nil {
			h.logger.Warn("failed to get component registry for status", slog.Any("error", err))
			return
		}
		mu.Lock()
		componentEntries = entries
		mu.Unlock()
	})

	wg.Wait()

	if natsInfo != nil {
		resp.NATS = &api.NATSInfo{URL: natsInfo.URL, Version: natsInfo.Version}
	}

	if streams != nil {
		resp.Streams = make([]api.StreamInfo, 0, len(streams))
		totalConsumers := 0
		for _, s := range streams {
			resp.Streams = append(resp.Streams, api.StreamInfo{
				Name:      s.Name,
				Messages:  int(s.Messages), //nolint:gosec // a stream's count fits an int
				Bytes:     int(s.Bytes),    //nolint:gosec // a stream's size fits an int
				Consumers: s.Consumers,
			})
			totalConsumers += s.Consumers
		}
		resp.Consumers = totalConsumers
	}

	if kvBuckets != nil {
		resp.KVBuckets = make([]api.KVBucketInfo, 0, len(kvBuckets))
		for _, b := range kvBuckets {
			resp.KVBuckets = append(resp.KVBuckets, api.KVBucketInfo{
				Name:  b.Name,
				Keys:  b.Keys,
				Bytes: int(b.Bytes), //nolint:gosec // a bucket's size fits an int
			})
		}
	}

	if objectStores != nil {
		resp.ObjectStores = make([]api.ObjectStoreInfo, 0, len(objectStores))
		for _, o := range objectStores {
			resp.ObjectStores = append(resp.ObjectStores, api.ObjectStoreInfo{
				Name: o.Name,
				Size: int(o.Size), //nolint:gosec // a store's size fits an int
			})
		}
	}

	if jobStats != nil {
		resp.Jobs = &api.JobStats{
			Total:      jobStats.Total,
			Submitted:  jobStats.Submitted,
			Processing: jobStats.Processing,
			Completed:  jobStats.Completed,
			Failed:     jobStats.Failed,
			DLQ:        jobStats.DLQ,
		}
	}

	if agentStats != nil {
		stats := api.AgentStats{Total: agentStats.Total, Ready: agentStats.Ready}
		for _, a := range agentStats.Agents {
			stats.Agents = append(stats.Agents, api.AgentDetail{
				Hostname:   a.Hostname,
				Labels:     a.Labels,
				Registered: a.Registered,
			})
		}
		resp.Agents = &stats
	}

	if componentEntries != nil {
		resp.Registry = make([]api.ComponentEntry, 0, len(componentEntries))
		for _, e := range componentEntries {
			entry := api.ComponentEntry{
				Type:       e.Type,
				Hostname:   e.Hostname,
				Status:     e.Status,
				Age:        e.Age,
				CPUPercent: float32(e.CPUPercent),
				MemBytes:   e.MemBytes,
				Version:    e.Version,
				Crypto:     e.Crypto,
			}
			if e.Crypto != nil {
				entry.FIPS = e.Crypto.Verdict()
			}
			if len(e.Conditions) > 0 {
				entry.Conditions = append([]types.ConditionType(nil), e.Conditions...)
			}
			// A sub-component is named for its component, such as
			// controller.api; under its component it is api.
			for _, name := range slices.Sorted(maps.Keys(e.SubComponents)) {
				sc := e.SubComponents[name]
				entry.Components = append(entry.Components, api.ComponentEntry{
					Type:    strings.TrimPrefix(name, e.Type+"."),
					Status:  sc.Status,
					Address: sc.Address,
					URL:     sc.URL,
				})
			}
			resp.Registry = append(resp.Registry, entry)
		}
	}
}
