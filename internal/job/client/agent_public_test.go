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
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type AgentPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	jobsClient     *client.Client
	ctx            context.Context
	mockNC         *jobmocks.MockNATSClient
	pub            ed25519.PublicKey
	priv           ed25519.PrivateKey
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
	s.mockNC = jobmocks.NewMockNATSClient(s.mockCtrl)
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)
	s.pub, s.priv = pub, priv
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
					Return(nil, jetstream.ErrKeyNotFound)
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
					Return(nil, jetstream.ErrKeyNotFound)
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
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state", "timeline.abc123.>").
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
					Return(nil, jetstream.ErrKeyNotFound)
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
					Return(nil, jetstream.ErrKeyNotFound)
				// GetAgentTimeline for hostname fallback
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state", "timeline.abc123.>").
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
					Return(nil, jetstream.ErrKeyNotFound)
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
					Return(nil, jetstream.ErrKeyNotFound)

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
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return(nil, errors.New("connection refused"))
				return s.newClientWithAllKVs("registry", "", "")
			},
			expectedErr: "connection refused",
		},
		{
			name:   "when the registry read fails returns the error, not a missing agent",
			target: "server1",
			setupClient: func() *client.Client {
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "registry", "agents.server1").
					Return(nil, errors.New("nats: timeout"))
				return s.newClientWithAllKVs("registry", "", "")
			},
			expectedErr: "read agent server1: nats: timeout",
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
					Return(nil, jetstream.ErrKeyNotFound)
				// GetAgentTimeline: returns one key
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state", "timeline.abc123.>").
					Return([]string{"timeline.abc123.drain.1000000000"}, nil)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "timeline.abc123.drain.1000000000").
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
					KVKeys(gomock.Any(), "state", "timeline.abc123.>").
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
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state", "timeline.abc123.>").
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
					Return(nil, jetstream.ErrKeyNotFound)
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
					Return(nil, jetstream.ErrKeyNotFound)
				// GetAgentTimeline for hostname fallback.
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state", "timeline.abc123.>").
					Return([]string{"timeline.abc123.drain.1000000000"}, nil)
				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "timeline.abc123.drain.1000000000").
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
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state", "timeline.abc123.>").
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
					Return(nil, jetstream.ErrKeyNotFound)

				s.mockNATSClient.EXPECT().
					KVGet(gomock.Any(), "state", "drain.abc123").
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state", "timeline.abc123.>").
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
					Return(nil, jetstream.ErrKeyNotFound)
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "state", "timeline.abc123.>").
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

// The keys of the tests: the controller's, which signs each job record,
// and the agents', which signs each response.
var (
	testControllerPub, testControllerPriv, _ = ed25519.GenerateKey(rand.Reader)
	testAgentPub, testAgentPriv, _           = ed25519.GenerateKey(rand.Reader)
)

// testSigner signs with the controller's key, and verifies a record with
// it, as the controller's job client does.
type testSigner struct{}

func (testSigner) Sign(data []byte) []byte { return ed25519.Sign(testControllerPriv, data) }

func (testSigner) Fingerprint() string { return "SHA256:test-controller" }

func (testSigner) ControllerPublicKey() ed25519.PublicKey { return testControllerPub }

func (testSigner) StoredRecordKeys() []ed25519.PublicKey {
	return []ed25519.PublicKey{testControllerPub}
}

// testAgentKeys is the admission store of the tests: each machine ID
// m-<hostname> is an agent admitted as <hostname>, with testAgentPub.
type testAgentKeys struct{}

func (testAgentKeys) RenameAgentKey(
	context.Context,
	string,
	string,
) error {
	return nil
}

func (testAgentKeys) LookupAgentKey(
	_ context.Context,
	machineID string,
) (*client.AgentKey, error) {
	return &client.AgentKey{
		MachineID: machineID,
		Hostname:  strings.TrimPrefix(machineID, "m-"),
		PublicKey: testAgentPub,
	}, nil
}

