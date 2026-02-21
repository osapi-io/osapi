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

package telemetry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/telemetry"
)

type InitTracerPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *InitTracerPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *InitTracerPublicTestSuite) TestInitTracer() {
	tests := []struct {
		name         string
		cfg          config.TracingConfig
		expectErr    bool
		errContains  string
		validateFunc func()
	}{
		{
			name: "when disabled returns noop provider",
			cfg: config.TracingConfig{
				Enabled: false,
			},
			expectErr: false,
			validateFunc: func() {
				// Noop provider should be set - creating a span should not panic
				_, span := otel.Tracer("test").Start(s.ctx, "test-span")
				defer span.End()
				s.False(span.SpanContext().IsValid())
			},
		},
		{
			name: "when stdout exporter configured",
			cfg: config.TracingConfig{
				Enabled:  true,
				Exporter: "stdout",
			},
			expectErr: false,
			validateFunc: func() {
				_, span := otel.Tracer("test").Start(s.ctx, "test-span")
				defer span.End()
				s.True(span.SpanContext().IsValid())
			},
		},
		{
			name: "when unsupported exporter returns error",
			cfg: config.TracingConfig{
				Enabled:  true,
				Exporter: "invalid",
			},
			expectErr:   true,
			errContains: "unsupported tracing exporter",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			shutdown, err := telemetry.InitTracer(s.ctx, "test-service", tc.cfg)
			if tc.expectErr {
				s.Error(err)
				s.Contains(err.Error(), tc.errContains)

				return
			}

			s.NoError(err)
			s.NotNil(shutdown)

			if tc.validateFunc != nil {
				tc.validateFunc()
			}

			s.NoError(shutdown(s.ctx))
		})
	}
}

func TestInitTracerPublicTestSuite(t *testing.T) {
	suite.Run(t, new(InitTracerPublicTestSuite))
}
