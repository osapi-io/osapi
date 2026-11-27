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
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type AgentTimelinePublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	ctx            context.Context
}

func (s *AgentTimelinePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNATSClient = jobmocks.NewMockNATSClient(s.mockCtrl)
	s.ctx = context.Background()
}

func (s *AgentTimelinePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *AgentTimelinePublicTestSuite) newClientWithState() *client.Client {
	opts := &client.Options{
		Timeout:     30 * time.Second,
		JobsBucket:  "jobs",
		StateBucket: "state",
	}
	c, err := newTrusted(s.mockNATSClient, opts)
	s.Require().NoError(err)

	return c
}

func (s *AgentTimelinePublicTestSuite) newClientWithoutState() *client.Client {
	opts := &client.Options{
		Timeout:    30 * time.Second,
		JobsBucket: "jobs",
	}
	c, err := newTrusted(s.mockNATSClient, opts)
	s.Require().NoError(err)

	return c
}

func (s *AgentTimelinePublicTestSuite) TestWriteAgentTimelineEvent() {
	tests := []struct {
		name         string
		hostname     string
		event        string
		message      string
		useState     bool
		marshalErr   bool
		setupMocks   func(*jobmocks.MockNATSClient)
		validateFunc func(error)
	}{
		{
			name:     "when write succeeds stores timeline event",
			hostname: "server1",
			event:    "drain",
			message:  "node marked for drain",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVPut(gomock.Any(), "state", gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context, _ string,
						key string,
						data []byte,
					) (uint64, error) {
						s.Contains(key, "timeline.machine_001.drain.")

						var te job.TimelineEvent
						err := json.Unmarshal(data, &te)
						s.NoError(err)
						s.Equal("drain", te.Event)
						s.Equal("server1", te.Hostname)
						s.Equal("node marked for drain", te.Message)
						s.NotZero(te.Timestamp)

						return 1, nil
					})
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:     "when KV put fails returns error",
			hostname: "server1",
			event:    "drain",
			message:  "drain requested",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVPut(gomock.Any(), "state", gomock.Any(), gomock.Any()).
					Return(uint64(0), errors.New("kv connection failed"))
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "write timeline event")
			},
		},
		{
			name:     "when there is no state bucket returns error",
			hostname: "server1",
			event:    "drain",
			message:  "drain requested",
			useState: false,
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "agent state bucket not configured")
			},
		},
		{
			name:       "when json marshal fails returns error",
			hostname:   "server1",
			event:      "drain",
			message:    "drain requested",
			useState:   true,
			marshalErr: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				// No KV expectations — marshal fails before Put
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "marshal timeline event")
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

			if tt.marshalErr {
				jobsClient.JSONMarshalFn = func(_ any) ([]byte, error) {
					return nil, errors.New("marshal failed")
				}
			}

			tt.validateFunc(jobsClient.WriteAgentTimelineEvent(
				s.ctx,
				"machine-001",
				tt.hostname,
				tt.event,
				tt.message,
			))
		})
	}
}

