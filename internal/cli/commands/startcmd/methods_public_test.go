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

package startcmd_test

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
	"github.com/osapi-io/osapi/internal/cli/commands/controllercmd"
	"github.com/osapi-io/osapi/internal/cli/commands/natsserver"
	"github.com/osapi-io/osapi/internal/cli/commands/startcmd"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/job"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

type fake struct{}

func (fake) Start()                                         {}
func (fake) Stop(context.Context)                           {}
func (fake) IsReady() error                                 { return nil }
func (fake) CheckHealth(context.Context) error              { return nil }
func (fake) SubComponents() map[string]job.SubComponentInfo { return nil }
func (fake) LastHeartbeatTime() time.Time                   { return time.Time{} }
func (fake) SetVersion(string)                              {}
func (fake) SetLogLevel(*slog.LevelVar)                     {}
func (fake) SetMeterProvider(*sdkmetric.MeterProvider)      {}
func (fake) StartHeartbeat(context.Context, string, config.ProcessConditions) {
}

// result is what Start made, and what its cleanup did.
type result struct {
	n       int
	err     error
	tracer  string
	cleanup func()
	closed  *bool
	stopped *bool
}

func (s *MethodsPublicTestSuite) TestStart() {
	boom := errors.New("boom")
	tests := []struct {
		name         string
		hostErr      error
		tracerErr    error
		natsErr      error
		controlErr   error
		agentErr     error
		validateFunc func(got result)
	}{
		{
			name: "when it starts, the controller, the agent and the NATS server run under one tracer",
			validateFunc: func(got result) {
				s.Require().NoError(got.err)
				s.Equal(3, got.n)
				s.Equal("osapi", got.tracer)
				got.cleanup()
				s.True(*got.closed, "the agent's connection closes")
				s.True(*got.stopped, "the tracer stops")
			},
		},
		{
			name:    "when the host is not supported, it is the error",
			hostErr: boom,
			validateFunc: func(got result) {
				s.EqualError(got.err, "unsupported host: boom")
			},
		},
		{
			name:      "when the tracer cannot start, it is the error",
			tracerErr: boom,
			validateFunc: func(got result) {
				s.EqualError(got.err, "start the tracer: boom")
			},
		},
		{
			name:    "when the NATS server cannot start, it is the error",
			natsErr: boom,
			validateFunc: func(got result) {
				s.EqualError(got.err, "start the NATS server: boom")
			},
		},
		{
			name:       "when the controller cannot start, it is the error",
			controlErr: boom,
			validateFunc: func(got result) {
				s.EqualError(got.err, "start the controller: boom")
			},
		},
		{
			name:     "when the agent cannot start, it is the error",
			agentErr: boom,
			validateFunc: func(got result) {
				s.EqualError(got.err, "start the agent: boom")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var closed, stopped bool
			var tracer string
			log := slog.New(slog.DiscardHandler)
			d := startcmd.Deps{
				NATS: natsserver.Deps{
					Logger: log,
					NewServer: func(
						context.Context, *slog.Logger, config.NATS, *sdkmetric.MeterProvider,
					) (natsserver.Server, error) {
						return fake{}, tc.natsErr
					},
				},
				Controller: controllercmd.Deps{
					Logger: log,
					NewController: func(
						context.Context, *slog.Logger, config.Config, ...controller.Option,
					) (controllercmd.Controller, error) {
						return fake{}, tc.controlErr
					},
				},
				Agent: agentcmd.Deps{
					Logger:    log,
					CheckHost: func() error { return tc.hostErr },
					Tracer: func(_ context.Context, name string) (func(context.Context) error, error) {
						tracer = name

						return func(context.Context) error { stopped = true; return nil }, tc.tracerErr
					},
					NewAgent: func(
						context.Context, *slog.Logger, avfs.VFS, config.Config, *sdkmetric.MeterProvider,
					) (agentcmd.Agent, func(), error) {
						return fake{}, func() { closed = true }, tc.agentErr
					},
				},
			}

			r, err := startcmd.Start(context.Background(), d)
			tc.validateFunc(result{
				n:       len(r.Components),
				err:     err,
				tracer:  tracer,
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
