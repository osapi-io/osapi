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
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type NetworkDNSPutByInterfacePublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinetwork.Network
	ctx           context.Context
}

func (s *NetworkDNSPutByInterfacePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinetwork.New(s.mockJobClient)
	s.ctx = context.Background()
}

func (s *NetworkDNSPutByInterfacePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NetworkDNSPutByInterfacePublicTestSuite) TestPutNetworkDNS() {
	servers := []string{"1.1.1.1", "8.8.8.8"}
	searchDomains := []string{"foo.bar"}
	interfaceName := "eth0"

	tests := []struct {
		name         string
		request      gen.PutNetworkDNSRequestObject
		mockError    error
		expectMock   bool
		validateFunc func(resp gen.PutNetworkDNSResponseObject)
	}{
		{
			name: "success",
			request: gen.PutNetworkDNSRequestObject{
				Body: &gen.PutNetworkDNSJSONRequestBody{
					Servers:       &servers,
					SearchDomains: &searchDomains,
					InterfaceName: interfaceName,
				},
			},
			expectMock: true,
			validateFunc: func(resp gen.PutNetworkDNSResponseObject) {
				_, ok := resp.(gen.PutNetworkDNS202Response)
				s.True(ok)
			},
		},
		{
			name: "validation error missing interface name",
			request: gen.PutNetworkDNSRequestObject{
				Body: &gen.PutNetworkDNSJSONRequestBody{
					Servers: &servers,
				},
			},
			expectMock: false,
			validateFunc: func(resp gen.PutNetworkDNSResponseObject) {
				r, ok := resp.(gen.PutNetworkDNS400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "InterfaceName")
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "job client error",
			request: gen.PutNetworkDNSRequestObject{
				Body: &gen.PutNetworkDNSJSONRequestBody{
					Servers:       &servers,
					SearchDomains: &searchDomains,
					InterfaceName: interfaceName,
				},
			},
			mockError:  assert.AnError,
			expectMock: true,
			validateFunc: func(resp gen.PutNetworkDNSResponseObject) {
				_, ok := resp.(gen.PutNetworkDNS500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.expectMock {
				s.mockJobClient.EXPECT().
					ModifyNetworkDNSAny(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(tt.mockError)
			}

			resp, err := s.handler.PutNetworkDNS(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestNetworkDNSPutByInterfacePublicTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkDNSPutByInterfacePublicTestSuite))
}
