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

package packages_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/packages"
)

type DebianPublicTestSuite struct {
	suite.Suite
}

// installedTree is the dpkg-query row of tree 2.1.1-2, installed.
const installedTree = "tree\t2.1.1-2\tamd64\tinstalled\ttree\t100\n"

func (s *DebianPublicTestSuite) TestInstall() {
	tests := []struct {
		name         string
		version      string
		setupMock    func(*hostmocks.MockExec)
		validateFunc func(bool, error)
	}{
		{
			name:    "a higher version upgrades",
			version: "2.1.1-3",
			setupMock: func(m *hostmocks.MockExec) {
				m.EXPECT().RunPrivilegedCmd(gomock.Any(), "apt-get", []string{
					"install", "-y", "-o", "DPkg::Lock::Timeout=300",
					"--allow-downgrades", "--", "tree=2.1.1-3",
				}).Return("", nil)
			},
			validateFunc: func(changed bool, err error) {
				s.NoError(err)
				s.True(changed)
			},
		},
		{
			name:    "a lower version downgrades",
			version: "2.1.1-1",
			setupMock: func(m *hostmocks.MockExec) {
				m.EXPECT().RunPrivilegedCmd(gomock.Any(), "apt-get", []string{
					"install", "-y", "-o", "DPkg::Lock::Timeout=300",
					"--allow-downgrades", "--", "tree=2.1.1-1",
				}).Return("", nil)
			},
			validateFunc: func(changed bool, err error) {
				s.NoError(err)
				s.True(changed)
			},
		},
		{
			name:      "the installed version is unchanged",
			version:   "2.1.1-2",
			setupMock: func(*hostmocks.MockExec) {},
			validateFunc: func(changed bool, err error) {
				s.NoError(err)
				s.False(changed)
			},
		},
		{
			name:      "an invalid version is refused",
			version:   "2.1;rm",
			setupMock: func(*hostmocks.MockExec) {},
			validateFunc: func(changed bool, err error) {
				s.ErrorContains(err, `"2.1;rm" is not a valid package version`)
				s.False(changed)
			},
		},
		{
			name:    "a failed install is an error",
			version: "2.1.1-3",
			setupMock: func(m *hostmocks.MockExec) {
				m.EXPECT().RunPrivilegedCmd(gomock.Any(), "apt-get", gomock.Any()).
					Return("", errors.New("exit status 100"))
			},
			validateFunc: func(changed bool, err error) {
				s.ErrorContains(err, "install tree=2.1.1-3: exit status 100")
				s.False(changed)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			exec := hostmocks.NewMockExec(ctrl)
			behaviortest.ReadsThroughRunCmd(exec)
			exec.EXPECT().RunCmd(gomock.Any(), "dpkg-query", gomock.Any()).
				Return(installedTree, nil).AnyTimes()
			tt.setupMock(exec)
			p, err := packages.ExportNewDebian(
				host.Host{Platform: behavior.PlatformDebian, Exec: exec},
			)
			s.Require().NoError(err)

			changed, err := p.Install(context.Background(), "tree", tt.version)
			tt.validateFunc(changed, err)
		})
	}
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