// trust gives c the keys of the controller, as the controller sets them at
// startup: the admission store, and a signer when opts names none.
func trust(
	c *client.Client,
	opts *client.Options,
) {
	if opts == nil || opts.PKISigner == nil {
		c.SetPKISigner(testSigner{})
	}
	c.SetAgentKeyStore(testAgentKeys{})
}

// envelope is payload signed with priv, from machineID.
func envelope(
	priv ed25519.PrivateKey,
	machineID string,
	payload []byte,
) []byte {
	data, _ := json.Marshal(job.SignedEnvelope{
		Payload:     payload,
		Signature:   ed25519.Sign(priv, payload),
		Fingerprint: "SHA256:test",
		MachineID:   machineID,
	})

	return data
}

// agentSigned is the response resp as the agent that it names writes it:
// signed, from m-<hostname>.
func agentSigned(
	resp []byte,
) []byte {
	var r struct {
		Hostname string `json:"hostname"`
	}
	_ = json.Unmarshal(resp, &r)

	return envelope(testAgentPriv, "m-"+r.Hostname, resp)
}

// controllerSigned is the job record as the controller writes it.
func controllerSigned(
	record []byte,
) []byte {
	return envelope(testControllerPriv, "", record)
}

// signingNATS is the NATS of the tests, with each value signed as the
// controller and the agents write it: a job record (jobs.<id>) by the
// controller, a response (responses.*) by its agent. The tests give the
// mock plain JSON.
type signingNATS struct {
	*jobmocks.MockNATSClient
}

// signed is the entry of key with its value signed as its writer signs it.
type signedEntry struct {
	jetstream.KeyValueEntry
	sign func([]byte) []byte
}

// Value is the value of the entry, signed. A value that a test gives as a
// signed envelope already is the value as the test wrote it.
func (e signedEntry) Value() []byte {
	value := e.KeyValueEntry.Value()
	var env job.SignedEnvelope
	if json.Unmarshal(value, &env) == nil && len(env.Signature) > 0 {
		return value
	}

	return e.sign(value)
}

func signerOf(
	key string,
) func([]byte) []byte {
	switch {
	case strings.HasPrefix(key, "jobs."):
		return controllerSigned
	case strings.HasPrefix(key, "responses."):
		return agentSigned
	default:
		return nil
	}
}

func (n signingNATS) KVGet(
	ctx context.Context,
	bucket string,
	key string,
) (jetstream.KeyValueEntry, error) {
	entry, err := n.MockNATSClient.KVGet(ctx, bucket, key)
	if sign := signerOf(key); entry != nil && sign != nil {
		return signedEntry{KeyValueEntry: entry, sign: sign}, err
	}

	return entry, err
}

func (n signingNATS) KVWatch(
	ctx context.Context,
	bucket string,
	pattern string,
	opts ...jetstream.WatchOpt,
) (<-chan jetstream.KeyValueEntry, error) {
	in, err := n.MockNATSClient.KVWatch(ctx, bucket, pattern, opts...)
	sign := signerOf(pattern)
	if in == nil || sign == nil {
		return in, err
	}
	out := make(chan jetstream.KeyValueEntry, cap(in))
	go func() {
		defer close(out)
		for e := range in {
			if e == nil {
				out <- nil

				continue
			}
			out <- signedEntry{KeyValueEntry: e, sign: sign}
		}
	}()

	return out, err
}

// newTrusted is a job client of the controller on nc: it signs, verifies,
// and reads the values of nc as its writers signed them.
func newTrusted(
	nc *jobmocks.MockNATSClient,
	opts *client.Options,
) (*client.Client, error) {
	c, err := client.New(slog.Default(), signingNATS{nc}, opts)
	if err == nil {
		trust(c, opts)
	}

	return c, err
}

// registry makes the registry bucket hold exactly this registration.
func (s *AgentPublicTestSuite) registry(
	reg job.AgentRegistration,
) {
	s.T().Helper()

	data, err := json.Marshal(reg)
	s.Require().NoError(err)

	s.mockNC.EXPECT().
		KVKeys(gomock.Any(), "registry").
		Return([]string{"agents." + reg.MachineID}, nil)

	entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
	entry.EXPECT().Value().Return(data)
	s.mockNC.EXPECT().
		KVGet(gomock.Any(), "registry", "agents."+reg.MachineID).
		Return(entry, nil)
}

