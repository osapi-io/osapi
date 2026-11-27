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

package certificate_test

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
	"github.com/osapi-io/osapi/internal/behaviors/certificate"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	managedfilemocks "github.com/osapi-io/osapi/internal/behaviors/managedfile/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/object"
	objectmocks "github.com/osapi-io/osapi/internal/behaviors/object/mocks"
)

const (
	caPath = "/usr/local/share/ca-certificates/osapi-internal-ca.crt"
	caPEM  = "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n"
)

// WriterPublicTestSuite covers what no example can: update-ca-certificates
// failing, which a fake host's commands do not do.
type WriterPublicTestSuite struct {
	suite.Suite

	files   *managedfilemocks.MockProvider
	objects *objectmocks.MockProvider
	exec    *hostmocks.MockExec
	fs      avfs.VFS
}

func (s *WriterPublicTestSuite) SetupSubTest() {
	ctrl := gomock.NewController(s.T())
	s.files = managedfilemocks.NewMockProvider(ctrl)
	s.objects = objectmocks.NewMockProvider(ctrl)
	s.exec = hostmocks.NewMockExec(ctrl)
	s.fs = memfs.New()
	s.Require().NoError(s.fs.MkdirAll("/usr/local/share/ca-certificates", 0o755))
}

func (s *WriterPublicTestSuite) writer() certificate.Provider {
	p, err := certificate.NewWriter(host.Host{
		Platform: behavior.PlatformDebian,
		FS:       s.fs,
		Exec:     s.exec,
		Required: map[string]any{
			managedfile.Behavior.Name: s.files,
			object.Behavior.Name:      s.objects,
		},
	})
	s.Require().NoError(err)

	return p
}

// marked matches a write of the CA whose record carries the mark of a
// rebuild that did not finish, beside the object it came from.
func marked() gomock.Matcher {
	return gomock.Cond(func(f managedfile.File) bool {
		return f.Path == caPath && string(f.Content) == caPEM &&
			f.Object == "internal-ca.pem" && f.ObjectSHA256 == "abc" &&
			maps.Equal(f.Metadata, map[string]string{
				"object": "internal-ca.pem", managedfile.Unapplied: "true",
			})
	})
}

// TestNoObjectStore: an agent with file state and no object store reads only.
// A deploy is unsupported, not a call on a nil object provider, which
// panicked.
func (s *WriterPublicTestSuite) TestNoObjectStore() {
	s.Run("a deploy is unsupported", func() {
		p, err := certificate.NewWriter(host.Host{
			Platform: behavior.PlatformDebian,
			FS:       s.fs,
			Exec:     s.exec,
			Required: map[string]any{managedfile.Behavior.Name: s.files},
		})
		s.Require().NoError(err)

		s.NotPanics(func() {
			_, err = certificate.Create(context.Background(), p,
				certificate.CreateInput{Name: "internal-ca", Object: "internal-ca.pem"})
		})
		s.ErrorIs(err, behavior.ErrUnsupported)
	})
}

// TestRetryAfterAFailedDeploy: a CA whose file landed and whose rebuild of
// the trust store failed is rebuilt by a retry with the same input. The
// record of the file is marked, so the retry's write changes; managedfile's
// own test shows that a marked record makes a write of the same content
// change. A call after a rebuild that worked changes nothing.
func (s *WriterPublicTestSuite) TestRetryAfterAFailedDeploy() {
	tests := []struct {
		name        string
		written     bool
		rebuild     error
		wantChanged bool
		wantError   string
	}{
		{
			name:      "when the rebuild fails, the record is marked",
			written:   true,
			rebuild:   errors.New("cannot link"),
			wantError: "update-ca-certificates: cannot link",
		},
		{
			name:        "when the retry's write changes the marked record, it rebuilds",
			written:     true,
			wantChanged: true,
		},
		{
			name: "when the trust store was rebuilt, it changes nothing",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.files.EXPECT().Read(gomock.Any(), caPath).Return(managedfile.Record{
				Path: caPath, Metadata: map[string]string{"object": "internal-ca.pem"},
			}, nil)
			s.objects.EXPECT().Object(gomock.Any(), "internal-ca.pem").
				Return(object.Object{Content: caPEM, SHA256: "abc"}, nil)
			s.files.EXPECT().Write(gomock.Any(), gomock.Cond(func(f managedfile.File) bool {
				return f.Metadata[managedfile.Unapplied] == ""
			})).Return(tc.written, nil)
			if tc.written {
				s.exec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "update-ca-certificates", gomock.Any()).
					Return("", tc.rebuild)
			}
			if tc.rebuild != nil {
				s.files.EXPECT().Write(gomock.Any(), marked()).Return(true, nil)
			}

			changed, err := s.writer().Deploy(
				context.Background(), "internal-ca", "internal-ca.pem", false)

			s.Equal(tc.wantChanged, changed)
			if tc.wantError != "" {
				s.EqualError(err, tc.wantError)

				return
			}
			s.NoError(err)
		})
	}
}

// TestRetryAfterAFailedRemove: when the rebuild fails after a remove, the
// file is put back with its record marked, so a retry removes it and
// rebuilds again. Without the file, nothing can be put back.
func (s *WriterPublicTestSuite) TestRetryAfterAFailedRemove() {
	tests := []struct {
		name      string
		onDisk    bool
		wantError string
	}{
		{
			name:      "when the file is there, it is put back",
			onDisk:    true,
			wantError: "update-ca-certificates: cannot link, so " + caPath + " was put back",
		},
		{
			name:      "when the file cannot be read, it is not put back",
			wantError: "update-ca-certificates: cannot link",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.onDisk {
				s.Require().NoError(s.fs.WriteFile(caPath, []byte(caPEM), 0o644))
			}
			s.files.EXPECT().Read(gomock.Any(), caPath).Return(managedfile.Record{
				Path: caPath, Metadata: map[string]string{"object": "internal-ca.pem"},
				ObjectSHA256: "abc",
			}, nil)
			s.files.EXPECT().Remove(gomock.Any(), caPath).Return(true, nil)
			s.exec.EXPECT().RunPrivilegedCmd(
				gomock.Any(), "update-ca-certificates", []string{"--fresh"},
			).Return("", errors.New("cannot link"))
			if tc.onDisk {
				s.files.EXPECT().Write(gomock.Any(), marked()).Return(true, nil)
			}

			changed, err := s.writer().Remove(context.Background(), "internal-ca")

			s.False(changed)
			s.EqualError(err, tc.wantError)
		})
	}
}

func TestWriterPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(WriterPublicTestSuite))
}
