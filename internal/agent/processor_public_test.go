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
	"log/slog"
	"testing"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/mocks"
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
			name: "unsupported job category",
			jobRequest: job.Request{
				Type:      types.JobTypeQuery,
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
				Type:      types.JobTypeQuery,
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
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(agent.ExportProcessJobOperation(s.testAgent, tt.jobRequest))
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
