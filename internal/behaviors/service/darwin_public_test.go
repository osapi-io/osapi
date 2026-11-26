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

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	managedfilemocks "github.com/osapi-io/osapi/internal/behaviors/managedfile/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/object"
	objectmocks "github.com/osapi-io/osapi/internal/behaviors/object/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/service"
)

// DarwinPublicTestSuite has the failures of the launchd provider that a fake
// host's examples do not give, such as a file-state bucket that cannot be
// read.
type DarwinPublicTestSuite struct {
	suite.Suite
}

func (s *DarwinPublicTestSuite) TestServiceWhenTheRecordCannotBeRead() {
	tests := []struct {
		name      string
		call      func(context.Context, service.Provider) error
		wantError string
	}{
		{
			name: "get fails with the file-state bucket",
			call: func(ctx context.Context, p service.Provider) error {
				_, err := p.Service(ctx, "app")

				return err
			},
			wantError: "bucket is down",
		},
		{
			name: "list fails with the file-state bucket",
			call: func(ctx context.Context, p service.Provider) error {
				_, err := p.Services(ctx)

				return err
			},
			wantError: "bucket is down",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			files := managedfilemocks.NewMockProvider(ctrl)
			files.EXPECT().Read(gomock.Any(), gomock.Any()).
				Return(managedfile.Record{}, errors.New("bucket is down")).AnyTimes()
			exec := hostmocks.NewMockExec(ctrl)
			exec.EXPECT().RunCmd(gomock.Any(), "launchctl", []string{"print", "system"}).
				Return("system = {\n\tservices = {\n\t\t     612      - \tio.osapi.app\n\t}\n}\n", nil).
				AnyTimes()
			p, err := service.Behavior.Provider(behavior.PlatformDarwin, host.Host{
				FS:   memfs.New(),
				Exec: exec,
				Required: map[string]any{
					managedfile.Behavior.Name: files,
					object.Behavior.Name:      objectmocks.NewMockProvider(ctrl),
				},
			})
			s.Require().NoError(err)

			s.ErrorContains(tc.call(context.Background(), p.(service.Provider)), tc.wantError)
		})
	}
}

// TestStartADisabledDaemon: launchd does not bootstrap a daemon that is
// disabled, so start refuses it before it runs launchctl.
func (s *DarwinPublicTestSuite) TestStartADisabledDaemon() {
	for _, action := range []string{"start", "restart"} {
		s.Run(action, func() {
			ctrl := gomock.NewController(s.T())
			files := managedfilemocks.NewMockProvider(ctrl)
			files.EXPECT().Read(gomock.Any(), "/Library/LaunchDaemons/io.osapi.app.plist").
				Return(managedfile.Record{Path: "/Library/LaunchDaemons/io.osapi.app.plist"}, nil).
				AnyTimes()
			exec := hostmocks.NewMockExec(ctrl)
			exec.EXPECT().RunCmd(gomock.Any(), "launchctl", []string{"print", "system"}).
				Return("system = {\n\tdisabled services = {\n\t\t\"io.osapi.app\" => disabled\n\t}\n}\n", nil).
				AnyTimes()
			exec.EXPECT().
				RunCmd(gomock.Any(), "launchctl", []string{"print", "system/io.osapi.app"}).
				Return("", errors.New("could not find service")).
				AnyTimes()
			p, err := service.Behavior.Provider(behavior.PlatformDarwin, host.Host{
				FS:   memfs.New(),
				Exec: exec,
				Required: map[string]any{
					managedfile.Behavior.Name: files,
					object.Behavior.Name:      objectmocks.NewMockProvider(ctrl),
				},
			})
			s.Require().NoError(err)

			changed, err := p.(service.Provider).Act(context.Background(), action, "app")
			s.False(changed)
			s.ErrorContains(err, "enable it first")
		})
	}
}

// TestStartALoadedDaemon: a daemon that launchd loaded and that does not run
// starts with a kickstart alone, and an error of launchctl is the action's.
func (s *DarwinPublicTestSuite) TestStartALoadedDaemon() {
	tests := []struct {
		name      string
		err       error
		wantError string
	}{
		{name: "it kickstarts the daemon"},
		{
			name:      "launchctl that fails",
			err:       errors.New("exit status 113"),
			wantError: "launchctl kickstart: exit status 113",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			files := managedfilemocks.NewMockProvider(ctrl)
			files.EXPECT().Read(gomock.Any(), "/Library/LaunchDaemons/io.osapi.app.plist").
				Return(managedfile.Record{Path: "/Library/LaunchDaemons/io.osapi.app.plist"}, nil).
				AnyTimes()
			exec := hostmocks.NewMockExec(ctrl)
			exec.EXPECT().RunCmd(gomock.Any(), "launchctl", []string{"print", "system"}).
				Return("system = {\n\tservices = {\n\t\t       0      0 \tio.osapi.app\n\t}\n}\n", nil).
				AnyTimes()
			exec.EXPECT().
				RunCmd(gomock.Any(), "launchctl", []string{"print", "system/io.osapi.app"}).
				Return("system/io.osapi.app = {\n}\n", nil).
				AnyTimes()
			exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "launchctl",
				[]string{"kickstart", "system/io.osapi.app"}).Return("", tc.err)
			p, err := service.Behavior.Provider(behavior.PlatformDarwin, host.Host{
				FS:   memfs.New(),
				Exec: exec,
				Required: map[string]any{
					managedfile.Behavior.Name: files,
					object.Behavior.Name:      objectmocks.NewMockProvider(ctrl),
				},
			})
			s.Require().NoError(err)

			_, err = p.(service.Provider).Act(context.Background(), "start", "app")
			if tc.wantError != "" {
				s.ErrorContains(err, tc.wantError)

				return
			}
			s.NoError(err)
		})
	}
}

func TestDarwinPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DarwinPublicTestSuite))
}
