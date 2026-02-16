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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apinetwork "github.com/retr0h/osapi/internal/api/network"
	"github.com/retr0h/osapi/internal/api/network/gen"
	"github.com/retr0h/osapi/internal/provider/network/dns"

	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type NetworkDNSGetByInterfacePublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinetwork.Network
	ctx           context.Context
}

func (s *NetworkDNSGetByInterfacePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinetwork.New(s.mockJobClient)
	s.ctx = context.Background()
}

func (s *NetworkDNSGetByInterfacePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NetworkDNSGetByInterfacePublicTestSuite) TestGetNetworkDNSByInterface() {
	tests := []struct {
		name         string
		request      gen.GetNetworkDNSByInterfaceRequestObject
		mockConfig   *dns.Config
		mockError    error
		validateFunc func(resp gen.GetNetworkDNSByInterfaceResponseObject)
	}{
		{
			name:    "success",
			request: gen.GetNetworkDNSByInterfaceRequestObject{InterfaceName: "eth0"},
			mockConfig: &dns.Config{
				DNSServers:    []string{"192.168.1.1", "8.8.8.8"},
				SearchDomains: []string{"example.com", "local.lan"},
			},
			validateFunc: func(resp gen.GetNetworkDNSByInterfaceResponseObject) {
				r, ok := resp.(gen.GetNetworkDNSByInterface200JSONResponse)
				s.True(ok)
				s.Equal([]string{"192.168.1.1", "8.8.8.8"}, *r.Servers)
				s.Equal([]string{"example.com", "local.lan"}, *r.SearchDomains)
			},
		},
		{
			name:      "job client error",
			request:   gen.GetNetworkDNSByInterfaceRequestObject{InterfaceName: "eth0"},
			mockError: assert.AnError,
			validateFunc: func(resp gen.GetNetworkDNSByInterfaceResponseObject) {
				_, ok := resp.(gen.GetNetworkDNSByInterface500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockJobClient.EXPECT().
				QueryNetworkDNS(gomock.Any(), gomock.Any(), tt.request.InterfaceName).
				Return(tt.mockConfig, tt.mockError)

			resp, err := s.handler.GetNetworkDNSByInterface(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestNetworkDNSGetByInterfacePublicTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkDNSGetByInterfacePublicTestSuite))
}