func (s *AgentPublicTestSuite) signed(
	reg job.AgentRegistration,
) job.AgentRegistration {
	s.T().Helper()

	reg.Signature = ed25519.Sign(s.priv, job.RegistrationSigningBytes(&reg))

	return reg
}

func (s *AgentPublicTestSuite) base() job.AgentRegistration {
	return job.AgentRegistration{
		MachineID:    "machine-001",
		Hostname:     "web-01",
		State:        "Ready",
		RegisteredAt: time.Now(),
	}
}

func (s *AgentPublicTestSuite) TestListAgentsTrust() {
	tests := []struct {
		name         string
		reg          func() job.AgentRegistration
		setupStore   func(*jobmocks.MockAgentKeyStore)
		noStore      bool
		validateFunc func(job.AgentInfo)
	}{
		{
			name: "a registration signed by the stored key is authoritative",
			reg:  func() job.AgentRegistration { return s.signed(s.base()) },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "web-01",
						PublicKey: s.pub,
					}, nil)
			},
			validateFunc: func(a job.AgentInfo) {
				s.True(a.Verified)
				s.True(a.KeyStored)
			},
		},
		{
			name: "an unsigned registration is not authoritative",
			reg:  func() job.AgentRegistration { return s.base() },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "web-01",
						PublicKey: s.pub,
					}, nil)
			},
			validateFunc: func(a job.AgentInfo) {
				s.False(a.Verified)
				s.True(a.KeyStored, "the key is stored; the registration just is not signed")
			},
		},
		{
			name: "a registration signed by another key is not authoritative",
			reg:  func() job.AgentRegistration { return s.signed(s.base()) },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				other, _, _ := ed25519.GenerateKey(rand.Reader)
				st.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "web-01",
						PublicKey: other,
					}, nil)
			},
			validateFunc: func(a job.AgentInfo) {
				s.False(a.Verified)
			},
		},
		{
			name: "an agent with no stored key is visible but not authoritative",
			reg:  func() job.AgentRegistration { return s.signed(s.base()) },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(nil, client.ErrResponseKeyUnknown)
			},
			validateFunc: func(a job.AgentInfo) {
				s.False(a.Verified)
				s.False(a.KeyStored)
				s.Equal("web-01", a.Hostname, "still listed, so a rollout can be staged")
			},
		},
		{
			name: "an unreadable store never yields an authoritative registration",
			reg:  func() job.AgentRegistration { return s.signed(s.base()) },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(nil, client.ErrResponseStoreUnavailable)
			},
			validateFunc: func(a job.AgentInfo) {
				s.False(a.Verified)
			},
		},
		{
			name:    "with no store wired no registration is trusted",
			reg:     func() job.AgentRegistration { return s.base() },
			noStore: true,
			validateFunc: func(a job.AgentInfo) {
				s.False(a.Verified)
				s.False(a.KeyStored)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			reg := tt.reg()
			s.registry(reg)

			c, err := client.New(slog.Default(), s.mockNC, &client.Options{
				Timeout:        30 * time.Second,
				JobsBucket:     "jobs",
				RegistryBucket: "registry",
				StreamName:     "JOBS",
			})
			s.Require().NoError(err)

			if !tt.noStore {
				store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
				tt.setupStore(store)
				c.SetAgentKeyStore(store)
			}

			agents, err := c.ListAgents(s.ctx)
			s.Require().NoError(err)
			s.Require().Len(agents, 1)

			tt.validateFunc(agents[0])
		})
	}
}

// newClient is a job client of the mocked NATS, with store.
func (s *AgentPublicTestSuite) newClient(
	store client.AgentKeyStore,
) *client.Client {
	s.T().Helper()

	c, err := client.New(slog.Default(), s.mockNC, &client.Options{
		Timeout:        30 * time.Second,
		JobsBucket:     "jobs",
		RegistryBucket: "registry",
		StreamName:     "JOBS",
	})
	s.Require().NoError(err)
	c.SetAgentKeyStore(store)

	return c
}

