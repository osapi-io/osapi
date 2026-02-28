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
	"github.com/retr0h/osapi/internal/provider/node/mem"
	"github.com/retr0h/osapi/internal/validation"
)

type NodeMemoryGetPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinode.Node
	ctx           context.Context
}

func (s *NodeMemoryGetPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *NodeMemoryGetPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinode.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *NodeMemoryGetPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NodeMemoryGetPublicTestSuite) TestGetNodeMemory() {
	tests := []struct {
		name         string
		request      gen.GetNodeMemoryRequestObject
		setupMock    func()
		validateFunc func(resp gen.GetNodeMemoryResponseObject)
	}{
		{
			name:    "success",
			request: gen.GetNodeMemoryRequestObject{Hostname: "_any"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeMemory(gomock.Any(), "_any").
					Return("550e8400-e29b-41d4-a716-446655440000", &mem.Stats{
						Total:  8192,
						Free:   4096,
						Cached: 2048,
					}, "agent1", nil)
			},
			validateFunc: func(resp gen.GetNodeMemoryResponseObject) {
				_, ok := resp.(gen.GetNodeMemory200JSONResponse)
				s.True(ok)
			},
		},
		{
			name:      "validation error empty hostname",
			request:   gen.GetNodeMemoryRequestObject{Hostname: ""},
			setupMock: func() {},
			validateFunc: func(resp gen.GetNodeMemoryResponseObject) {
				r, ok := resp.(gen.GetNodeMemory400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name:    "job client error",
			request: gen.GetNodeMemoryRequestObject{Hostname: "_any"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeMemory(gomock.Any(), "_any").
					Return("", nil, "", assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeMemoryResponseObject) {
				_, ok := resp.(gen.GetNodeMemory500JSONResponse)
				s.True(ok)
			},
		},
		{
			name:    "broadcast all success",
			request: gen.GetNodeMemoryRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeMemoryBroadcast(gomock.Any(), "_all").
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*mem.Stats{
						"server1": {Total: 8192, Free: 4096, Cached: 2048},
						"server2": {Total: 16384, Free: 8192, Cached: 4096},
					}, map[string]string{}, nil)
			},
			validateFunc: func(resp gen.GetNodeMemoryResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name:    "broadcast all with errors",
			request: gen.GetNodeMemoryRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeMemoryBroadcast(gomock.Any(), "_all").
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*mem.Stats{
						"server1": {Total: 8192, Free: 4096, Cached: 2048},
					}, map[string]string{
						"server2": "some error",
					}, nil)
			},
			validateFunc: func(resp gen.GetNodeMemoryResponseObject) {
				r, ok := resp.(gen.GetNodeMemory200JSONResponse)
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
			request: gen.GetNodeMemoryRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeMemoryBroadcast(gomock.Any(), "_all").
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeMemoryResponseObject) {
				_, ok := resp.(gen.GetNodeMemory500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.GetNodeMemory(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestNodeMemoryGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(NodeMemoryGetPublicTestSuite))
}
