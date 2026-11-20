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

	"github.com/avfs/avfs"
	"github.com/stretchr/testify/suite"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/osapi-io/osapi/internal/cli/commands/agentcmd"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
)

// result is what Start made, and what its cleanup did.
type result struct {
	n       int
	agent   *fakeAgent
	err     error
	cleanup func()
	closed  *bool
	stopped *bool
}

type MethodsPublicTestSuite struct {
	suite.Suite
}

// fakeAgent records what Components set on it.
type fakeAgent struct {
	version string
	metered bool
}

func (*fakeAgent) Start()                       {}
func (*fakeAgent) Stop(context.Context)         {}
func (*fakeAgent) IsReady() error               { return nil }
func (*fakeAgent) LastHeartbeatTime() time.Time { return time.Time{} }
func (*fakeAgent) SubComponents() map[string]job.SubComponentInfo {
	return nil
}
func (f *fakeAgent) SetVersion(v string)                       { f.version = v }
func (f *fakeAgent) SetMeterProvider(*sdkmetric.MeterProvider) { f.metered = true }

func (s *MethodsPublicTestSuite) TestStart() {
	tests := []struct {
		name         string
		metrics      bool
		hostErr      error
		tracerErr    error
		agentErr     error
		validateFunc func(got result)
	}{
		{
			name: "when it starts, the agent is the one component, and the cleanup closes, then stops the tracer",
			validateFunc: func(got result) {
				s.Require().NoError(got.err)
				s.Equal(1, got.n)
				s.Equal("v1.2.3", got.agent.version)
				s.False(got.agent.metered)
				got.cleanup()
				s.True(*got.closed)
				s.True(*got.stopped)
			},
		},
		{
			name:    "when the metrics server is on, it is a component, and the agent meters on it",
			metrics: true,
			validateFunc: func(got result) {
				s.Require().NoError(got.err)
				s.Equal(2, got.n)
				s.True(got.agent.metered)
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
			a := &fakeAgent{}
			var closed, stopped bool
			cfg := config.Config{}
			cfg.Agent.Metrics = config.MetricsServer{Enabled: tc.metrics, Host: "127.0.0.1"}
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

			r, err := agentcmd.Start(context.Background(), d)
			tc.validateFunc(result{
				n:       len(r.Components),
				agent:   a,
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
