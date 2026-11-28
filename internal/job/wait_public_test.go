// Copyright (c) 2025 John Dewey

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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/job"
)

type WaitPublicTestSuite struct {
	suite.Suite
}

// TestWait: a job waits the longer of the client's timeout and the wait that
// its context gives.
func (s *WaitPublicTestSuite) TestWait() {
	ctx := context.Background()
	for _, tc := range []struct {
		name string
		ctx  context.Context
		want time.Duration
	}{
		{name: "with no wait, the timeout", ctx: ctx, want: 30 * time.Second},
		{name: "with a longer wait, the wait", ctx: job.WithWait(ctx, 90*time.Second), want: 90 * time.Second},
		{name: "with a shorter wait, the timeout", ctx: job.WithWait(ctx, time.Second), want: 30 * time.Second},
	} {
		s.Run(tc.name, func() {
			s.Equal(tc.want, job.Wait(tc.ctx, 30*time.Second))
		})
	}
}

func TestWaitPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(WaitPublicTestSuite))
}
