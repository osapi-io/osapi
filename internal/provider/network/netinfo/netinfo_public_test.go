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

package netinfo_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/network/netinfo"
)

type GetInterfacesPublicTestSuite struct {
	suite.Suite
}

func (suite *GetInterfacesPublicTestSuite) SetupTest() {}

func (suite *GetInterfacesPublicTestSuite) TearDownTest() {}

func (suite *GetInterfacesPublicTestSuite) TestGetInterfaces() {
	tests := []struct {
		name         string
		setupMock    func() func() ([]net.Interface, error)
		wantErr      bool
		wantErrType  error
		validateFunc func(result []job.NetworkInterface)
	}{
		{
			name: "when GetInterfaces Ok",
			setupMock: func() func() ([]net.Interface, error) {
				return func() ([]net.Interface, error) {
					return []net.Interface{
						{
							Index:        1,
							MTU:          65536,
							Name:         "lo",
							HardwareAddr: nil,
							Flags:        net.FlagUp | net.FlagLoopback,
						},
						{
							Index:        2,
							MTU:          1500,
							Name:         "eth0",
							HardwareAddr: net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
							Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
						},
						{
							Index:        3,
							MTU:          1500,
							Name:         "eth1",
							HardwareAddr: net.HardwareAddr{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
							Flags:        net.FlagBroadcast | net.FlagMulticast, // not up
						},
					}, nil
				}
			},
			wantErr: false,
			validateFunc: func(result []job.NetworkInterface) {
				suite.Require().Len(result, 1)
				suite.Equal("eth0", result[0].Name)
				suite.Equal("00:11:22:33:44:55", result[0].MAC)
			},
		},
		{
			name: "when no non-loopback interfaces exist",
			setupMock: func() func() ([]net.Interface, error) {
				return func() ([]net.Interface, error) {
					return []net.Interface{
						{
							Index: 1,
							MTU:   65536,
							Name:  "lo",
							Flags: net.FlagUp | net.FlagLoopback,
						},
					}, nil
				}
			},
			wantErr: false,
			validateFunc: func(result []job.NetworkInterface) {
				suite.Empty(result)
			},
		},
		{
			name: "when net.Interfaces errors",
			setupMock: func() func() ([]net.Interface, error) {
				return func() ([]net.Interface, error) {
					return nil, assert.AnError
				}
			},
			wantErr:     true,
			wantErrType: assert.AnError,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			n := netinfo.New()

			if tc.setupMock != nil {
				n.InterfacesFn = tc.setupMock()
			}

			got, err := n.GetInterfaces()

			if tc.wantErr {
				suite.Error(err)
				suite.ErrorContains(err, tc.wantErrType.Error())
				suite.Nil(got)
			} else {
				suite.NoError(err)

				if tc.validateFunc != nil {
					tc.validateFunc(got)
				}
			}
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestGetInterfacesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(GetInterfacesPublicTestSuite))
}
