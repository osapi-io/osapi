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

type NodeHostnameGetPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinode.Node
	ctx           context.Context
}

func (s *NodeHostnameGetPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.WorkerTarget, error) {
		return []validation.WorkerTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *NodeHostnameGetPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinode.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *NodeHostnameGetPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NodeHostnameGetPublicTestSuite) TestGetNodeHostname() {
	tests := []struct {
		name         string
		request      gen.GetNodeHostnameRequestObject
		setupMock    func()
		validateFunc func(resp gen.GetNodeHostnameResponseObject)
	}{
		{
			name:    "success",
			request: gen.GetNodeHostnameRequestObject{},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeHostname(gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", "my-hostname", &job.WorkerInfo{
						Hostname: "worker1",
						Labels:   map[string]string{"group": "web"},
					}, nil)
			},
			validateFunc: func(resp gen.GetNodeHostnameResponseObject) {
				r, ok := resp.(gen.GetNodeHostname200JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal("my-hostname", r.Results[0].Hostname)
				s.Require().NotNil(r.Results[0].Labels)
				s.Equal(map[string]string{"group": "web"}, *r.Results[0].Labels)
			},
		},
		{
			name:    "empty hostname falls back to worker hostname",
			request: gen.GetNodeHostnameRequestObject{},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeHostname(gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", "", &job.WorkerInfo{
						Hostname: "worker1",
					}, nil)
			},
			validateFunc: func(resp gen.GetNodeHostnameResponseObject) {
				r, ok := resp.(gen.GetNodeHostname200JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal("worker1", r.Results[0].Hostname)
			},
		},
		{
			name: "validation error empty target_hostname",
			request: gen.GetNodeHostnameRequestObject{
				Params: gen.GetNodeHostnameParams{TargetHostname: strPtr("")},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.GetNodeHostnameResponseObject) {
				r, ok := resp.(gen.GetNodeHostname400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "TargetHostname")
				s.Contains(*r.Error, "min")
			},
		},
		{
			name:    "job client error",
			request: gen.GetNodeHostnameRequestObject{},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeHostname(gomock.Any(), gomock.Any()).
					Return("", "", nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeHostnameResponseObject) {
				_, ok := resp.(gen.GetNodeHostname500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "broadcast all success",
			request: gen.GetNodeHostnameRequestObject{
				Params: gen.GetNodeHostnameParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeHostnameBroadcast(gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*job.WorkerInfo{
						"server1": {Hostname: "host1", Labels: map[string]string{"group": "web"}},
						"server2": {Hostname: "host2"},
					}, map[string]string{}, nil)
			},
			validateFunc: func(resp gen.GetNodeHostnameResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name: "broadcast all with errors",
			request: gen.GetNodeHostnameRequestObject{
				Params: gen.GetNodeHostnameParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeHostnameBroadcast(gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*job.WorkerInfo{
						"server1": {Hostname: "host1"},
					}, map[string]string{
						"server2": "interface not found",
					}, nil)
			},
			validateFunc: func(resp gen.GetNodeHostnameResponseObject) {
				r, ok := resp.(gen.GetNodeHostname200JSONResponse)
				s.True(ok)
				s.Len(r.Results, 2)
				var foundError bool
				for _, h := range r.Results {
					if h.Error != nil {
						foundError = true
						s.Equal("server2", h.Hostname)
						s.Equal("interface not found", *h.Error)
					}
				}
				s.True(foundError)
			},
		},
		{
			name: "broadcast all error",
			request: gen.GetNodeHostnameRequestObject{
				Params: gen.GetNodeHostnameParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeHostnameBroadcast(gomock.Any(), gomock.Any()).
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeHostnameResponseObject) {
				_, ok := resp.(gen.GetNodeHostname500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.GetNodeHostname(s.ctx, tt.request)
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

func TestNodeHostnameGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(NodeHostnameGetPublicTestSuite))
}
