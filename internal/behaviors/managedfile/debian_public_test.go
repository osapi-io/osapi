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

package managedfile_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	providermocks "github.com/osapi-io/osapi/internal/provider/mocks"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

const path = "/etc/osapi/example.conf"

type DebianPublicTestSuite struct {
	suite.Suite

	ctrl  *gomock.Controller
	state *providermocks.MockFileState
}

func (s *DebianPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.state = providermocks.NewMockFileState(s.ctrl)
}

func (s *DebianPublicTestSuite) TearDownSubTest() {
	managedfile.ResetMarshalJSON()
}

func failMarshal(any) ([]byte, error) { return nil, errors.New("cannot encode") }

func (s *DebianPublicTestSuite) TestNewDebian() {
	tests := []struct {
		name    string
		h       host.Host
		wantErr error
	}{
		{
			name:    "when the host has no file state store, it declines",
			h:       host.Host{Platform: "debian"},
			wantErr: behavior.ErrUnsupported,
		},
		{
			name: "when it has one, it builds",
			h: host.Host{
				Platform:  "debian",
				FS:        memfs.New(),
				FileState: providermocks.NewMockFileState(gomock.NewController(s.T())),
			},
		},
		{
			name:    "when a Mac has no file state store, it declines",
			h:       host.Host{Platform: "darwin"},
			wantErr: behavior.ErrUnsupported,
		},
		{
			name: "when a Mac has one, it builds",
			h: host.Host{
				Platform:  "darwin",
				FS:        memfs.New(),
				FileState: providermocks.NewMockFileState(gomock.NewController(s.T())),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			p, err := host.Build(managedfile.Behavior, tc.h)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)

				return
			}
			s.NoError(err)
			s.NotNil(p)
		})
	}
}

func (s *DebianPublicTestSuite) TestWrite() {
	tests := []struct {
		name    string
		wantErr string
	}{
		{name: "when the record cannot be encoded", wantErr: "marshal state: cannot encode"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			managedfile.SetMarshalJSON(failMarshal)
			p := managedfile.NewDebian(memfs.New(), s.state, "file-state", nil, "web-01")

			_, err := p.Write(
				context.Background(),
				managedfile.File{Path: path, Content: []byte("on\n")},
			)

			s.EqualError(err, tc.wantErr)
		})
	}
}

// TestWriteOfTheSameContent: a file that has the content changes only when
// osapi deployed it with other metadata. A path with no record, an unreadable
// one or an undeployed one stays as it is.
func (s *DebianPublicTestSuite) TestWriteOfTheSameContent() {
	undeployed, _ := json.Marshal(
		managedfile.State{Path: path, UndeployedAt: "2026-11-22T00:00:00Z"},
	)
	tests := []struct {
		name   string
		record []byte
		err    error
	}{
		{name: "when the path has no record, it does not change", err: errors.New("not found")},
		{name: "when the record cannot be read, it does not change", record: []byte("{")},
		{name: "when the record is undeployed, it does not change", record: undeployed},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			fs := memfs.New()
			s.Require().NoError(fs.MkdirAll("/etc/osapi", 0o755))
			s.Require().NoError(fs.WriteFile(path, []byte("on\n"), 0o644))
			var entry *jobmocks.MockKeyValueEntry
			if tc.err == nil {
				entry = jobmocks.NewMockKeyValueEntry(s.ctrl)
				entry.EXPECT().Value().Return(tc.record)
			}
			s.state.EXPECT().KVGet(gomock.Any(), "file-state", gomock.Any()).Return(entry, tc.err)
			p := managedfile.NewDebian(fs, s.state, "file-state", nil, "web-01")

			changed, err := p.Write(
				context.Background(),
				managedfile.File{
					Path:     path,
					Content:  []byte("on\n"),
					Metadata: map[string]string{"a": "b"},
				},
			)

			s.NoError(err)
			s.False(changed)
		})
	}
}

