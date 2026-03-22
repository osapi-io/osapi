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

// Package ops provides a lightweight HTTP server for per-component
// Prometheus metrics with isolated registries.
package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	prometheusExporter "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// prometheusNewFn is injectable for testing the exporter creation error path.
var prometheusNewFn = prometheusExporter.New

// New creates a new ops server on the given port with an isolated
// Prometheus registry and OTEL MeterProvider.
func New(
	port int,
	logger *slog.Logger,
) *Server {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(
		collectors.ProcessCollectorOpts{},
	))

	exporter, err := prometheusNewFn(
		prometheusExporter.WithRegisterer(reg),
	)
	if err != nil {
		logger.Error("failed to create prometheus exporter", "error", err)

		return nil
	}

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{Registry: reg},
	))

	return &Server{
		httpServer: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		},
		logger:        logger,
		registry:      reg,
		meterProvider: mp,
	}
}

// MeterProvider returns the isolated OTEL MeterProvider for this server.
// Components use this to create instruments that appear on this server's
// /metrics endpoint.
func (s *Server) MeterProvider() *sdkmetric.MeterProvider {
	return s.meterProvider
}

// Registry returns the isolated Prometheus registry for this server.
func (s *Server) Registry() *prometheus.Registry {
	return s.registry
}

// Start starts the HTTP server in a background goroutine.
func (s *Server) Start() {
	go func() {
		s.logger.Info("ops server started", "addr", s.httpServer.Addr)

		if err := s.httpServer.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			s.logger.Error("ops server error", "error", err)
		}
	}()
}

// Stop gracefully shuts down the HTTP server and meter provider.
func (s *Server) Stop(ctx context.Context) {
	if err := s.meterProvider.Shutdown(ctx); err != nil {
		s.logger.Error("meter provider shutdown error", "error", err)
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("ops server shutdown error", "error", err)
	}

	s.logger.Info("ops server stopped")
}
