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

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job/mocks"
	commandMocks "github.com/retr0h/osapi/internal/provider/command/mocks"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/dns/mocks"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/node/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/node/mem/mocks"
)

type HandlerTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	agent         *Agent
}

func (s *HandlerTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)

	appFs := afero.NewMemMapFs()
	appConfig := config.Config{
		NATS: config.NATS{
			Stream: config.NATSStream{
				Name: "test-stream",
			},
		},
		Node: config.Node{
			Agent: config.NodeAgent{
				Hostname:   "test-worker",
				QueueGroup: "test-queue",
				MaxJobs:    5,
			},
		},
	}

	// Create mock providers using the same pattern as processor tests
	hostMock := hostMocks.NewDefaultMockProvider(s.mockCtrl)
	diskMock := diskMocks.NewDefaultMockProvider(s.mockCtrl)
	memMock := memMocks.NewDefaultMockProvider(s.mockCtrl)
	loadMock := loadMocks.NewDefaultMockProvider(s.mockCtrl)

	// Use plain DNS mock with appropriate expectations
	dnsMock := dnsMocks.NewPlainMockProvider(s.mockCtrl)
	dnsMock.EXPECT().GetResolvConfByInterface(gomock.Any()).Return(&dns.Config{
		DNSServers:    []string{"192.168.1.1", "8.8.8.8"},
		SearchDomains: []string{"example.com"},
	}, nil).AnyTimes()
	dnsMock.EXPECT().
		UpdateResolvConfByInterface(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&dns.Result{Changed: true}, nil).
		AnyTimes()

	// Use plain ping mock with appropriate expectations
	pingMock := pingMocks.NewPlainMockProvider(s.mockCtrl)
	pingMock.EXPECT().Do(gomock.Any()).Return(&ping.Result{
		PacketsSent:     3,
		PacketsReceived: 3,
		PacketLoss:      0,
	}, nil).AnyTimes()

	commandMock := commandMocks.NewDefaultMockProvider(s.mockCtrl)

	s.agent = New(
		appFs,
		appConfig,
		slog.Default(),
		s.mockJobClient,
		"test-stream",
		hostMock,
		diskMock,
		memMock,
		loadMock,
		dnsMock,
		pingMock,
		commandMock,
		nil,
	)
}

