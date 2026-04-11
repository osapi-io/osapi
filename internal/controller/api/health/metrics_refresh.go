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

	"github.com/retr0h/osapi/internal/controller/api/health/gen"
)

// StartMetricsRefresh starts a background goroutine that refreshes the
// metrics cache on a fixed interval. This ensures the health endpoint
// always serves from cache and never blocks on NATS queries.
func (h *Health) StartMetricsRefresh(
	ctx context.Context,
) {
	if h.Metrics == nil {
		return
	}

	interval := h.MetricsRefreshInterval
	if interval == 0 {
		interval = metricsCacheTTL
	}

	h.logger.Info(
		"metrics cache refresh started",
		slog.String("interval", interval.String()),
	)

	go func() {
		h.refreshMetricsCache(ctx)
		h.logger.Info(
			"metrics cache initialized",
			slog.String("next_in", interval.String()),
		)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				h.logger.Info("metrics cache refresh stopped")

				return
			case <-ticker.C:
				h.refreshMetricsCache(ctx)
				h.logger.Debug(
					"metrics cache refreshed",
					slog.String("next_in", interval.String()),
				)
			}
		}
	}()
}

// refreshMetricsCache fetches fresh metrics and updates the cache.
func (h *Health) refreshMetricsCache(
	ctx context.Context,
) {
	resp := gen.StatusResponse{
		Components: make(map[string]gen.ComponentHealth),
	}

	h.populateMetrics(ctx, &resp)

	h.metricsCacheMu.Lock()
	h.metricsCache = &cachedMetrics{
		resp:      resp,
		fetchedAt: time.Now(),
	}
	h.metricsCacheMu.Unlock()
}
