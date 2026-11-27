// Copyright (c) 2025 John Dewey

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
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type AgentPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	jobsClient     *client.Client
	ctx            context.Context
}

func (s *AgentPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNATSClient = jobmocks.NewMockNATSClient(s.mockCtrl)
	s.ctx = context.Background()

	opts := &client.Options{
		Timeout:    30 * time.Second,
		JobsBucket: "jobs",
	}
	var err error
	s.jobsClient, err = newTrusted(s.mockNATSClient, opts)
	s.Require().NoError(err)
}

func (s *AgentPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *AgentPublicTestSuite) TearDownSubTest() {
	client.ResetSigningMarshalFn()
}

func (s *AgentPublicTestSuite) TestWriteStatusEvent() {
	tests := []struct {
		name         string
		jobID        string
		event        types.AgentJobStatus
		hostname     string
		data         map[string]interface{}
		kvError      error
		setupMocks   func()
		validateFunc func(error)
	}{
		{
			name:     "successful status event with data",
			jobID:    "job-123",
			event:    "started",
			hostname: "agent-1",
			data:     map[string]interface{}{"key": "value", "count": 42},
			setupMocks: func() {
				s.mockNATSClient.EXPECT().
					KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:     "successful status event without data",
			jobID:    "job-456",
			event:    "completed",
			hostname: "agent-2",
			data:     nil,
			setupMocks: func() {
				s.mockNATSClient.EXPECT().
					KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:     "hostname with special characters",
			jobID:    "job-789",
			event:    "failed",
			hostname: "agent.host-name@domain.com",
			data:     map[string]interface{}{"error": "timeout"},
			setupMocks: func() {
				s.mockNATSClient.EXPECT().
					KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:     "KV put error",
			jobID:    "job-error",
			event:    "started",
			hostname: "agent-1",
			data:     map[string]interface{}{"key": "value"},
			setupMocks: func() {
				s.mockNATSClient.EXPECT().
					KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
					Return(uint64(0), errors.New("kv connection failed"))
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "failed to write status event")
			},
		},
		{
			name:     "empty job ID",
			jobID:    "",
			event:    "started",
			hostname: "agent-1",
			data:     map[string]interface{}{"key": "value"},
			setupMocks: func() {
				s.mockNATSClient.EXPECT().
					KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:       "unmarshalable data",
			jobID:      "job-marshal",
			event:      "started",
			hostname:   "agent-1",
			data:       map[string]interface{}{"fn": make(chan int)},
			setupMocks: func() {},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "failed to marshal status event")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			tt.validateFunc(
				s.jobsClient.WriteStatusEvent(s.ctx, tt.jobID, tt.event, tt.hostname, tt.data),
			)
		})
	}
}

func (s *AgentPublicTestSuite) TestWriteJobResponse() {
	tests := []struct {
		name         string
		jobID        string
		hostname     string
		responseData []byte
		status       types.AgentJobStatus
		errorMsg     string
		errorCode    job.ErrorCode
		changed      *bool
		steps        []types.Step
		kvError      error
		validateFunc func(error)
	}{
		{
			name:         "successful job response completed",
			jobID:        "job-123",
			hostname:     "agent-1",
			responseData: []byte(`{"result": "success", "count": 42}`),
			status:       "completed",
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:         "successful job response with error",
			jobID:        "job-456",
			hostname:     "agent-2",
			responseData: []byte(`{"error": "processing failed"}`),
			status:       "failed",
			errorMsg:     "job execution failed",
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:         "when the failure has a cause, the stored response keeps its code",
			jobID:        "job-404",
			hostname:     "agent-2",
			responseData: []byte{},
			status:       "failed",
			errorMsg:     `user "nobody": not found`,
			errorCode:    job.ErrorCodeNotFound,
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:         "the steps of the method are stored with the response",
			jobID:        "job-steps",
			hostname:     "agent-1",
			responseData: []byte(`{"changed": true}`),
			status:       "completed",
			steps: []types.Step{{
				Behavior: "netplan", Method: "update",
				Steps: []types.Step{{Behavior: "managedfile", Method: "update"}},
			}},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:         "empty response data",
			jobID:        "job-789",
			hostname:     "agent-3",
			responseData: []byte{},
			status:       "completed",
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:         "hostname with special characters",
			jobID:        "job-special",
			hostname:     "agent.host-name@domain.com",
			responseData: []byte(`{"data": "test"}`),
			status:       "completed",
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:         "KV put error",
			jobID:        "job-error",
			hostname:     "agent-1",
			responseData: []byte(`{"result": "success"}`),
			status:       "completed",
			kvError:      errors.New("storage failure"),
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "failed to store job response")
			},
		},
		{
			name:     "large response data",
			jobID:    "job-large",
			hostname: "agent-1",
			responseData: []byte(
				`{"data": "large_data_payload_with_repeated_content_` + strings.Repeat(
					"x",
					500,
				) + `"}`,
			),
			status: "completed",
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var stored job.Response
			s.mockNATSClient.EXPECT().
				KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _, _ string, value []byte) (uint64, error) {
					// The agent stores its response signed.
					var envelope job.SignedEnvelope
					s.Require().NoError(json.Unmarshal(value, &envelope))
					s.Require().NoError(json.Unmarshal(envelope.Payload, &stored))
					return uint64(0), tt.kvError
				})

			tt.validateFunc(s.jobsClient.WriteJobResponse(
				s.ctx,
				tt.jobID,
				tt.hostname,
				tt.responseData,
				tt.status,
				tt.errorMsg,
				tt.errorCode,
				tt.changed,
				tt.steps,
			))
			s.Equal(tt.steps, stored.Steps, "the steps are in what the KV holds")
			s.Equal(
				tt.errorCode,
				stored.ErrorCode,
				"the controller reads the cause from the stored response",
			)
		})
	}
}

