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

package agent_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"

	"github.com/osapi-io/osapi/internal/agent"
	agentmocks "github.com/osapi-io/osapi/internal/agent/mocks"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/provider"
	diskMocks "github.com/osapi-io/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/osapi-io/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/osapi-io/osapi/internal/provider/node/load/mocks"
	memoryMocks "github.com/osapi-io/osapi/internal/provider/node/memory/mocks"
	"github.com/osapi-io/osapi/internal/secret/seal"
	"github.com/osapi-io/osapi/internal/telemetry/metrics"
	processMocks "github.com/osapi-io/osapi/internal/telemetry/process/mocks"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// newTestMsg creates a jetstream.Msg mock that returns the given subject,
// data, and nil headers. All three methods are expected any number of times,
// mirroring how the handler accesses message fields.
func newTestMsg(
	ctrl *gomock.Controller,
	subject string,
	data []byte,
) jetstream.Msg {
	m := agentmocks.NewMockMsg(ctrl)
	m.EXPECT().Subject().Return(subject).AnyTimes()
	m.EXPECT().Data().Return(data).AnyTimes()
	m.EXPECT().Headers().Return(nil).AnyTimes()
	// InProgress fires only if a test's operation outlives the (very long,
	// default-derived) keepalive interval; TermWithReason fires only for
	// pre-execution failures the handler decides can never be fixed by
	// redelivery. AnyTimes() permits both zero and multiple calls.
	m.EXPECT().InProgress().Return(nil).AnyTimes()
	m.EXPECT().TermWithReason(gomock.Any()).Return(nil).AnyTimes()
	return m
}

// expectNotAnswered sets up the standard "no prior response from this
// agent" HasJobResponse expectation shared by most handleJobMessage cases.
func (s *HandlerPublicTestSuite) expectNotAnswered(
	jobKey string,
) {
	s.mockJobClient.EXPECT().
		HasJobResponse(gomock.Any(), jobKey, gomock.Any()).
		Return(false, nil)
}

type HandlerPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	testAgent     *agent.Agent
}

func (s *HandlerPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)

	s.testAgent = newTestAgent(newTestAgentParams{
		jobClient:       s.mockJobClient,
		streamName:      "test-stream",
		hostProvider:    hostMocks.NewDefaultMockProvider(s.mockCtrl),
		diskProvider:    diskMocks.NewDefaultMockProvider(s.mockCtrl),
		memoryProvider:  memoryMocks.NewDefaultMockProvider(s.mockCtrl),
		loadProvider:    loadMocks.NewDefaultMockProvider(s.mockCtrl),
		processProvider: processMocks.NewDefaultMockProvider(s.mockCtrl),
	})

	// Set up a real MeterProvider so the instrumentation nil-guards are exercised.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err == nil {
		port := l.Addr().(*net.TCPAddr).Port
		_ = l.Close()
		srv, err := metrics.New("127.0.0.1", port, slog.Default())
		s.Require().NoError(err)
		s.testAgent.SetMeterProvider(srv.MeterProvider())
	}
}

func (s *HandlerPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *HandlerPublicTestSuite) TestWriteStatusEvent() {
	tests := []struct {
		name         string
		jobID        string
		event        types.AgentJobStatus
		data         map[string]interface{}
		setupMocks   func()
		validateFunc func(error)
	}{
		{
			name:  "when successful status event write",
			jobID: "test-job-123",
			event: "started",
			data:  map[string]interface{}{"agent_version": "1.0.0", "pid": 12345},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-123", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:  "when status event write with nil data",
			jobID: "test-job-456",
			event: "completed",
			data:  nil,
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-456", types.AgentJobStatusCompleted, gomock.Any(), nil).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:  "when status event write failure",
			jobID: "test-job-789",
			event: "failed",
			data:  map[string]interface{}{"error": "processing failed"},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-789", types.AgentJobStatusFailed, gomock.Any(), gomock.Any()).
					Return(errors.New("KV storage failed"))
			},
			validateFunc: func(err error) {
				s.Error(err)
				if "KV storage failed" != "" {
					s.Contains(err.Error(), "KV storage failed")
				}
			},
		},
		{
			name:  "when empty job ID",
			jobID: "",
			event: "started",
			data:  map[string]interface{}{},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			tt.validateFunc(agent.ExportWriteStatusEvent(
				context.Background(),
				s.testAgent,
				tt.jobID,
				tt.event,
				tt.data,
			))
		})
	}
}