func (s *AgentTimelinePublicTestSuite) TestGetAgentTimeline() {
	now := time.Now()
	earlier := now.Add(-10 * time.Minute)
	later := now.Add(10 * time.Minute)

	tests := []struct {
		name          string
		hostname      string
		useState      bool
		setupMocks    func(*jobmocks.MockNATSClient)
		expectError   bool
		errorMsg      string
		expectedCount int
		validateFunc  func([]job.TimelineEvent)
	}{
		{
			name:     "when events exist returns sorted events",
			hostname: "server1",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{
						"timeline.machine_001.drain.1000000000",
						"timeline.machine_001.undrain.2000000000",
						"agents.server1",
					}, nil)

				drainEvent, _ := json.Marshal(job.TimelineEvent{
					Timestamp: later,
					Event:     "drain",
					Hostname:  "server1",
					Message:   "drain requested",
				})
				entry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry1.EXPECT().Value().Return(drainEvent)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "timeline.machine_001.drain.1000000000").
					Return(entry1, nil)

				undrainEvent, _ := json.Marshal(job.TimelineEvent{
					Timestamp: earlier,
					Event:     "undrain",
					Hostname:  "server1",
					Message:   "undrain requested",
				})
				entry2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry2.EXPECT().Value().Return(undrainEvent)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "timeline.machine_001.undrain.2000000000").
					Return(entry2, nil)
			},
			expectedCount: 2,
			validateFunc: func(events []job.TimelineEvent) {
				// Should be sorted by timestamp (earlier first)
				s.Equal("undrain", events[0].Event)
				s.Equal("drain", events[1].Event)
			},
		},
		{
			name:     "when the bucket is empty returns an empty timeline",
			hostname: "server1",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{}, nil)
			},
			expectedCount: 0,
		},
		{
			name:     "when the bucket cannot be read the failure is reported",
			hostname: "server1",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return(nil, errors.New("nats: connection closed"))
			},
			expectError: true,
			// An agent with no history and a backend nobody can read are not the
			// same answer.
			errorMsg: "list timeline keys",
		},
		{
			name:        "when there is no state bucket returns error",
			hostname:    "server1",
			useState:    false,
			expectError: true,
			errorMsg:    "agent state bucket not configured",
		},
		{
			name:     "when Get fails for a key skips it",
			hostname: "server1",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{
						"timeline.machine_001.drain.1000000000",
						"timeline.machine_001.undrain.2000000000",
					}, nil)

				drainEvent, _ := json.Marshal(job.TimelineEvent{
					Timestamp: now,
					Event:     "drain",
					Hostname:  "server1",
					Message:   "drain requested",
				})
				entry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry1.EXPECT().Value().Return(drainEvent)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "timeline.machine_001.drain.1000000000").
					Return(entry1, nil)

				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "timeline.machine_001.undrain.2000000000").
					Return(nil, errors.New("key not found"))
			},
			expectedCount: 1,
			validateFunc: func(events []job.TimelineEvent) {
				s.Equal("drain", events[0].Event)
			},
		},
		{
			name:     "when unmarshal fails for a key skips it",
			hostname: "server1",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{
						"timeline.machine_001.drain.1000000000",
						"timeline.machine_001.undrain.2000000000",
					}, nil)

				drainEvent, _ := json.Marshal(job.TimelineEvent{
					Timestamp: now,
					Event:     "drain",
					Hostname:  "server1",
					Message:   "drain requested",
				})
				entry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry1.EXPECT().Value().Return(drainEvent)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "timeline.machine_001.drain.1000000000").
					Return(entry1, nil)

				entry2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry2.EXPECT().Value().Return([]byte("invalid json"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "timeline.machine_001.undrain.2000000000").
					Return(entry2, nil)
			},
			expectedCount: 1,
			validateFunc: func(events []job.TimelineEvent) {
				s.Equal("drain", events[0].Event)
			},
		},
		{
			name:     "when keys exist for other hostnames filters them out",
			hostname: "server1",
			useState: true,
			setupMocks: func(_ *jobmocks.MockNATSClient) {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{
						"timeline.machine_001.drain.1000000000",
						"timeline.server2.drain.2000000000",
						"agents.server1",
					}, nil)

				drainEvent, _ := json.Marshal(job.TimelineEvent{
					Timestamp: now,
					Event:     "drain",
					Hostname:  "server1",
					Message:   "drain requested",
				})
				entry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry1.EXPECT().Value().Return(drainEvent)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "timeline.machine_001.drain.1000000000").
					Return(entry1, nil)
			},
			expectedCount: 1,
			validateFunc: func(events []job.TimelineEvent) {
				s.Equal("server1", events[0].Hostname)
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

			events, err := jobsClient.GetAgentTimeline(s.ctx, "machine-001", tt.hostname)

			if tt.expectError {
				s.Error(err)
				s.Contains(err.Error(), tt.errorMsg)
			} else {
				s.NoError(err)
				s.Len(events, tt.expectedCount)
				if tt.validateFunc != nil {
					tt.validateFunc(events)
				}
			}
		})
	}
}

func TestAgentTimelinePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AgentTimelinePublicTestSuite))
}
