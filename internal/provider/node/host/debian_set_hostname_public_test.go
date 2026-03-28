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

package host_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/exec/mocks"
	"github.com/retr0h/osapi/internal/provider/node/host"
)

type DebianSetHostnamePublicTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
}

func (suite *DebianSetHostnamePublicTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
}

func (suite *DebianSetHostnamePublicTestSuite) SetupSubTest() {
	suite.SetupTest()
}

func (suite *DebianSetHostnamePublicTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *DebianSetHostnamePublicTestSuite) TestSetHostname() {
	tests := []struct {
		name        string
		hostname    string
		setupMock   func() *mocks.MockManager
		want        *host.SetHostnameResult
		wantErr     bool
		wantErrType error
	}{
		{
			name:     "when hostname changes",
			hostname: "new-host",
			setupMock: func() *mocks.MockManager {
				mock := mocks.NewPlainMockManager(suite.ctrl)
				gomock.InOrder(
					mock.EXPECT().
						RunCmd("hostnamectl", []string{"hostname"}).
						Return("old-host", nil),
					mock.EXPECT().
						RunCmd("hostnamectl", []string{"set-hostname", "new-host"}).
						Return("", nil),
				)
				return mock
			},
			want:    &host.SetHostnameResult{Changed: true},
			wantErr: false,
		},
		{
			name:     "when hostname already set",
			hostname: "existing-host",
			setupMock: func() *mocks.MockManager {
				mock := mocks.NewPlainMockManager(suite.ctrl)
				mock.EXPECT().
					RunCmd("hostnamectl", []string{"hostname"}).
					Return("existing-host", nil)
				return mock
			},
			want:    &host.SetHostnameResult{Changed: false},
			wantErr: false,
		},
		{
			name:     "when hostname already set with trailing newline",
			hostname: "existing-host",
			setupMock: func() *mocks.MockManager {
				mock := mocks.NewPlainMockManager(suite.ctrl)
				mock.EXPECT().
					RunCmd("hostnamectl", []string{"hostname"}).
					Return("existing-host\n", nil)
				return mock
			},
			want:    &host.SetHostnameResult{Changed: false},
			wantErr: false,
		},
		{
			name:     "when hostnamectl hostname errors",
			hostname: "new-host",
			setupMock: func() *mocks.MockManager {
				mock := mocks.NewPlainMockManager(suite.ctrl)
				mock.EXPECT().
					RunCmd("hostnamectl", []string{"hostname"}).
					Return("", assert.AnError)
				return mock
			},
			wantErr:     true,
			wantErrType: assert.AnError,
		},
		{
			name:     "when hostnamectl set-hostname errors",
			hostname: "new-host",
			setupMock: func() *mocks.MockManager {
				mock := mocks.NewPlainMockManager(suite.ctrl)
				gomock.InOrder(
					mock.EXPECT().
						RunCmd("hostnamectl", []string{"hostname"}).
						Return("old-host", nil),
					mock.EXPECT().
						RunCmd("hostnamectl", []string{"set-hostname", "new-host"}).
						Return("", assert.AnError),
				)
				return mock
			},
			wantErr:     true,
			wantErrType: assert.AnError,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			mock := tc.setupMock()
			debian := host.NewDebianProvider(mock)

			got, err := debian.SetHostname(tc.hostname)

			if tc.wantErr {
				suite.Error(err)
				suite.ErrorContains(err, tc.wantErrType.Error())
				suite.Nil(got)
			} else {
				suite.NoError(err)
				suite.Equal(tc.want, got)
			}
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestDebianSetHostnamePublicTestSuite(t *testing.T) {
	suite.Run(t, new(DebianSetHostnamePublicTestSuite))
}
