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

package uptime_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/uptime"
)

type DarwinPublicTestSuite struct {
	suite.Suite
}

// exec answers sysctl with out and err.
type exec struct {
	host.Exec
	out string
	err error
}

func (e exec) RunCmd(context.Context, string, []string) (string, error) {
	return e.out, e.err
}

func (s *DarwinPublicTestSuite) TestUptime() {
	now := time.Unix(1767323045, 0)
	tests := []struct {
		name    string
		exec    exec
		want    time.Duration
		wantErr string
	}{
		{
			name: "when sysctl gives the boot time, the uptime is the host's time since it",
			exec: exec{out: "{ sec = 1767229145, usec = 0 } Thu Jan  1 00:59:05 2026"},
			want: 93900 * time.Second,
		},
		{
			name:    "when sysctl fails, it is the error",
			exec:    exec{err: errors.New("no sysctl")},
			wantErr: "read the boot time: no sysctl",
		},
		{
			name:    "when sysctl gives no boot time, it is the error",
			exec:    exec{out: "nothing"},
			wantErr: "no boot time",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			p, err := uptime.Behavior.Provider(behavior.PlatformDarwin, host.Host{
				Exec: tc.exec,
				Now:  func() time.Time { return now },
			})
			s.Require().NoError(err)

			got, err := p.(uptime.Provider).Uptime(context.Background())
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func TestDarwinPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DarwinPublicTestSuite))
}
