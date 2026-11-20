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
	"testing"
	"time"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/uptime"
)

type LinuxPublicTestSuite struct {
	suite.Suite
}

func (s *LinuxPublicTestSuite) TestUptime() {
	tests := []struct {
		name    string
		file    string
		want    time.Duration
		wantErr string
	}{
		{
			name: "when /proc/uptime has the seconds since boot, it is them",
			file: "93900.42 180001.15\n",
			want: 93900*time.Second + 420*time.Millisecond,
		},
		{name: "when /proc/uptime is missing, it is the error", wantErr: "read /proc/uptime"},
		{name: "when /proc/uptime is empty, it is the error", file: " ", wantErr: "is empty"},
		{
			name:    "when /proc/uptime has no number, it is the error",
			file:    "up 0",
			wantErr: "invalid syntax",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			fs := memfs.New()
			if tc.file != "" {
				s.Require().NoError(fs.MkdirAll("/proc", 0o755))
				s.Require().NoError(fs.WriteFile("/proc/uptime", []byte(tc.file), 0o444))
			}
			p, err := uptime.Behavior.Provider(behavior.PlatformLinux, host.Host{FS: fs})
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

func TestLinuxPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LinuxPublicTestSuite))
}
