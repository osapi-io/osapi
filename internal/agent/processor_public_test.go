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
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/mocks"
	commandMocks "github.com/osapi-io/osapi/internal/provider/command/mocks"
	fileMocks "github.com/osapi-io/osapi/internal/provider/file/mocks"
	"github.com/osapi-io/osapi/internal/provider/network/ping"
	pingMocks "github.com/osapi-io/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/osapi-io/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/osapi-io/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/osapi-io/osapi/internal/provider/node/load/mocks"
	memoryMocks "github.com/osapi-io/osapi/internal/provider/node/memory/mocks"
	processMocks "github.com/osapi-io/osapi/internal/telemetry/process/mocks"
)

type ProcessorPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	testAgent     *agent.Agent
}

func (s *ProcessorPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)

	appFs := memfs.New()
	appConfig := config.Config{
		NATS: config.NATS{
			Stream: config.NATSStream{Name: "test-stream"},
		},
		Agent: config.AgentConfig{
			Hostname:   "test-agent",
			QueueGroup: "test-queue",
			MaxJobs:    5,
		},
	}

	// Create mock providers
	hostMock := hostMocks.NewDefaultMockProvider(s.mockCtrl)
	diskMock := diskMocks.NewDefaultMockProvider(s.mockCtrl)
	memMock := memoryMocks.NewDefaultMockProvider(s.mockCtrl)
	loadMock := loadMocks.NewDefaultMockProvider(s.mockCtrl)

	// Use plain ping mock to avoid hardcoded address expectations
	pingMock := pingMocks.NewPlainMockProvider(s.mockCtrl)
	// Set up expectations for addresses used in tests
	pingMock.EXPECT().Do(gomock.Any(), "127.0.0.1").Return(&ping.Result{
		PacketsSent:     3,
		PacketsReceived: 3,
		PacketLoss:      0,
	}, nil).AnyTimes()
	pingMock.EXPECT().Do(gomock.Any(), "8.8.8.8").Return(&ping.Result{
		PacketsSent:     3,
		PacketsReceived: 3,
		PacketLoss:      0,
	}, nil).AnyTimes()

	commandMock := commandMocks.NewDefaultMockProvider(s.mockCtrl)
	fMock := fileMocks.NewDefaultMockProvider(s.mockCtrl)

	s.testAgent = newTestAgent(newTestAgentParams{
		appFs:           appFs,
		appConfig:       appConfig,
		logger:          slog.Default(),
		jobClient:       s.mockJobClient,
		streamName:      "test-stream",
		hostProvider:    hostMock,
		diskProvider:    diskMock,
		memoryProvider:  memMock,
		loadProvider:    loadMock,
		pingProvider:    pingMock,
		commandProvider: commandMock,
		fileProvider:    fMock,
		processProvider: processMocks.NewDefaultMockProvider(s.mockCtrl),
	})
}

func (s *ProcessorPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ProcessorPublicTestSuite) TestProcessJobOperation() {
	tests := []struct {
		name         string
		jobRequest   job.Request
		validateFunc func(json.RawMessage, error)
	}{
		{
			name: "successful network ping operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "network",
				Operation: "ping.do",
				Data:      json.RawMessage(`{"address": "8.8.8.8"}`),
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.NoError(err)
				s.NotNil(result)

				var response map[string]interface{}
				decodeErr := json.Unmarshal(result, &response)
				s.NoError(decodeErr)
				s.Equal(false, response["changed"])
			},
		},
		{
			name: "successful command exec operation",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "command",
				Operation: "exec.execute",
				Data:      json.RawMessage(`{"command":"ls","args":["-la"]}`),
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.NoError(err)
				s.NotNil(result)

				var response map[string]interface{}
				decodeErr := json.Unmarshal(result, &response)
				s.NoError(decodeErr)
				s.Contains(response, "stdout")
			},
		},
		{
			name: "successful command shell operation",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "command",
				Operation: "shell.execute",
				Data:      json.RawMessage(`{"command":"echo hello"}`),
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.NoError(err)
				s.NotNil(result)

				var response map[string]interface{}
				decodeErr := json.Unmarshal(result, &response)
				s.NoError(decodeErr)
				s.Contains(response, "stdout")
			},
		},
		{
			name: "successful file deploy operation",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "file",
				Operation: "deploy.execute",
				Data: json.RawMessage(
					`{"object_name":"app.conf","path":"/etc/mock/file.conf","content_type":"raw"}`,
				),
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.NoError(err)
				s.NotNil(result)

				var response map[string]interface{}
				decodeErr := json.Unmarshal(result, &response)
				s.NoError(decodeErr)
				s.Equal(true, response["changed"])
			},
		},
		{
			name: "successful file status operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "file",
				Operation: "status.get",
				Data:      json.RawMessage(`{"path":"/etc/mock/file.conf"}`),
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.NoError(err)
				s.NotNil(result)

				var response map[string]interface{}
				decodeErr := json.Unmarshal(result, &response)
				s.NoError(decodeErr)
				s.Equal("in-sync", response["status"])
			},
		},
		{
			name: "docker category routes to docker processor",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "docker",
				Operation: "create.execute",
				Data:      json.RawMessage(`{"image":"nginx:latest"}`),
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "docker runtime not available")
				s.Nil(result)
			},
		},
		{
			name: "unsupported job category",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "unsupported",
				Operation: "test.get",
				Data:      json.RawMessage(`{}`),
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "unsupported job category: unsupported")
				s.Nil(result)
			},
		},
		{
			name: "unsupported node operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "node",
				Operation: "unsupported.get",
				Data:      json.RawMessage(`{}`),
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "unsupported node operation: unsupported.get")
				s.Nil(result)
			},
		},
		{
			name: "unsupported network operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "network",
				Operation: "unsupported.get",
				Data:      json.RawMessage(`{}`),
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "unsupported network operation: unsupported.get")
				s.Nil(result)
			},
		},
		{
			name: "network ping missing address",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "network",
				Operation: "ping.do",
				Data:      json.RawMessage(`{}`),
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "missing ping address")
				s.Nil(result)
			},
		},
		{
			name: "network ping invalid data format",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "network",
				Operation: "ping.do",
				Data:      json.RawMessage(`invalid json`),
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "failed to parse ping data")
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(agent.ExportProcessJobOperation(s.testAgent, tt.jobRequest))
		})
	}
}