func (s *HandlerPublicTestSuite) TestHandleJobMessage() {
	tests := []struct {
		name         string
		setupMsg     func(ctrl *gomock.Controller) jetstream.Msg
		setupMocks   func()
		validateFunc func(error)
	}{
		{
			name: "when successful job processing",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("test-job-123"))
			},
			setupMocks: func() {
				s.expectNotAnswered("test-job-123")

				// Mock job data retrieval
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.test-job-123").
					Return(signedJob(`{
						"id": "test-job-123",
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "1.1.1.1"}
						}
					}`), nil)

				// Mock status event writes
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-123", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				// The version of the agent is in its registration, not in each
				// started event.
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-123", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, _ types.AgentJobStatus, _ string, data map[string]interface{}) error {
						s.NotContains(data, "agent_version")
						s.Contains(data, "pid")

						return nil
					})

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-123", types.AgentJobStatusCompleted, gomock.Any(), gomock.Any()).
					Return(nil)

				// Mock response write
				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "test-job-123", gomock.Any(), gomock.Any(), types.AgentJobStatusCompleted, "", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "when job processing fails",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("test-job-456"))
			},
			setupMocks: func() {
				s.expectNotAnswered("test-job-456")

				// Mock job data retrieval
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.test-job-456").
					Return(signedJob(`{
						"id": "test-job-456",
						"operation": {
							"type": "node.unsupported.get",
							"data": {}
						}
					}`), nil)

				// Mock status event writes
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-456", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-456", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-456", types.AgentJobStatusFailed, gomock.Any(), gomock.Any()).
					Return(nil)

				// Mock response write for failed job
				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "test-job-456", gomock.Any(), gomock.Any(), types.AgentJobStatusFailed, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				// The operation ran and its failure is already recorded in
				// the response and status event, so the message is acked
				// (nil) rather than left to redeliver.
				s.NoError(err)
			},
		},
		{
			name: "when invalid subject format",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "invalid", []byte("test-job-789"))
			},
			setupMocks: func() {
				s.expectNotAnswered("test-job-789")
				// No further mocks needed: an unparsable subject is
				// terminated (msg.TermWithReason, stubbed in newTestMsg)
				// rather than retried.
			},
			validateFunc: func(err error) {
				s.Error(err)
				if "failed to parse subject" != "" {
					s.Contains(err.Error(), "failed to parse subject")
				}
			},
		},
		{
			name: "when job not found",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("nonexistent-job"))
			},
			setupMocks: func() {
				s.expectNotAnswered("nonexistent-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.nonexistent-job").
					Return(nil, errors.New("job not found"))
			},
			validateFunc: func(err error) {
				s.Error(err)
				if "job not found" != "" {
					s.Contains(err.Error(), "job not found")
				}
			},
		},
		{
			name: "when invalid job data format",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("invalid-job"))
			},
			setupMocks: func() {
				s.expectNotAnswered("invalid-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.invalid-job").
					Return(signedJob(`invalid json`), nil)
			},
			validateFunc: func(err error) {
				s.Error(err)
				if "failed to parse job data" != "" {
					s.Contains(err.Error(), "failed to parse job data")
				}
			},
		},
		{
			name: "when missing job ID",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("missing-id-job"))
			},
			setupMocks: func() {
				s.expectNotAnswered("missing-id-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.missing-id-job").
					Return(signedJob(`{
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "1.1.1.1"}
						}
					}`), nil)
			},
			validateFunc: func(err error) {
				s.Error(err)
				if "invalid job format: missing id" != "" {
					s.Contains(err.Error(), "invalid job format: missing id")
				}
			},
		},
		{
			name: "when missing operation",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("missing-op-job"))
			},
			setupMocks: func() {
				s.expectNotAnswered("missing-op-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.missing-op-job").
					Return(signedJob(`{
						"id": "missing-op-job"
					}`), nil)
			},
			validateFunc: func(err error) {
				s.Error(err)
				if "invalid job format: missing operation" != "" {
					s.Contains(err.Error(), "invalid job format: missing operation")
				}
			},
		},
		{
			name: "when missing operation type",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("missing-type-job"))
			},
			setupMocks: func() {
				s.expectNotAnswered("missing-type-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.missing-type-job").
					Return(signedJob(`{
						"id": "missing-type-job",
						"operation": {
							"data": {}
						}
					}`), nil)
			},
			validateFunc: func(err error) {
				s.Error(err)
				if "invalid operation format: missing type field" != "" {
					s.Contains(err.Error(), "invalid operation format: missing type field")
				}
			},
		},
		{
			name: "when invalid operation type format",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("invalid-type-job"))
			},
			setupMocks: func() {
				s.expectNotAnswered("invalid-type-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.invalid-type-job").
					Return(signedJob(`{
						"id": "invalid-type-job",
						"operation": {
							"type": "invalid",
							"data": {}
						}
					}`), nil)
			},
			validateFunc: func(err error) {
				s.Error(err)
				if "invalid operation type format" != "" {
					s.Contains(err.Error(), "invalid operation type format")
				}
			},
		},
		{
			name: "when acknowledged write error logged",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("ack-err-job"))
			},
			setupMocks: func() {
				s.expectNotAnswered("ack-err-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.ack-err-job").
					Return(signedJob(`{
						"id": "ack-err-job",
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "1.1.1.1"}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "ack-err-job", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(errors.New("ack write failed"))

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "ack-err-job", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "ack-err-job", types.AgentJobStatusCompleted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "ack-err-job", gomock.Any(), gomock.Any(), types.AgentJobStatusCompleted, "", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "when started write error logged",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("start-err-job"))
			},
			setupMocks: func() {
				s.expectNotAnswered("start-err-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.start-err-job").
					Return(signedJob(`{
						"id": "start-err-job",
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "1.1.1.1"}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "start-err-job", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "start-err-job", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(errors.New("started write failed"))

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "start-err-job", types.AgentJobStatusCompleted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "start-err-job", gomock.Any(), gomock.Any(), types.AgentJobStatusCompleted, "", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "when completed write error logged",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("comp-err-job"))
			},
			setupMocks: func() {
				s.expectNotAnswered("comp-err-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.comp-err-job").
					Return(signedJob(`{
						"id": "comp-err-job",
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "1.1.1.1"}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "comp-err-job", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "comp-err-job", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "comp-err-job", types.AgentJobStatusCompleted, gomock.Any(), gomock.Any()).
					Return(errors.New("completed write failed"))

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "comp-err-job", gomock.Any(), gomock.Any(), types.AgentJobStatusCompleted, "", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "when failed write error logged",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("fail-err-job"))
			},
			setupMocks: func() {
				s.expectNotAnswered("fail-err-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.fail-err-job").
					Return(signedJob(`{
						"id": "fail-err-job",
						"operation": {
							"type": "node.unsupported.get",
							"data": {}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fail-err-job", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fail-err-job", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fail-err-job", types.AgentJobStatusFailed, gomock.Any(), gomock.Any()).
					Return(errors.New("failed write failed"))

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "fail-err-job", gomock.Any(), gomock.Any(), types.AgentJobStatusFailed, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				// The response was persisted despite the failed-event write
				// error (which is only logged), so the job is still terminal.
				s.NoError(err)
			},
		},
		{
			name: "when processor returns ErrUnsupported sets skipped status",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("skip-job"))
			},
			setupMocks: func() {
				// The node processor returns ErrUnsupported for this one job.
				agent.SetAgentProcessorOnce(s.testAgent, "node",
					fmt.Errorf("linux: %w", provider.ErrUnsupported))

				s.expectNotAnswered("skip-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.skip-job").
					Return(signedJob(`{
						"id": "skip-job",
						"operation": {
							"type": "node.ntp.get",
							"data": {}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "skip-job", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "skip-job", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "skip-job", types.AgentJobStatusSkipped, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "skip-job", gomock.Any(), gomock.Any(), types.AgentJobStatusSkipped, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "when skipped write error logged",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("skip-err-job"))
			},
			setupMocks: func() {
				agent.SetAgentProcessorOnce(s.testAgent, "node",
					fmt.Errorf("linux: %w", provider.ErrUnsupported))

				s.expectNotAnswered("skip-err-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.skip-err-job").
					Return(signedJob(`{
						"id": "skip-err-job",
						"operation": {
							"type": "node.ntp.get",
							"data": {}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "skip-err-job", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "skip-err-job", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "skip-err-job", types.AgentJobStatusSkipped, gomock.Any(), gomock.Any()).
					Return(errors.New("skipped write failed"))

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "skip-err-job", gomock.Any(), gomock.Any(), types.AgentJobStatusSkipped, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "when fact reference resolved in job data",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("fact-resolve-job"))
			},
			setupMocks: func() {
				agent.SetAgentCachedFacts(s.testAgent, &job.FactsRegistration{
					PrimaryInterface: "eth0",
				})

				s.expectNotAnswered("fact-resolve-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.fact-resolve-job").
					Return(signedJob(`{
						"id": "fact-resolve-job",
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "@fact.interface.primary"}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fact-resolve-job", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fact-resolve-job", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fact-resolve-job", types.AgentJobStatusCompleted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "fact-resolve-job", gomock.Any(), gomock.Any(), types.AgentJobStatusCompleted, "", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "when fact reference with nil cached facts writes error to KV",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("fact-nil-job"))
			},
			setupMocks: func() {
				agent.SetAgentCachedFacts(s.testAgent, nil)

				s.expectNotAnswered("fact-nil-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.fact-nil-job").
					Return(signedJob(`{
						"id": "fact-nil-job",
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "@fact.interface.primary"}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fact-nil-job", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fact-nil-job", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fact-nil-job", types.AgentJobStatusFailed, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "fact-nil-job", gomock.Any(), gomock.Any(), types.AgentJobStatusFailed, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				// Fact resolution fails before the operation executes, but
				// the failure is still recorded via WriteJobResponse above,
				// so the message is acked rather than retried.
				s.NoError(err)
			},
		},
		{
			name: "when unresolvable fact reference writes error to KV",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("fact-fail-job"))
			},
			setupMocks: func() {
				agent.SetAgentCachedFacts(s.testAgent, &job.FactsRegistration{})

				s.expectNotAnswered("fact-fail-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.fact-fail-job").
					Return(signedJob(`{
						"id": "fact-fail-job",
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "@fact.nonexistent"}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fact-fail-job", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fact-fail-job", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fact-fail-job", types.AgentJobStatusFailed, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "fact-fail-job", gomock.Any(), gomock.Any(), types.AgentJobStatusFailed, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				// Same as the nil-facts case: recorded, so acked.
				s.NoError(err)
			},
		},
		{
			name: "when response storage failure",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("storage-fail-job"))
			},
			setupMocks: func() {
				s.expectNotAnswered("storage-fail-job")

				// Mock successful job processing
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.storage-fail-job").
					Return(signedJob(`{
						"id": "storage-fail-job",
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "1.1.1.1"}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "storage-fail-job", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "storage-fail-job", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				// Response is written before the status event; failure
				// here prevents the completed event from being written.
				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "storage-fail-job", gomock.Any(), gomock.Any(), types.AgentJobStatusCompleted, "", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("storage failure"))

				// The handler falls back to recording the storage failure
				// itself as a failed status event, rather than leaving the
				// message unacked (which would redeliver and re-execute
				// the already-completed operation).
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "storage-fail-job", types.AgentJobStatusFailed, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				// The operation ran exactly once (enforced by the mock
				// expectations above having no AnyTimes()/extra calls); a
				// storage failure after execution must still ack rather
				// than trigger redelivery and a second execution.
				s.NoError(err)
			},
		},
		{
			name: "when response storage failure and the failed-event write also fails",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(
					ctrl,
					"jobs.query.host.test-agent",
					[]byte("storage-fail-double-job"),
				)
			},
			setupMocks: func() {
				s.expectNotAnswered("storage-fail-double-job")

				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.storage-fail-double-job").
					Return(signedJob(`{
						"id": "storage-fail-double-job",
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "1.1.1.1"}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "storage-fail-double-job", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "storage-fail-double-job", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "storage-fail-double-job", gomock.Any(), gomock.Any(), types.AgentJobStatusCompleted, "", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("storage failure"))

				// Even the fallback failed-event write fails; this is only
				// logged, and the message is still acked rather than
				// redelivered for a second execution.
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "storage-fail-double-job", types.AgentJobStatusFailed, gomock.Any(), gomock.Any()).
					Return(errors.New("status write also failed"))
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "when the message cannot be terminated the original error still propagates",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				m := agentmocks.NewMockMsg(ctrl)
				m.EXPECT().Subject().Return("invalid").AnyTimes()
				m.EXPECT().Data().Return([]byte("term-err-job")).AnyTimes()
				m.EXPECT().Headers().Return(nil).AnyTimes()
				m.EXPECT().TermWithReason(gomock.Any()).Return(errors.New("term failed"))
				return m
			},
			setupMocks: func() {
				s.expectNotAnswered("term-err-job")
			},
			validateFunc: func(err error) {
				// TermWithReason failing is only logged; the original
				// parse error is still returned so the message is neither
				// silently dropped nor mistaken for success.
				s.Error(err)
				s.Contains(err.Error(), "failed to parse subject")
			},
		},
		{
			name: "when job already answered by this agent skips redelivered execution",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(
					ctrl,
					"jobs.query.host.test-agent",
					[]byte("already-answered-job"),
				)
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					HasJobResponse(gomock.Any(), "already-answered-job", gomock.Any()).
					Return(true, nil)
				// No GetJobData, WriteStatusEvent, or WriteJobResponse call
				// is expected: the operation must not run a second time.
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "when checking for an existing response fails proceeds with execution",
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("check-err-job"))
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					HasJobResponse(gomock.Any(), "check-err-job", gomock.Any()).
					Return(false, errors.New("kv unavailable"))

				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.check-err-job").
					Return(signedJob(`{
						"id": "check-err-job",
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "1.1.1.1"}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "check-err-job", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "check-err-job", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "check-err-job", types.AgentJobStatusCompleted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "check-err-job", gomock.Any(), gomock.Any(), types.AgentJobStatusCompleted, "", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			msg := tt.setupMsg(s.mockCtrl)
			tt.validateFunc(agent.ExportHandleJobMessage(context.Background(), s.testAgent, msg))
		})
	}
}

// TestHandleJobMessageConcurrently is the jobs of one consumer at once, as
// MaxInFlight (agent.max_jobs) lets nats-client run them: each one answers,
// and -race finds no state that the jobs share without a lock.
func (s *HandlerPublicTestSuite) TestHandleJobMessageConcurrently() {
	tests := []struct {
		name string
		jobs int
	}{
		{name: "when ten jobs run at once", jobs: 10},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			msgs := make([]jetstream.Msg, tt.jobs)
			for i := range tt.jobs {
				id := fmt.Sprintf("concurrent-job-%d", i)
				msgs[i] = newTestMsg(s.mockCtrl, "jobs.query.host.test-agent", []byte(id))
				s.expectNotAnswered(id)
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs."+id).
					Return(signedJob(`{"id": "`+id+`", "operation": {"type": "node.unsupported.get", "data": {}}}`), nil)
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), id, gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					Times(3)
				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), id, gomock.Any(), gomock.Any(), types.AgentJobStatusFailed, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			}

			errs := make([]error, tt.jobs)
			var wg sync.WaitGroup
			for i, msg := range msgs {
				wg.Add(1)
				go func() {
					defer wg.Done()
					errs[i] = agent.ExportHandleJobMessage(context.Background(), s.testAgent, msg)
				}()
			}
			wg.Wait()

			for i, err := range errs {
				s.NoError(err, "job %d", i)
			}
		})
	}
}

