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

package disk_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/disk"
	"github.com/osapi-io/osapi/internal/behaviors/disk/mocks"
)

type ProviderPublicTestSuite struct {
	suite.Suite
}

// exec answers each command by its line: an output, or an error.
type exec struct {
	host.Exec
	out map[string]string
	err map[string]error
}

func (e exec) RunCmd(_ context.Context, name string, args []string) (string, error) {
	line := strings.Join(append([]string{name}, args...), " ")

	return e.out[line], e.err[line]
}

// build is the behavior's provider on platform, with a file system that has
// files, and commands that answer as e says.
func build(
	s *suite.Suite,
	platform string,
	files map[string]string,
	e exec,
) any {
	fs := memfs.New()
	for path, content := range files {
		s.Require().NoError(fs.MkdirAll(filepath.Dir(path), 0o755))
		s.Require().NoError(fs.WriteFile(path, []byte(content), 0o444))
	}
	p, err := disk.Behavior.Provider(platform, host.Host{FS: fs, Exec: e})
	s.Require().NoError(err)

	return p
}

func (s *ProviderPublicTestSuite) TestErrors() {
	linux := "df -kP -t ext4 -t xfs -t btrfs"
	tests := []struct {
		name string
		exec exec
		want string
	}{
		{
			name: "when df fails",
			exec: exec{err: map[string]error{linux: errors.New("boom")}},
			want: "run df: boom",
		},
		{
			name: "when a size is not a number",
			exec: exec{
				out: map[string]string{linux: "header\n/dev/sda1 x 1 1 1% /\n"},
			},
			want: "df:",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			p := build(&s.Suite, behavior.PlatformLinux, nil, tc.exec).(disk.Provider)
			_, err := p.Disks(context.Background())
			s.ErrorContains(err, tc.want)
		})
	}
}

// TestShortLine leaves out a line that has fewer fields than df prints.
func (s *ProviderPublicTestSuite) TestShortLine() {
	e := exec{
		out: map[string]string{
			"df -kP -T apfs,hfs": "header\nshort line\n/dev/d 1 1 1 1% /Volumes/My Disk\n",
		},
	}
	p := build(&s.Suite, behavior.PlatformDarwin, nil, e).(disk.Provider)
	got, err := p.Disks(context.Background())
	s.Require().NoError(err)
	s.Equal([]disk.Disk{{Name: "/Volumes/My Disk", Total: 1024, Used: 1024, Free: 1024}}, got)
}

func (s *ProviderPublicTestSuite) TestGetError() {
	ctrl := gomock.NewController(s.T())
	m := mocks.NewMockProvider(ctrl)
	m.EXPECT().Disks(gomock.Any()).Return(nil, errors.New("boom"))
	_, err := disk.Get(context.Background(), m, disk.GetInput{})
	s.EqualError(err, "boom")
}

func TestProviderPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProviderPublicTestSuite))
}
