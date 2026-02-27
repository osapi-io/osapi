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
	"time"

	"github.com/retr0h/osapi/internal/api/health/gen"
)

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
		"nats": natsComponent,
		"kv":   kvComponent,
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
		h.populateMetrics(ctx, &resp)
	}

	if overall != "ok" {
		return gen.GetHealthStatus503JSONResponse(resp)
	}

	return gen.GetHealthStatus200JSONResponse(resp)
}

// populateMetrics enriches the response with system metrics. Each call is
// independent — if one fails, log and skip it (graceful degradation).
func (h *Health) populateMetrics(
	ctx context.Context,
	resp *gen.StatusResponse,
) {
	if natsInfo, err := h.Metrics.GetNATSInfo(ctx); err != nil {
		h.logger.Warn("failed to get NATS info for status", "error", err)
	} else {
		resp.Nats = &gen.NATSInfo{
			Url:     natsInfo.URL,
			Version: natsInfo.Version,
		}
	}

	if streams, err := h.Metrics.GetStreamInfo(ctx); err != nil {
		h.logger.Warn("failed to get stream info for status", "error", err)
	} else {
		streamInfos := make([]gen.StreamInfo, 0, len(streams))
		for _, s := range streams {
			streamInfos = append(streamInfos, gen.StreamInfo{
				Name:      s.Name,
				Messages:  int(s.Messages),
				Bytes:     int(s.Bytes),
				Consumers: s.Consumers,
			})
		}
		resp.Streams = &streamInfos
	}

	if kvBuckets, err := h.Metrics.GetKVInfo(ctx); err != nil {
		h.logger.Warn("failed to get KV info for status", "error", err)
	} else {
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

	if jobStats, err := h.Metrics.GetJobStats(ctx); err != nil {
		h.logger.Warn("failed to get job stats for status", "error", err)
	} else {
		resp.Jobs = &gen.JobStats{
			Total:       jobStats.Total,
			Unprocessed: jobStats.Unprocessed,
			Processing:  jobStats.Processing,
			Completed:   jobStats.Completed,
			Failed:      jobStats.Failed,
			Dlq:         jobStats.DLQ,
		}
	}

	if agentStats, err := h.Metrics.GetAgentStats(ctx); err != nil {
		h.logger.Warn("failed to get agent stats for status", "error", err)
	} else {
		resp.Agents = &gen.AgentStats{
			Total: agentStats.Total,
			Ready: agentStats.Ready,
		}
	}

	if consumerStats, err := h.Metrics.GetConsumerStats(ctx); err != nil {
		h.logger.Warn("failed to get consumer stats for status", "error", err)
	} else {
		resp.Consumers = &gen.ConsumerStats{
			Total: consumerStats.Total,
		}
	}
}