func (s *HandlerPublicTestSuite) TestHandleJobMessageModifyJobs() {
	tests := []struct {
		name         string
		subject      string
		jobData      string
		setupMocks   func()
		validateFunc func(error)
	}{
		{
			name:    "when modify job type identification",
			subject: "jobs.modify.host.test-agent",
			jobData: `{
				"id": "modify-job-123",
				"operation": {
					"type": "network.ping.do",
					"data": {"address": "1.1.1.1"}
				}
			}`,
			setupMocks: func() {
				s.expectNotAnswered("modify-job-123")

				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.modify-job-123").
					Return(signedJob(`{
						"id": "modify-job-123",
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "1.1.1.1"}
						}
					}`), nil)

				// Mock status events with specific job ID.
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "modify-job-123", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()

				// Mock response write — verify job ID and completed status.
				s.mockJobClient.EXPECT().
					WriteJobResponse(
						gomock.Any(),
						"modify-job-123",
						gomock.Any(),
						gomock.Any(),
						types.AgentJobStatusCompleted,
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			msg := newTestMsg(s.mockCtrl, tt.subject, []byte("modify-job-123"))
			tt.validateFunc(agent.ExportHandleJobMessage(context.Background(), s.testAgent, msg))
		})
	}
}

func (s *HandlerPublicTestSuite) TestExtractChanged() {
	tests := []struct {
		name         string
		data         json.RawMessage
		validateFunc func(*bool)
	}{
		{
			name: "when empty data returns nil",
			data: nil,
			validateFunc: func(got *bool) {
				s.Equal((*bool)(nil), got)
			},
		},
		{
			name: "when invalid JSON returns nil",
			data: json.RawMessage(`not json`),
			validateFunc: func(got *bool) {
				s.Equal((*bool)(nil), got)
			},
		},
		{
			name: "when changed key missing returns nil",
			data: json.RawMessage(`{"success":true}`),
			validateFunc: func(got *bool) {
				s.Equal((*bool)(nil), got)
			},
		},
		{
			name: "when changed is non-bool returns nil",
			data: json.RawMessage(`{"changed":"yes"}`),
			validateFunc: func(got *bool) {
				s.Equal((*bool)(nil), got)
			},
		},
		{
			name: "when changed is true returns true",
			data: json.RawMessage(`{"changed":true}`),
			validateFunc: func(got *bool) {
				s.Equal(ptr.To(true), got)
			},
		},
		{
			name: "when changed is false returns false",
			data: json.RawMessage(`{"changed":false}`),
			validateFunc: func(got *bool) {
				s.Equal(ptr.To(false), got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := agent.ExportExtractChanged(tt.data)

			tt.validateFunc(got)
		})
	}
}

func (s *HandlerPublicTestSuite) TestUnwrapJobEnvelope() {
	// Generate a keypair for signing.
	controllerPub, controllerPriv, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)

	tests := []struct {
		name         string
		setupPKI     func() *pki.Manager
		data         func() []byte
		expectError  bool
		errorMsg     string
		validateFunc func(string)
	}{
		{
			name: "when the agent has no keys refuses each job",
			setupPKI: func() *pki.Manager {
				return nil
			},
			data: func() []byte {
				return signedJob(`{"id":"test"}`)
			},
			expectError: true,
			errorMsg:    "agent not enrolled",
		},
		{
			name: "when valid signed envelope with correct controller key",
			setupPKI: func() *pki.Manager {
				m := pki.New(memfs.New(), "/tmp/pki", "agent")
				s.Require().NoError(m.LoadOrGenerate())
				m.SetControllerPublicKey(controllerPub)
				return m
			},
			data: func() []byte {
				payload := []byte(`{"id":"signed-test"}`)
				sig := ed25519.Sign(controllerPriv, payload)
				envelope := job.SignedEnvelope{
					Payload:     payload,
					Signature:   sig,
					Fingerprint: "SHA256:controller-fp",
				}
				data, _ := json.Marshal(envelope)
				return data
			},
			expectError: false,
			validateFunc: func(got string) {
				s.Equal(`{"id":"signed-test"}`, got)
			},
		},
		{
			name: "when signed envelope with invalid signature",
			setupPKI: func() *pki.Manager {
				m := pki.New(memfs.New(), "/tmp/pki", "agent")
				s.Require().NoError(m.LoadOrGenerate())
				m.SetControllerPublicKey(controllerPub)
				return m
			},
			data: func() []byte {
				payload := []byte(`{"id":"bad-sig"}`)
				envelope := job.SignedEnvelope{
					Payload: payload,
					Signature: []byte(
						"invalid-signature-data-that-is-long-enough-for-ed25519-64-bytes!!",
					),
					Fingerprint: "SHA256:bad-fp",
				}
				data, _ := json.Marshal(envelope)
				return data
			},
			expectError: true,
			errorMsg:    "invalid controller signature",
			validateFunc: func(got string) {
				s.Equal("", got)
			},
		},
		{
			name: "when signed envelope without cached controller key is rejected distinctly",
			setupPKI: func() *pki.Manager {
				m := pki.New(memfs.New(), "/tmp/pki", "agent")
				s.Require().NoError(m.LoadOrGenerate())
				// No controller public key set — agent has not enrolled.
				return m
			},
			data: func() []byte {
				payload := []byte(`{"id":"no-ctrl-key"}`)
				sig := ed25519.Sign(controllerPriv, payload)
				envelope := job.SignedEnvelope{
					Payload:     payload,
					Signature:   sig,
					Fingerprint: "SHA256:controller-fp",
				}
				data, _ := json.Marshal(envelope)
				return data
			},
			expectError: true,
			errorMsg:    "no cached controller public key",
			validateFunc: func(got string) {
				s.Equal("", got)
			},
		},
		{
			name: "when raw JSON with PKI enabled is rejected as unsigned",
			setupPKI: func() *pki.Manager {
				m := pki.New(memfs.New(), "/tmp/pki", "agent")
				s.Require().NoError(m.LoadOrGenerate())
				m.SetControllerPublicKey(controllerPub)
				return m
			},
			data: func() []byte {
				return []byte(
					`{"id":"raw-job","operation":{"type":"network.ping.do","data":{"address":"1.1.1.1"}}}`,
				)
			},
			expectError: true,
			errorMsg:    "not a signed envelope",
			validateFunc: func(got string) {
				s.Equal("", got)
			},
		},
		{
			name: "when invalid JSON with PKI enabled is rejected as unsigned",
			setupPKI: func() *pki.Manager {
				m := pki.New(memfs.New(), "/tmp/pki", "agent")
				s.Require().NoError(m.LoadOrGenerate())
				return m
			},
			data: func() []byte {
				return []byte(`not json at all`)
			},
			expectError: true,
			errorMsg:    "not a signed envelope",
			validateFunc: func(got string) {
				s.Equal("", got)
			},
		},
		{
			name: "when signed with the previous controller key during rotation grace",
			setupPKI: func() *pki.Manager {
				m := pki.New(memfs.New(), "/tmp/pki", "agent")
				s.Require().NoError(m.LoadOrGenerate())
				// Simulate a rotation: controllerPub is now "previous",
				// and a freshly generated key is current.
				m.SetControllerPublicKey(controllerPub)
				newPub, _, genErr := ed25519.GenerateKey(rand.Reader)
				s.Require().NoError(genErr)
				m.RotateControllerKey(newPub)
				return m
			},
			data: func() []byte {
				payload := []byte(`{"id":"rotation-grace"}`)
				// Signed with the now-previous key.
				sig := ed25519.Sign(controllerPriv, payload)
				envelope := job.SignedEnvelope{
					Payload:     payload,
					Signature:   sig,
					Fingerprint: "SHA256:controller-fp",
				}
				data, _ := json.Marshal(envelope)
				return data
			},
			expectError: false,
			validateFunc: func(got string) {
				s.Equal(`{"id":"rotation-grace"}`, got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			pkiMgr := tt.setupPKI()
			agent.SetAgentPKIManager(s.testAgent, pkiMgr)
			defer agent.SetAgentPKIManager(s.testAgent, nil)

			data := tt.data()
			result, err := agent.ExportUnwrapJobEnvelope(s.testAgent, data)

			if tt.expectError {
				s.Error(err)
				if tt.errorMsg != "" {
					s.Contains(err.Error(), tt.errorMsg)
				}
				return
			}

			s.NoError(err)
			tt.validateFunc(string(result))
		})
	}
}

// TestHandleJobMessageWithSealedSecrets: an agent opens the job key that
// the controller sealed to it, and fails a job that has no key for it.
func (s *HandlerPublicTestSuite) TestHandleJobMessageWithSealedSecrets() {
	controllerPub, controllerPriv, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)
	m := pki.New(memfs.New(), "/tmp/pki", "agent", pki.WithEncryptionKey())
	s.Require().NoError(m.LoadOrGenerate())
	m.SetControllerPublicKey(controllerPub)
	// noXKey is an enrolled agent with no P-256 encryption key.
	noXKey := pki.New(memfs.New(), "/tmp/pki", "agent")
	s.Require().NoError(noXKey.LoadOrGenerate())
	noXKey.SetControllerPublicKey(controllerPub)

	// jobData is the job jobID with its job key sealed to the agent m.
	jobData := func(jobID string) []byte {
		sealer, err := seal.New(jobID)
		s.Require().NoError(err)
		key, err := sealer.KeyFor(m.EncryptionPublicKey())
		s.Require().NoError(err)
		payload, _ := json.Marshal(map[string]any{
			"id": jobID,
			"operation": map[string]any{
				"type":        "network.ping.do",
				"data":        map[string]any{"address": "1.1.1.1"},
				"sealed_keys": map[string]string{"m-1": key},
				"sealed_from": sealer.From(),
			},
		})
		envelope, _ := json.Marshal(job.SignedEnvelope{
			Payload:     payload,
			Signature:   ed25519.Sign(controllerPriv, payload),
			Fingerprint: "SHA256:ctrl",
		})

		return envelope
	}

	tests := []struct {
		name       string
		jobID      string
		pki        *pki.Manager
		wantStatus types.AgentJobStatus
	}{
		{
			name:       "when the job key is sealed to this agent, the job runs",
			jobID:      "sealed-ok",
			pki:        m,
			wantStatus: "completed",
		},
		{
			name:       "when the agent has no encryption key, the job fails and does not run",
			jobID:      "sealed-none",
			pki:        noXKey,
			wantStatus: "failed",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			agent.SetAgentPKIManager(s.testAgent, tt.pki)
			agent.SetAgentMachineID(s.testAgent, "m-1")
			defer agent.SetAgentPKIManager(s.testAgent, nil)

			s.expectNotAnswered(tt.jobID)
			s.mockJobClient.EXPECT().
				GetJobData(gomock.Any(), "jobs."+tt.jobID).
				Return(jobData(tt.jobID), nil)
			for _, event := range []types.AgentJobStatus{types.AgentJobStatusAcknowledged, types.AgentJobStatusStarted, tt.wantStatus} {
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), tt.jobID, event, gomock.Any(), gomock.Any()).
					Return(nil)
			}
			s.mockJobClient.EXPECT().
				WriteJobResponse(gomock.Any(), tt.jobID, gomock.Any(), gomock.Any(), tt.wantStatus, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil)

			msg := newTestMsg(s.mockCtrl, "jobs.query.host.test-agent", []byte(tt.jobID))
			s.NoError(agent.ExportHandleJobMessage(context.Background(), s.testAgent, msg))
		})
	}
}

