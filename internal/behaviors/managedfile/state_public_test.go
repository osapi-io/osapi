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

package managedfile_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
)

type StatePublicTestSuite struct {
	suite.Suite
}

func (s *StatePublicTestSuite) TestStateKey() {
	tests := []struct {
		name     string
		hostname string
		path     string
		want     string
	}{
		{
			name:     "when a path is keyed, the key is the host and the path's checksum",
			hostname: "web-01",
			path:     "/etc/motd",
			want:     "web-01.a5017c23a9f6ed2d1b4e08408fedf24e39aab24bacd1d1597c1432e80497614b",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := managedfile.StateKey(tc.hostname, tc.path)

			s.Equal(tc.want, got)
		})
	}
}

func TestStatePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(StatePublicTestSuite))
}
