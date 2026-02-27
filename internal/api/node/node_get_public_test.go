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
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	apinode "github.com/retr0h/osapi/internal/api/node"
	"github.com/retr0h/osapi/internal/api/node/gen"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
)

type NodeGetPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinode.Node
	ctx           context.Context
}

func (s *NodeGetPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinode.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *NodeGetPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NodeGetPublicTestSuite) TestGetNodeDetails() {
	tests := []struct {
		name         string
		hostname     string
		mockAgent    *jobtypes.AgentInfo
		mockError    error
		validateFunc func(resp gen.GetNodeDetailsResponseObject)
	}{
		{
			name:     "success returns agent details",
			hostname: "server1",
			mockAgent: &jobtypes.AgentInfo{
				Hostname:     "server1",
				Labels:       map[string]string{"group": "web"},
				RegisteredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				StartedAt:    time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
				OSInfo:       &host.OSInfo{Distribution: "Ubuntu", Version: "24.04"},
				Uptime:       5 * time.Hour,
				LoadAverages: &load.AverageStats{Load1: 0.5, Load5: 0.3, Load15: 0.2},
				MemoryStats:  &mem.Stats{Total: 8388608, Free: 4194304, Cached: 2097152},
			},
			validateFunc: func(resp gen.GetNodeDetailsResponseObject) {
				r, ok := resp.(gen.GetNodeDetails200JSONResponse)
				s.True(ok)
				s.Equal("server1", r.Hostname)
				s.Equal(gen.Ready, r.Status)
				s.NotNil(r.Labels)
				s.NotNil(r.OsInfo)
				s.Equal("Ubuntu", r.OsInfo.Distribution)
				s.NotNil(r.LoadAverage)
				s.NotNil(r.Memory)
				s.NotNil(r.Uptime)
			},
		},
		{
			name:      "agent not found returns 404",
			hostname:  "unknown",
			mockError: fmt.Errorf("agent not found: unknown"),
			validateFunc: func(resp gen.GetNodeDetailsResponseObject) {
				_, ok := resp.(gen.GetNodeDetails404JSONResponse)
				s.True(ok)
			},
		},
		{
			name:      "client error returns 500",
			hostname:  "server1",
			mockError: fmt.Errorf("connection failed"),
			validateFunc: func(resp gen.GetNodeDetailsResponseObject) {
				_, ok := resp.(gen.GetNodeDetails500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockJobClient.EXPECT().
				GetAgent(gomock.Any(), tt.hostname).
				Return(tt.mockAgent, tt.mockError)

			resp, err := s.handler.GetNodeDetails(s.ctx, gen.GetNodeDetailsRequestObject{
				Hostname: tt.hostname,
			})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestNodeGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(NodeGetPublicTestSuite))
}
