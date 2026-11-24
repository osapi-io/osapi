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
// Package password is the password behavior: the password of a user, under
package password_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/password"
	"github.com/osapi-io/osapi/internal/behaviors/password/mocks"
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
	tests := []struct {
		name string
		err  error
		want password.Result
	}{
		{
			name: "when the provider sets the password, it changed",
			want: password.Result{Name: "alice", Changed: true},
		},
		{
			name: "when the provider fails",
			err:  errors.New("chpasswd failed"),
			want: password.Result{Name: "alice"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.mock.EXPECT().Set(gomock.Any(), "alice", "$6$x$y").Return(tc.err)

			got, err := password.Update(
				context.Background(),
				s.mock,
				password.UpdateInput{Name: "alice", PasswordHash: "$6$x$y"},
			)

			if tc.err != nil {
				s.ErrorIs(err, tc.err)
			} else {
				s.Require().NoError(err)
			}
			s.Equal(tc.want, got)
		})
	}
}

// TestHashPassword: the controller step leaves the job only the hash of the
// password, and a call with no password stays as it is.
func (s *MethodsPublicTestSuite) TestHashPassword() {
	tests := []struct {
		name     string
		in       password.UpdateInput
		hashErr  error
		wantHash bool
		wantErr  string
	}{
		{
			name:     "when the call has a password, the job has its hash",
			in:       password.UpdateInput{Name: "alice", Password: "pw"},
			wantHash: true,
		},
		{
			name: "when the call has no password, it stays as it is",
			in:   password.UpdateInput{Name: "alice"},
		},
		{
			name:    "when the hash fails",
			in:      password.UpdateInput{Name: "alice", Password: "pw"},
			hashErr: errors.New("no salt"),
			wantErr: "no salt",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.hashErr != nil {
				defer password.SetHashFn(func(string) (string, error) { return "", tc.hashErr })()
			}

			got, err := password.HashPassword(context.Background(), tc.in)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Empty(got.Password)
			s.Equal(tc.wantHash, strings.HasPrefix(got.PasswordHash, "$6$"))
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
