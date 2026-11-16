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

package client_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type AgentDrainPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	ctx            context.Context
}

func (s *AgentDrainPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNATSClient = jobmocks.NewMockNATSClient(s.mockCtrl)
	s.ctx = context.Background()
}

func (s *AgentDrainPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *AgentDrainPublicTestSuite) newClientWithState() *client.Client {
	opts := &client.Options{
		Timeout:     30 * time.Second,
		JobsBucket:  "jobs",
		StateBucket: "state",
	}
	c, err := client.New(slog.Default(), s.mockNATSClient, opts)
	s.Require().NoError(err)

	return c
}

func (s *AgentDrainPublicTestSuite) newClientWithoutState() *client.Client {
	opts := &client.Options{
		Timeout:    30 * time.Second,
		JobsBucket: "jobs",
	}
	c, err := client.New(slog.Default(), s.mockNATSClient, opts)
	s.Require().NoError(err)

	return c
}

func (s *AgentDrainPublicTestSuite) TestCheckDrainFlag() {
	tests := []struct {
		name         string
		hostname     string
		useState     bool
		setupMocks   func(*jobmocks.MockNATSClient)
		validateFunc func(bool)
	}{
		{
			name:     "when drain flag exists returns true",
			hostname: "server1",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.server1").
					Return(entry, nil)
			},
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:     "when drain flag missing returns false",
			hostname: "server1",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.server1").
					Return(nil, errors.New("key not found"))
			},
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:     "when there is no state bucket returns false",
			hostname: "server1",
			useState: false,
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var jobsClient *client.Client
			if tt.useState {
				if tt.setupMocks != nil {
					tt.setupMocks(s.mockNATSClient)
				}
				jobsClient = s.newClientWithState()
			} else {
				jobsClient = s.newClientWithoutState()
			}

			result := jobsClient.CheckDrainFlag(s.ctx, tt.hostname)
			tt.validateFunc(result)
		})
	}
}

func (s *AgentDrainPublicTestSuite) TestSetDrainFlag() {
	tests := []struct {
		name         string
		hostname     string
		useState     bool
		setupMocks   func(*jobmocks.MockNATSClient)
		validateFunc func(error)
	}{
		{
			name:     "when write succeeds sets drain flag",
			hostname: "server1",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVPut(gomock.Any(), "state", "drain.server1", []byte("1")).
					Return(uint64(1), nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:     "when KV put fails returns error",
			hostname: "server1",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVPut(gomock.Any(), "state", "drain.server1", []byte("1")).
					Return(uint64(0), errors.New("kv connection failed"))
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "set drain flag")
			},
		},
		{
			name:     "when there is no state bucket returns error",
			hostname: "server1",
			useState: false,
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "agent state bucket not configured")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var jobsClient *client.Client
			if tt.useState {
				if tt.setupMocks != nil {
					tt.setupMocks(s.mockNATSClient)
				}
				jobsClient = s.newClientWithState()
			} else {
				jobsClient = s.newClientWithoutState()
			}

			tt.validateFunc(jobsClient.SetDrainFlag(s.ctx, tt.hostname))
		})
	}
}

func (s *AgentDrainPublicTestSuite) TestDeleteDrainFlag() {
	tests := []struct {
		name         string
		hostname     string
		useState     bool
		setupMocks   func(*jobmocks.MockNATSClient)
		validateFunc func(error)
	}{
		{
			name:     "when delete succeeds removes drain flag",
			hostname: "server1",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVDelete(gomock.Any(), "state", "drain.server1").
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:     "when KV delete fails returns error",
			hostname: "server1",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVDelete(gomock.Any(), "state", "drain.server1").
					Return(errors.New("kv connection failed"))
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "delete drain flag")
			},
		},
		{
			name:     "when there is no state bucket returns error",
			hostname: "server1",
			useState: false,
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "agent state bucket not configured")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var jobsClient *client.Client
			if tt.useState {
				if tt.setupMocks != nil {
					tt.setupMocks(s.mockNATSClient)
				}
				jobsClient = s.newClientWithState()
			} else {
				jobsClient = s.newClientWithoutState()
			}

			tt.validateFunc(jobsClient.DeleteDrainFlag(s.ctx, tt.hostname))
		})
	}
}

func (s *AgentDrainPublicTestSuite) TestOverlayDrainState() {
	tests := []struct {
		name         string
		useState     bool
		setupMocks   func(*jobmocks.MockNATSClient)
		validateFunc func(string)
	}{
		{
			name:     "when drain flag exists sets state to Cordoned",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(entry, nil)
			},
			validateFunc: func(got string) {
				s.Equal(job.AgentStateCordoned, got)
			},
		},
		{
			name:     "when drain flag missing keeps original state",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(nil, errors.New("key not found"))
			},
			validateFunc: func(got string) {
				s.Equal("", got)
			},
		},
		{
			name:     "when there is no state bucket keeps original state",
			useState: false,
			validateFunc: func(got string) {
				s.Equal("", got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Set up the registry KV to return agent data with machine_id
			entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
			entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
			s.mockNATSClient.EXPECT().
				KVGet(gomock.Any(), "registry", "agents.abc123").
				Return(entry, nil)

			opts := &client.Options{
				Timeout:        30 * time.Second,
				JobsBucket:     "jobs",
				RegistryBucket: "registry",
			}

			if tt.useState {
				if tt.setupMocks != nil {
					tt.setupMocks(s.mockNATSClient)
				}
				opts.StateBucket = "state"
				// GetAgent also calls GetAgentTimeline which uses stateKV
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{}, nil)
			}

			jobsClient, err := client.New(
				slog.Default(),
				s.mockNATSClient,
				opts,
			)
			s.Require().NoError(err)

			info, err := jobsClient.GetAgent(s.ctx, "abc123")
			s.NoError(err)
			tt.validateFunc(info.State)
		})
	}
}

func TestAgentDrainPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AgentDrainPublicTestSuite))
}
