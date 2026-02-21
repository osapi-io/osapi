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
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apinetwork "github.com/retr0h/osapi/internal/api/network"
	"github.com/retr0h/osapi/internal/api/network/gen"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/network/ping"
)

type NetworkPingPostPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinetwork.Network
	ctx           context.Context
}

func (s *NetworkPingPostPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinetwork.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *NetworkPingPostPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NetworkPingPostPublicTestSuite) TestPostNetworkPing() {
	tests := []struct {
		name         string
		request      gen.PostNetworkPingRequestObject
		setupMock    func()
		validateFunc func(resp gen.PostNetworkPingResponseObject)
	}{
		{
			name: "success",
			request: gen.PostNetworkPingRequestObject{
				Body: &gen.PostNetworkPingJSONRequestBody{
					Address: "1.1.1.1",
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkPing(gomock.Any(), gomock.Any(), "1.1.1.1").
					Return("550e8400-e29b-41d4-a716-446655440000", &ping.Result{
						PacketsSent:     3,
						PacketsReceived: 3,
						PacketLoss:      0,
						MinRTT:          10 * time.Millisecond,
						AvgRTT:          15 * time.Millisecond,
						MaxRTT:          20 * time.Millisecond,
					}, "worker1", nil)
			},
			validateFunc: func(resp gen.PostNetworkPingResponseObject) {
				r, ok := resp.(gen.PostNetworkPing200JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal(3, *r.Results[0].PacketsSent)
				s.Equal(3, *r.Results[0].PacketsReceived)
				s.Equal(0.0, *r.Results[0].PacketLoss)
				s.Equal("worker1", r.Results[0].Hostname)
			},
		},
		{
			name: "validation error invalid address",
			request: gen.PostNetworkPingRequestObject{
				Body: &gen.PostNetworkPingJSONRequestBody{
					Address: "not-an-ip",
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNetworkPingResponseObject) {
				r, ok := resp.(gen.PostNetworkPing400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "Address")
				s.Contains(*r.Error, "ip")
			},
		},
		{
			name: "validation error empty target_hostname",
			request: gen.PostNetworkPingRequestObject{
				Body: &gen.PostNetworkPingJSONRequestBody{
					Address: "1.1.1.1",
				},
				Params: gen.PostNetworkPingParams{TargetHostname: strPtr("")},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNetworkPingResponseObject) {
				r, ok := resp.(gen.PostNetworkPing400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "TargetHostname")
				s.Contains(*r.Error, "min")
			},
		},
		{
			name: "job client error",
			request: gen.PostNetworkPingRequestObject{
				Body: &gen.PostNetworkPingJSONRequestBody{
					Address: "1.1.1.1",
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkPing(gomock.Any(), gomock.Any(), "1.1.1.1").
					Return("", nil, "", assert.AnError)
			},
			validateFunc: func(resp gen.PostNetworkPingResponseObject) {
				_, ok := resp.(gen.PostNetworkPing500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "broadcast all success",
			request: gen.PostNetworkPingRequestObject{
				Body: &gen.PostNetworkPingJSONRequestBody{
					Address: "1.1.1.1",
				},
				Params: gen.PostNetworkPingParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkPingBroadcast(gomock.Any(), gomock.Any(), "1.1.1.1").
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*ping.Result{
						"server1": {
							PacketsSent:     3,
							PacketsReceived: 3,
							PacketLoss:      0,
							MinRTT:          10 * time.Millisecond,
							AvgRTT:          15 * time.Millisecond,
							MaxRTT:          20 * time.Millisecond,
						},
					}, map[string]string{}, nil)
			},
			validateFunc: func(resp gen.PostNetworkPingResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name: "broadcast all with errors",
			request: gen.PostNetworkPingRequestObject{
				Body: &gen.PostNetworkPingJSONRequestBody{
					Address: "1.1.1.1",
				},
				Params: gen.PostNetworkPingParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkPingBroadcast(gomock.Any(), gomock.Any(), "1.1.1.1").
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*ping.Result{
						"server1": {
							PacketsSent:     3,
							PacketsReceived: 3,
							PacketLoss:      0,
							MinRTT:          10 * time.Millisecond,
							AvgRTT:          15 * time.Millisecond,
							MaxRTT:          20 * time.Millisecond,
						},
					}, map[string]string{
						"server2": "host unreachable",
					}, nil)
			},
			validateFunc: func(resp gen.PostNetworkPingResponseObject) {
				r, ok := resp.(gen.PostNetworkPing200JSONResponse)
				s.True(ok)
				s.Len(r.Results, 2)
				var foundError bool
				for _, res := range r.Results {
					if res.Error != nil {
						foundError = true
						s.Equal("server2", res.Hostname)
						s.Equal("host unreachable", *res.Error)
					}
				}
				s.True(foundError)
			},
		},
		{
			name: "broadcast all error",
			request: gen.PostNetworkPingRequestObject{
				Body: &gen.PostNetworkPingJSONRequestBody{
					Address: "1.1.1.1",
				},
				Params: gen.PostNetworkPingParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkPingBroadcast(gomock.Any(), gomock.Any(), "1.1.1.1").
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.PostNetworkPingResponseObject) {
				_, ok := resp.(gen.PostNetworkPing500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.PostNetworkPing(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestNetworkPingPostPublicTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkPingPostPublicTestSuite))
}
