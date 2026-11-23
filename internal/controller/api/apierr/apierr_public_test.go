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

package apierr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller/api/apierr"
	"github.com/osapi-io/osapi/internal/job"
)

// ApierrPublicTestSuite covers the one question a handler asks of a failed job.
type ApierrPublicTestSuite struct {
	suite.Suite
}

func (s *ApierrPublicTestSuite) TestIsMissing() {
	tests := []struct {
		name         string
		err          error
		validateFunc func(bool)
	}{
		{
			name:         "a thing the host does not have",
			err:          fmt.Errorf(`user "nobody": %w`, job.ErrNotFound),
			validateFunc: func(ok bool) { s.True(ok) },
		},
		{
			name:         "a thing osapi does not manage",
			err:          fmt.Errorf("sysctl entry: %w", job.ErrNotManaged),
			validateFunc: func(ok bool) { s.True(ok) },
		},
		{
			name:         "a program that is not installed",
			err:          fmt.Errorf("package: %w", job.ErrNotInstalled),
			validateFunc: func(ok bool) { s.True(ok) },
		},
		{
			name: "an operation this OS family does not have",
			err:  fmt.Errorf("timezone: %w", job.ErrUnsupported),
			validateFunc: func(ok bool) {
				// Skipped is not missing: the endpoint answers 200 with a
				// skipped result rather than 404.
				s.False(ok)
			},
		},
		{
			name:         "an ordinary failure",
			err:          errors.New("apt-get returned 100"),
			validateFunc: func(ok bool) { s.False(ok) },
		},
		{
			name:         "a failed job carrying the cause on its response",
			err:          job.NewResponseError(&job.Response{ErrorCode: job.ErrorCodeNotFound}),
			validateFunc: func(ok bool) { s.True(ok) },
		},
		{
			name:         "no error at all",
			err:          nil,
			validateFunc: func(ok bool) { s.False(ok) },
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(apierr.IsMissing(tt.err))
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestApierrPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ApierrPublicTestSuite))
}
