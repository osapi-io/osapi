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

// metricTimeout is the per-metric context deadline. Individual metric
// collectors that exceed this are logged and skipped so one slow NATS
// call doesn't stall the entire response.
const metricTimeout = 3 * time.Second

// GetHealthStatus returns per-component health status with system metrics (authenticated).
func (h *Health) GetHealthStatus(
	ctx context.Context,
	_ gen.GetHealthStatusRequestObject,
) (gen.GetHealthStatusResponseObject, error) {
	checker, ok := h.Checker.(*NATSChecker)
	if !ok {
		return h.buildStatusResponse(ctx, nil, nil), nil
	}

	natsErr := checker.CheckNATS()
	kvErr := checker.CheckKV()

	return h.buildStatusResponse(ctx, natsErr, kvErr), nil
}

// buildStatusResponse constructs the status response from component checks and metrics.
func (h *Health) buildStatusResponse(
	ctx context.Context,
	natsErr error,
	kvErr error,
) gen.GetHealthStatusResponseObject {
	natsComponent := gen.ComponentHealth{Status: "ok"}
	if natsErr != nil {
		errMsg := natsErr.Error()
		natsComponent = gen.ComponentHealth{Status: "error", Error: &errMsg}
	}

	kvComponent := gen.ComponentHealth{Status: "ok"}
	if kvErr != nil {
		errMsg := kvErr.Error()
		kvComponent = gen.ComponentHealth{Status: "error", Error: &errMsg}
	}

	uptime := time.Since(h.StartTime).Round(time.Second).String()

	components := map[string]gen.ComponentHealth{
		"controller.nats (connectivity)": natsComponent,
		"controller.kv (connectivity)":   kvComponent,
	}

	for name, info := range h.SubComponents {
		ch := gen.ComponentHealth{Status: info.Status}
		if info.Address != "" {
			ch.Address = &info.Address
		}

		components[name] = ch
	}

	overall := "ok"
	if natsErr != nil || kvErr != nil {
		overall = "degraded"
	}

	resp := gen.StatusResponse{
		Status:     overall,
		Components: components,
		Version:    h.Version,
		Uptime:     uptime,
	}

	if h.Metrics != nil {
		h.populateMetricsWithCache(ctx, &resp)
	}

	if overall != "ok" {
		return gen.GetHealthStatus503JSONResponse(resp)
	}

	return gen.GetHealthStatus200JSONResponse(resp)
}

// populateMetricsWithCache serves cached metrics if fresh enough,
// otherwise fetches new metrics and caches them.
func (h *Health) populateMetricsWithCache(
	ctx context.Context,
	resp *gen.StatusResponse,
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
	resp *gen.StatusResponse,
) {
	resp.Nats = cached.resp.Nats
	resp.Streams = cached.resp.Streams
	resp.KvBuckets = cached.resp.KvBuckets
	resp.ObjectStores = cached.resp.ObjectStores
	resp.Jobs = cached.resp.Jobs
	resp.Agents = cached.resp.Agents
	resp.Consumers = cached.resp.Consumers
	resp.Registry = cached.resp.Registry

	for k, v := range cached.resp.Components {
		if _, exists := resp.Components[k]; !exists {
			resp.Components[k] = v
		}
	}
}

