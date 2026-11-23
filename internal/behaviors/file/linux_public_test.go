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

package file_test

import (
	"context"
	"errors"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/file"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	mfmocks "github.com/osapi-io/osapi/internal/behaviors/managedfile/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/object"
	objmocks "github.com/osapi-io/osapi/internal/behaviors/object/mocks"
)

type LinuxPublicTestSuite struct {
	suite.Suite

	ctrl    *gomock.Controller
	objects *objmocks.MockProvider
	files   *mfmocks.MockProvider
}

func (s *LinuxPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.objects = objmocks.NewMockProvider(s.ctrl)
	s.files = mfmocks.NewMockProvider(s.ctrl)
}

func (s *LinuxPublicTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *LinuxPublicTestSuite) TestDeploy() {
	tests := []struct {
		name         string
		in           file.UpdateInput
		facts        map[string]any
		setup        func()
		validateFunc func(file.UpdateResult, error)
	}{
		{
			name: "when the object has no content type it is written as it is",
			in:   file.UpdateInput{ObjectName: "a", Path: "/etc/a"},
			setup: func() {
				s.objects.EXPECT().
					Object(gomock.Any(), "a").
					Return(object.Object{Name: "a", Content: "x\n", SHA256: "sum"}, nil)
				s.files.EXPECT().Write(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, f managedfile.File) (bool, error) {
						s.Equal("raw", f.ContentType)
						s.Equal("sum", f.ObjectSHA256)
						s.Equal("a", f.Object)
						s.Equal("x\n", string(f.Content))

						return true, nil
					})
			},
			validateFunc: func(r file.UpdateResult, err error) {
				s.Require().NoError(err)
				s.True(r.Changed)
			},
		},
		{
			name:  "when a template reads the facts it renders them",
			in:    file.UpdateInput{ObjectName: "t", Path: "/etc/t", ContentType: "template"},
			facts: map[string]any{"architecture": "arm64"},
			setup: func() {
				s.objects.EXPECT().Object(gomock.Any(), "t").Return(
					object.Object{
						Name: "t", Content: "{{ .Facts.architecture }}", SHA256: "object",
					}, nil)
				s.files.EXPECT().Write(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, f managedfile.File) (bool, error) {
						s.Equal("arm64", string(f.Content))
						s.Equal("object", f.ObjectSHA256)

						return false, nil
					})
			},
			validateFunc: func(r file.UpdateResult, err error) {
				s.Require().NoError(err)
				s.False(r.Changed)
			},
		},
		{
			name: "when the object is not there it fails",
			in:   file.UpdateInput{ObjectName: "a", Path: "/etc/a"},
			setup: func() {
				s.objects.EXPECT().
					Object(gomock.Any(), "a").
					Return(object.Object{}, errors.New("not found"))
			},
			validateFunc: func(_ file.UpdateResult, err error) { s.ErrorContains(err, "not found") },
		},
		{
			name: "when the template does not render it fails",
			in:   file.UpdateInput{ObjectName: "t", Path: "/etc/t", ContentType: "template"},
			setup: func() {
				s.objects.EXPECT().
					Object(gomock.Any(), "t").
					Return(object.Object{Content: "{{ .Vars.missing }}"}, nil)
			},
			validateFunc: func(_ file.UpdateResult, err error) { s.ErrorContains(err, "render template") },
		},
		{
			name: "when the write fails it fails",
			in:   file.UpdateInput{ObjectName: "a", Path: "/etc/a"},
			setup: func() {
				s.objects.EXPECT().
					Object(gomock.Any(), "a").
					Return(object.Object{Content: "x"}, nil)
				s.files.EXPECT().
					Write(gomock.Any(), gomock.Any()).
					Return(false, errors.New("disk full"))
			},
			validateFunc: func(_ file.UpdateResult, err error) { s.ErrorContains(err, "disk full") },
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			p := file.NewLinux(s.objects, s.files, memfs.New(), "web-01", tc.facts)
			tc.validateFunc(p.Deploy(context.Background(), tc.in))
		})
	}
}

func (s *LinuxPublicTestSuite) TestUndeploy() {
	s.files.EXPECT().Remove(gomock.Any(), "/etc/a").Return(false, errors.New("busy"))
	p := file.NewLinux(s.objects, s.files, memfs.New(), "web-01", nil)
	_, err := p.Undeploy(context.Background(), "/etc/a")
	s.ErrorContains(err, "busy")
}

func (s *LinuxPublicTestSuite) TestStatus() {
	tests := []struct {
		name      string
		onDisk    string
		record    managedfile.Record
		recordErr error
		wantState string
		wantErr   string
	}{
		{
			name:   "when the file changed after the deploy it drifted",
			onDisk: "changed by hand\n",
			record: managedfile.Record{
				SHA256: "0083dacc561dfd01081d8554269e201f4b60fe597ab9892977bfde81da46ef60",
			},
			wantState: file.Drifted,
		},
		{
			name:      "when osapi undeployed it it is missing",
			record:    managedfile.Record{Undeployed: true},
			wantState: file.Missing,
		},
		{
			name:      "when the record is there and the file is gone it is missing",
			record:    managedfile.Record{SHA256: "x"},
			wantState: file.Missing,
		},
		{
			name:      "when the record cannot be read it fails",
			recordErr: managedfile.ErrUnreadable,
			wantErr:   "record unreadable",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			fs := memfs.New()
			if tc.onDisk != "" {
				s.Require().NoError(fs.MkdirAll("/etc", 0o755))
				s.Require().NoError(fs.WriteFile("/etc/a", []byte(tc.onDisk), 0o644))
			}
			s.files.EXPECT().Read(gomock.Any(), "/etc/a").Return(tc.record, tc.recordErr)
			got, err := file.NewLinux(s.objects, s.files, fs, "web-01", nil).
				Status(context.Background(), "/etc/a")
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.wantState, got.State)
		})
	}
}

func TestLinuxPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LinuxPublicTestSuite))
}
