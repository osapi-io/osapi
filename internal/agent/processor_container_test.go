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
	"time"

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
			name: "successful container create with ports and volumes",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerCreate,
				Data: json.RawMessage(
					`{"image":"nginx:latest","name":"web","ports":[{"host":8080,"container":80}],"volumes":[{"host":"/data","container":"/var/data"}]}`,
				),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Create(gomock.Any(), runtime.CreateParams{
						Image: "nginx:latest",
						Name:  "web",
						Ports: []runtime.PortMapping{
							{Host: 8080, Container: 80},
						},
						Volumes: []runtime.VolumeMapping{
							{Host: "/data", Container: "/var/data"},
						},
					}).
					Return(&runtime.Container{
						ID:    "def456",
						Name:  "web",
						Image: "nginx:latest",
						State: "created",
					}, nil)
			},
			validate: func(result json.RawMessage) {
				var r runtime.Container
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Equal("def456", r.ID)
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
		// --- start operation ---
		{
			name: "successful container start",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerStart,
				Data:      json.RawMessage(`{"id":"abc123"}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Start(gomock.Any(), "abc123").
					Return(nil)
			},
			validate: func(result json.RawMessage) {
				var r map[string]interface{}
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Contains(r, "message")
			},
		},
		{
			name: "start with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerStart,
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *containerMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "unmarshal start data",
		},
		{
			name: "provider error on start",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerStart,
				Data:      json.RawMessage(`{"id":"abc123"}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Start(gomock.Any(), "abc123").
					Return(errors.New("start failed"))
			},
			expectError: true,
			errorMsg:    "start failed",
		},
		// --- stop operation ---
		{
			name: "successful container stop with timeout",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerStop,
				Data:      json.RawMessage(`{"id":"abc123","timeout":10}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Stop(gomock.Any(), "abc123", gomock.Any()).
					Return(nil)
			},
			validate: func(result json.RawMessage) {
				var r map[string]interface{}
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Contains(r, "message")
			},
		},
		{
			name: "successful container stop without timeout",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerStop,
				Data:      json.RawMessage(`{"id":"abc123"}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Stop(gomock.Any(), "abc123", (*time.Duration)(nil)).
					Return(nil)
			},
			validate: func(result json.RawMessage) {
				var r map[string]interface{}
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Contains(r, "message")
			},
		},
		{
			name: "stop with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerStop,
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *containerMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "unmarshal stop data",
		},
		{
			name: "provider error on stop",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerStop,
				Data:      json.RawMessage(`{"id":"abc123","timeout":10}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Stop(gomock.Any(), "abc123", gomock.Any()).
					Return(errors.New("stop failed"))
			},
			expectError: true,
			errorMsg:    "stop failed",
		},
		// --- remove operation ---
		{
			name: "successful container remove",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerRemove,
				Data:      json.RawMessage(`{"id":"abc123","force":true}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Remove(gomock.Any(), "abc123", true).
					Return(nil)
			},
			validate: func(result json.RawMessage) {
				var r map[string]interface{}
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Contains(r, "message")
			},
		},
		{
			name: "remove with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerRemove,
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *containerMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "unmarshal remove data",
		},
		{
			name: "provider error on remove",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerRemove,
				Data:      json.RawMessage(`{"id":"abc123","force":false}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Remove(gomock.Any(), "abc123", false).
					Return(errors.New("remove failed"))
			},
			expectError: true,
			errorMsg:    "remove failed",
		},
		// --- list operation ---
		{
			name: "successful container list",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerList,
				Data:      json.RawMessage(`{"state":"running","limit":10}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					List(gomock.Any(), runtime.ListParams{
						State: "running",
						Limit: 10,
					}).
					Return([]runtime.Container{
						{ID: "abc123"},
					}, nil)
			},
			validate: func(result json.RawMessage) {
				var r []runtime.Container
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Len(r, 1)
				s.Equal("abc123", r[0].ID)
			},
		},
		{
			name: "list with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerList,
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *containerMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "unmarshal list data",
		},
		{
			name: "provider error on list",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerList,
				Data:      json.RawMessage(`{"state":"all"}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("list failed"))
			},
			expectError: true,
			errorMsg:    "list failed",
		},
		// --- inspect operation ---
		{
			name: "successful container inspect",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerInspect,
				Data:      json.RawMessage(`{"id":"abc123"}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Inspect(gomock.Any(), "abc123").
					Return(&runtime.ContainerDetail{
						Container: runtime.Container{ID: "abc123"},
					}, nil)
			},
			validate: func(result json.RawMessage) {
				var r runtime.ContainerDetail
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Equal("abc123", r.ID)
			},
		},
		{
			name: "inspect with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerInspect,
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *containerMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "unmarshal inspect data",
		},
		{
			name: "provider error on inspect",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerInspect,
				Data:      json.RawMessage(`{"id":"abc123"}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Inspect(gomock.Any(), "abc123").
					Return(nil, errors.New("inspect failed"))
			},
			expectError: true,
			errorMsg:    "inspect failed",
		},
		// --- exec operation ---
		{
			name: "successful container exec",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerExec,
				Data:      json.RawMessage(`{"id":"abc123","command":["ls","-la"]}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Exec(gomock.Any(), "abc123", runtime.ExecParams{
						Command: []string{"ls", "-la"},
					}).
					Return(&runtime.ExecResult{
						Stdout:   "output",
						ExitCode: 0,
					}, nil)
			},
			validate: func(result json.RawMessage) {
				var r runtime.ExecResult
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Equal("output", r.Stdout)
				s.Equal(0, r.ExitCode)
			},
		},
		{
			name: "exec with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerExec,
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *containerMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "unmarshal exec data",
		},
		{
			name: "provider error on exec",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerExec,
				Data:      json.RawMessage(`{"id":"abc123","command":["ls"]}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Exec(gomock.Any(), "abc123", gomock.Any()).
					Return(nil, errors.New("exec failed"))
			},
			expectError: true,
			errorMsg:    "exec failed",
		},
		// --- pull operation ---
		{
			name: "successful container pull",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerPull,
				Data:      json.RawMessage(`{"image":"nginx:latest"}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Pull(gomock.Any(), "nginx:latest").
					Return(&runtime.PullResult{
						ImageID: "sha256:abc",
					}, nil)
			},
			validate: func(result json.RawMessage) {
				var r runtime.PullResult
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Equal("sha256:abc", r.ImageID)
			},
		},
		{
			name: "pull with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerPull,
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *containerMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "unmarshal pull data",
		},
		{
			name: "provider error on pull",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "container",
				Operation: job.OperationContainerPull,
				Data:      json.RawMessage(`{"image":"nginx:latest"}`),
			},
			setupMock: func(m *containerMocks.MockProvider) {
				m.EXPECT().
					Pull(gomock.Any(), "nginx:latest").
					Return(nil, errors.New("pull failed"))
			},
			expectError: true,
			errorMsg:    "pull failed",
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
