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

package controllercmd_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/controllercmd"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/job"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

type fakeController struct{}

func (fakeController) Start()                            {}
func (fakeController) Stop(context.Context)              {}
func (fakeController) CheckHealth(context.Context) error { return nil }
func (fakeController) SubComponents() map[string]job.SubComponentInfo {
	return nil
}

// result is what Start made, and what its cleanup did.
type result struct {
	n       int
	opts    int
	err     error
	cleanup func()
	stopped *bool
}

func (s *MethodsPublicTestSuite) TestStart() {
	tests := []struct {
		name         string
		metrics      bool
		tracerErr    error
		newErr       error
		validateFunc func(got result)
	}{
		{
			name: "when it starts, the controller is the one component, and the cleanup stops the tracer",
			validateFunc: func(got result) {
				s.Require().NoError(got.err)
				s.Equal(1, got.n)
				s.Equal(1, got.opts, "the version is the one option")
				got.cleanup()
				s.True(*got.stopped)
			},
		},
		{
			name:    "when the metrics server is on, it is a component, and the API meters on it",
			metrics: true,
			validateFunc: func(got result) {
				s.Require().NoError(got.err)
				s.Equal(2, got.n)
				s.Equal(2, got.opts)
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
			name:   "when the controller cannot start, it is the error",
			newErr: errors.New("no NATS"),
			validateFunc: func(got result) {
				s.EqualError(got.err, "start the controller: no NATS")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var stopped bool
			var opts int
			cfg := config.Config{}
			cfg.Controller.Metrics = config.MetricsServer{Enabled: tc.metrics, Host: "127.0.0.1"}
			d := controllercmd.Deps{
				Config:  cfg,
				Logger:  slog.New(slog.DiscardHandler),
				Version: "v1.2.3",
				Tracer: func(context.Context, string) (func(context.Context) error, error) {
					return func(context.Context) error { stopped = true; return nil }, tc.tracerErr
				},
				NewController: func(
					_ context.Context, _ *slog.Logger, _ config.Config, o ...controller.Option,
				) (controllercmd.Controller, error) {
					opts = len(o)

					return fakeController{}, tc.newErr
				},
			}

			r, err := controllercmd.Start(context.Background(), d)
			tc.validateFunc(result{
				n:       len(r.Components),
				opts:    opts,
				err:     err,
				cleanup: func() { r.Cleanup(context.Background()) },
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
