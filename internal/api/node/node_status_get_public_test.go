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
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apinode "github.com/retr0h/osapi/internal/api/node"
	"github.com/retr0h/osapi/internal/api/node/gen"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/validation"
)

type NodeStatusGetPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinode.Node
	ctx           context.Context
}

func (s *NodeStatusGetPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *NodeStatusGetPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinode.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *NodeStatusGetPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NodeStatusGetPublicTestSuite) TestGetNodeStatus() {
	tests := []struct {
		name         string
		request      gen.GetNodeStatusRequestObject
		setupMock    func()
		validateFunc func(resp gen.GetNodeStatusResponseObject)
	}{
		{
			name:    "success",
			request: gen.GetNodeStatusRequestObject{Hostname: "_any"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeStatus(gomock.Any(), "_any").
					Return("550e8400-e29b-41d4-a716-446655440000", &jobtypes.NodeStatusResponse{
						Hostname: "test-host",
						Uptime:   time.Hour,
					}, nil)
			},
			validateFunc: func(resp gen.GetNodeStatusResponseObject) {
				_, ok := resp.(gen.GetNodeStatus200JSONResponse)
				s.True(ok)
			},
		},
		{
			name:      "validation error empty hostname",
			request:   gen.GetNodeStatusRequestObject{Hostname: ""},
			setupMock: func() {},
			validateFunc: func(resp gen.GetNodeStatusResponseObject) {
				r, ok := resp.(gen.GetNodeStatus400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name:    "job client error",
			request: gen.GetNodeStatusRequestObject{Hostname: "_any"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeStatus(gomock.Any(), "_any").
					Return("", nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeStatusResponseObject) {
				_, ok := resp.(gen.GetNodeStatus500JSONResponse)
				s.True(ok)
			},
		},
		{
			name:    "broadcast all success",
			request: gen.GetNodeStatusRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeStatusBroadcast(gomock.Any(), "_all").
					Return("550e8400-e29b-41d4-a716-446655440000", []*jobtypes.NodeStatusResponse{
						{Hostname: "server1", Uptime: time.Hour},
						{Hostname: "server2", Uptime: 2 * time.Hour},
					}, map[string]string{}, nil)
			},
			validateFunc: func(resp gen.GetNodeStatusResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name:    "broadcast all with errors",
			request: gen.GetNodeStatusRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeStatusBroadcast(gomock.Any(), "_all").
					Return("550e8400-e29b-41d4-a716-446655440000", []*jobtypes.NodeStatusResponse{
						{Hostname: "server1", Uptime: time.Hour},
					}, map[string]string{
						"server2": "disk full",
					}, nil)
			},
			validateFunc: func(resp gen.GetNodeStatusResponseObject) {
				r, ok := resp.(gen.GetNodeStatus200JSONResponse)
				s.True(ok)
				s.Len(r.Results, 2)
				var foundError bool
				for _, res := range r.Results {
					if res.Error != nil {
						foundError = true
						s.Equal("server2", res.Hostname)
						s.Equal("disk full", *res.Error)
					}
				}
				s.True(foundError)
			},
		},
		{
			name:    "broadcast all error",
			request: gen.GetNodeStatusRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeStatusBroadcast(gomock.Any(), "_all").
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeStatusResponseObject) {
				_, ok := resp.(gen.GetNodeStatus500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.GetNodeStatus(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestNodeStatusGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(NodeStatusGetPublicTestSuite))
}
