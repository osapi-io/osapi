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

package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type RowStatusPublicTestSuite struct {
	suite.Suite
}

func (s *RowStatusPublicTestSuite) TestValid() {
	tests := []struct {
		name   string
		status types.RowStatus
		want   bool
	}{
		{
			name:   "when the host did not answer, timeout is a status",
			status: types.RowStatusTimeout,
			want:   true,
		},
		{name: "when the word is skip, it is not a status", status: "skip", want: false},
		{name: "when the status is empty, it is not a status", status: "", want: false},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.status.Valid())
		})
	}
}

func TestRowStatusPublicTestSuite(t *testing.T) {
	suite.Run(t, new(RowStatusPublicTestSuite))
}
