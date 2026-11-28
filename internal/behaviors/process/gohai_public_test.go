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

package process_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/process"
)

type GohaiPublicTestSuite struct {
	suite.Suite
}

// TestGet: the provider refuses a PID outside the rule's bounds before it
// converts it to an int32, as a caller that skips the rule would send it.
func (s *GohaiPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		pid          int
		validateFunc func(process.Entry, error)
	}{
		{
			name: "when the PID is above pid_max, which int32 wraps",
			pid:  4294967297,
			validateFunc: func(_ process.Entry, err error) {
				s.ErrorIs(err, behavior.ErrInvalid)
				s.ErrorContains(err, "a PID is from 1 to 4194304")
			},
		},
		{
			name: "when the PID is below 1",
			pid:  0,
			validateFunc: func(_ process.Entry, err error) {
				s.ErrorIs(err, behavior.ErrInvalid)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			p, err := process.ExportNewGohai(host.Host{})
			s.Require().NoError(err)

			tc.validateFunc(p.Get(context.Background(), tc.pid))
		})
	}
}

func TestGohaiPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(GohaiPublicTestSuite))
}
