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

package behavior_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

type ChainPublicTestSuite struct {
	suite.Suite
}

func (s *ChainPublicTestSuite) TearDownTest() {
	behavior.Reset()
}

func (s *ChainPublicTestSuite) TestRequires() {
	files := behavior.Define("files", "x")
	state := behavior.Define("state", "x")

	b := behavior.Define("sysctl", "x", behavior.Requires(files), behavior.Requires(state))

	s.Equal([]*behavior.Behavior{files, state}, b.Requires)
}

func (s *ChainPublicTestSuite) TestInternal() {
	s.True(behavior.Define("files", "x", behavior.Internal).Internal)
	s.False(behavior.Define("sysctl", "x").Internal)
}

func (s *ChainPublicTestSuite) TestRequiredBy() {
	tests := []struct {
		name  string
		setup func() *behavior.Behavior
		want  []string
	}{
		{
			name: "when behaviors require it, each is listed once",
			setup: func() *behavior.Behavior {
				files := behavior.Register(behavior.Define("files", "x"))
				behavior.Register(behavior.Define("sysctl", "x", behavior.Requires(files, files)))
				behavior.Register(behavior.Define("cron", "x", behavior.Requires(files)))
				behavior.Register(behavior.Define("dns", "x"))

				return files
			},
			want: []string{"cron", "sysctl"},
		},
		{
			name: "when nothing requires it",
			setup: func() *behavior.Behavior {
				return behavior.Register(behavior.Define("files", "x"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			behavior.Reset()
			b := tc.setup()

			var got []string
			for _, r := range behavior.RequiredBy(b) {
				got = append(got, r.Name)
			}
			s.Equal(tc.want, got)
		})
	}
}

func TestChainPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ChainPublicTestSuite))
}
