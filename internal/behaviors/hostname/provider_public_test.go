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

package hostname_test

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
	"github.com/osapi-io/osapi/internal/behaviors/hostname"
	"github.com/osapi-io/osapi/internal/behaviors/hostname/mocks"
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

func (e exec) RunPrivilegedCmd(ctx context.Context, name string, args []string) (string, error) {
	return e.RunCmd(ctx, name, args)
}

// build is the provider on platform, with files, and commands that answer
// as e says.
func (s *ProviderPublicTestSuite) build(
	platform string,
	files map[string]string,
	e exec,
	container bool,
) hostname.Provider {
	fs := memfs.New()
	for path, content := range files {
		s.Require().NoError(fs.MkdirAll(filepath.Dir(path), 0o755))
		s.Require().NoError(fs.WriteFile(path, []byte(content), 0o444))
	}
	p, err := hostname.Behavior.Provider(platform, host.Host{FS: fs, Exec: e, Container: container})
	s.Require().NoError(err)

	return p.(hostname.Provider)
}

func (s *ProviderPublicTestSuite) TestHostnameErrors() {
	tests := []struct {
		name     string
		platform string
		exec     exec
		want     string
	}{
		{
			name: "when /proc/sys/kernel/hostname is missing", platform: behavior.PlatformLinux,
			want: "read /proc/sys/kernel/hostname",
		},
		{
			name: "when sysctl fails", platform: behavior.PlatformDarwin,
			exec: exec{err: map[string]error{"sysctl -n kern.hostname": errors.New("boom")}},
			want: "read kern.hostname: boom",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			_, err := s.build(tc.platform, nil, tc.exec, false).Hostname(context.Background())
			s.ErrorContains(err, tc.want)
		})
	}
}

func (s *ProviderPublicTestSuite) TestSetHostnameErrors() {
	tests := []struct {
		name      string
		platform  string
		exec      exec
		container bool
		want      error
		wantText  string
	}{
		{
			name: "when the host is a container", platform: behavior.PlatformLinux,
			container: true, want: behavior.ErrUnsupported,
		},
		{
			name: "when the host is macOS", platform: behavior.PlatformDarwin,
			want: behavior.ErrUnsupported,
		},
		{
			name: "when hostnamectl cannot read the name", platform: behavior.PlatformLinux,
			exec:     exec{err: map[string]error{"hostnamectl hostname": errors.New("boom")}},
			wantText: "read the hostname: boom",
		},
		{
			name: "when hostnamectl cannot set the name", platform: behavior.PlatformLinux,
			exec: exec{
				out: map[string]string{"hostnamectl hostname": "web-01\n"},
				err: map[string]error{"hostnamectl set-hostname web-02": errors.New("boom")},
			},
			wantText: "set the hostname: boom",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			p := s.build(tc.platform, nil, tc.exec, tc.container)
			_, err := p.SetHostname(context.Background(), "web-02")
			if tc.want != nil {
				s.ErrorIs(err, tc.want)

				return
			}
			s.ErrorContains(err, tc.wantText)
		})
	}
}

func (s *ProviderPublicTestSuite) TestMethodErrors() {
	ctrl := gomock.NewController(s.T())
	m := mocks.NewMockProvider(ctrl)
	m.EXPECT().Hostname(gomock.Any()).Return("", errors.New("boom"))
	m.EXPECT().SetHostname(gomock.Any(), "web-02").Return(false, errors.New("boom"))

	_, err := hostname.Get(context.Background(), m, hostname.GetInput{})
	s.EqualError(err, "boom")
	_, err = hostname.Update(context.Background(), m, hostname.UpdateInput{Name: "web-02"})
	s.EqualError(err, "boom")
}

func TestProviderPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProviderPublicTestSuite))
}
