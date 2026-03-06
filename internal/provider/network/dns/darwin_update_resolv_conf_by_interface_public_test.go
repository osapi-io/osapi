// Copyright (c) 2026 John Dewey
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package dns_test

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	execMocks "github.com/retr0h/osapi/internal/exec/mocks"
	"github.com/retr0h/osapi/internal/provider/network/dns"
)

const (
	darwinHardwarePorts = `Hardware Port: Wi-Fi
Device: en0
Ethernet Address: a4:83:e7:1a:2b:3c

Hardware Port: Thunderbolt Ethernet
Device: en1
Ethernet Address: 00:11:22:33:44:55
`
	darwinScutilExisting = `
DNS configuration

resolver #1
  nameserver[0] : 192.168.1.1
  nameserver[1] : 8.8.8.8
  search domain[0] : old.example.com
  if_index : 6 (en0)
`
	darwinScutilSameConfig = `
DNS configuration

resolver #1
  nameserver[0] : 8.8.8.8
  nameserver[1] : 9.9.9.9
  search domain[0] : example.com
  if_index : 6 (en0)
`
)

type DarwinUpdateResolvConfByInterfacePublicTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	logger *slog.Logger
}

func (suite *DarwinUpdateResolvConfByInterfacePublicTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *DarwinUpdateResolvConfByInterfacePublicTestSuite) SetupSubTest() {
	suite.SetupTest()
}

