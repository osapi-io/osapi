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

package natsserver_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/osapi-io/osapi/internal/cli/commands/natsserver/mocks"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/suite"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/osapi-io/osapi/internal/cli/commands/natsserver"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/cryptoreport"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestStart() {
	tests := []struct {
		name           string
		metrics        bool
		badMetricsCert bool
		newErr         error
		validateFunc   func(n int, version string, err error)
	}{
		{
			name:           "when the metrics server is on and its certificate does not load, it is the error",
			metrics:        true,
			badMetricsCert: true,
			validateFunc: func(_ int, _ string, err error) {
				s.ErrorContains(err, "metrics server: tls")
			},
		},
		{
			name: "when the metrics server is off, the NATS server is the one component",
			validateFunc: func(n int, version string, err error) {
				s.Require().NoError(err)
				s.Equal(1, n)
				s.Equal("v1.2.3", version, "the heartbeat has the version")
			},
		},
		{
			name:    "when the metrics server is on, it is a component too",
			metrics: true,
			validateFunc: func(n int, _ string, err error) {
				s.Require().NoError(err)
				s.Equal(2, n)
			},
		},
		{
			name:   "when the NATS server cannot start, it is the error",
			newErr: errors.New("port in use"),
			validateFunc: func(_ int, _ string, err error) {
				s.EqualError(err, "start the NATS server: port in use")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var version string
			var report *cryptoreport.Report
			srv := mocks.NewMockServer(gomock.NewController(s.T()))
			srv.EXPECT().StartHeartbeat(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, v string, _ config.ProcessConditions, r *cryptoreport.Report) {
					version, report = v, r
				}).
				AnyTimes()
			cfg := config.Config{}
			cfg.NATS.Server.Metrics = config.MetricsServer{Enabled: tc.metrics, Host: "127.0.0.1"}
			if tc.badMetricsCert {
				cfg.NATS.Server.Metrics.TLS = config.ServerTLS{
					Enabled:  true,
					CertFile: "/no/such.crt",
					KeyFile:  "/no/such.key",
				}
			}
			d := natsserver.Deps{
				Config:  cfg,
				Logger:  slog.New(slog.DiscardHandler),
				Version: "v1.2.3",
				NewServer: func(
					context.Context,
					*slog.Logger,
					config.NATS,
					*sdkmetric.MeterProvider,
				) (natsserver.Server, error) {
					return srv, tc.newErr
				},
			}

			r, err := natsserver.Start(context.Background(), d)
			tc.validateFunc(len(r.Components), version, err)
			if err == nil {
				s.Require().NotNil(report, "the heartbeat has the crypto of the process")
				s.Equal("NATS operator, account and user JWTs", report.Algorithms[0].Name)
			}
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