// populateMetrics enriches the response with system metrics. All calls
// run concurrently with per-metric timeouts. If one fails, log and skip.
func (h *Health) populateMetrics(
	ctx context.Context,
	resp *gen.StatusResponse,
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
		name string,
		fn func(ctx context.Context),
	) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metricCtx, cancel := context.WithTimeout(ctx, metricTimeout)
			defer cancel()
			fn(metricCtx)
		}()
		_ = name
	}

	collect("nats", func(fnCtx context.Context) {
		info, err := h.Metrics.GetNATSInfo(fnCtx)
		if err != nil {
			h.logger.Warn("failed to get NATS info for status", slog.Any("error", err))
			return
		}
		mu.Lock()
		natsInfo = info
		mu.Unlock()
	})

	collect("streams", func(fnCtx context.Context) {
		s, err := h.Metrics.GetStreamInfo(fnCtx)
		if err != nil {
			h.logger.Warn("failed to get stream info for status", slog.Any("error", err))
			return
		}
		mu.Lock()
		streams = s
		mu.Unlock()
	})

	collect("kv", func(fnCtx context.Context) {
		b, err := h.Metrics.GetKVInfo(fnCtx)
		if err != nil {
			h.logger.Warn("failed to get KV info for status", slog.Any("error", err))
			return
		}
		mu.Lock()
		kvBuckets = b
		mu.Unlock()
	})

	collect("object-stores", func(fnCtx context.Context) {
		o, err := h.Metrics.GetObjectStoreInfo(fnCtx)
		if err != nil {
			h.logger.Warn("failed to get Object Store info for status", slog.Any("error", err))
			return
		}
		mu.Lock()
		objectStores = o
		mu.Unlock()
	})

	collect("jobs", func(fnCtx context.Context) {
		j, err := h.Metrics.GetJobStats(fnCtx)
		if err != nil {
			h.logger.Warn("failed to get job stats for status", slog.Any("error", err))
			return
		}
		mu.Lock()
		jobStats = j
		mu.Unlock()
	})

	collect("agents", func(fnCtx context.Context) {
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

	collect("registry", func(fnCtx context.Context) {
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
		resp.Nats = &gen.NATSInfo{
			Url:     natsInfo.URL,
			Version: natsInfo.Version,
		}
	}

	if streams != nil {
		streamInfos := make([]gen.StreamInfo, 0, len(streams))
		totalConsumers := 0
		for _, s := range streams {
			streamInfos = append(streamInfos, gen.StreamInfo{
				Name:      s.Name,
				Messages:  int(s.Messages),
				Bytes:     int(s.Bytes),
				Consumers: s.Consumers,
			})
			totalConsumers += s.Consumers
		}
		resp.Streams = &streamInfos

		resp.Consumers = &gen.ConsumerStats{
			Total: totalConsumers,
		}
	}

	if kvBuckets != nil {
		bucketInfos := make([]gen.KVBucketInfo, 0, len(kvBuckets))
		for _, b := range kvBuckets {
			bucketInfos = append(bucketInfos, gen.KVBucketInfo{
				Name:  b.Name,
				Keys:  b.Keys,
				Bytes: int(b.Bytes),
			})
		}
		resp.KvBuckets = &bucketInfos
	}

	if objectStores != nil {
		infos := make([]gen.ObjectStoreInfo, 0, len(objectStores))
		for _, o := range objectStores {
			infos = append(infos, gen.ObjectStoreInfo{
				Name: o.Name,
				Size: int(o.Size),
			})
		}
		resp.ObjectStores = &infos
	}

	if jobStats != nil {
		resp.Jobs = &gen.JobStats{
			Total:       jobStats.Total,
			Unprocessed: jobStats.Unprocessed,
			Processing:  jobStats.Processing,
			Completed:   jobStats.Completed,
			Failed:      jobStats.Failed,
			Dlq:         jobStats.DLQ,
		}
	}

	if agentStats != nil {
		stats := gen.AgentStats{
			Total: agentStats.Total,
			Ready: agentStats.Ready,
		}
		if len(agentStats.Agents) > 0 {
			details := make([]gen.AgentDetail, 0, len(agentStats.Agents))
			for _, a := range agentStats.Agents {
				d := gen.AgentDetail{
					Hostname:   a.Hostname,
					Registered: a.Registered,
				}
				if a.Labels != "" {
					d.Labels = &a.Labels
				}
				details = append(details, d)
			}
			stats.Agents = &details
		}
		resp.Agents = &stats
	}

	if componentEntries != nil {
		entries := make([]gen.ComponentEntry, 0, len(componentEntries))
		for _, e := range componentEntries {
			entry := gen.ComponentEntry{}
			entry.Type = &e.Type
			entry.Hostname = &e.Hostname
			entry.Status = &e.Status
			age := e.Age
			entry.Age = &age
			cpu := float32(e.CPUPercent)
			entry.CpuPercent = &cpu
			mem := e.MemBytes
			entry.MemBytes = &mem
			if len(e.Conditions) > 0 {
				conds := make([]string, len(e.Conditions))
				copy(conds, e.Conditions)
				entry.Conditions = &conds
			}
			entries = append(entries, entry)

			for name, sc := range e.SubComponents {
				ch := gen.ComponentHealth{Status: sc.Status}
				if sc.Address != "" {
					ch.Address = &sc.Address
				}
				resp.Components[name] = ch
			}
		}
		resp.Registry = &entries
	}
}