func (s *HandlerTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *HandlerTestSuite) TestWriteStatusEvent() {
	tests := []struct {
		name        string
		jobID       string
		event       string
		data        map[string]interface{}
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:  "successful status event write",
			jobID: "test-job-123",
			event: "started",
			data:  map[string]interface{}{"worker_version": "1.0.0", "pid": 12345},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-123", "started", gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:  "status event write with nil data",
			jobID: "test-job-456",
			event: "completed",
			data:  nil,
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-456", "completed", gomock.Any(), nil).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:  "status event write failure",
			jobID: "test-job-789",
			event: "failed",
			data:  map[string]interface{}{"error": "processing failed"},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-789", "failed", gomock.Any(), gomock.Any()).
					Return(errors.New("KV storage failed"))
			},
			expectError: true,
			errorMsg:    "KV storage failed",
		},
		{
			name:  "empty job ID",
			jobID: "",
			event: "started",
			data:  map[string]interface{}{},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "", "started", gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			err := s.agent.writeStatusEvent(context.Background(), tt.jobID, tt.event, tt.data)

			if tt.expectError {
				s.Error(err)
				if tt.errorMsg != "" {
					s.Contains(err.Error(), tt.errorMsg)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *HandlerTestSuite) TestHandleJobMessage() {
	tests := []struct {
		name        string
		msg         jetstream.Msg
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful job processing",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("test-job-123"),
			},
			setupMocks: func() {
				// Mock job data retrieval
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.test-job-123").
					Return([]byte(`{
						"id": "test-job-123",
						"operation": {
							"type": "node.hostname.get",
							"data": {}
						}
					}`), nil)

				// Mock status event writes
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-123", "acknowledged", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-123", "started", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-123", "completed", gomock.Any(), gomock.Any()).
					Return(nil)

				// Mock response write
				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "test-job-123", gomock.Any(), gomock.Any(), "completed", "", gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "job processing with failure",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("test-job-456"),
			},
			setupMocks: func() {
				// Mock job data retrieval
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.test-job-456").
					Return([]byte(`{
						"id": "test-job-456",
						"operation": {
							"type": "node.unsupported.get",
							"data": {}
						}
					}`), nil)

				// Mock status event writes
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-456", "acknowledged", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-456", "started", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-456", "failed", gomock.Any(), gomock.Any()).
					Return(nil)

				// Mock response write for failed job
				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "test-job-456", gomock.Any(), gomock.Any(), "failed", gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: true,
			errorMsg:    "job processing failed",
		},
		{
			name: "invalid subject format",
			msg: &mockJetStreamMsg{
				subject: "invalid",
				data:    []byte("test-job-789"),
			},
			setupMocks: func() {
				// No mocks needed as it should fail early
			},
			expectError: true,
			errorMsg:    "failed to parse subject",
		},
		{
			name: "job not found",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("nonexistent-job"),
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.nonexistent-job").
					Return(nil, errors.New("job not found"))
			},
			expectError: true,
			errorMsg:    "job not found",
		},
		{
			name: "invalid job data format",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("invalid-job"),
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.invalid-job").
					Return([]byte(`invalid json`), nil)
			},
			expectError: true,
			errorMsg:    "failed to parse job data",
		},
		{
			name: "missing job ID",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("missing-id-job"),
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.missing-id-job").
					Return([]byte(`{
						"operation": {
							"type": "node.hostname.get",
							"data": {}
						}
					}`), nil)
			},
			expectError: true,
			errorMsg:    "invalid job format: missing id",
		},
		{
			name: "missing operation",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("missing-op-job"),
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.missing-op-job").
					Return([]byte(`{
						"id": "missing-op-job"
					}`), nil)
			},
			expectError: true,
			errorMsg:    "invalid job format: missing operation",
		},
		{
			name: "missing operation type",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("missing-type-job"),
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.missing-type-job").
					Return([]byte(`{
						"id": "missing-type-job",
						"operation": {
							"data": {}
						}
					}`), nil)
			},
			expectError: true,
			errorMsg:    "invalid operation format: missing type field",
		},
		{
			name: "invalid operation type format",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("invalid-type-job"),
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.invalid-type-job").
					Return([]byte(`{
						"id": "invalid-type-job",
						"operation": {
							"type": "invalid",
							"data": {}
						}
					}`), nil)
			},
			expectError: true,
			errorMsg:    "invalid operation type format",
		},
		{
			name: "acknowledged write error logged",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("ack-err-job"),
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.ack-err-job").
					Return([]byte(`{
						"id": "ack-err-job",
						"operation": {
							"type": "node.hostname.get",
							"data": {}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "ack-err-job", "acknowledged", gomock.Any(), gomock.Any()).
					Return(errors.New("ack write failed"))

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "ack-err-job", "started", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "ack-err-job", "completed", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "ack-err-job", gomock.Any(), gomock.Any(), "completed", "", gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "started write error logged",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("start-err-job"),
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.start-err-job").
					Return([]byte(`{
						"id": "start-err-job",
						"operation": {
							"type": "node.hostname.get",
							"data": {}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "start-err-job", "acknowledged", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "start-err-job", "started", gomock.Any(), gomock.Any()).
					Return(errors.New("started write failed"))

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "start-err-job", "completed", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "start-err-job", gomock.Any(), gomock.Any(), "completed", "", gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "completed write error logged",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("comp-err-job"),
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.comp-err-job").
					Return([]byte(`{
						"id": "comp-err-job",
						"operation": {
							"type": "node.hostname.get",
							"data": {}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "comp-err-job", "acknowledged", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "comp-err-job", "started", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "comp-err-job", "completed", gomock.Any(), gomock.Any()).
					Return(errors.New("completed write failed"))

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "comp-err-job", gomock.Any(), gomock.Any(), "completed", "", gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "failed write error logged",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("fail-err-job"),
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.fail-err-job").
					Return([]byte(`{
						"id": "fail-err-job",
						"operation": {
							"type": "node.unsupported.get",
							"data": {}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fail-err-job", "acknowledged", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fail-err-job", "started", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "fail-err-job", "failed", gomock.Any(), gomock.Any()).
					Return(errors.New("failed write failed"))

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "fail-err-job", gomock.Any(), gomock.Any(), "failed", gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: true,
			errorMsg:    "job processing failed",
		},
		{
			name: "response storage failure",
			msg: &mockJetStreamMsg{
				subject: "jobs.query.test-worker",
				data:    []byte("storage-fail-job"),
			},
			setupMocks: func() {
				// Mock successful job processing
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.storage-fail-job").
					Return([]byte(`{
						"id": "storage-fail-job",
						"operation": {
							"type": "node.hostname.get",
							"data": {}
						}
					}`), nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "storage-fail-job", "acknowledged", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "storage-fail-job", "started", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "storage-fail-job", "completed", gomock.Any(), gomock.Any()).
					Return(nil)

				// Mock response write failure
				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "storage-fail-job", gomock.Any(), gomock.Any(), "completed", "", gomock.Any()).
					Return(errors.New("storage failure"))
			},
			expectError: true,
			errorMsg:    "failed to store job response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			err := s.agent.handleJobMessage(tt.msg)

			if tt.expectError {
				s.Error(err)
				if tt.errorMsg != "" {
					s.Contains(err.Error(), tt.errorMsg)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *HandlerTestSuite) TestHandleJobMessageModifyJobs() {
	tests := []struct {
		name        string
		subject     string
		jobData     string
		setupMocks  func()
		expectError bool
	}{
		{
			name:    "modify job type identification",
			subject: "jobs.modify.test-worker",
			jobData: `{
				"id": "modify-job-123",
				"operation": {
					"type": "network.dns.update",
					"data": {
						"servers": ["8.8.8.8"],
						"search_domains": ["example.com"],
						"interface": "eth0"
					}
				}
			}`,
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.modify-job-123").
					Return([]byte(`{
						"id": "modify-job-123",
						"operation": {
							"type": "network.dns.update",
							"data": {
								"servers": ["8.8.8.8"],
								"search_domains": ["example.com"],
								"interface": "eth0"
							}
						}
					}`), nil)

				// Mock all the status events and response
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()

				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			msg := &mockJetStreamMsg{
				subject: tt.subject,
				data:    []byte("modify-job-123"),
			}

			err := s.agent.handleJobMessage(msg)

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *HandlerTestSuite) TestExtractChanged() {
	tests := []struct {
		name string
		data json.RawMessage
		want *bool
	}{
		{
			name: "when empty data returns nil",
			data: nil,
			want: nil,
		},
		{
			name: "when invalid JSON returns nil",
			data: json.RawMessage(`not json`),
			want: nil,
		},
		{
			name: "when changed key missing returns nil",
			data: json.RawMessage(`{"success":true}`),
			want: nil,
		},
		{
			name: "when changed is non-bool returns nil",
			data: json.RawMessage(`{"changed":"yes"}`),
			want: nil,
		},
		{
			name: "when changed is true returns true",
			data: json.RawMessage(`{"changed":true}`),
			want: boolPtr(true),
		},
		{
			name: "when changed is false returns false",
			data: json.RawMessage(`{"changed":false}`),
			want: boolPtr(false),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := extractChanged(tt.data)

			s.Equal(tt.want, got)
		})
	}
}

func boolPtr(b bool) *bool { return &b }

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
