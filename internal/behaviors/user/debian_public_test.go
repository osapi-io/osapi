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

package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/user"
)

type DebianPublicTestSuite struct {
	suite.Suite
}

// TestCreateWhenThePasswordFails: when chpasswd fails after useradd, the
// user is removed again, so that a second run adds it with its password
// instead of finding it and reporting changed: false.
func (s *DebianPublicTestSuite) TestCreateWhenThePasswordFails() {
	chpasswdErr := errors.New("chpasswd: PAM failure")
	tests := []struct {
		name        string
		userdelErr  error
		wantChanged bool
	}{
		{name: "the user is removed and nothing changed"},
		{
			name:        "when userdel fails too, the user stays and both errors return",
			userdelErr:  errors.New("userdel: user is logged in"),
			wantChanged: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			fs := memfs.New()
			s.Require().NoError(fs.MkdirAll("/etc", 0o755))
			s.Require().NoError(fs.WriteFile("/etc/passwd",
				[]byte("root:x:0:0:root:/root:/bin/sh\n"), 0o644))
			s.Require().NoError(fs.WriteFile("/etc/group", []byte("root:x:0:\n"), 0o644))
			exec := hostmocks.NewMockExec(ctrl)
			behaviortest.ReadsThroughRunCmd(exec)
			exec.EXPECT().
				RunCmd(gomock.Any(), gomock.Any(), gomock.Any()).
				Return("", nil).
				AnyTimes()
			gomock.InOrder(
				exec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "useradd", []string{"--create-home", "--", "osapitest"}).
					Return("", nil),
				exec.EXPECT().
					RunPrivilegedCmdWithStdin(gomock.Any(), "chpasswd", []string{"-e"}, gomock.Any()).
					Return("", chpasswdErr),
				exec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "userdel", []string{"-r", "--", "osapitest"}).
					Return("", tc.userdelErr),
			)
			p, err := user.NewDebian(host.Host{FS: fs, Exec: exec, Platform: "debian"})
			s.Require().NoError(err)

			changed, err := p.Create(context.Background(), user.CreateInput{
				Name:         "osapitest",
				PasswordHash: "$6$salt$hash",
			})

			s.ErrorIs(err, chpasswdErr)
			s.Equal(tc.wantChanged, changed)
			if tc.userdelErr != nil {
				s.ErrorIs(err, tc.userdelErr)
			}
		})
	}
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
