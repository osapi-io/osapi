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

package agent_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/provider"
)

// ClassifyPublicTestSuite covers the one place a provider's error becomes a code
// on the response, which is what lets the controller answer without reading the
// sentence back.
type ClassifyPublicTestSuite struct {
	suite.Suite
}

func (s *ClassifyPublicTestSuite) TestClassifyProviderError() {
	tests := []struct {
		name string
		err  error
		want job.ErrorCode
	}{
		{
			name: "a thing the host does not have",
			err:  fmt.Errorf(`user "nobody": %w`, provider.ErrNotFound),
			want: job.ErrorCodeNotFound,
		},
		{
			name: "a thing osapi did not deploy",
			err:  fmt.Errorf("cron entry %q: %w", "backup", provider.ErrNotManaged),
			want: job.ErrorCodeNotManaged,
		},
		{
			name: "a program the host does not have",
			err:  fmt.Errorf("package %q: %w", "nginx", provider.ErrNotInstalled),
			want: job.ErrorCodeNotInstalled,
		},
		{
			name: "an operation this OS family does not have",
			err:  fmt.Errorf("timezone: %w", provider.ErrUnsupported),
			want: job.ErrorCodeUnsupported,
		},
		{
			name: "a failure with no cause the API distinguishes",
			err:  errors.New("apt-get returned 100"),
			want: "",
		},
		{
			name: "a cause wrapped two errors deep",
			err:  fmt.Errorf("get user: %w", fmt.Errorf("read passwd: %w", provider.ErrNotFound)),
			want: job.ErrorCodeNotFound,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, agent.ExportClassifyProviderError(tt.err))
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestClassifyPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ClassifyPublicTestSuite))
}