func (s *AgentPublicTestSuite) TestWriteJobResponseWithPKISigner() {
	signer, _ := newSigner(gomock.NewController(s.T()))

	tests := []struct {
		name         string
		jobID        string
		hostname     string
		responseData []byte
		status       types.AgentJobStatus
		errorMsg     string
		changed      *bool
		setup        func(*jobmocks.MockNATSClient)
		validateFunc func(error)
	}{
		{
			name:         "when PKI signer signs response before KV write",
			jobID:        "job-pki-123",
			hostname:     "agent-1",
			responseData: []byte(`{"result": "signed"}`),
			status:       "completed",
			setup: func(nats *jobmocks.MockNATSClient) {
				nats.EXPECT().
					KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:         "when PKI signer and KV write fails",
			jobID:        "job-pki-456",
			hostname:     "agent-2",
			responseData: []byte(`{"result": "fail"}`),
			status:       "completed",
			setup: func(nats *jobmocks.MockNATSClient) {
				nats.EXPECT().
					KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
					Return(uint64(0), errors.New("storage failure"))
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "failed to store job response")
			},
		},
		{
			name:         "when signing fails nothing is written",
			jobID:        "job-sign-fail",
			hostname:     "agent-1",
			responseData: []byte(`{"result": "test"}`),
			status:       "completed",
			setup: func(_ *jobmocks.MockNATSClient) {
				client.SetSigningMarshalFn(func(_ any) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "failed to sign response")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockNATS := jobmocks.NewMockNATSClient(s.mockCtrl)

			opts := &client.Options{
				Timeout:    30 * time.Second,
				JobsBucket: "jobs",
				PKISigner:  signer,
			}
			pkiClient, err := client.New(slog.Default(), mockNATS, opts)
			s.Require().NoError(err)

			tt.setup(mockNATS)

			tt.validateFunc(pkiClient.WriteJobResponse(
				s.ctx,
				tt.jobID,
				tt.hostname,
				tt.responseData,
				tt.status,
				tt.errorMsg,
				"",
				tt.changed,
				nil,
			))
		})
	}
}