func (suite *DarwinUpdateResolvConfByInterfacePublicTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *DarwinUpdateResolvConfByInterfacePublicTestSuite) TestUpdateResolvConfByInterface() {
	tests := []struct {
		name          string
		setupMock     func() *execMocks.MockManager
		servers       []string
		searchDomains []string
		interfaceName string
		want          *dns.UpdateResult
		wantErr       bool
		wantErrType   error
	}{
		{
			name: "when update succeeds with new servers and domains",
			setupMock: func() *execMocks.MockManager {
				mock := execMocks.NewPlainMockManager(suite.ctrl)

				mock.EXPECT().
					RunCmd("scutil", []string{"--dns"}).
					Return(darwinScutilExisting, nil)

				mock.EXPECT().
					RunCmd("networksetup", []string{"-listallhardwareports"}).
					Return(darwinHardwarePorts, nil)

				mock.EXPECT().
					RunCmd("networksetup", []string{"-setdnsservers", "Wi-Fi", "8.8.8.8", "9.9.9.9"}).
					Return("", nil)

				mock.EXPECT().
					RunCmd("networksetup", []string{"-setsearchdomains", "Wi-Fi", "example.com"}).
					Return("", nil)

				return mock
			},
			servers:       []string{"8.8.8.8", "9.9.9.9"},
			searchDomains: []string{"example.com"},
			interfaceName: "en0",
			want:          &dns.UpdateResult{Changed: true},
		},
		{
			name: "when configuration unchanged returns no-op",
			setupMock: func() *execMocks.MockManager {
				mock := execMocks.NewPlainMockManager(suite.ctrl)

				mock.EXPECT().
					RunCmd("scutil", []string{"--dns"}).
					Return(darwinScutilSameConfig, nil)

				return mock
			},
			servers:       []string{"8.8.8.8", "9.9.9.9"},
			searchDomains: []string{"example.com"},
			interfaceName: "en0",
			want:          &dns.UpdateResult{Changed: false},
		},
		{
			name: "when no servers or domains provided",
			setupMock: func() *execMocks.MockManager {
				return execMocks.NewPlainMockManager(suite.ctrl)
			},
			servers:       []string{},
			searchDomains: []string{},
			interfaceName: "en0",
			wantErr:       true,
			wantErrType:   fmt.Errorf("no DNS servers or search domains provided"),
		},
		{
			name: "when scutil errors",
			setupMock: func() *execMocks.MockManager {
				mock := execMocks.NewPlainMockManager(suite.ctrl)

				mock.EXPECT().
					RunCmd("scutil", []string{"--dns"}).
					Return("", assert.AnError)

				return mock
			},
			servers:       []string{"8.8.8.8"},
			searchDomains: []string{},
			interfaceName: "en0",
			wantErr:       true,
			wantErrType:   assert.AnError,
		},
		{
			name: "when interface not found in hardware ports",
			setupMock: func() *execMocks.MockManager {
				mock := execMocks.NewPlainMockManager(suite.ctrl)

				mock.EXPECT().
					RunCmd("scutil", []string{"--dns"}).
					Return(darwinScutilExisting, nil)

				mock.EXPECT().
					RunCmd("networksetup", []string{"-listallhardwareports"}).
					Return(darwinHardwarePorts, nil)

				return mock
			},
			servers:       []string{"8.8.8.8"},
			searchDomains: []string{},
			interfaceName: "en99",
			wantErr:       true,
			wantErrType:   fmt.Errorf("no network service found for interface"),
		},
		{
			name: "when setdnsservers errors",
			setupMock: func() *execMocks.MockManager {
				mock := execMocks.NewPlainMockManager(suite.ctrl)

				mock.EXPECT().
					RunCmd("scutil", []string{"--dns"}).
					Return(darwinScutilExisting, nil)

				mock.EXPECT().
					RunCmd("networksetup", []string{"-listallhardwareports"}).
					Return(darwinHardwarePorts, nil)

				mock.EXPECT().
					RunCmd("networksetup", []string{"-setdnsservers", "Wi-Fi", "8.8.8.8"}).
					Return("", assert.AnError)

				return mock
			},
			servers:       []string{"8.8.8.8"},
			searchDomains: []string{},
			interfaceName: "en0",
			wantErr:       true,
			wantErrType:   assert.AnError,
		},
		{
			name: "when setsearchdomains errors",
			setupMock: func() *execMocks.MockManager {
				mock := execMocks.NewPlainMockManager(suite.ctrl)

				mock.EXPECT().
					RunCmd("scutil", []string{"--dns"}).
					Return(darwinScutilExisting, nil)

				mock.EXPECT().
					RunCmd("networksetup", []string{"-listallhardwareports"}).
					Return(darwinHardwarePorts, nil)

				mock.EXPECT().
					RunCmd("networksetup", []string{"-setdnsservers", "Wi-Fi", "8.8.8.8"}).
					Return("", nil)

				mock.EXPECT().
					RunCmd("networksetup", []string{"-setsearchdomains", "Wi-Fi", "new.example.com"}).
					Return("", assert.AnError)

				return mock
			},
			servers:       []string{"8.8.8.8"},
			searchDomains: []string{"new.example.com"},
			interfaceName: "en0",
			wantErr:       true,
			wantErrType:   assert.AnError,
		},
		{
			name: "when preserving existing servers when only domains specified",
			setupMock: func() *execMocks.MockManager {
				mock := execMocks.NewPlainMockManager(suite.ctrl)

				mock.EXPECT().
					RunCmd("scutil", []string{"--dns"}).
					Return(darwinScutilExisting, nil)

				mock.EXPECT().
					RunCmd("networksetup", []string{"-listallhardwareports"}).
					Return(darwinHardwarePorts, nil)

				mock.EXPECT().
					RunCmd("networksetup", []string{"-setdnsservers", "Wi-Fi", "192.168.1.1", "8.8.8.8"}).
					Return("", nil)

				mock.EXPECT().
					RunCmd("networksetup", []string{"-setsearchdomains", "Wi-Fi", "new.example.com"}).
					Return("", nil)

				return mock
			},
			servers:       []string{},
			searchDomains: []string{"new.example.com"},
			interfaceName: "en0",
			want:          &dns.UpdateResult{Changed: true},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			mock := tc.setupMock()

			darwin := dns.NewDarwinProvider(suite.logger, mock)
			got, err := darwin.UpdateResolvConfByInterface(
				tc.servers,
				tc.searchDomains,
				tc.interfaceName,
			)

			if !tc.wantErr {
				suite.NoError(err)
				suite.Equal(tc.want, got)
			} else {
				suite.Error(err)
				suite.Contains(err.Error(), tc.wantErrType.Error())
			}
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestDarwinUpdateResolvConfByInterfacePublicTestSuite(t *testing.T) {
	suite.Run(t, new(DarwinUpdateResolvConfByInterfacePublicTestSuite))
}