// registryOf makes the registry bucket hold regs, for each read.
func (s *AgentPublicTestSuite) registryOf(
	regs ...job.AgentRegistration,
) {
	s.T().Helper()

	keys := make([]string, 0, len(regs))
	for _, reg := range regs {
		keys = append(keys, "agents."+reg.MachineID)
		data, err := json.Marshal(reg)
		s.Require().NoError(err)
		entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
		entry.EXPECT().Value().Return(data).AnyTimes()
		s.mockNC.EXPECT().
			KVGet(gomock.Any(), "registry", "agents."+reg.MachineID).
			Return(entry, nil).AnyTimes()
	}
	s.mockNC.EXPECT().KVKeys(gomock.Any(), "registry").Return(keys, nil).AnyTimes()
}

// TestHostnameFollows is a host whose hostname changed: its record follows a
// signed registration, unless another admitted machine that is online holds
// the name (GHSA-j73r).
func (s *AgentPublicTestSuite) TestHostnameFollows() {
	renamed := func() job.AgentRegistration {
		reg := s.base()
		reg.Hostname = "web-02"

		return s.signed(reg)
	}
	record := func(machineID, hostname string) *client.AgentKey {
		return &client.AgentKey{MachineID: machineID, Hostname: hostname, PublicKey: s.pub}
	}
	tests := []struct {
		name         string
		regs         func() []job.AgentRegistration
		setupStore   func(*jobmocks.MockAgentKeyStore)
		wantVerified bool
	}{
		{
			name: "when no other machine holds the new name, the record follows it",
			regs: func() []job.AgentRegistration { return []job.AgentRegistration{renamed()} },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().LookupAgentKey(gomock.Any(), "machine-001").
					Return(record("machine-001", "web-01"), nil).AnyTimes()
				st.EXPECT().RenameAgentKey(gomock.Any(), "machine-001", "web-02").Return(nil)
			},
			wantVerified: true,
		},
		{
			name: "when another admitted machine online holds the name, the claim is refused",
			regs: func() []job.AgentRegistration {
				holder := job.AgentRegistration{
					MachineID: "machine-002",
					Hostname:  "web-02",
					State:     "Ready",
				}

				return []job.AgentRegistration{renamed(), holder}
			},
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().LookupAgentKey(gomock.Any(), "machine-001").
					Return(record("machine-001", "web-01"), nil).AnyTimes()
				st.EXPECT().LookupAgentKey(gomock.Any(), "machine-002").
					Return(record("machine-002", "web-02"), nil).AnyTimes()
			},
		},
		{
			name: "when the record cannot be renamed, the claim is refused",
			regs: func() []job.AgentRegistration { return []job.AgentRegistration{renamed()} },
			setupStore: func(st *jobmocks.MockAgentKeyStore) {
				st.EXPECT().LookupAgentKey(gomock.Any(), "machine-001").
					Return(record("machine-001", "web-01"), nil).AnyTimes()
				st.EXPECT().RenameAgentKey(gomock.Any(), "machine-001", "web-02").
					Return(errors.New("kv: down"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.mockCtrl = gomock.NewController(s.T())
			s.mockNC = jobmocks.NewMockNATSClient(s.mockCtrl)
			s.registryOf(tc.regs()...)
			store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
			tc.setupStore(store)
			c := s.newClient(store)

			agents, err := c.ListAgents(s.ctx)

			s.Require().NoError(err)
			s.Equal(tc.wantVerified, agents[0].Verified)
		})
	}
}

// auditLog keeps each audit entry that the client writes.
type auditLog struct {
	entries []audit.Entry
	err     error
}

func (l *auditLog) Write(
	_ context.Context,
	e audit.Entry,
) error {
	l.entries = append(l.entries, e)

	return l.err
}

