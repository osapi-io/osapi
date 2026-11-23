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

package power_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/power"
	"github.com/osapi-io/osapi/internal/behaviors/power/mocks"
)

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	mock *mocks.MockProvider
}

func (s *MethodsPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mock = mocks.NewMockProvider(s.ctrl)
}

func (s *MethodsPublicTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *MethodsPublicTestSuite) TestMethods() {
	tests := []struct {
		name    string
		call    func(context.Context, power.Provider, power.Input) (power.Result, error)
		expect  func(in power.Input) *gomock.Call
		out     power.Result
		err     error
		wantErr string
	}{
		{
			name: "when reboot asks the provider",
			call: power.Reboot,
			expect: func(in power.Input) *gomock.Call {
				return s.mock.EXPECT().Reboot(gomock.Any(), in)
			},
			out: power.Result{Action: "reboot", Delay: 60, Changed: true},
		},
		{
			name: "when shutdown fails, the error",
			call: power.Shutdown,
			expect: func(in power.Input) *gomock.Call {
				return s.mock.EXPECT().Shutdown(gomock.Any(), in)
			},
			err:     errors.New("boom"),
			wantErr: "boom",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			in := power.Input{Delay: 60}
			tc.expect(in).Return(tc.out, tc.err)
			got, err := tc.call(context.Background(), s.mock, in)
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.NoError(err)
			s.Equal(tc.out, got)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
