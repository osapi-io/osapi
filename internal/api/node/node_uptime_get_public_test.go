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

package node_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apinode "github.com/retr0h/osapi/internal/api/node"
	"github.com/retr0h/osapi/internal/api/node/gen"
	"github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/validation"
)

type NodeUptimeGetPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinode.Node
	ctx           context.Context
}

func (s *NodeUptimeGetPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *NodeUptimeGetPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinode.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *NodeUptimeGetPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NodeUptimeGetPublicTestSuite) TestGetNodeUptime() {
	tests := []struct {
		name         string
		request      gen.GetNodeUptimeRequestObject
		setupMock    func()
		validateFunc func(resp gen.GetNodeUptimeResponseObject)
	}{
		{
			name:    "success",
			request: gen.GetNodeUptimeRequestObject{Hostname: "_any"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeUptime(gomock.Any(), "_any").
					Return("550e8400-e29b-41d4-a716-446655440000", &job.NodeUptimeResponse{
						Uptime: "3 days, 2 hours, 15 minutes",
					}, "agent1", nil)
			},
			validateFunc: func(resp gen.GetNodeUptimeResponseObject) {
				_, ok := resp.(gen.GetNodeUptime200JSONResponse)
				s.True(ok)
			},
		},
		{
			name:      "validation error empty hostname",
			request:   gen.GetNodeUptimeRequestObject{Hostname: ""},
			setupMock: func() {},
			validateFunc: func(resp gen.GetNodeUptimeResponseObject) {
				r, ok := resp.(gen.GetNodeUptime400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name:    "job client error",
			request: gen.GetNodeUptimeRequestObject{Hostname: "_any"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeUptime(gomock.Any(), "_any").
					Return("", nil, "", assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeUptimeResponseObject) {
				_, ok := resp.(gen.GetNodeUptime500JSONResponse)
				s.True(ok)
			},
		},
		{
			name:    "broadcast all success",
			request: gen.GetNodeUptimeRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeUptimeBroadcast(gomock.Any(), "_all").
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*job.NodeUptimeResponse{
						"server1": {Uptime: "3 days, 2 hours, 15 minutes"},
						"server2": {Uptime: "1 day, 5 hours, 30 minutes"},
					}, map[string]string{}, nil)
			},
			validateFunc: func(resp gen.GetNodeUptimeResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name:    "broadcast all with errors",
			request: gen.GetNodeUptimeRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeUptimeBroadcast(gomock.Any(), "_all").
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*job.NodeUptimeResponse{
						"server1": {Uptime: "3 days, 2 hours, 15 minutes"},
					}, map[string]string{
						"server2": "some error",
					}, nil)
			},
			validateFunc: func(resp gen.GetNodeUptimeResponseObject) {
				r, ok := resp.(gen.GetNodeUptime200JSONResponse)
				s.True(ok)
				s.Len(r.Results, 2)
				var foundError bool
				for _, res := range r.Results {
					if res.Error != nil {
						foundError = true
						s.Equal("server2", res.Hostname)
						s.Equal("some error", *res.Error)
					}
				}
				s.True(foundError)
			},
		},
		{
			name:    "broadcast all error",
			request: gen.GetNodeUptimeRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeUptimeBroadcast(gomock.Any(), "_all").
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeUptimeResponseObject) {
				_, ok := resp.(gen.GetNodeUptime500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.GetNodeUptime(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestNodeUptimeGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(NodeUptimeGetPublicTestSuite))
}
