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

package command_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
)

type ProcessPublicTestSuite struct {
	suite.Suite
}

type ctxKey struct{}

// fakeComponent is a component that records what Serve did with it.
type fakeComponent struct {
	mu      sync.Mutex
	started bool
	stopCtx context.Context
	stopErr error
}

func (p *fakeComponent) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.started = true
}

func (p *fakeComponent) Stop(
	ctx context.Context,
) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopCtx = ctx
	p.stopErr = ctx.Err()
}

func (s *ProcessPublicTestSuite) TestServe() {
	tests := []struct {
		name         string
		setup        func(context.Context, localDeps) (command.Running, error)
		method       string
		uses         bool
		validateFunc func(err error, setupCtx context.Context, p *fakeComponent, cleaned context.Context)
	}{
		{
			name:   "when the setup works, the parts run until the context ends, then stop and clean up",
			method: "start",
			uses:   true,
			validateFunc: func(err error, setupCtx context.Context, p *fakeComponent, cleaned context.Context) {
				s.Require().NoError(err)
				s.Equal("root", setupCtx.Value(ctxKey{}), "setup gets the command's context")
				s.True(p.started)
				s.Require().NotNil(p.stopCtx)
				s.NoError(p.stopErr, "the stop context is not the cancelled one")
				s.Equal("root", p.stopCtx.Value(ctxKey{}), "the stop context keeps its values")
				_, ok := p.stopCtx.Deadline()
				s.True(ok, "the stop has ShutdownTimeout")
				s.Equal(p.stopCtx, cleaned)
			},
		},
		{
			name:   "when the setup fails, its error is the command's, and nothing starts",
			method: "start",
			uses:   true,
			setup: func(context.Context, localDeps) (command.Running, error) {
				return command.Running{}, errors.New("no port")
			},
			validateFunc: func(err error, _ context.Context, p *fakeComponent, _ context.Context) {
				s.EqualError(err, "no port")
				s.False(p.started)
			},
		},
		{
			name:   "when the method is not the group's, it is an error",
			method: "stop",
			uses:   true,
			validateFunc: func(err error, _ context.Context, _ *fakeComponent, _ context.Context) {
				s.ErrorContains(err, "has no method stop")
			},
		},
		{
			name:   "when nothing gives the deps, it is an error",
			method: "start",
			validateFunc: func(err error, _ context.Context, _ *fakeComponent, _ context.Context) {
				s.ErrorContains(err, "is not given")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			p := &fakeComponent{}
			var setupCtx, cleaned context.Context
			setup := tc.setup
			if setup == nil {
				setup = func(ctx context.Context, _ localDeps) (command.Running, error) {
					setupCtx = ctx

					return command.Running{
						Components: []command.Component{p},
						Cleanup:    func(ctx context.Context) { cleaned = ctx },
					}, nil
				}
			}
			g := command.Define("zzproc", "Process", command.Uses[localDeps](),
				command.Process("start").Short("Start it").Setup(setup))
			if tc.uses {
				command.Provide(g, func() localDeps { return localDeps{} })
			}

			ctx, cancel := context.WithCancel(
				context.WithValue(context.Background(), ctxKey{}, "root"),
			)
			cmd := &cobra.Command{}
			cmd.SetContext(ctx)
			time.AfterFunc(10*time.Millisecond, cancel)
			defer cancel()

			tc.validateFunc(command.Serve(cmd, g, tc.method), setupCtx, p, cleaned)
		})
	}
}

// TestServeNoContext runs a process whose command has no context: Serve
// gives the setup a background one.
func (s *ProcessPublicTestSuite) TestServeNoContext() {
	g := command.Define("zzproc", "Process", command.Uses[struct{}](),
		command.Process("start").Short("Start it").
			Setup(func(ctx context.Context, _ struct{}) (command.Running, error) {
				s.NotNil(ctx)

				return command.Running{}, errors.New("stop here")
			}))

	s.EqualError(command.Serve(&cobra.Command{}, g, "start"), "stop here")
}

// TestProcessOutput holds that a process, which prints nothing, breaks no
// rule of the output.
func (s *ProcessPublicTestSuite) TestProcessOutput() {
	g := command.Define("zzprocout", "Process", command.Uses[struct{}](),
		command.Process("start").Short("Start it").
			Setup(func(context.Context, struct{}) (command.Running, error) {
				return command.Running{}, nil
			}))
	command.AddCommands(g, &cobra.Command{Use: "start"})

	s.True(g.Method("start").IsProcess())
	s.Nil(g.Method("start").Output())
	for _, p := range command.ValidateOutput() {
		s.NotEqual("zzprocout", p.Group, p.String())
	}
}

func (s *ProcessPublicTestSuite) TestThen() {
	var order []string
	add := func(name string) func(context.Context) {
		return func(context.Context) { order = append(order, name) }
	}

	command.Then(command.Running{Cleanup: add("close")}, add("trace")).Cleanup(context.Background())
	command.Then(command.Running{}, add("only")).Cleanup(context.Background())

	s.Equal([]string{"close", "trace", "only"}, order)
}

func (s *ProcessPublicTestSuite) TestFlag() {
	m := command.Process("start").Flag("agent.max_jobs", "agent.hostname")
	s.Equal([]string{"agent.max_jobs", "agent.hostname"}, m.ConfigKeys())
}

func TestProcessPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProcessPublicTestSuite))
}
