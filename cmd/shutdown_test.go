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

package cmd

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

// fakeComponent records the calls that serve makes.
type fakeComponent struct {
	mu          sync.Mutex
	started     int
	stopped     int
	hasDeadline bool
}

func (f *fakeComponent) Start() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.started++
}

func (f *fakeComponent) Stop(
	ctx context.Context,
) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stopped++
	_, f.hasDeadline = ctx.Deadline()
}

type ShutdownTestSuite struct {
	suite.Suite
}

func TestShutdownTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ShutdownTestSuite))
}

func (s *ShutdownTestSuite) TestServe() {
	tests := []struct {
		name       string
		components int
		cleanups   int
	}{
		{
			name:       "when context ends stops one component",
			components: 1,
		},
		{
			name:       "when context ends stops each component and runs each cleanup",
			components: 3,
			cleanups:   2,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			fakes := make([]*fakeComponent, tc.components)
			components := make([]component, tc.components)
			for i := range fakes {
				fakes[i] = &fakeComponent{}
				components[i] = fakes[i]
			}

			var order []int
			var deadlines []bool
			cleanups := make([]func(context.Context), tc.cleanups)
			for i := range cleanups {
				cleanups[i] = func(ctx context.Context) {
					order = append(order, i)
					_, ok := ctx.Deadline()
					deadlines = append(deadlines, ok)
				}
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			serve(ctx, components, cleanups...)

			for _, f := range fakes {
				s.Equal(1, f.started)
				s.Equal(1, f.stopped)
				s.True(f.hasDeadline, "a stop must be bounded")
			}
			for i := range order {
				s.Equal(i, order[i], "cleanups run in order")
			}
			s.Len(order, tc.cleanups)
			for _, ok := range deadlines {
				s.True(
					ok,
					"cleanup must be bounded: an unreachable collector otherwise blocks SIGTERM",
				)
			}
		})
	}
}