func (s *AgentPublicTestSuite) TestConsumeJobs() {
	tests := []struct {
		name          string
		streamName    string
		consumerName  string
		handler       natsclient.JetStreamMessageHandler
		opts          *natsclient.ConsumeOptions
		consumeError  error
		invokeHandler bool
		validateFunc  func(error)
	}{
		{
			name:         "handler invoked per message",
			streamName:   "test-stream",
			consumerName: "test-consumer",
			handler: func(_ context.Context, _ jetstream.Msg) error {
				return nil
			},
			invokeHandler: true,
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:         "successful job consumption with options",
			streamName:   "jobs-stream",
			consumerName: "agent-consumer",
			handler: func(_ context.Context, _ jetstream.Msg) error {
				return nil
			},
			opts: &natsclient.ConsumeOptions{
				MaxInFlight: 5,
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:         "NATS consume error",
			streamName:   "test-stream",
			consumerName: "test-consumer",
			handler: func(_ context.Context, _ jetstream.Msg) error {
				return nil
			},
			consumeError: errors.New("stream not found"),
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "stream not found")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.invokeHandler {
				s.mockNATSClient.EXPECT().
					Consume(gomock.Any(), tt.streamName, tt.consumerName, gomock.Any(), tt.opts).
					DoAndReturn(func(
						_ context.Context,
						_ string,
						_ string,
						handlerFn natsclient.JetStreamMessageHandler,
						_ *natsclient.ConsumeOptions,
					) error {
						return handlerFn(context.Background(), nil)
					})
			} else {
				s.mockNATSClient.EXPECT().
					Consume(gomock.Any(), tt.streamName, tt.consumerName, gomock.Any(), tt.opts).
					Return(tt.consumeError)
			}

			tt.validateFunc(s.jobsClient.ConsumeJobs(
				s.ctx,
				tt.streamName,
				tt.consumerName,
				tt.handler,
				tt.opts,
			))
		})
	}
}

func (s *AgentPublicTestSuite) TestGetJobData() {
	tests := []struct {
		name         string
		jobKey       string
		setupMocks   func()
		validateFunc func([]byte, error)
	}{
		{
			name:   "successful get job data",
			jobKey: "jobs.job-123",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(`{"test": "data"}`))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "jobs", "jobs.job-123").
					Return(mockEntry, nil)
			},
			validateFunc: func(data []byte, err error) {
				s.NoError(err)
				// The job as the controller wrote it: the agent verifies it.
				s.Equal(controllerSigned([]byte(`{"test": "data"}`)), data)
			},
		},
		{
			name:   "job not found error",
			jobKey: "jobs.nonexistent",
			setupMocks: func() {
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "jobs", "jobs.nonexistent").
					Return(nil, errors.New("key not found"))
			},
			validateFunc: func(_ []byte, err error) {
				s.Error(err)
				s.Contains(err.Error(), "failed to get job data for key jobs.nonexistent")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			tt.validateFunc(s.jobsClient.GetJobData(s.ctx, tt.jobKey))
		})
	}
}

func (s *AgentPublicTestSuite) TestHasJobResponse() {
	tests := []struct {
		name         string
		jobID        string
		hostname     string
		setupMocks   func()
		validateFunc func(bool, error)
	}{
		{
			name:     "when a response exists for this agent",
			jobID:    "job-1",
			hostname: "agent-1",
			setupMocks: func() {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "jobs", "responses.job-1.agent-1.>").
					Return([]string{
						"responses.job-1.agent-1.1700000000000000000",
					}, nil)
			},
			validateFunc: func(got bool, err error) {
				s.NoError(err)
				s.True(got)
			},
		},
		{
			name:     "when no response exists for this agent",
			jobID:    "job-2",
			hostname: "agent-1",
			setupMocks: func() {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "jobs", "responses.job-2.agent-1.>").
					Return([]string{}, nil)
			},
			validateFunc: func(got bool, err error) {
				s.NoError(err)
				s.False(got)
			},
		},
		{
			name:     "when the hostname needs sanitizing for the key filter",
			jobID:    "job-3",
			hostname: "agent.one",
			setupMocks: func() {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "jobs", "responses.job-3.agent_one.>").
					Return([]string{}, nil)
			},
			validateFunc: func(got bool, err error) {
				s.NoError(err)
				s.False(got)
			},
		},
		{
			name:     "when the KV listing fails",
			jobID:    "job-4",
			hostname: "agent-1",
			setupMocks: func() {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "jobs", "responses.job-4.agent-1.>").
					Return(nil, errors.New("kv unavailable"))
			},
			validateFunc: func(got bool, err error) {
				s.Error(err)
				s.False(got)
				s.Contains(err.Error(), "failed to check for existing job response")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			got, err := s.jobsClient.HasJobResponse(s.ctx, tt.jobID, tt.hostname)
			tt.validateFunc(got, err)
		})
	}
}

