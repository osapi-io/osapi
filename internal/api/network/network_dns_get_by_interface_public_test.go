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

package network_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apinetwork "github.com/retr0h/osapi/internal/api/network"
	"github.com/retr0h/osapi/internal/api/network/gen"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/validation"
)

type NetworkDNSGetByInterfacePublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinetwork.Network
	ctx           context.Context
}

func (s *NetworkDNSGetByInterfacePublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.WorkerTarget, error) {
		return []validation.WorkerTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *NetworkDNSGetByInterfacePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinetwork.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *NetworkDNSGetByInterfacePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NetworkDNSGetByInterfacePublicTestSuite) TestGetNetworkDNSByInterface() {
	tests := []struct {
		name         string
		request      gen.GetNetworkDNSByInterfaceRequestObject
		setupMock    func()
		validateFunc func(resp gen.GetNetworkDNSByInterfaceResponseObject)
	}{
		{
			name:    "success",
			request: gen.GetNetworkDNSByInterfaceRequestObject{InterfaceName: "eth0"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkDNS(gomock.Any(), gomock.Any(), "eth0").
					Return("550e8400-e29b-41d4-a716-446655440000", &dns.Config{
						DNSServers:    []string{"192.168.1.1", "8.8.8.8"},
						SearchDomains: []string{"example.com", "local.lan"},
					}, "worker1", nil)
			},
			validateFunc: func(resp gen.GetNetworkDNSByInterfaceResponseObject) {
				r, ok := resp.(gen.GetNetworkDNSByInterface200JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal([]string{"192.168.1.1", "8.8.8.8"}, *r.Results[0].Servers)
				s.Equal([]string{"example.com", "local.lan"}, *r.Results[0].SearchDomains)
				s.Equal("worker1", r.Results[0].Hostname)
			},
		},
		{
			name:      "validation error non-alphanum interface name",
			request:   gen.GetNetworkDNSByInterfaceRequestObject{InterfaceName: "eth-0!"},
			setupMock: func() {},
			validateFunc: func(resp gen.GetNetworkDNSByInterfaceResponseObject) {
				r, ok := resp.(gen.GetNetworkDNSByInterface400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "InterfaceName")
				s.Contains(*r.Error, "alphanum")
			},
		},
		{
			name:      "validation error empty interface name",
			request:   gen.GetNetworkDNSByInterfaceRequestObject{InterfaceName: ""},
			setupMock: func() {},
			validateFunc: func(resp gen.GetNetworkDNSByInterfaceResponseObject) {
				r, ok := resp.(gen.GetNetworkDNSByInterface400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "InterfaceName")
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "validation error empty target_hostname",
			request: gen.GetNetworkDNSByInterfaceRequestObject{
				InterfaceName: "eth0",
				Params:        gen.GetNetworkDNSByInterfaceParams{TargetHostname: strPtr("")},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.GetNetworkDNSByInterfaceResponseObject) {
				r, ok := resp.(gen.GetNetworkDNSByInterface400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "TargetHostname")
				s.Contains(*r.Error, "min")
			},
		},
		{
			name:    "job client error",
			request: gen.GetNetworkDNSByInterfaceRequestObject{InterfaceName: "eth0"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkDNS(gomock.Any(), gomock.Any(), "eth0").
					Return("", nil, "", assert.AnError)
			},
			validateFunc: func(resp gen.GetNetworkDNSByInterfaceResponseObject) {
				_, ok := resp.(gen.GetNetworkDNSByInterface500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "broadcast all success",
			request: gen.GetNetworkDNSByInterfaceRequestObject{
				InterfaceName: "eth0",
				Params:        gen.GetNetworkDNSByInterfaceParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkDNSBroadcast(gomock.Any(), gomock.Any(), "eth0").
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*dns.Config{
						"server1": {
							DNSServers:    []string{"8.8.8.8"},
							SearchDomains: []string{"example.com"},
						},
					}, map[string]string{}, nil)
			},
			validateFunc: func(resp gen.GetNetworkDNSByInterfaceResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name: "broadcast all with errors",
			request: gen.GetNetworkDNSByInterfaceRequestObject{
				InterfaceName: "eth0",
				Params:        gen.GetNetworkDNSByInterfaceParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkDNSBroadcast(gomock.Any(), gomock.Any(), "eth0").
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*dns.Config{
						"server1": {
							DNSServers:    []string{"8.8.8.8"},
							SearchDomains: []string{"example.com"},
						},
					}, map[string]string{
						"server2": "interface eth0 does not exist",
					}, nil)
			},
			validateFunc: func(resp gen.GetNetworkDNSByInterfaceResponseObject) {
				r, ok := resp.(gen.GetNetworkDNSByInterface200JSONResponse)
				s.True(ok)
				s.Len(r.Results, 2)
				var foundError bool
				for _, res := range r.Results {
					if res.Error != nil {
						foundError = true
						s.Equal("server2", res.Hostname)
						s.Equal("interface eth0 does not exist", *res.Error)
					}
				}
				s.True(foundError)
			},
		},
		{
			name: "broadcast all error",
			request: gen.GetNetworkDNSByInterfaceRequestObject{
				InterfaceName: "eth0",
				Params:        gen.GetNetworkDNSByInterfaceParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkDNSBroadcast(gomock.Any(), gomock.Any(), "eth0").
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetNetworkDNSByInterfaceResponseObject) {
				_, ok := resp.(gen.GetNetworkDNSByInterface500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.GetNetworkDNSByInterface(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func strPtr(
	s string,
) *string {
	return &s
}

func TestNetworkDNSGetByInterfacePublicTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkDNSGetByInterfacePublicTestSuite))
}
