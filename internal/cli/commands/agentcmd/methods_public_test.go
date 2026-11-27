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

package agentcmd_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/osapi-io/osapi/internal/cli/commands/agentcmd/mocks"
	"go.uber.org/mock/gomock"

	"github.com/avfs/avfs"
	"github.com/stretchr/testify/suite"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/osapi-io/osapi/internal/cli/commands/agentcmd"
	"github.com/osapi-io/osapi/internal/config"
)

// result is what Start made, and what its cleanup did.
type result struct {
	leveled bool
	n       int
	version string
	metered bool
	err     error
	cleanup func()
	closed  *bool
	stopped *bool
}

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestStart() {
	tests := []struct {
		name           string
		metrics        bool
		badMetricsCert bool
		level          bool
		hostErr        error
		tracerErr      error
		agentErr       error
		validateFunc   func(got result)
	}{
		{
			name: "when it starts, the agent is the one component, and the cleanup closes, then stops the tracer",
			validateFunc: func(got result) {
				s.Require().NoError(got.err)
				s.Equal(1, got.n)
				s.Equal("v1.2.3", got.version)
				s.False(got.metered)
				got.cleanup()
				s.True(*got.closed)
				s.True(*got.stopped)
			},
		},
		{
			name:  "when it has the logger's level, the agent takes it",
			level: true,
			validateFunc: func(got result) {
				s.Require().NoError(got.err)
				s.True(got.leveled)
				got.cleanup()
			},
		},
		{
			name:           "when the metrics server is on and its certificate does not load, it is the error",
			metrics:        true,
			badMetricsCert: true,
			validateFunc: func(got result) {
				s.ErrorContains(got.err, "metrics server: tls")
			},
		},
		{
			name:    "when the metrics server is on, it is a component, and the agent meters on it",
			metrics: true,
			validateFunc: func(got result) {
				s.Require().NoError(got.err)
				s.Equal(2, got.n)
				s.True(got.metered)
			},
		},
		{
			name:    "when the host is not supported, it is the error",
			hostErr: errors.New("plan9"),
			validateFunc: func(got result) {
				s.EqualError(got.err, "unsupported host: plan9")
			},
		},
		{
			name:      "when the tracer cannot start, it is the error",
			tracerErr: errors.New("no exporter"),
			validateFunc: func(got result) {
				s.EqualError(got.err, "start the tracer: no exporter")
			},
		},
		{
			name:     "when the agent cannot start, it is the error",
			agentErr: errors.New("no NATS"),
			validateFunc: func(got result) {
				s.EqualError(got.err, "start the agent: no NATS")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var version string
			var metered, leveled bool
			a := mocks.NewMockAgent(gomock.NewController(s.T()))
			a.EXPECT().SetVersion(gomock.Any()).Do(func(v string) { version = v }).AnyTimes()
			a.EXPECT().SubComponents().Return(nil).AnyTimes()
			a.EXPECT().SetLogLevel(gomock.Any()).
				Do(func(*slog.LevelVar) { leveled = true }).AnyTimes()
			a.EXPECT().IsReady().Return(nil).AnyTimes()
			a.EXPECT().LastHeartbeatTime().Return(time.Time{}).AnyTimes()
			a.EXPECT().SetMeterProvider(gomock.Any()).
				Do(func(*sdkmetric.MeterProvider) { metered = true }).AnyTimes()
			var closed, stopped bool
			cfg := config.Config{}
			cfg.Agent.Metrics = config.MetricsServer{Enabled: tc.metrics, Host: "127.0.0.1"}
			if tc.badMetricsCert {
				cfg.Agent.Metrics.TLS = config.ServerTLS{
					Enabled:  true,
					CertFile: "/no/such.crt",
					KeyFile:  "/no/such.key",
				}
			}
			d := agentcmd.Deps{
				Config:    cfg,
				Logger:    slog.New(slog.DiscardHandler),
				Version:   "v1.2.3",
				CheckHost: func() error { return tc.hostErr },
				Tracer: func(context.Context, string) (func(context.Context) error, error) {
					return func(context.Context) error { stopped = true; return nil }, tc.tracerErr
				},
				NewAgent: func(
					context.Context, *slog.Logger, avfs.VFS, config.Config, *sdkmetric.MeterProvider,
				) (agentcmd.Agent, func(), error) {
					if tc.agentErr != nil {
						return nil, nil, tc.agentErr
					}

					return a, func() { closed = true }, nil
				},
			}

			if tc.level {
				d.LogLevel = new(slog.LevelVar)
			}
			r, err := agentcmd.Start(context.Background(), d)
			tc.validateFunc(result{
				n:       len(r.Components),
				version: version,
				metered: metered,
				leveled: leveled,
				err:     err,
				cleanup: func() { r.Cleanup(context.Background()) },
				closed:  &closed,
				stopped: &stopped,
			})
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