func (s *AgentPublicTestSuite) TestCreateOrUpdateConsumer() {
	refused := errors.New("deliver policy can not be updated")
	cfg := jetstream.ConsumerConfig{
		Durable:       "test-consumer",
		DeliverPolicy: jetstream.DeliverNewPolicy,
	}
	existing := func(policy jetstream.DeliverPolicy) jetstream.Consumer {
		c := jobmocks.NewMockConsumer(s.mockCtrl)
		c.EXPECT().CachedInfo().Return(&jetstream.ConsumerInfo{
			Config: jetstream.ConsumerConfig{DeliverPolicy: policy},
		})

		return c
	}

	tests := []struct {
		name         string
		withStore    bool
		setupMocks   func(store *jobmocks.MockConsumerStore)
		validateFunc func(error)
	}{
		{
			name: "successful consumer creation",
			setupMocks: func(_ *jobmocks.MockConsumerStore) {
				s.mockNATSClient.EXPECT().
					ConsumerCreateOrUpdate(gomock.Any(), "test-stream", cfg).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "consumer creation error without a consumer store",
			setupMocks: func(_ *jobmocks.MockConsumerStore) {
				s.mockNATSClient.EXPECT().
					ConsumerCreateOrUpdate(gomock.Any(), "test-stream", cfg).
					Return(errors.New("consumer creation failed"))
			},
			validateFunc: func(err error) {
				s.ErrorContains(err, "consumer creation failed")
			},
		},
		{
			name:      "when the deliver policy changed the consumer is made again",
			withStore: true,
			setupMocks: func(store *jobmocks.MockConsumerStore) {
				gomock.InOrder(
					s.mockNATSClient.EXPECT().
						ConsumerCreateOrUpdate(gomock.Any(), "test-stream", cfg).
						Return(refused),
					store.EXPECT().
						Consumer(gomock.Any(), "test-stream", "test-consumer").
						Return(existing(jetstream.DeliverAllPolicy), nil),
					store.EXPECT().
						DeleteConsumer(gomock.Any(), "test-stream", "test-consumer").
						Return(nil),
					s.mockNATSClient.EXPECT().
						ConsumerCreateOrUpdate(gomock.Any(), "test-stream", cfg).
						Return(nil),
				)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:      "when the consumer is made again and the create fails",
			withStore: true,
			setupMocks: func(store *jobmocks.MockConsumerStore) {
				gomock.InOrder(
					s.mockNATSClient.EXPECT().
						ConsumerCreateOrUpdate(gomock.Any(), "test-stream", cfg).
						Return(refused),
					store.EXPECT().
						Consumer(gomock.Any(), "test-stream", "test-consumer").
						Return(existing(jetstream.DeliverAllPolicy), nil),
					store.EXPECT().
						DeleteConsumer(gomock.Any(), "test-stream", "test-consumer").
						Return(nil),
					s.mockNATSClient.EXPECT().
						ConsumerCreateOrUpdate(gomock.Any(), "test-stream", cfg).
						Return(errors.New("create failed")),
				)
			},
			validateFunc: func(err error) {
				s.ErrorContains(err, "create failed")
			},
		},
		{
			name:      "when the delete of the old consumer fails",
			withStore: true,
			setupMocks: func(store *jobmocks.MockConsumerStore) {
				s.mockNATSClient.EXPECT().
					ConsumerCreateOrUpdate(gomock.Any(), "test-stream", cfg).
					Return(refused)
				store.EXPECT().
					Consumer(gomock.Any(), "test-stream", "test-consumer").
					Return(existing(jetstream.DeliverAllPolicy), nil)
				store.EXPECT().
					DeleteConsumer(gomock.Any(), "test-stream", "test-consumer").
					Return(errors.New("delete failed"))
			},
			validateFunc: func(err error) {
				s.ErrorContains(err, "delete consumer test-consumer to change its deliver policy")
				s.ErrorContains(err, "delete failed")
			},
		},
		{
			name:      "when the deliver policy is the same the error is returned",
			withStore: true,
			setupMocks: func(store *jobmocks.MockConsumerStore) {
				s.mockNATSClient.EXPECT().
					ConsumerCreateOrUpdate(gomock.Any(), "test-stream", cfg).
					Return(errors.New("max consumers"))
				store.EXPECT().
					Consumer(gomock.Any(), "test-stream", "test-consumer").
					Return(existing(jetstream.DeliverNewPolicy), nil)
			},
			validateFunc: func(err error) {
				s.ErrorContains(err, "max consumers")
			},
		},
		{
			name:      "when the consumer lookup fails the error is returned",
			withStore: true,
			setupMocks: func(store *jobmocks.MockConsumerStore) {
				s.mockNATSClient.EXPECT().
					ConsumerCreateOrUpdate(gomock.Any(), "test-stream", cfg).
					Return(errors.New("max consumers"))
				store.EXPECT().
					Consumer(gomock.Any(), "test-stream", "test-consumer").
					Return(nil, jetstream.ErrConsumerNotFound)
			},
			validateFunc: func(err error) {
				s.ErrorContains(err, "max consumers")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			store := jobmocks.NewMockConsumerStore(s.mockCtrl)
			jobsClient := s.jobsClient
			if tt.withStore {
				var err error
				jobsClient, err = newTrusted(s.mockNATSClient, &client.Options{
					JobsBucket: "jobs",
					Consumers:  store,
				})
				s.Require().NoError(err)
			}
			tt.setupMocks(store)

			tt.validateFunc(jobsClient.CreateOrUpdateConsumer(s.ctx, "test-stream", cfg))
		})
	}
}

func (s *AgentPublicTestSuite) TestSanitizeKeyForNATS() {
	tests := []struct {
		name         string
		input        string
		validateFunc func(string)
	}{
		{
			name:  "valid characters only",
			input: "validKey123",
			validateFunc: func(got string) {
				s.Equal("validKey123", got)
			},
		},
		{
			name:  "alphanumeric with underscores and hyphens",
			input: "valid_key-123",
			validateFunc: func(got string) {
				s.Equal("valid_key-123", got)
			},
		},
		{
			name:  "hostname with dots",
			input: "server.example.com",
			validateFunc: func(got string) {
				s.Equal("server_example_com", got)
			},
		},
		{
			name:  "hostname with special characters",
			input: "agent.host-name@domain.com",
			validateFunc: func(got string) {
				s.Equal("agent_host-name_domain_com", got)
			},
		},
		{
			name:  "email-like string",
			input: "user@domain.com",
			validateFunc: func(got string) {
				s.Equal("user_domain_com", got)
			},
		},
		{
			name:  "string with spaces",
			input: "agent node 1",
			validateFunc: func(got string) {
				s.Equal("agent_node_1", got)
			},
		},
		{
			name:  "string with mixed special characters",
			input: "agent#1!@#$%^&*()",
			validateFunc: func(got string) {
				s.Equal("agent_1__________", got)
			},
		},
		{
			name:  "empty string",
			input: "",
			validateFunc: func(got string) {
				s.Equal("", got)
			},
		},
		{
			name:  "only special characters",
			input: "!@#$%^&*()",
			validateFunc: func(got string) {
				s.Equal("__________", got)
			},
		},
		{
			name:  "path-like string",
			input: "/path/to/resource",
			validateFunc: func(got string) {
				s.Equal("_path_to_resource", got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := client.ExportSanitizeKeyForNATS(tt.input)
			tt.validateFunc(got)
		})
	}
}

// newClientWithAllKVs creates a client with the registry, facts, and state
// buckets that it names. An empty name is a bucket that is not configured.
func (s *AgentPublicTestSuite) newClientWithAllKVs(
	registryBucket string,
	factsBucket string,
	stateBucket string,
) *client.Client {
	opts := &client.Options{
		Timeout:        30 * time.Second,
		JobsBucket:     "jobs",
		RegistryBucket: registryBucket,
		FactsBucket:    factsBucket,
		StateBucket:    stateBucket,
	}
	c, err := newTrusted(s.mockNATSClient, opts)
	s.Require().NoError(err)
	return c
}

// agentRegistrationJSON returns valid agent registration JSON for the given hostname.
func agentRegistrationJSON(
	hostname string,
) []byte {
	return agentRegistrationJSONWithMachineID(hostname, "abc123")
}

// agentRegistrationJSONWithMachineID returns valid agent registration JSON
// with a specific machine ID.
func agentRegistrationJSONWithMachineID(
	hostname, machineID string,
) []byte {
	return []byte(fmt.Sprintf(
		`{"machine_id":%q,"hostname":%q,"registered_at":"2026-01-01T00:00:00Z"}`,
		machineID, hostname,
	))
}

// factsRegistrationJSON returns valid facts registration JSON with sample data.
func factsRegistrationJSON() []byte {
	facts := job.FactsRegistration{
		Architecture:     "x86_64",
		KernelVersion:    "5.15.0",
		CPUCount:         4,
		FQDN:             "server1.example.com",
		ServiceMgr:       "systemd",
		PackageMgr:       "apt",
		PrimaryInterface: "eth0",
		Facts:            map[string]any{"os_family": "debian"},
	}
	data, _ := json.Marshal(facts)
	return data
}

func (s *AgentPublicTestSuite) TestListAgents() {
	tests := []struct {
		name         string
		setupClient  func() *client.Client
		expectedErr  string
		validateFunc func([]job.AgentInfo)
	}{
		{
			name: "when there is no registry bucket returns error",
			setupClient: func() *client.Client {
				// s.jobsClient has no registryKV configured
				return s.jobsClient
			},
			expectedErr: "agent registry not configured",
		},
		{
			name: "when Keys returns non-empty error returns error",
			setupClient: func() *client.Client {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return(nil, errors.New("connection refused"))
				return s.newClientWithAllKVs("registry", "", "")
			},
			expectedErr: "failed to list registry keys",
		},
		{
			name: "when the registry is empty returns empty slice",
			setupClient: func() *client.Client {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return([]string{}, nil)
				return s.newClientWithAllKVs("registry", "", "")
			},
			validateFunc: func(agents []job.AgentInfo) {
				s.Empty(agents)
			},
		},
		{
			name: "when key does not start with agents. prefix skips it",
			setupClient: func() *client.Client {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return([]string{"controllers.api", "agents.server1"}, nil)
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.server1").
					Return(entry, nil)
				return s.newClientWithAllKVs("registry", "", "")
			},
			validateFunc: func(agents []job.AgentInfo) {
				s.Len(agents, 1)
				s.Equal("server1", agents[0].Hostname)
			},
		},
		{
			name: "when Get fails for a key continues to next",
			setupClient: func() *client.Client {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return([]string{"agents.server1", "agents.server2"}, nil)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.server1").
					Return(nil, errors.New("key not found"))
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server2"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.server2").
					Return(entry, nil)
				return s.newClientWithAllKVs("registry", "", "")
			},
			validateFunc: func(agents []job.AgentInfo) {
				s.Len(agents, 1)
				s.Equal("server2", agents[0].Hostname)
			},
		},
		{
			name: "when unmarshal fails for a key continues to next",
			setupClient: func() *client.Client {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return([]string{"agents.server1", "agents.server2"}, nil)
				badEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				badEntry.EXPECT().Value().Return([]byte("not-valid-json"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.server1").
					Return(badEntry, nil)
				goodEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				goodEntry.EXPECT().Value().Return(agentRegistrationJSON("server2"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.server2").
					Return(goodEntry, nil)
				return s.newClientWithAllKVs("registry", "", "")
			},
			validateFunc: func(agents []job.AgentInfo) {
				s.Len(agents, 1)
				s.Equal("server2", agents[0].Hostname)
			},
		},
		{
			name: "when drain flag set and not applied overlays state to Draining",
			setupClient: func() *client.Client {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return([]string{"agents.abc123"}, nil)
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(entry, nil)

				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(jobmocks.NewMockKeyValueEntry(s.mockCtrl), nil)

				return s.newClientWithAllKVs("registry", "", "state")
			},
			validateFunc: func(agents []job.AgentInfo) {
				s.Len(agents, 1)
				s.Equal(types.AgentDraining, agents[0].State)
			},
		},
		{
			name: "when facts available merges into agent info",
			setupClient: func() *client.Client {
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return([]string{"agents.abc123"}, nil)
				regEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				regEntry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(regEntry, nil)

				factsEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				factsEntry.EXPECT().Value().Return(factsRegistrationJSON())
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "facts", "facts.abc123").
					Return(factsEntry, nil)

				return s.newClientWithAllKVs("registry", "facts", "")
			},
			validateFunc: func(agents []job.AgentInfo) {
				s.Len(agents, 1)
				s.Equal("x86_64", agents[0].Architecture)
				s.Equal("5.15.0", agents[0].KernelVersion)
				s.Equal(4, agents[0].CPUCount)
				s.Equal("server1.example.com", agents[0].FQDN)
				s.Equal("systemd", agents[0].ServiceMgr)
				s.Equal("apt", agents[0].PackageMgr)
				s.Equal("eth0", agents[0].PrimaryInterface)
				s.NotNil(agents[0].Facts)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			c := tt.setupClient()

			agents, err := c.ListAgents(s.ctx)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				s.Nil(agents)
			} else {
				s.NoError(err)
				if tt.validateFunc != nil {
					tt.validateFunc(agents)
				}
			}
		})
	}
}

func (s *AgentPublicTestSuite) TestGetAgent() {
	tests := []struct {
		name         string
		target       string
		setupClient  func() *client.Client
		expectedErr  string
		validateFunc func(*job.AgentInfo)
	}{
		{
			name:   "when there is no registry bucket returns error",
			target: "server1",
			setupClient: func() *client.Client {
				return s.jobsClient
			},
			expectedErr: "agent registry not configured",
		},
		{
			name:   "when direct machine ID lookup succeeds",
			target: "abc123",
			setupClient: func() *client.Client {
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(entry, nil)

				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(nil, errors.New("key not found"))
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{}, nil)

				return s.newClientWithAllKVs("registry", "", "state")
			},
			validateFunc: func(info *job.AgentInfo) {
				s.NotNil(info)
				s.Equal("server1", info.Hostname)
				s.Equal("abc123", info.MachineID)
			},
		},
		{
			name:   "when unmarshal fails on direct lookup returns error",
			target: "abc123",
			setupClient: func() *client.Client {
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return([]byte("not-valid-json"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(entry, nil)
				return s.newClientWithAllKVs("registry", "", "")
			},
			expectedErr: "failed to unmarshal agent registration",
		},
		{
			name:   "when hostname fallback finds agent",
			target: "server1",
			setupClient: func() *client.Client {
				// Direct lookup fails (server1 is not a machine ID key).
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.server1").
					Return(nil, errors.New("key not found"))
				// Fallback: ListAgents scans all keys.
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return([]string{"agents.abc123"}, nil)
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(entry, nil)

				// overlayDrainState in ListAgents
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(nil, errors.New("key not found"))
				// GetAgentTimeline for hostname fallback
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{}, nil)

				return s.newClientWithAllKVs("registry", "", "state")
			},
			validateFunc: func(info *job.AgentInfo) {
				s.NotNil(info)
				s.Equal("server1", info.Hostname)
				s.Equal("abc123", info.MachineID)
			},
		},
		{
			name:   "when hostname fallback finds no match returns error",
			target: "nonexistent",
			setupClient: func() *client.Client {
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.nonexistent").
					Return(nil, errors.New("key not found"))
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return([]string{"agents.abc123"}, nil)
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(entry, nil)

				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(nil, errors.New("key not found"))

				return s.newClientWithAllKVs("registry", "", "state")
			},
			expectedErr: "agent nonexistent: not found",
		},
		{
			name:   "when hostname fallback ListAgents fails returns error",
			target: "server1",
			setupClient: func() *client.Client {
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.server1").
					Return(nil, errors.New("key not found"))
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return(nil, errors.New("connection refused"))
				return s.newClientWithAllKVs("registry", "", "")
			},
			expectedErr: "agent server1: not found",
		},
		{
			name:   "when timeline available sets timeline on direct lookup",
			target: "abc123",
			setupClient: func() *client.Client {
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(entry, nil)

				timelineEvent, _ := json.Marshal(job.TimelineEvent{
					Timestamp: time.Now(),
					Event:     "drain",
					Hostname:  "server1",
					Message:   "drain requested",
				})
				timelineEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				timelineEntry.EXPECT().Value().Return(timelineEvent)

				// overlayDrainState: no drain flag
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(nil, errors.New("key not found"))
				// GetAgentTimeline: returns one key
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{"timeline.server1.drain.1000000000"}, nil)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "timeline.server1.drain.1000000000").
					Return(timelineEntry, nil)

				return s.newClientWithAllKVs("registry", "", "state")
			},
			validateFunc: func(info *job.AgentInfo) {
				s.NotNil(info)
				s.Equal("server1", info.Hostname)
				s.Len(info.Timeline, 1)
				s.Equal("drain", info.Timeline[0].Event)
			},
		},
		{
			name:   "when drain flag set and not applied state is Draining",
			target: "abc123",
			setupClient: func() *client.Client {
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(entry, nil)

				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(jobmocks.NewMockKeyValueEntry(s.mockCtrl), nil)
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{}, nil)

				return s.newClientWithAllKVs("registry", "", "state")
			},
			validateFunc: func(info *job.AgentInfo) {
				s.NotNil(info)
				s.Equal(types.AgentDraining, info.State)
			},
		},
		{
			name:   "when facts available merges into agent info",
			target: "abc123",
			setupClient: func() *client.Client {
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(entry, nil)

				factsEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				factsEntry.EXPECT().Value().Return(factsRegistrationJSON())
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "facts", "facts.abc123").
					Return(factsEntry, nil)

				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(nil, errors.New("key not found"))
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{}, nil)

				return s.newClientWithAllKVs("registry", "facts", "state")
			},
			validateFunc: func(info *job.AgentInfo) {
				s.NotNil(info)
				s.Equal("x86_64", info.Architecture)
				s.Equal("5.15.0", info.KernelVersion)
				s.Equal(4, info.CPUCount)
				s.Equal("server1.example.com", info.FQDN)
			},
		},
		{
			name:   "when timeline available on hostname fallback sets timeline",
			target: "server1",
			setupClient: func() *client.Client {
				// Direct lookup fails (server1 is not a machine ID).
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.server1").
					Return(nil, errors.New("key not found"))
				// ListAgents scan.
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return([]string{"agents.abc123"}, nil)
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(entry, nil)

				timelineEvent, _ := json.Marshal(job.TimelineEvent{
					Timestamp: time.Now(),
					Event:     "drain",
					Hostname:  "server1",
					Message:   "drain requested",
				})
				timelineEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				timelineEntry.EXPECT().Value().Return(timelineEvent)

				// overlayDrainState in ListAgents.
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(nil, errors.New("key not found"))
				// GetAgentTimeline for hostname fallback.
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{"timeline.server1.drain.1000000000"}, nil)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "timeline.server1.drain.1000000000").
					Return(timelineEntry, nil)

				return s.newClientWithAllKVs("registry", "", "state")
			},
			validateFunc: func(info *job.AgentInfo) {
				s.NotNil(info)
				s.Equal("server1", info.Hostname)
				s.Len(info.Timeline, 1)
				s.Equal("drain", info.Timeline[0].Event)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			c := tt.setupClient()

			info, err := c.GetAgent(s.ctx, tt.target)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				s.Nil(info)
			} else {
				s.NoError(err)
				if tt.validateFunc != nil {
					tt.validateFunc(info)
				}
			}
		})
	}
}

