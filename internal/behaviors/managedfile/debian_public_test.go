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
	state *jobmocks.MockKeyValue
}

func (s *DebianPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.state = jobmocks.NewMockKeyValue(s.ctrl)
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
				FileState: jobmocks.NewMockKeyValue(gomock.NewController(s.T())),
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
			p := managedfile.NewDebian(memfs.New(), s.state, nil, "web-01")

			_, err := p.Write(
				context.Background(),
				managedfile.File{Path: path, Content: []byte("on\n")},
			)

			s.EqualError(err, tc.wantErr)
		})
	}
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
			s.state.EXPECT().Get(gomock.Any(), gomock.Any()).Return(entry, nil)
			managedfile.SetMarshalJSON(failMarshal)
			p := managedfile.NewDebian(memfs.New(), s.state, nil, "web-01")

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
