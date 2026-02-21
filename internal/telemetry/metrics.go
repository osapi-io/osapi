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

package telemetry

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/retr0h/osapi/internal/config"
)

// prometheusNewFn is the function used to create Prometheus exporters.
// It is a package-level variable so tests can replace it to simulate errors.
var prometheusNewFn = prometheus.New

// DefaultMetricsPath is the default HTTP path for the Prometheus scrape endpoint.
const DefaultMetricsPath = "/metrics"

// InitMeter initializes the OpenTelemetry meter provider with a Prometheus exporter.
// It returns the HTTP handler for the scrape endpoint, the resolved path,
// a shutdown function, and any initialization error.
func InitMeter(
	cfg config.MetricsConfig,
) (http.Handler, string, func(context.Context) error, error) {
	path := cfg.Path
	if path == "" {
		path = DefaultMetricsPath
	}

	exporter, err := prometheusNewFn()
	if err != nil {
		return nil, "", nil, fmt.Errorf("creating prometheus exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(mp)

	return promhttp.Handler(), path, mp.Shutdown, nil
}