// TestWriteOfANewObject: a file that has the content changes when osapi
// deployed it from another object, or from the object as it was, such as a
// template whose object changed and renders the same file. The record then
// keeps the object's checksum, which the stale check compares, and a
// secret's record keeps none.
func (s *DebianPublicTestSuite) TestWriteOfANewObject() {
	record := func(st managedfile.State) []byte {
		data, _ := json.Marshal(st)

		return data
	}
	tests := []struct {
		name        string
		record      []byte
		secret      bool
		wantChanged bool
		wantSHA     string
	}{
		{
			name: "when the record has the object as it is, it does not change",
			record: record(managedfile.State{
				Path: path, ObjectName: "a.tmpl", ContentType: "template", ObjectSHA256: "new",
			}),
		},
		{
			name: "when the object changed, the record gets its checksum",
			record: record(managedfile.State{
				Path: path, ObjectName: "a.tmpl", ContentType: "template", ObjectSHA256: "old",
			}),
			wantChanged: true,
			wantSHA:     "new",
		},
		{
			name: "when the record names another object, it changes",
			record: record(managedfile.State{
				Path: path, ObjectName: "b.tmpl", ContentType: "template", ObjectSHA256: "new",
			}),
			wantChanged: true,
			wantSHA:     "new",
		},
		{
			name: "when the record has another content type, it changes",
			record: record(managedfile.State{
				Path: path, ObjectName: "a.tmpl", ContentType: "raw", ObjectSHA256: "new",
			}),
			wantChanged: true,
			wantSHA:     "new",
		},
		{
			name: "when the record was written before object checksums, it gets one",
			record: record(managedfile.State{
				Path: path, ObjectName: "a.tmpl", ContentType: "template",
			}),
			wantChanged: true,
			wantSHA:     "new",
		},
		{
			name: "when the content is a secret, the record keeps no object checksum",
			record: record(managedfile.State{
				Path: path, ObjectName: "a.tmpl", ContentType: "template",
			}),
			secret: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			fs := memfs.New()
			s.Require().NoError(fs.MkdirAll("/etc/osapi", 0o755))
			s.Require().NoError(fs.WriteFile(path, []byte("on\n"), 0o644))
			entry := jobmocks.NewMockKeyValueEntry(s.ctrl)
			entry.EXPECT().Value().Return(tc.record)
			s.state.EXPECT().KVGet(gomock.Any(), "file-state", gomock.Any()).Return(entry, nil)
			if tc.wantChanged {
				s.state.EXPECT().KVPut(gomock.Any(), "file-state", gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _ string, data []byte) (uint64, error) {
						var st managedfile.State
						s.Require().NoError(json.Unmarshal(data, &st))
						s.Equal(tc.wantSHA, st.ObjectSHA256)
						s.Equal("a.tmpl", st.ObjectName)

						return 1, nil
					})
			}
			p := managedfile.NewDebian(fs, s.state, "file-state", nil, "web-01")

			changed, err := p.Write(
				context.Background(),
				managedfile.File{
					Path:         path,
					Content:      []byte("on\n"),
					Object:       "a.tmpl",
					ContentType:  "template",
					ObjectSHA256: "new",
					Secret:       tc.secret,
				},
			)

			s.NoError(err)
			s.Equal(tc.wantChanged, changed)
		})
	}
}

// TestWriteOfAnUnappliedFile: a file whose record a caller marked, as its
// step after the write did not finish, changes on a write of the same
// content, so the caller runs the step again. The new record drops the mark.
func (s *DebianPublicTestSuite) TestWriteOfAnUnappliedFile() {
	meta := map[string]string{"setting": "on"}
	record, _ := json.Marshal(managedfile.State{
		Path:     path,
		SHA256:   "b6ac5bc39377e44ee783dff946f6703a994f5eb31a438e69dc0735b40060d2e8",
		Metadata: managedfile.MarkUnapplied(managedfile.UpdateInput{Metadata: meta}).Metadata,
	})
	fs := memfs.New()
	s.Require().NoError(fs.MkdirAll("/etc/osapi", 0o755))
	s.Require().NoError(fs.WriteFile(path, []byte("on\n"), 0o644))
	ctrl := gomock.NewController(s.T())
	state := providermocks.NewMockFileState(ctrl)
	entry := jobmocks.NewMockKeyValueEntry(ctrl)
	entry.EXPECT().Value().Return(record)
	state.EXPECT().KVGet(gomock.Any(), "file-state", gomock.Any()).Return(entry, nil)
	state.EXPECT().KVPut(gomock.Any(), "file-state", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, data []byte) (uint64, error) {
			var st managedfile.State
			s.Require().NoError(json.Unmarshal(data, &st))
			s.Equal(meta, st.Metadata)

			return 1, nil
		})
	p := managedfile.NewDebian(fs, state, "file-state", nil, "web-01")

	changed, err := p.Write(
		context.Background(),
		managedfile.File{Path: path, Content: []byte("on\n"), Metadata: meta},
	)

	s.NoError(err)
	s.True(changed)
}

// TestReadOfAnObjectChecksum: the record gives the checksum of the object
// that the file came from, so a caller that puts the file back keeps it.
func (s *DebianPublicTestSuite) TestReadOfAnObjectChecksum() {
	record, _ := json.Marshal(managedfile.State{Path: path, ObjectSHA256: "abc"})
	ctrl := gomock.NewController(s.T())
	state := providermocks.NewMockFileState(ctrl)
	entry := jobmocks.NewMockKeyValueEntry(ctrl)
	entry.EXPECT().Value().Return(record)
	state.EXPECT().KVGet(gomock.Any(), "file-state", gomock.Any()).Return(entry, nil)
	p := managedfile.NewDebian(memfs.New(), state, "file-state", nil, "web-01")

	got, err := p.Read(context.Background(), path)

	s.NoError(err)
	s.Equal("abc", got.ObjectSHA256)
}

func (s *DebianPublicTestSuite) TestRemove() {
	tests := []struct {
		name    string
		wantErr string
	}{
		{name: "when the record cannot be encoded", wantErr: "marshal state: cannot encode"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			record, _ := json.Marshal(managedfile.State{Path: path})
			entry := jobmocks.NewMockKeyValueEntry(s.ctrl)
			entry.EXPECT().Value().Return(record)
			s.state.EXPECT().KVGet(gomock.Any(), "file-state", gomock.Any()).Return(entry, nil)
			managedfile.SetMarshalJSON(failMarshal)
			p := managedfile.NewDebian(memfs.New(), s.state, "file-state", nil, "web-01")

			_, err := p.Remove(context.Background(), path)

			s.EqualError(err, tc.wantErr)
		})
	}
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
