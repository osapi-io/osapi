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

package job_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/job"
)

// ErrorsPublicTestSuite covers the failure a job becomes on the controller: the
// cause survives as a type, so a handler asks rather than reads.
type ErrorsPublicTestSuite struct {
	suite.Suite
}

func (s *ErrorsPublicTestSuite) TestResponseError() {
	tests := []struct {
		name         string
		resp         *job.Response
		validateFunc func(error)
	}{
		{
			name: "a missing thing",
			resp: &job.Response{
				Error:     `user "nobody": not found`,
				ErrorCode: job.ErrorCodeNotFound,
				Hostname:  "web-01",
			},
			validateFunc: func(err error) {
				s.Require().ErrorIs(err, job.ErrNotFound)
				s.Require().NotErrorIs(err, job.ErrNotManaged)
				s.Equal(`job failed: user "nobody": not found`, err.Error())
			},
		},
		{
			name: "a thing osapi does not manage",
			resp: &job.Response{
				Error:     "sysctl entry: not managed by osapi",
				ErrorCode: job.ErrorCodeNotManaged,
			},
			validateFunc: func(err error) {
				s.Require().ErrorIs(err, job.ErrNotManaged)
				s.Require().NotErrorIs(err, job.ErrNotFound)
			},
		},
		{
			name: "a program the host does not have",
			resp: &job.Response{
				ErrorCode: job.ErrorCodeNotInstalled,
			},
			validateFunc: func(err error) {
				s.Require().ErrorIs(err, job.ErrNotInstalled)
			},
		},
		{
			name: "an operation this OS family does not have",
			resp: &job.Response{
				ErrorCode: job.ErrorCodeUnsupported,
			},
			validateFunc: func(err error) {
				s.Require().ErrorIs(err, job.ErrUnsupported)
			},
		},
		{
			name: "a failure with no cause the API distinguishes",
			resp: &job.Response{
				Error: "apt-get returned 100",
			},
			validateFunc: func(err error) {
				s.Require().Error(err)
				s.Require().NotErrorIs(err, job.ErrNotFound)
				s.Require().NotErrorIs(err, job.ErrNotManaged)
				s.Require().NotErrorIs(err, job.ErrNotInstalled)
				s.Require().NotErrorIs(err, job.ErrUnsupported)
			},
		},
		{
			name: "a code this controller does not know, from a newer agent",
			resp: &job.Response{
				Error:     "something new",
				ErrorCode: job.ErrorCode("sideways"),
			},
			validateFunc: func(err error) {
				// A rollout, not a reason to answer 404.
				s.Require().Error(err)
				s.Require().NotErrorIs(err, job.ErrNotFound)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(job.NewResponseError(tt.resp))
		})
	}
}

// TestResponseErrorNil covers the one call that has nothing to describe.
func (s *ErrorsPublicTestSuite) TestResponseErrorNil() {
	s.Nil(job.NewResponseError(nil))

	var e *job.ResponseError
	s.False(errors.Is(e, job.ErrNotFound))
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestErrorsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ErrorsPublicTestSuite))
}