func (s *HandlerPublicTestSuite) TestHandleJobMessageWithSignedEnvelope() {
	// Generate controller keypair for signing job data.
	controllerPub, controllerPriv, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)

	tests := []struct {
		name         string
		setupPKI     func()
		cleanupPKI   func()
		setupMsg     func(ctrl *gomock.Controller) jetstream.Msg
		setupMocks   func()
		validateFunc func(error)
	}{
		{
			name: "when signed job data processed successfully",
			setupPKI: func() {
				m := pki.New(memfs.New(), "/tmp/pki", "agent")
				s.Require().NoError(m.LoadOrGenerate())
				m.SetControllerPublicKey(controllerPub)
				agent.SetAgentPKIManager(s.testAgent, m)
			},
			cleanupPKI: func() {
				agent.SetAgentPKIManager(s.testAgent, nil)
			},
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("signed-job-ok"))
			},
			setupMocks: func() {
				// Create signed job data.
				payload := []byte(
					`{"id":"signed-job-ok","operation":{"type":"network.ping.do","data":{"address":"1.1.1.1"}}}`,
				)
				sig := ed25519.Sign(controllerPriv, payload)
				envelope := job.SignedEnvelope{
					Payload:     payload,
					Signature:   sig,
					Fingerprint: "SHA256:ctrl",
				}
				envelopeJSON, _ := json.Marshal(envelope)

				s.expectNotAnswered("signed-job-ok")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.signed-job-ok").
					Return(envelopeJSON, nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "signed-job-ok", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "signed-job-ok", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "signed-job-ok", types.AgentJobStatusCompleted, gomock.Any(), gomock.Any()).
					Return(nil)
				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "signed-job-ok", gomock.Any(), gomock.Any(), types.AgentJobStatusCompleted, "", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "when signed job data with invalid signature fails",
			setupPKI: func() {
				m := pki.New(memfs.New(), "/tmp/pki", "agent")
				s.Require().NoError(m.LoadOrGenerate())
				m.SetControllerPublicKey(controllerPub)
				agent.SetAgentPKIManager(s.testAgent, m)
			},
			cleanupPKI: func() {
				agent.SetAgentPKIManager(s.testAgent, nil)
			},
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("bad-sig-job"))
			},
			setupMocks: func() {
				// Create job data with invalid signature.
				payload := []byte(
					`{"id":"bad-sig-job","operation":{"type":"network.ping.do","data":{"address":"1.1.1.1"}}}`,
				)
				envelope := job.SignedEnvelope{
					Payload: payload,
					Signature: []byte(
						"bad-signature-data-that-is-at-least-64-bytes-long-for-ed25519!!!",
					),
					Fingerprint: "SHA256:bad",
				}
				envelopeJSON, _ := json.Marshal(envelope)

				s.expectNotAnswered("bad-sig-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.bad-sig-job").
					Return(envelopeJSON, nil)
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "job signature verification failed")
			},
		},
		{
			name: "when agent has not completed enrollment rejects distinctly",
			setupPKI: func() {
				m := pki.New(memfs.New(), "/tmp/pki", "agent")
				s.Require().NoError(m.LoadOrGenerate())
				// No controller public key — agent is pending, not enrolled.
				agent.SetAgentPKIManager(s.testAgent, m)
			},
			cleanupPKI: func() {
				agent.SetAgentPKIManager(s.testAgent, nil)
			},
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("not-enrolled-job"))
			},
			setupMocks: func() {
				payload := []byte(
					`{"id":"not-enrolled-job","operation":{"type":"network.ping.do","data":{"address":"1.1.1.1"}}}`,
				)
				sig := ed25519.Sign(controllerPriv, payload)
				envelope := job.SignedEnvelope{
					Payload:     payload,
					Signature:   sig,
					Fingerprint: "SHA256:ctrl",
				}
				envelopeJSON, _ := json.Marshal(envelope)

				s.expectNotAnswered("not-enrolled-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.not-enrolled-job").
					Return(envelopeJSON, nil)
				// No WriteStatusEvent/WriteJobResponse expectations: an
				// unenrolled agent must refuse before touching the
				// operation, distinctly from a bad-signature rejection.
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "no cached controller public key")
			},
		},
		{
			name: "when job data is not a signed envelope rejects distinctly",
			setupPKI: func() {
				m := pki.New(memfs.New(), "/tmp/pki", "agent")
				s.Require().NoError(m.LoadOrGenerate())
				m.SetControllerPublicKey(controllerPub)
				agent.SetAgentPKIManager(s.testAgent, m)
			},
			cleanupPKI: func() {
				agent.SetAgentPKIManager(s.testAgent, nil)
			},
			setupMsg: func(ctrl *gomock.Controller) jetstream.Msg {
				return newTestMsg(ctrl, "jobs.query.host.test-agent", []byte("unwrapped-job"))
			},
			setupMocks: func() {
				// Plain, unsigned job data: PKI is enabled, so this is
				// rejected for missing an envelope rather than parsed.
				s.expectNotAnswered("unwrapped-job")
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.unwrapped-job").
					Return([]byte(`{"id":"unwrapped-job","operation":{"type":"network.ping.do","data":{"address":"1.1.1.1"}}}`), nil)

				// No WriteStatusEvent/WriteJobResponse expectations: an
				// unsigned envelope must be refused before touching the
				// operation.
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "job data is not a signed envelope")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupPKI()
			defer tt.cleanupPKI()
			tt.setupMocks()

			msg := tt.setupMsg(s.mockCtrl)
			tt.validateFunc(agent.ExportHandleJobMessage(context.Background(), s.testAgent, msg))
		})
	}
}

