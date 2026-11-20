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

package memory_test

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
	"github.com/osapi-io/osapi/internal/behaviors/memory"
	"github.com/osapi-io/osapi/internal/behaviors/memory/mocks"
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
	p, err := memory.Behavior.Provider(platform, host.Host{FS: fs, Exec: e})
	s.Require().NoError(err)

	return p
}

func (s *ProviderPublicTestSuite) TestErrors() {
	tests := []struct {
		name     string
		platform string
		files    map[string]string
		exec     exec
		want     string
	}{
		{
			name:     "when /proc/meminfo is missing",
			platform: behavior.PlatformLinux,
			want:     "read /proc/meminfo",
		},
		{
			name:     "when /proc/meminfo has no MemTotal",
			platform: behavior.PlatformLinux,
			files: map[string]string{
				"/proc/meminfo": "MemFree: x kB\nshort\n",
			},
			want: "has no MemTotal",
		},
		{
			name:     "when sysctl fails",
			platform: behavior.PlatformDarwin,
			exec: exec{
				err: map[string]error{"sysctl -n hw.memsize": errors.New("boom")},
			},
			want: "read hw.memsize: boom",
		},
		{
			name: "when sysctl gives no number", platform: behavior.PlatformDarwin,
			exec: exec{out: map[string]string{"sysctl -n hw.memsize": "lots"}}, want: "hw.memsize",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			p := build(&s.Suite, tc.platform, tc.files, tc.exec).(memory.Provider)
			_, err := p.Memory(context.Background())
			s.ErrorContains(err, tc.want)
		})
	}
}

func (s *ProviderPublicTestSuite) TestGetError() {
	ctrl := gomock.NewController(s.T())
	m := mocks.NewMockProvider(ctrl)
	m.EXPECT().Memory(gomock.Any()).Return(memory.Usage{}, errors.New("boom"))
	_, err := memory.Get(context.Background(), m, memory.GetInput{})
	s.EqualError(err, "boom")
}

func TestProviderPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProviderPublicTestSuite))
}
