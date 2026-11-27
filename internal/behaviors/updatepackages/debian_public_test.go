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

package updatepackages_test

import (
	"context"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/updatepackages"
)

type DebianPublicTestSuite struct {
	suite.Suite
}

func (s *DebianPublicTestSuite) TestRefresh() {
	const release = "/var/lib/apt/lists/archive.ubuntu.com_ubuntu_dists_noble_InRelease"
	for name, tc := range map[string]struct {
		before string
		after  string
		want   bool
	}{
		"when a repository publishes new data": {before: "Date: Mon\n", after: "Date: Tue\n", want: true},
		"when a repository is first fetched":   {after: "Date: Tue\n", want: true},
		"when each repository is unchanged":    {before: "Date: Mon\n", after: "Date: Mon\n"},
	} {
		s.Run(name, func() {
			fs := memfs.New()
			s.Require().NoError(fs.MkdirAll("/var/lib/apt/lists", 0o755))
			if tc.before != "" {
				s.Require().NoError(fs.WriteFile(release, []byte(tc.before), 0o644))
			}
			exec := hostmocks.NewMockExec(gomock.NewController(s.T()))
			exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "apt-get", []string{"update"}).
				DoAndReturn(func(context.Context, string, []string) (string, error) {
					// apt-get prints Get: on each run; only the file tells.
					return "Get:1 http://archive.ubuntu.com/ubuntu noble InRelease\n",
						fs.WriteFile(release, []byte(tc.after), 0o644)
				})
			p, err := updatepackages.ExportNewDebian(host.Host{Exec: exec, FS: fs})
			s.Require().NoError(err)

			got, err := p.Refresh(context.Background())
			s.NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

// TestUpdatesReadInC: apt list runs as a read command, which sets LC_ALL=C,
// so its "upgradable from" is the English that parse reads in each locale.
// A run as a plain command took the host's locale, and a translated apt
// shifted the fields, so the updates were left out.
func (s *DebianPublicTestSuite) TestUpdatesReadInC() {
	exec := hostmocks.NewMockExec(gomock.NewController(s.T()))
	exec.EXPECT().RunReadCmd(gomock.Any(), "apt", []string{"list", "--upgradable"}).
		Return([]byte("curl/noble-updates 8.5.0-2ubuntu10.6 amd64 [upgradable from: 8.5.0-2ubuntu10.5]\n"), nil)
	p, err := updatepackages.ExportNewDebian(host.Host{Exec: exec, FS: memfs.New()})
	s.Require().NoError(err)

	got, err := p.Updates(context.Background())
	s.Require().NoError(err)
	s.Equal([]updatepackages.Available{{
		Name: "curl", CurrentVersion: "8.5.0-2ubuntu10.5", NewVersion: "8.5.0-2ubuntu10.6",
	}}, got)
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