// TestKeepAliveInterval is the keepalive of a job: it pings before the
// server delivers the job again, which with a back_off is back_off[0], not
// ack_wait.
func (s *HandlerPublicTestSuite) TestKeepAliveInterval() {
	tests := []struct {
		name     string
		consumer config.AgentConsumer
		want     time.Duration
	}{
		{
			name:     "when there is no back_off, half of ack_wait",
			consumer: config.AgentConsumer{AckWait: "2m"},
			want:     time.Minute,
		},
		{
			name: "when back_off is shorter than ack_wait, half of its shortest entry",
			consumer: config.AgentConsumer{
				AckWait: "2m",
				BackOff: []string{"30s", "2m", "5m"},
			},
			want: 15 * time.Second,
		},
		{
			name:     "when back_off is longer than ack_wait, half of ack_wait",
			consumer: config.AgentConsumer{AckWait: "20s", BackOff: []string{"30s"}},
			want:     10 * time.Second,
		},
		{
			name:     "when an entry does not parse, it is not the wait",
			consumer: config.AgentConsumer{AckWait: "2m", BackOff: []string{"soon", "0s", "40s"}},
			want:     20 * time.Second,
		},
		{
			name:     "when ack_wait does not parse, half of the default",
			consumer: config.AgentConsumer{AckWait: "later"},
			want:     time.Minute,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, agent.ExportKeepAliveInterval(tt.consumer))
		})
	}
}