func (s *AgentPublicTestSuite) TestMergeFacts() {
	tests := []struct {
		name         string
		setupClient  func() *client.Client
		target       string
		expectedErr  string
		validateFunc func(*job.AgentInfo)
	}{
		{
			name:   "when there is no facts bucket does nothing",
			target: "abc123",
			setupClient: func() *client.Client {
				// No factsKV: mergeFacts is a no-op.
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(entry, nil)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(nil, errors.New("key not found"))
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{}, nil)
				return s.newClientWithAllKVs("registry", "", "state")
			},
			validateFunc: func(info *job.AgentInfo) {
				s.NotNil(info)
				s.Empty(info.Architecture)
				s.Empty(info.KernelVersion)
			},
		},
		{
			name:   "when factsKV Get fails does nothing",
			target: "abc123",
			setupClient: func() *client.Client {
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(entry, nil)

				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "facts", "facts.abc123").
					Return(nil, errors.New("key not found"))

				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(nil, errors.New("key not found"))
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{}, nil)

				return s.newClientWithAllKVs("registry", "facts", "state")
			},
			validateFunc: func(info *job.AgentInfo) {
				s.NotNil(info)
				s.Empty(info.Architecture)
			},
		},
		{
			name:   "when factsKV returns invalid JSON does nothing",
			target: "abc123",
			setupClient: func() *client.Client {
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(agentRegistrationJSON("server1"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.abc123").
					Return(entry, nil)

				factsEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				factsEntry.EXPECT().Value().Return([]byte("not-valid-json"))
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "facts", "facts.abc123").
					Return(factsEntry, nil)

				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(nil, errors.New("key not found"))
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state").
					Return([]string{}, nil)

				return s.newClientWithAllKVs("registry", "facts", "state")
			},
			validateFunc: func(info *job.AgentInfo) {
				s.NotNil(info)
				s.Empty(info.Architecture)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			c := tt.setupClient()

			info, err := c.GetAgent(s.ctx, tt.target)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
			} else {
				s.NoError(err)
				if tt.validateFunc != nil {
					tt.validateFunc(info)
				}
			}
		})
	}
}

func TestAgentPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AgentPublicTestSuite))
}
