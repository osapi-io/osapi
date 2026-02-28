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
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/validation"
)

type NodeOSGetPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinode.Node
	ctx           context.Context
}

func (s *NodeOSGetPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *NodeOSGetPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinode.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *NodeOSGetPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NodeOSGetPublicTestSuite) TestGetNodeOS() {
	tests := []struct {
		name         string
		request      gen.GetNodeOSRequestObject
		setupMock    func()
		validateFunc func(resp gen.GetNodeOSResponseObject)
	}{
		{
			name:    "success",
			request: gen.GetNodeOSRequestObject{Hostname: "_any"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeOS(gomock.Any(), "_any").
					Return("550e8400-e29b-41d4-a716-446655440000", &host.OSInfo{
						Distribution: "Ubuntu",
						Version:      "22.04",
					}, "agent1", nil)
			},
			validateFunc: func(resp gen.GetNodeOSResponseObject) {
				_, ok := resp.(gen.GetNodeOS200JSONResponse)
				s.True(ok)
			},
		},
		{
			name:      "validation error empty hostname",
			request:   gen.GetNodeOSRequestObject{Hostname: ""},
			setupMock: func() {},
			validateFunc: func(resp gen.GetNodeOSResponseObject) {
				r, ok := resp.(gen.GetNodeOS400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name:    "job client error",
			request: gen.GetNodeOSRequestObject{Hostname: "_any"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeOS(gomock.Any(), "_any").
					Return("", nil, "", assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeOSResponseObject) {
				_, ok := resp.(gen.GetNodeOS500JSONResponse)
				s.True(ok)
			},
		},
		{
			name:    "broadcast all success",
			request: gen.GetNodeOSRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeOSBroadcast(gomock.Any(), "_all").
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*host.OSInfo{
						"server1": {Distribution: "Ubuntu", Version: "22.04"},
						"server2": {Distribution: "CentOS", Version: "8.3"},
					}, map[string]string{}, nil)
			},
			validateFunc: func(resp gen.GetNodeOSResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name:    "broadcast all with errors",
			request: gen.GetNodeOSRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeOSBroadcast(gomock.Any(), "_all").
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*host.OSInfo{
						"server1": {Distribution: "Ubuntu", Version: "22.04"},
					}, map[string]string{
						"server2": "some error",
					}, nil)
			},
			validateFunc: func(resp gen.GetNodeOSResponseObject) {
				r, ok := resp.(gen.GetNodeOS200JSONResponse)
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
			request: gen.GetNodeOSRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeOSBroadcast(gomock.Any(), "_all").
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeOSResponseObject) {
				_, ok := resp.(gen.GetNodeOS500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.GetNodeOS(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestNodeOSGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(NodeOSGetPublicTestSuite))
}