func (s *HandlerPublicTestSuite) TestStartInProgressKeepAlive() {
	tests := []struct {
		name         string
		inProgressFn func() error
		cancelCtx    bool
		validateFunc func(calls *atomic.Int32)
	}{
		{
			name:         "when the operation outlives the interval it extends the ack deadline",
			inProgressFn: func() error { return nil },
			validateFunc: func(calls *atomic.Int32) {
				s.GreaterOrEqual(calls.Load(), int32(1))
			},
		},
		{
			name:         "when stopped no further InProgress calls are made",
			inProgressFn: func() error { return nil },
			validateFunc: func(calls *atomic.Int32) {
				afterStop := calls.Load()
				time.Sleep(30 * time.Millisecond)
				// stop() has already returned by the time validateFunc
				// runs, which blocks until the keepalive goroutine exits —
				// so no call arrives after this point and no goroutine
				// leaks past the test.
				s.Equal(afterStop, calls.Load())
			},
		},
		{
			name:         "when InProgress fails the keepalive logs and keeps ticking",
			inProgressFn: func() error { return errors.New("ack deadline extend failed") },
			validateFunc: func(calls *atomic.Int32) {
				s.GreaterOrEqual(calls.Load(), int32(1))
			},
		},
		{
			name:         "when the context is already canceled the keepalive exits immediately",
			cancelCtx:    true,
			inProgressFn: func() error { return nil },
			validateFunc: func(_ *atomic.Int32) {
				// No assertion beyond reaching here without hanging: stop()
				// returning proves the goroutine exited via ctx.Done()
				// rather than waiting for a tick.
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			agent.SetInProgressInterval(5 * time.Millisecond)
			defer agent.ResetInProgressInterval()

			var calls atomic.Int32
			mockMsg := agentmocks.NewMockMsg(s.mockCtrl)
			mockMsg.EXPECT().InProgress().DoAndReturn(func() error {
				calls.Add(1)
				return tt.inProgressFn()
			}).AnyTimes()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancelCtx {
				cancel()
			}

			stop := agent.ExportStartInProgressKeepAlive(ctx, s.testAgent, mockMsg)

			// Always let the keepalive goroutine reach its select before
			// stop() closes the stop channel. Without this, a canceled
			// context and a closed stop channel are both ready and select
			// picks at random, so the ctx.Done() path is only sometimes

			// taken -- which made its coverage flap between runs.
			time.Sleep(30 * time.Millisecond)

			stop()

			tt.validateFunc(&calls)
		})
	}
}

func TestHandlerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HandlerPublicTestSuite))
}

func (s *HandlerPublicTestSuite) TestClassifyProviderError() {
	tests := []struct {
		name string
		err  error
		want job.ErrorCode
	}{
		{
			name: "a thing the host does not have",
			err:  fmt.Errorf(`user "nobody": %w`, provider.ErrNotFound),
			want: job.ErrorCodeNotFound,
		},
		{
			name: "a thing osapi did not deploy",
			err:  fmt.Errorf("cron entry %q: %w", "backup", provider.ErrNotManaged),
			want: job.ErrorCodeNotManaged,
		},
		{
			name: "a program the host does not have",
			err:  fmt.Errorf("package %q: %w", "nginx", provider.ErrNotInstalled),
			want: job.ErrorCodeNotInstalled,
		},
		{
			name: "an input the host refuses",
			err:  fmt.Errorf("/tmp: a directory, not a file: %w", provider.ErrInvalid),
			want: job.ErrorCodeInvalid,
		},
		{
			name: "an operation this OS family does not have",
			err:  fmt.Errorf("timezone: %w", provider.ErrUnsupported),
			want: job.ErrorCodeUnsupported,
		},
		{
			name: "a failure with no cause the API distinguishes",
			err:  errors.New("apt-get returned 100"),
			want: "",
		},
		{
			name: "a cause wrapped two errors deep",
			err:  fmt.Errorf("get user: %w", fmt.Errorf("read passwd: %w", provider.ErrNotFound)),
			want: job.ErrorCodeNotFound,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, agent.ExportClassifyProviderError(tt.err))
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.

// TestOpenJob: a running agent opens a job sealed to its keys, also after
// a rotation moved the keys on disk.
func (s *HandlerPublicTestSuite) TestOpenJob() {
	// sealedTo is a job sealed to the public encryption key pub.
	sealedTo := func(pub string) job.Request {
		sealer, err := seal.New("job-1")
		s.Require().NoError(err)
		key, err := sealer.KeyFor(pub)
		s.Require().NoError(err)

		return job.Request{
			JobID:      "job-1",
			SealedKeys: map[string]string{"m-1": key},
			SealedFrom: sealer.From(),
		}
	}
	// rotate moves new keys in place on disk, as osapi pki key rotate
	// does, and returns their public encryption key.
	rotate := func(fs avfs.VFS) string {
		disk := pki.New(fs, keyDir, "agent", pki.WithEncryptionKey())
		s.Require().NoError(disk.Load())
		next, err := disk.Next()
		s.Require().NoError(err)
		s.Require().NoError(disk.Promote(next))

		return next.EncryptionPublicKey()
	}

	tests := []struct {
		name string
		// arrange changes the disk after the agent loaded its keys, and
		// returns the job to open.
		arrange      func(fs avfs.VFS, m *pki.Manager) job.Request
		validateFunc func(err error, before string, m *pki.Manager)
	}{
		{
			name: "a job sealed to the new key opens: the agent reads the new keys",
			arrange: func(fs avfs.VFS, _ *pki.Manager) job.Request {
				return sealedTo(rotate(fs))
			},
			validateFunc: func(err error, before string, m *pki.Manager) {
				s.Require().NoError(err)
				s.NotEqual(before, m.Fingerprint())
			},
		},
		{
			name: "a job sealed to the old key just before the rotation opens",
			arrange: func(fs avfs.VFS, m *pki.Manager) job.Request {
				req := sealedTo(m.EncryptionPublicKey())
				rotate(fs)

				return req
			},
			validateFunc: func(err error, _ string, _ *pki.Manager) {
				s.Require().NoError(err)
			},
		},
		{
			name: "a job sealed to a key that is neither does not open",
			arrange: func(fs avfs.VFS, _ *pki.Manager) job.Request {
				rotate(fs)
				other := pki.New(fs, "/other", "agent", pki.WithEncryptionKey())
				s.Require().NoError(other.LoadOrGenerate())

				return sealedTo(other.EncryptionPublicKey())
			},
			validateFunc: func(err error, _ string, _ *pki.Manager) {
				s.Error(err)
			},
		},
		{
			name: "keys that are gone from the disk: the agent keeps the keys it has",
			arrange: func(fs avfs.VFS, m *pki.Manager) job.Request {
				req := sealedTo(m.EncryptionPublicKey())
				s.Require().NoError(fs.Remove(keyDir + "/agent.pub"))

				return req
			},
			validateFunc: func(err error, _ string, _ *pki.Manager) {
				s.Require().NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			fs := memfs.New()
			a := newTestAgent(newTestAgentParams{
				appFs:           fs,
				jobClient:       s.mockJobClient,
				streamName:      "test-stream",
				hostProvider:    hostMocks.NewDefaultMockProvider(s.mockCtrl),
				diskProvider:    diskMocks.NewDefaultMockProvider(s.mockCtrl),
				memoryProvider:  memoryMocks.NewDefaultMockProvider(s.mockCtrl),
				loadProvider:    loadMocks.NewDefaultMockProvider(s.mockCtrl),
				processProvider: processMocks.NewDefaultMockProvider(s.mockCtrl),
			})
			agent.SetAgentMachineID(a, "m-1")
			m := pki.New(fs, keyDir, "agent", pki.WithEncryptionKey())
			s.Require().NoError(m.LoadOrGenerate())
			agent.SetAgentPKIManager(a, m)
			before := m.Fingerprint()

			req := tt.arrange(fs, m)
			_, err := agent.ExportOpenJob(a, req)
			tt.validateFunc(err, before, m)
		})
	}
}
