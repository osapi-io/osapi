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
	"fmt"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apinetwork "github.com/retr0h/osapi/internal/api/network"
	"github.com/retr0h/osapi/internal/api/network/gen"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/validation"
)

type NetworkDNSPutByInterfacePublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinetwork.Network
	ctx           context.Context
}

func (s *NetworkDNSPutByInterfacePublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.WorkerTarget, error) {
		return []validation.WorkerTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *NetworkDNSPutByInterfacePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinetwork.New(slog.Default(), s.mockJobClient)
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
		setupMock    func()
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
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyNetworkDNS(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", "worker1", true, nil)
			},
			validateFunc: func(resp gen.PutNetworkDNSResponseObject) {
				r, ok := resp.(gen.PutNetworkDNS202JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal("worker1", r.Results[0].Hostname)
				s.Equal(gen.Ok, r.Results[0].Status)
				s.Require().NotNil(r.Results[0].Changed)
				s.True(*r.Results[0].Changed)
			},
		},
		{
			name: "validation error missing interface name",
			request: gen.PutNetworkDNSRequestObject{
				Body: &gen.PutNetworkDNSJSONRequestBody{
					Servers: &servers,
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PutNetworkDNSResponseObject) {
				r, ok := resp.(gen.PutNetworkDNS400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "InterfaceName")
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "validation error empty target_hostname",
			request: gen.PutNetworkDNSRequestObject{
				Body: &gen.PutNetworkDNSJSONRequestBody{
					Servers:       &servers,
					SearchDomains: &searchDomains,
					InterfaceName: interfaceName,
				},
				Params: gen.PutNetworkDNSParams{TargetHostname: strPtr("")},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PutNetworkDNSResponseObject) {
				r, ok := resp.(gen.PutNetworkDNS400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "TargetHostname")
				s.Contains(*r.Error, "min")
			},
		},
		{
			name: "validation error unknown target_hostname",
			request: gen.PutNetworkDNSRequestObject{
				Body: &gen.PutNetworkDNSJSONRequestBody{
					Servers:       &servers,
					SearchDomains: &searchDomains,
					InterfaceName: interfaceName,
				},
				Params: gen.PutNetworkDNSParams{TargetHostname: strPtr("nonexistent")},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PutNetworkDNSResponseObject) {
				r, ok := resp.(gen.PutNetworkDNS400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "valid_target")
				s.Contains(*r.Error, "not found")
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
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyNetworkDNS(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", "", false, assert.AnError)
			},
			validateFunc: func(resp gen.PutNetworkDNSResponseObject) {
				_, ok := resp.(gen.PutNetworkDNS500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "broadcast all success",
			request: gen.PutNetworkDNSRequestObject{
				Body: &gen.PutNetworkDNSJSONRequestBody{
					Servers:       &servers,
					SearchDomains: &searchDomains,
					InterfaceName: interfaceName,
				},
				Params: gen.PutNetworkDNSParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyNetworkDNSBroadcast(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]error{
						"server1": nil,
						"server2": nil,
					}, map[string]bool{
						"server1": true,
						"server2": true,
					}, nil)
			},
			validateFunc: func(resp gen.PutNetworkDNSResponseObject) {
				r, ok := resp.(gen.PutNetworkDNS202JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 2)
				for _, result := range r.Results {
					s.Require().NotNil(result.Changed)
					s.True(*result.Changed)
				}
			},
		},
		{
			name: "broadcast all with partial failure",
			request: gen.PutNetworkDNSRequestObject{
				Body: &gen.PutNetworkDNSJSONRequestBody{
					Servers:       &servers,
					SearchDomains: &searchDomains,
					InterfaceName: interfaceName,
				},
				Params: gen.PutNetworkDNSParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyNetworkDNSBroadcast(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]error{
						"server1": nil,
						"server2": fmt.Errorf("disk full"),
					}, map[string]bool{
						"server1": true,
						"server2": false,
					}, nil)
			},
			validateFunc: func(resp gen.PutNetworkDNSResponseObject) {
				r, ok := resp.(gen.PutNetworkDNS202JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 2)
				for _, result := range r.Results {
					s.Require().NotNil(result.Changed)
				}
			},
		},
		{
			name: "broadcast all error",
			request: gen.PutNetworkDNSRequestObject{
				Body: &gen.PutNetworkDNSJSONRequestBody{
					Servers:       &servers,
					SearchDomains: &searchDomains,
					InterfaceName: interfaceName,
				},
				Params: gen.PutNetworkDNSParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyNetworkDNSBroadcast(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.PutNetworkDNSResponseObject) {
				_, ok := resp.(gen.PutNetworkDNS500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.PutNetworkDNS(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestNetworkDNSPutByInterfacePublicTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkDNSPutByInterfacePublicTestSuite))
}
