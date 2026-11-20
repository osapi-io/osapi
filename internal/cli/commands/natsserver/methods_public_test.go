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

	"github.com/stretchr/testify/suite"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/osapi-io/osapi/internal/cli/commands/natsserver"
	"github.com/osapi-io/osapi/internal/config"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

// fakeServer records the heartbeat that Start began.
type fakeServer struct {
	version string
}

func (*fakeServer) Start()               {}
func (*fakeServer) Stop(context.Context) {}
func (f *fakeServer) StartHeartbeat(_ context.Context, v string, _ config.ProcessConditions) {
	f.version = v
}

func (s *MethodsPublicTestSuite) TestStart() {
	tests := []struct {
		name         string
		metrics      bool
		newErr       error
		validateFunc func(n int, srv *fakeServer, err error)
	}{
		{
			name: "when the metrics server is off, the NATS server is the one component",
			validateFunc: func(n int, srv *fakeServer, err error) {
				s.Require().NoError(err)
				s.Equal(1, n)
				s.Equal("v1.2.3", srv.version, "the heartbeat has the version")
			},
		},
		{
			name:    "when the metrics server is on, it is a component too",
			metrics: true,
			validateFunc: func(n int, _ *fakeServer, err error) {
				s.Require().NoError(err)
				s.Equal(2, n)
			},
		},
		{
			name:   "when the NATS server cannot start, it is the error",
			newErr: errors.New("port in use"),
			validateFunc: func(_ int, _ *fakeServer, err error) {
				s.EqualError(err, "start the NATS server: port in use")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			srv := &fakeServer{}
			cfg := config.Config{}
			cfg.NATS.Server.Metrics = config.MetricsServer{Enabled: tc.metrics, Host: "127.0.0.1"}
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
			tc.validateFunc(len(r.Components), srv, err)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
