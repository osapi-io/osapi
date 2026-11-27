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

package wiring_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/wiring"
)

type WiringPublicTestSuite struct {
	suite.Suite
}

func TestWiringPublicTestSuite(t *testing.T) {
	suite.Run(t, new(WiringPublicTestSuite))
}

func (s *WiringPublicTestSuite) TestTake() {
	tests := []struct {
		name    string
		reports []error
		want    []string
	}{
		{name: "when nothing is reported it is nil"},
		{
			name:    "when two are reported it joins both",
			reports: []error{errors.New("one"), errors.New("two")},
			want:    []string{"one", "two"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_ = wiring.Take()
			for _, err := range tt.reports {
				wiring.Report(err)
			}
			err := wiring.Take()
			if len(tt.want) == 0 {
				s.NoError(err)
				return
			}
			for _, w := range tt.want {
				s.ErrorContains(err, w)
			}
			s.NoError(wiring.Take(), "Take clears what it returned")
		})
	}
}