func (s *ProcessorPublicTestSuite) TestNetworkOperations() {
	tests := []struct {
		name      string
		operation string
		// jobType defaults to a query.
		jobType job.Type
		data    string
		// createAgent builds an agent with the providers a row needs;
		// nil uses the suite's agent.
		createAgent  func() *agent.Agent
		validateFunc func(json.RawMessage, error)
	}{
		{
			name:      "ping with valid address",
			operation: "ping.do",
			data:      `{"address": "127.0.0.1"}`,
			validateFunc: func(result json.RawMessage, err error) {
				s.NoError(err)
				s.NotNil(result)

				var response map[string]interface{}
				decodeErr := json.Unmarshal(result, &response)
				s.NoError(decodeErr)
				s.Equal(false, response["changed"])
			},
		},
		{
			name:      "ping without address",
			operation: "ping.execute",
			data:      `{}`,
			validateFunc: func(_ json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "missing ping address")
			},
		},
		{
			name:      "unsupported network operation",
			operation: "unknown.get",
			data:      `{}`,
			validateFunc: func(_ json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "unsupported network operation")
			},
		},
		{
			name:      "ping provider error",
			operation: "ping.do",
			jobType:   job.TypeQuery,
			data:      `{"address": "8.8.8.8"}`,
			createAgent: func() *agent.Agent {
				pingMock := pingMocks.NewPlainMockProvider(s.mockCtrl)
				pingMock.EXPECT().
					Do(gomock.Any(), "8.8.8.8").
					Return(nil, errors.New("ping timeout"))
				return newTestAgent(newTestAgentParams{
					appFs:           memfs.New(),
					appConfig:       config.Config{},
					jobClient:       s.mockJobClient,
					hostProvider:    hostMocks.NewPlainMockProvider(s.mockCtrl),
					diskProvider:    diskMocks.NewPlainMockProvider(s.mockCtrl),
					memoryProvider:  memoryMocks.NewPlainMockProvider(s.mockCtrl),
					loadProvider:    loadMocks.NewPlainMockProvider(s.mockCtrl),
					pingProvider:    pingMock,
					commandProvider: commandMocks.NewPlainMockProvider(s.mockCtrl),
				})
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "ping failed")
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := s.testAgent
			if tt.createAgent != nil {
				a = tt.createAgent()
			}

			jobType := tt.jobType
			if jobType == "" {
				jobType = job.TypeQuery
			}

			request := job.Request{
				Type:      jobType,
				Category:  "network",
				Operation: tt.operation,
				Data:      json.RawMessage(tt.data),
			}

			tt.validateFunc(agent.ExportProcessNetworkOperation(a, request))
		})
	}
}

func (s *ProcessorPublicTestSuite) TestProviderFactoryMethods() {
	tests := []struct {
		name         string
		getProvider  func() interface{}
		validateFunc func(interface{})
	}{
		{
			name:        "getHostProvider",
			getProvider: func() interface{} { return agent.ExportGetHostProvider(s.testAgent) },
			validateFunc: func(provider interface{}) {
				s.NotNil(provider)
			},
		},
		{
			name:        "getDiskProvider",
			getProvider: func() interface{} { return agent.ExportGetDiskProvider(s.testAgent) },
			validateFunc: func(provider interface{}) {
				s.NotNil(provider)
			},
		},
		{
			name:        "getMemProvider",
			getProvider: func() interface{} { return agent.ExportGetMemProvider(s.testAgent) },
			validateFunc: func(provider interface{}) {
				s.NotNil(provider)
			},
		},
		{
			name:        "getLoadProvider",
			getProvider: func() interface{} { return agent.ExportGetLoadProvider(s.testAgent) },
			validateFunc: func(provider interface{}) {
				s.NotNil(provider)
			},
		},
		{
			name:        "getPingProvider",
			getProvider: func() interface{} { return agent.ExportGetPingProvider(s.testAgent) },
			validateFunc: func(provider interface{}) {
				s.NotNil(provider)
			},
		},
		{
			name:        "getCommandProvider",
			getProvider: func() interface{} { return agent.ExportGetCommandProvider(s.testAgent) },
			validateFunc: func(provider interface{}) {
				s.NotNil(provider)
			},
		},
		{
			name:        "getFileProvider",
			getProvider: func() interface{} { return agent.ExportGetFileProvider(s.testAgent) },
			validateFunc: func(provider interface{}) {
				s.NotNil(provider)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(tt.getProvider())
		})
	}
}

func TestProcessorPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProcessorPublicTestSuite))
}
