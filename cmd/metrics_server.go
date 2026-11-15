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

package cmd

import (
	"log/slog"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/telemetry/metrics"
)

// newMetricsServer makes the metrics server that cfg turns on, without
// starting it. It returns nil when cfg turns it off.
func newMetricsServer(
	cfg config.MetricsServer,
	log *slog.Logger,
) *metrics.Server {
	if !cfg.Enabled {
		return nil
	}

	return metrics.New(cfg.Host, cfg.Port, log)
}

// meterProviderOf returns the meter provider of s, or nil when there is no
// server.
func meterProviderOf(
	s *metrics.Server,
) *sdkmetric.MeterProvider {
	if s == nil {
		return nil
	}

	return s.MeterProvider()
}
