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

package ntp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/ntp"
	"github.com/osapi-io/osapi/internal/behaviors/ntp/mocks"
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

func (s *MethodsPublicTestSuite) TestGet() {
	status := ntp.Status{Synchronized: true, Servers: []string{"0.pool.ntp.org"}}
	tests := []struct {
		name string
		err  error
	}{
		{name: "when the provider reads the status"},
		{name: "when the provider fails", err: errors.New("chronyc: not running")},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.mock.EXPECT().Get(gomock.Any()).Return(status, tc.err)

			got, err := ntp.Get(context.Background(), s.mock, ntp.GetInput{})

			if tc.err != nil {
				s.ErrorIs(err, tc.err)

				return
			}
			s.Require().NoError(err)
			s.Equal(status, got)
		})
	}
}

func (s *MethodsPublicTestSuite) TestWrites() {
	servers := []string{"0.pool.ntp.org"}
	boom := errors.New("write failed")
	tests := []struct {
		name    string
		call    func() (ntp.Result, error)
		expect  func(changed bool, err error)
		changed bool
		err     error
	}{
		{
			name: "when create writes the servers",
			call: func() (ntp.Result, error) {
				return ntp.Create(context.Background(), s.mock, ntp.CreateInput{Servers: servers})
			},
			expect: func(changed bool, err error) {
				s.mock.EXPECT().Create(gomock.Any(), servers).Return(changed, err)
			},
			changed: true,
		},
		{
			name: "when update fails",
			call: func() (ntp.Result, error) {
				return ntp.Update(context.Background(), s.mock, ntp.UpdateInput{Servers: servers})
			},
			expect: func(changed bool, err error) {
				s.mock.EXPECT().Update(gomock.Any(), servers).Return(changed, err)
			},
			err: boom,
		},
		{
			name: "when delete removes the servers",
			call: func() (ntp.Result, error) {
				return ntp.Delete(context.Background(), s.mock, ntp.DeleteInput{})
			},
			expect: func(changed bool, err error) {
				s.mock.EXPECT().Delete(gomock.Any()).Return(changed, err)
			},
			changed: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.expect(tc.changed, tc.err)

			got, err := tc.call()

			s.ErrorIs(err, tc.err)
			s.Equal(ntp.Result{Changed: tc.changed}, got)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
