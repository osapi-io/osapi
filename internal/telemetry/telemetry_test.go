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
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"

	"github.com/retr0h/osapi/internal/config"
)

// noopClient implements otlptrace.Client for testing.
type noopClient struct{}

func (noopClient) Start(_ context.Context) error { return nil }
func (noopClient) Stop(_ context.Context) error  { return nil }
func (noopClient) UploadTraces(
	_ context.Context,
	_ []*tracepb.ResourceSpans,
) error {
	return nil
}

type InitTracerTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *InitTracerTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *InitTracerTestSuite) TestInitTracerResourceError() {
	tests := []struct {
		name string
	}{
		{
			name: "when resource creation fails returns error",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			original := resourceNewFn
			defer func() { resourceNewFn = original }()

			resourceNewFn = func(
				_ context.Context,
				_ ...resource.Option,
			) (*resource.Resource, error) {
				return nil, errors.New("resource creation failed")
			}

			cfg := config.TracingConfig{
				Enabled: true,
			}

			shutdown, err := InitTracer(s.ctx, "test-service", cfg)

			s.Error(err)
			s.Nil(shutdown)
			s.Contains(err.Error(), "creating resource")
		})
	}
}

func (s *InitTracerTestSuite) TestInitTracerStdoutExporter() {
	tests := []struct {
		name         string
		stubFn       func(...stdouttrace.Option) (*stdouttrace.Exporter, error)
		validateFunc func(func(context.Context) error, error)
	}{
		{
			name: "when stdout exporter creation fails returns error",
			stubFn: func(_ ...stdouttrace.Option) (*stdouttrace.Exporter, error) {
				return nil, errors.New("stdout exporter failed")
			},
			validateFunc: func(shutdown func(context.Context) error, err error) {
				s.Error(err)
				s.Nil(shutdown)
				s.Contains(err.Error(), "creating stdout exporter")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			original := stdouttraceNewFn
			defer func() { stdouttraceNewFn = original }()

			stdouttraceNewFn = tc.stubFn

			cfg := config.TracingConfig{
				Enabled:  true,
				Exporter: "stdout",
			}

			shutdown, err := InitTracer(s.ctx, "test-service", cfg)
			tc.validateFunc(shutdown, err)
		})
	}
}

func (s *InitTracerTestSuite) TestInitTracerOTLPExporter() {
	tests := []struct {
		name         string
		stubFn       func(context.Context, ...otlptracegrpc.Option) (*otlptrace.Exporter, error)
		validateFunc func(func(context.Context) error, error)
	}{
		{
			name: "when OTLP exporter configured creates valid provider",
			stubFn: func(_ context.Context, _ ...otlptracegrpc.Option) (*otlptrace.Exporter, error) {
				return otlptrace.NewUnstarted(noopClient{}), nil
			},
			validateFunc: func(shutdown func(context.Context) error, err error) {
				s.NoError(err)
				s.NotNil(shutdown)

				_, span := otel.Tracer("test").Start(s.ctx, "test-span")
				defer span.End()
				s.True(span.SpanContext().IsValid())

				s.NoError(shutdown(s.ctx))
			},
		},
		{
			name: "when OTLP exporter creation fails returns error",
			stubFn: func(_ context.Context, _ ...otlptracegrpc.Option) (*otlptrace.Exporter, error) {
				return nil, errors.New("otlp exporter failed")
			},
			validateFunc: func(shutdown func(context.Context) error, err error) {
				s.Error(err)
				s.Nil(shutdown)
				s.Contains(err.Error(), "creating OTLP exporter")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			original := otlptraceNewFn
			defer func() { otlptraceNewFn = original }()

			otlptraceNewFn = tc.stubFn

			cfg := config.TracingConfig{
				Enabled:      true,
				Exporter:     "otlp",
				OTLPEndpoint: "localhost:4317",
			}

			shutdown, err := InitTracer(s.ctx, "test-service", cfg)
			tc.validateFunc(shutdown, err)
		})
	}
}

func TestInitTracerTestSuite(t *testing.T) {
	suite.Run(t, new(InitTracerTestSuite))
}
