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
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile/mocks"
)

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	mock *mocks.MockProvider
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mock = mocks.NewMockProvider(s.ctrl)
}

func (s *MethodsPublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

func (s *MethodsPublicTestSuite) TestUpdate() {
	meta := map[string]string{"owner": "sysctl"}
	tests := []struct {
		name     string
		in       managedfile.UpdateInput
		wantFile *managedfile.File
		changed  bool
		err      error
		want     managedfile.Result
		wantErr  string
	}{
		{
			name: "when the provider writes the file",
			in: managedfile.UpdateInput{
				Path:     "/etc/motd",
				Content:  "hello\n",
				Metadata: meta,
			},
			wantFile: &managedfile.File{
				Path:     "/etc/motd",
				Content:  []byte("hello\n"),
				Metadata: meta,
			},
			changed: true,
			want:    managedfile.Result{Path: "/etc/motd", Changed: true},
		},
		{
			name: "when a mode, owner and group are given, they reach the provider",
			in: managedfile.UpdateInput{
				Path: "/etc/myapp/token", Content: "s\n", Mode: "0600",
				Owner: "root", Group: "root", Secret: true,
			},
			wantFile: &managedfile.File{
				Path: "/etc/myapp/token", Content: []byte("s\n"), Mode: 0o600,
				Owner: "root", Group: "root", Secret: true,
			},
			want: managedfile.Result{Path: "/etc/myapp/token"},
		},
		{
			name:     "when the provider fails",
			in:       managedfile.UpdateInput{Path: "/etc/motd", Content: "hello\n"},
			wantFile: &managedfile.File{Path: "/etc/motd", Content: []byte("hello\n")},
			err:      errors.New("disk full"),
			wantErr:  "disk full",
		},
		{
			name:    "when the mode is not octal, nothing is written",
			in:      managedfile.UpdateInput{Path: "/etc/motd", Content: "x", Mode: "rw-------"},
			wantErr: "invalid mode",
		},
		{
			name:    "when the mode is out of range, nothing is written",
			in:      managedfile.UpdateInput{Path: "/etc/motd", Content: "x", Mode: "17777"},
			wantErr: "invalid mode",
		},
		{
			name:    "when the group is not a group name, nothing is written",
			in:      managedfile.UpdateInput{Path: "/etc/motd", Content: "x", Group: "-g"},
			wantErr: "invalid owner or group",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.wantFile != nil {
				s.mock.EXPECT().Write(gomock.Any(), *tc.wantFile).Return(tc.changed, tc.err)
			}

			got, err := managedfile.Update(context.Background(), s.mock, tc.in)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *MethodsPublicTestSuite) TestDelete() {
	tests := []struct {
		name    string
		changed bool
		err     error
		want    managedfile.Result
	}{
		{
			name:    "when the provider removes the file",
			changed: true,
			want:    managedfile.Result{Path: "/etc/motd", Changed: true},
		},
		{
			name: "when there was no file to remove",
			want: managedfile.Result{Path: "/etc/motd"},
		},
		{
			name: "when the provider fails",
			err:  managedfile.ErrUnreadable,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.mock.EXPECT().Remove(gomock.Any(), "/etc/motd").Return(tc.changed, tc.err)

			got, err := managedfile.Delete(
				context.Background(),
				s.mock,
				managedfile.DeleteInput{Path: "/etc/motd"},
			)

			if tc.err != nil {
				s.ErrorIs(err, tc.err)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name   string
		record managedfile.Record
		err    error
	}{
		{
			name:   "when osapi keeps a record of the file",
			record: managedfile.Record{Path: "/etc/motd"},
		},
		{
			name: "when osapi does not manage the file",
			err:  managedfile.ErrNotManaged,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.mock.EXPECT().Read(gomock.Any(), "/etc/motd").Return(tc.record, tc.err)

			got, err := managedfile.Get(
				context.Background(),
				s.mock,
				managedfile.GetInput{Path: "/etc/motd"},
			)

			if tc.err != nil {
				s.ErrorIs(err, tc.err)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.record, got)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
