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

package uptime_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/uptime"
	"github.com/osapi-io/osapi/internal/behaviors/uptime/mocks"
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
	tests := []struct {
		name    string
		uptime  time.Duration
		err     error
		want    uptime.Entry
		wantErr string
	}{
		{
			name:   "when the host has run for days, each unit is plural but one",
			uptime: 26*time.Hour + 5*time.Minute,
			want:   uptime.Entry{Uptime: "1 day, 2 hours, 5 minutes", Seconds: 93900},
		},
		{
			name:   "when the host just booted, each unit is zero",
			uptime: 30 * time.Second,
			want:   uptime.Entry{Uptime: "0 days, 0 hours, 0 minutes", Seconds: 30},
		},
		{
			name:   "when the host has run one hour and one minute",
			uptime: time.Hour + time.Minute,
			want:   uptime.Entry{Uptime: "0 days, 1 hour, 1 minute", Seconds: 3660},
		},
		{
			name:    "when the provider fails, it is the error",
			err:     errors.New("no /proc"),
			wantErr: "no /proc",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.mock.EXPECT().Uptime(gomock.Any()).Return(tc.uptime, tc.err)

			got, err := uptime.Get(context.Background(), s.mock, uptime.GetInput{})
			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
