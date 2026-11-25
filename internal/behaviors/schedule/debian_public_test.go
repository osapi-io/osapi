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

package schedule_test

import (
	"context"
	"testing"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	filemocks "github.com/osapi-io/osapi/internal/behaviors/managedfile/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/object"
	objectmocks "github.com/osapi-io/osapi/internal/behaviors/object/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/schedule"
)

type DebianPublicTestSuite struct {
	suite.Suite

	ctrl    *gomock.Controller
	files   *filemocks.MockProvider
	objects *objectmocks.MockProvider
}

func (s *DebianPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.files = filemocks.NewMockProvider(s.ctrl)
	s.objects = objectmocks.NewMockProvider(s.ctrl)
}

// provider is the Debian provider on fs, with facts.
func (s *DebianPublicTestSuite) provider(
	fs *memfs.MemFS,
	facts func() map[string]any,
) schedule.Provider {
	p, err := schedule.NewDebian(host.Host{
		FS:       fs,
		Hostname: "web-01",
		Facts:    facts,
		Required: map[string]any{
			managedfile.Behavior.Name: s.files,
			object.Behavior.Name:      s.objects,
		},
	})
	s.Require().NoError(err)

	return p
}

func (s *DebianPublicTestSuite) TestList() {
	tests := []struct {
		name  string
		setup func(fs *memfs.MemFS)
	}{
		{
			name: "when /etc/cron.d holds a directory, it is not an entry",
			setup: func(fs *memfs.MemFS) {
				s.Require().NoError(fs.MkdirAll("/etc/cron.d/sub", 0o755))
			},
		},
		{
			name: "when osapi dropped the record of a file, it is not osapi's",
			setup: func(fs *memfs.MemFS) {
				s.Require().NoError(fs.MkdirAll("/etc/cron.d", 0o755))
				s.Require().NoError(fs.WriteFile("/etc/cron.d/old", []byte("x"), 0o644))
				s.files.EXPECT().Read(gomock.Any(), "/etc/cron.d/old").
					Return(managedfile.Record{Path: "/etc/cron.d/old", Undeployed: true}, nil)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			fs := memfs.New()
			tc.setup(fs)
			got, err := s.provider(fs, nil).List(context.Background())
			s.Require().NoError(err)
			s.Empty(got)
		})
	}
}

// TestCreateRendersTheFacts: a template reads the agent's facts, as the
// file deployer's did. The record keeps the object and its checksum, for the
// stale check.
func (s *DebianPublicTestSuite) TestCreateRendersTheFacts() {
	s.files.EXPECT().Read(gomock.Any(), gomock.Any()).
		Return(managedfile.Record{}, managedfile.ErrNotManaged).AnyTimes()
	s.objects.EXPECT().Object(gomock.Any(), "arch.tmpl").
		Return(object.Object{Name: "arch.tmpl", Content: "# {{ .Facts.architecture }}\n", ContentType: "template", SHA256: "object"}, nil)
	s.files.EXPECT().Write(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, f managedfile.File) (bool, error) {
			s.Equal("/etc/cron.d/arch", f.Path)
			s.Equal("# arm64\n", string(f.Content))
			s.Equal("arch.tmpl", f.Object)
			s.Equal(types.ContentTemplate, f.ContentType)
			s.Equal("object", f.ObjectSHA256)

			return true, nil
		})

	changed, err := s.provider(memfs.New(), func() map[string]any {
		return map[string]any{"architecture": "arm64"}
	}).Create(context.Background(), schedule.CreateInput{Name: "arch", Object: "arch.tmpl", Schedule: "0 2 * * *"})
	s.Require().NoError(err)
	s.True(changed)
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
