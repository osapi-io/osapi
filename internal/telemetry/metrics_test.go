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
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/exporters/prometheus"

	"github.com/retr0h/osapi/internal/config"
)

type InitMeterTestSuite struct {
	suite.Suite
}

func (s *InitMeterTestSuite) TestInitMeter() {
	tests := []struct {
		name         string
		cfg          config.MetricsConfig
		setupFn      func()
		cleanupFn    func()
		validateFunc func(http.Handler, string, func(context.Context) error, error)
	}{
		{
			name: "when path is empty uses default /metrics",
			cfg:  config.MetricsConfig{},
			validateFunc: func(
				handler http.Handler,
				path string,
				shutdown func(context.Context) error,
				err error,
			) {
				s.NoError(err)
				s.NotNil(handler)
				s.Equal(DefaultMetricsPath, path)
				s.NotNil(shutdown)
				s.NoError(shutdown(context.Background()))
			},
		},
		{
			name: "when path is configured uses custom path",
			cfg:  config.MetricsConfig{Path: "/custom/metrics"},
			validateFunc: func(
				handler http.Handler,
				path string,
				shutdown func(context.Context) error,
				err error,
			) {
				s.NoError(err)
				s.NotNil(handler)
				s.Equal("/custom/metrics", path)
				s.NotNil(shutdown)
				s.NoError(shutdown(context.Background()))
			},
		},
		{
			name: "when prometheus exporter creation fails returns error",
			cfg:  config.MetricsConfig{},
			setupFn: func() {
				prometheusNewFn = func(
					_ ...prometheus.Option,
				) (*prometheus.Exporter, error) {
					return nil, errors.New("prometheus exporter failed")
				}
			},
			validateFunc: func(
				handler http.Handler,
				path string,
				shutdown func(context.Context) error,
				err error,
			) {
				s.Error(err)
				s.Nil(handler)
				s.Empty(path)
				s.Nil(shutdown)
				s.Contains(err.Error(), "creating prometheus exporter")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			original := prometheusNewFn
			defer func() { prometheusNewFn = original }()

			if tc.setupFn != nil {
				tc.setupFn()
			}

			handler, path, shutdown, err := InitMeter(tc.cfg)
			tc.validateFunc(handler, path, shutdown, err)
		})
	}
}

func TestInitMeterTestSuite(t *testing.T) {
	suite.Run(t, new(InitMeterTestSuite))
}
