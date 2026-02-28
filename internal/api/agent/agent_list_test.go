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

package agent

import (
	"math"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AgentListInternalTestSuite struct {
	suite.Suite
}

func (s *AgentListInternalTestSuite) TestUint64ToInt() {
	maxInt := int(^uint(0) >> 1)

	tests := []struct {
		name string
		val  uint64
		want int
	}{
		{
			name: "when zero",
			val:  0,
			want: 0,
		},
		{
			name: "when normal value",
			val:  42,
			want: 42,
		},
		{
			name: "when max int value",
			val:  uint64(maxInt),
			want: maxInt,
		},
		{
			name: "when overflow",
			val:  math.MaxUint64,
			want: maxInt,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := uint64ToInt(tt.val)
			s.Equal(tt.want, got)
		})
	}
}

func TestAgentListInternalTestSuite(t *testing.T) {
	suite.Run(t, new(AgentListInternalTestSuite))
}
