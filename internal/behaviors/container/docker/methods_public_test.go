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

package docker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/container/docker"
	"github.com/osapi-io/osapi/internal/behaviors/container/docker/mocks"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

// TestMethods: each method hands its input to the provider and gives back
// what it answers, with the container's name in a write's result.
func (s *MethodsPublicTestSuite) TestMethods() {
	boom := errors.New("boom")
	tests := []struct {
		name string
		run  func(m *mocks.MockProvider) (any, error)
		want any
		err  error
	}{
		{
			name: "list hands over the state",
			run: func(m *mocks.MockProvider) (any, error) {
				m.EXPECT().List(gomock.Any(), "running").Return([]docker.Entry{{Name: "a"}}, nil)
				return docker.List(context.Background(), m, docker.ListInput{State: "running"})
			},
			want: []docker.Entry{{Name: "a"}},
		},
		{
			name: "get reads one by its name",
			run: func(m *mocks.MockProvider) (any, error) {
				m.EXPECT().Get(gomock.Any(), "a").Return(docker.Entry{Name: "a"}, nil)
				return docker.Get(context.Background(), m, docker.GetInput{Name: "a"})
			},
			want: docker.Entry{Name: "a"},
		},
		{
			name: "create answers what the provider made",
			run: func(m *mocks.MockProvider) (any, error) {
				in := docker.CreateInput{Name: "a", Image: "alpine"}
				m.EXPECT().
					Create(gomock.Any(), in).
					Return(docker.Result{Name: "a", Changed: true}, nil)
				return docker.Create(context.Background(), m, in)
			},
			want: docker.Result{Name: "a", Changed: true},
		},
		{
			name: "delete names the container and whether it changed",
			run: func(m *mocks.MockProvider) (any, error) {
				m.EXPECT().Delete(gomock.Any(), "a").Return(true, nil)
				return docker.Delete(
					context.Background(),
					m,
					docker.DeleteInput{Name: "a"},
				)
			},
			want: docker.Result{Name: "a", Changed: true},
		},
		{
			name: "start passes the provider's error",
			run: func(m *mocks.MockProvider) (any, error) {
				m.EXPECT().Start(gomock.Any(), "a").Return(false, boom)
				return docker.Start(context.Background(), m, docker.StartInput{Name: "a"})
			},
			err: boom,
		},
		{
			name: "stop hands over the timeout",
			run: func(m *mocks.MockProvider) (any, error) {
				m.EXPECT().Stop(gomock.Any(), "a", 5).Return(true, nil)
				return docker.Stop(context.Background(), m, docker.StopInput{Name: "a", Timeout: 5})
			},
			want: docker.Result{Name: "a", Changed: true},
		},
		{
			name: "exec answers what the command wrote",
			run: func(m *mocks.MockProvider) (any, error) {
				in := docker.ExecInput{Name: "a", Command: []string{"true"}}
				m.EXPECT().
					Exec(gomock.Any(), in).
					Return(docker.ExecResult{ExitCode: 3, Changed: true}, nil)
				return docker.Exec(context.Background(), m, in)
			},
			want: docker.ExecResult{ExitCode: 3, Changed: true},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got, err := tc.run(mocks.NewMockProvider(gomock.NewController(s.T())))
			if tc.err != nil {
				s.ErrorIs(err, tc.err)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
