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
	"maps"
	"testing"

	"github.com/avfs/avfs"
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

const (
	unitPath = "/etc/systemd/system/osapi-app.service"
	unitText = "[Service]\nExecStart=/usr/local/bin/app\n"
)

// DebianPublicTestSuite covers what no example can: systemctl
// daemon-reload failing, which a fake host's commands do not do.
type DebianPublicTestSuite struct {
	suite.Suite

	files   *managedfilemocks.MockProvider
	objects *objectmocks.MockProvider
	exec    *hostmocks.MockExec
	fs      avfs.VFS
}

func (s *DebianPublicTestSuite) SetupSubTest() {
	ctrl := gomock.NewController(s.T())
	s.files = managedfilemocks.NewMockProvider(ctrl)
	s.objects = objectmocks.NewMockProvider(ctrl)
	s.exec = hostmocks.NewMockExec(ctrl)
	s.fs = memfs.New()
	s.Require().NoError(s.fs.MkdirAll("/etc/systemd/system", 0o755))
}

func (s *DebianPublicTestSuite) provider() service.Provider {
	p, err := service.Behavior.Provider(behavior.PlatformDebian, host.Host{
		FS:   s.fs,
		Exec: s.exec,
		Required: map[string]any{
			managedfile.Behavior.Name: s.files,
			object.Behavior.Name:      s.objects,
		},
	})
	s.Require().NoError(err)

	return p.(service.Provider)
}

// marked matches a write of the unit whose record carries the mark of a
// reload that did not finish, beside the object it came from.
func marked() gomock.Matcher {
	return gomock.Cond(func(f managedfile.File) bool {
		return f.Path == unitPath && string(f.Content) == unitText &&
			f.Object == "app.service" && f.ObjectSHA256 == "abc" &&
			maps.Equal(f.Metadata, map[string]string{
				"object": "app.service", managedfile.Unapplied: "true",
			})
	})
}

// TestRetryAfterAFailedDeploy: a unit whose file landed and whose reload of
// systemd failed is reloaded by a retry with the same input. The record of
// the file is marked, so the retry's write changes; managedfile's own test
// shows that a marked record makes a write of the same content change. A
// call after a reload that worked changes nothing.
func (s *DebianPublicTestSuite) TestRetryAfterAFailedDeploy() {
	tests := []struct {
		name        string
		written     bool
		reload      error
		wantChanged bool
		wantError   string
	}{
		{
			name:      "when the reload fails, the record is marked",
			written:   true,
			reload:    errors.New("bus is down"),
			wantError: "systemctl daemon-reload: bus is down",
		},
		{
			name:        "when the retry's write changes the marked record, it reloads",
			written:     true,
			wantChanged: true,
		},
		{
			name: "when systemd was reloaded, it changes nothing",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.files.EXPECT().Read(gomock.Any(), unitPath).Return(managedfile.Record{
				Path: unitPath, Metadata: map[string]string{"object": "app.service"},
			}, nil)
			s.objects.EXPECT().Object(gomock.Any(), "app.service").
				Return(object.Object{Content: unitText, SHA256: "abc"}, nil)
			s.files.EXPECT().Write(gomock.Any(), gomock.Cond(func(f managedfile.File) bool {
				return f.Metadata[managedfile.Unapplied] == ""
			})).Return(tc.written, nil)
			if tc.written {
				s.exec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "systemctl", []string{"daemon-reload"}).
					Return("", tc.reload)
			}
			if tc.reload != nil {
				s.files.EXPECT().Write(gomock.Any(), marked()).Return(true, nil)
			}

			changed, err := s.provider().Deploy(
				context.Background(), service.Unit{Name: "app", Object: "app.service"}, false)

			s.Equal(tc.wantChanged, changed)
			if tc.wantError != "" {
				s.EqualError(err, tc.wantError)

				return
			}
			s.NoError(err)
		})
	}
}

// TestRetryAfterAFailedRemove: when the reload fails after a remove, the
// unit file is put back with its record marked, so a retry removes it and
// reloads again. Without the file, nothing can be put back.
func (s *DebianPublicTestSuite) TestRetryAfterAFailedRemove() {
	tests := []struct {
		name      string
		onDisk    bool
		wantError string
	}{
		{
			name:      "when the file is there, it is put back",
			onDisk:    true,
			wantError: "systemctl daemon-reload: bus is down, so " + unitPath + " was put back",
		},
		{
			name:      "when the file cannot be read, it is not put back",
			wantError: "systemctl daemon-reload: bus is down",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.onDisk {
				s.Require().NoError(s.fs.WriteFile(unitPath, []byte(unitText), 0o644))
			} else {
				// A directory at the path is there, and cannot be read as
				// a file.
				s.Require().NoError(s.fs.Mkdir(unitPath, 0o755))
			}
			s.files.EXPECT().Read(gomock.Any(), unitPath).Return(managedfile.Record{
				Path: unitPath, Metadata: map[string]string{"object": "app.service"},
				ObjectSHA256: "abc",
			}, nil)
			s.exec.EXPECT().RunPrivilegedCmd(
				gomock.Any(), "systemctl", []string{"disable", "--now", "osapi-app.service"},
			)
			s.files.EXPECT().Remove(gomock.Any(), unitPath).Return(true, nil)
			s.exec.EXPECT().RunPrivilegedCmd(gomock.Any(), "systemctl", []string{"daemon-reload"}).
				Return("", errors.New("bus is down"))
			if tc.onDisk {
				s.files.EXPECT().Write(gomock.Any(), marked()).Return(true, nil)
			}

			changed, err := s.provider().Remove(context.Background(), "app")

			s.False(changed)
			s.EqualError(err, tc.wantError)
		})
	}
}

// TestRemoveWithoutTheUnitFile: a record of a unit whose file is not on the
// host, such as the record that a VM built again with the same name keeps,
// is dropped, and the remove changes nothing. systemctl is not run, because
// it fails on a unit file that is not there.
func (s *DebianPublicTestSuite) TestRemoveWithoutTheUnitFile() {
	tests := []struct {
		name      string
		remove    error
		wantError string
	}{
		{
			name: "when the unit file is not there, the record is dropped",
		},
		{
			name:      "when the record cannot be dropped, it fails",
			remove:    errors.New("update state: kv is down"),
			wantError: "update state: kv is down",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.files.EXPECT().Read(gomock.Any(), unitPath).Return(managedfile.Record{
				Path: unitPath, Metadata: map[string]string{"object": "app.service"},
			}, nil)
			s.files.EXPECT().Remove(gomock.Any(), unitPath).Return(false, tc.remove)

			changed, err := s.provider().Remove(context.Background(), "app")

			s.False(changed)
			if tc.wantError != "" {
				s.EqualError(err, tc.wantError)

				return
			}
			s.NoError(err)
		})
	}
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
