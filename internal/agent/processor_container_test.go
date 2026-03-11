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

package agent

import (
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/mocks"
	commandMocks "github.com/retr0h/osapi/internal/provider/command/mocks"
	containerProv "github.com/retr0h/osapi/internal/provider/container"
	containerMocks "github.com/retr0h/osapi/internal/provider/container/mocks"
	"github.com/retr0h/osapi/internal/provider/container/runtime"
	fileMocks "github.com/retr0h/osapi/internal/provider/file/mocks"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/dns/mocks"
	netinfoMocks "github.com/retr0h/osapi/internal/provider/network/netinfo/mocks"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/node/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/node/mem/mocks"
)

type ProcessorContainerTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
}

func (s *ProcessorContainerTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
}

func (s *ProcessorContainerTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ProcessorContainerTestSuite) newAgentWithContainerMock(
	containerMock containerProv.Provider,
) *Agent {
	return New(
		afero.NewMemMapFs(),
		config.Config{},
		slog.Default(),
		s.mockJobClient,
		"test-stream",
		hostMocks.NewPlainMockProvider(s.mockCtrl),
		diskMocks.NewPlainMockProvider(s.mockCtrl),
		memMocks.NewPlainMockProvider(s.mockCtrl),
		loadMocks.NewPlainMockProvider(s.mockCtrl),
		dnsMocks.NewPlainMockProvider(s.mockCtrl),
		pingMocks.NewPlainMockProvider(s.mockCtrl),
		netinfoMocks.NewPlainMockProvider(s.mockCtrl),
		commandMocks.NewPlainMockProvider(s.mockCtrl),
		fileMocks.NewPlainMockProvider(s.mockCtrl),
		containerMock,
		nil,
		nil,
	)
}

func (s *ProcessorContainerTestSuite) TestProcessContainerOperation() {
	tests := []struct {
		name        string
		jobRequest  job.Request
		setupMock   func(*containerMocks.MockProvider)
		expectError bool
		errorMsg    string
		validate    func(json.RawMessage)
	}{
		{
			name: "nil provider returns error",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerCreate,
				Data:      json.RawMessage(`{"image":"nginx:latest"}`),
			},
			setupMock:   nil,
			expectError: true,
			errorMsg:    "container runtime not available",
		},
		{
			name: "successful container create",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerCreate,
				Data: json.RawMessage(
					`{"image":"nginx:latest","name":"web","auto_start":true}`,
				),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Create(gomock.Any(), runtime.CreateParams{
						Image:     "nginx:latest",
						Name:      "web",
						AutoStart: true,
					}).
					Return(&runtime.Container{
						ID:    "abc123",
						Name:  "web",
						Image: "nginx:latest",
						State: "created",
					}, nil)
			},
			validate: func(result json.RawMessage) {
				var r runtime.Container
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Equal("abc123", r.ID)
				s.Equal("web", r.Name)
			},
		},
		{
			name: "unsupported container operation",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: "container.unknown",
				Data:      json.RawMessage(`{}`),
			},
			setupMock:   func(_ *containerMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "unsupported container operation",
		},
		{
			name: "create with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerCreate,
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *containerMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "unmarshal create data",
		},
		{
			name: "provider error on create",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerCreate,
				Data:      json.RawMessage(`{"image":"nginx:latest"}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("docker error"))
			},
			expectError: true,
			errorMsg:    "docker error",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var a *Agent
			if tt.setupMock != nil {
				containerMock := containerMocks.NewMockProvider(s.mockCtrl)
				tt.setupMock(containerMock)
				a = s.newAgentWithContainerMock(containerMock)
			} else {
				// nil provider case
				a = s.newAgentWithContainerMock(nil)
			}

			result, err := a.processContainerOperation(tt.jobRequest)

			if tt.expectError {
				s.Error(err)
				s.Contains(err.Error(), tt.errorMsg)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				if tt.validate != nil {
					tt.validate(result)
				}
			}
		})
	}
}

func TestProcessorContainerTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorContainerTestSuite))
}
