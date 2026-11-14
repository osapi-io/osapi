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

package behaviorcmdtest_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior/behaviorcmdtest"
	_ "github.com/osapi-io/osapi/internal/behaviors/sysctl/gen"
)

type BehaviorcmdtestPublicTestSuite struct {
	suite.Suite
}

func (s *BehaviorcmdtestPublicTestSuite) TestSent() {
	tests := []struct {
		name string
		args []string
		want behaviorcmdtest.Request
	}{
		{
			name: "when a command runs, the request it sends is returned",
			args: []string{"sysctl", "get", "--target", "server1", "--key", "vm.swappiness"},
			want: behaviorcmdtest.Request{
				Method: "GET",
				Path:   "/api/next/node/server1/sysctl/vm.swappiness",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, behaviorcmdtest.Sent(s.T(), tc.args))
		})
	}
}

func TestBehaviorcmdtestPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(BehaviorcmdtestPublicTestSuite))
}