// TestHostnameRenameAudited: a rename that an agent reported, with no
// request, is an audit entry as the machine (osapi-io/osapi#1453). An audit
// log that fails does not undo the rename.
func (s *AgentPublicTestSuite) TestHostnameRenameAudited() {
	for _, fail := range []bool{false, true} {
		s.Run(fmt.Sprintf("the audit log fails: %v", fail), func() {
			s.mockCtrl = gomock.NewController(s.T())
			s.mockNC = jobmocks.NewMockNATSClient(s.mockCtrl)
			reg := s.base()
			reg.Hostname = "web-02"
			s.registryOf(s.signed(reg))
			store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
			store.EXPECT().LookupAgentKey(gomock.Any(), "machine-001").
				Return(&client.AgentKey{
					MachineID: "machine-001", Hostname: "web-01", PublicKey: s.pub,
				}, nil).AnyTimes()
			store.EXPECT().RenameAgentKey(gomock.Any(), "machine-001", "web-02").Return(nil)
			c := s.newClient(store)
			log := &auditLog{}
			if fail {
				log.err = errors.New("stream down")
			}
			c.SetAudit(log)

			agents, err := c.ListAgents(s.ctx)

			s.Require().NoError(err)
			s.True(agents[0].Verified)
			s.Require().Len(log.entries, 1)
			e := log.entries[0]
			s.Equal(audit.ActorAgent, e.Actor)
			s.Equal("machine-001", e.User)
			s.Equal("agent.rename", e.Action)
			s.Equal("hostname changed from web-01 to web-02", e.RequestSummary)
		})
	}
}

// TestHostnameRegistryUnreadable is a rename while the registry cannot be
// read for the check: the claim is refused, not taken on a guess.
func (s *AgentPublicTestSuite) TestHostnameRegistryUnreadable() {
	reg := s.base()
	reg.Hostname = "web-02"
	data, err := json.Marshal(s.signed(reg))
	s.Require().NoError(err)
	entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
	entry.EXPECT().Value().Return(data).AnyTimes()
	gomock.InOrder(
		s.mockNC.EXPECT().
			KVKeys(gomock.Any(), "registry").
			Return([]string{"agents.machine-001"}, nil),
		s.mockNC.EXPECT().KVKeys(gomock.Any(), "registry").Return(nil, errors.New("kv: down")),
	)
	s.mockNC.EXPECT().
		KVGet(gomock.Any(), "registry", "agents.machine-001").
		Return(entry, nil).
		AnyTimes()
	store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
	store.EXPECT().LookupAgentKey(gomock.Any(), "machine-001").
		Return(&client.AgentKey{MachineID: "machine-001", Hostname: "web-01", PublicKey: s.pub}, nil).
		AnyTimes()

	agents, err := s.newClient(store).ListAgents(s.ctx)

	s.Require().NoError(err)
	s.False(agents[0].Verified)
}

// TestHostnameEntryUnreadable is a registry entry that cannot be read
// during the check: it is skipped, and the rename goes on.
func (s *AgentPublicTestSuite) TestHostnameEntryUnreadable() {
	reg := s.base()
	reg.Hostname = "web-02"
	data, err := json.Marshal(s.signed(reg))
	s.Require().NoError(err)
	entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
	entry.EXPECT().Value().Return(data).AnyTimes()
	s.mockNC.EXPECT().KVKeys(gomock.Any(), "registry").
		Return([]string{"agents.machine-001", "agents.machine-002"}, nil).AnyTimes()
	s.mockNC.EXPECT().
		KVGet(gomock.Any(), "registry", "agents.machine-001").
		Return(entry, nil).
		AnyTimes()
	s.mockNC.EXPECT().KVGet(gomock.Any(), "registry", "agents.machine-002").
		Return(nil, errors.New("kv: gone")).AnyTimes()
	store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
	store.EXPECT().LookupAgentKey(gomock.Any(), "machine-001").
		Return(&client.AgentKey{MachineID: "machine-001", Hostname: "web-01", PublicKey: s.pub}, nil).
		AnyTimes()
	store.EXPECT().RenameAgentKey(gomock.Any(), "machine-001", "web-02").Return(nil)

	agents, err := s.newClient(store).ListAgents(s.ctx)

	s.Require().NoError(err)
	s.True(agents[0].Verified)
}
