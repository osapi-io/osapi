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

package agent_test

import (
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/agent"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/mocks"
	commandMocks "github.com/retr0h/osapi/internal/provider/command/mocks"
	fileProv "github.com/retr0h/osapi/internal/provider/file"
	fileMocks "github.com/retr0h/osapi/internal/provider/file/mocks"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/dns/mocks"
	netinfoMocks "github.com/retr0h/osapi/internal/provider/network/netinfo/mocks"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/node/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/node/mem/mocks"
)

type ProcessorFilePublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
}

func (s *ProcessorFilePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
}

func (s *ProcessorFilePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ProcessorFilePublicTestSuite) newAgentWithFileMock(
	fMock fileProv.Provider,
) *agent.Agent {
	return agent.New(
		memfs.New(),
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
		fMock,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func (s *ProcessorFilePublicTestSuite) TestProcessFileOperation() {
	tests := []struct {
		name        string
		jobRequest  job.Request
		setupMock   func(*fileMocks.MockProvider)
		expectError bool
		errorMsg    string
		validate    func(json.RawMessage)
	}{
		{
			name: "successful deploy operation",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "file",
				Operation: "deploy.execute",
				Data: json.RawMessage(
					`{"object_name":"app.conf","path":"/etc/app/app.conf","mode":"0644","content_type":"raw"}`,
				),
			},
			setupMock: func(m *fileMocks.MockProvider) {
				m.EXPECT().
					Deploy(gomock.Any(), fileProv.DeployRequest{
						ObjectName:  "app.conf",
						Path:        "/etc/app/app.conf",
						Mode:        "0644",
						ContentType: "raw",
					}).
					Return(&fileProv.DeployResult{
						Changed: true,
						SHA256:  "abc123def456",
						Path:    "/etc/app/app.conf",
					}, nil)
			},
			validate: func(result json.RawMessage) {
				var r fileProv.DeployResult
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.True(r.Changed)
				s.Equal("abc123def456", r.SHA256)
				s.Equal("/etc/app/app.conf", r.Path)
			},
		},
		{
			name: "successful status operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "file",
				Operation: "status.get",
				Data:      json.RawMessage(`{"path":"/etc/app/app.conf"}`),
			},
			setupMock: func(m *fileMocks.MockProvider) {
				m.EXPECT().
					Status(gomock.Any(), fileProv.StatusRequest{
						Path: "/etc/app/app.conf",
					}).
					Return(&fileProv.StatusResult{
						Path:   "/etc/app/app.conf",
						Status: "in-sync",
						SHA256: "abc123def456",
					}, nil)
			},
			validate: func(result json.RawMessage) {
				var r fileProv.StatusResult
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Equal("in-sync", r.Status)
				s.Equal("/etc/app/app.conf", r.Path)
				s.Equal("abc123def456", r.SHA256)
			},
		},
		{
			name: "unsupported file operation",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "file",
				Operation: "unknown.execute",
				Data:      json.RawMessage(`{}`),
			},
			setupMock:   func(_ *fileMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "unsupported file operation",
		},
		{
			name: "deploy with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "file",
				Operation: "deploy.execute",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *fileMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "failed to parse file deploy data",
		},
		{
			name: "status with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "file",
				Operation: "status.get",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *fileMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "failed to parse file status data",
		},
		{
			name: "deploy provider error",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "file",
				Operation: "deploy.execute",
				Data: json.RawMessage(
					`{"object_name":"app.conf","path":"/etc/app/app.conf","content_type":"raw"}`,
				),
			},
			setupMock: func(m *fileMocks.MockProvider) {
				m.EXPECT().
					Deploy(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("object not found"))
			},
			expectError: true,
			errorMsg:    "file deploy failed",
		},
		{
			name: "status provider error",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "file",
				Operation: "status.get",
				Data:      json.RawMessage(`{"path":"/etc/app/app.conf"}`),
			},
			setupMock: func(m *fileMocks.MockProvider) {
				m.EXPECT().
					Status(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("state KV unavailable"))
			},
			expectError: true,
			errorMsg:    "file status failed",
		},
		{
			name: "successful undeploy operation",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "file",
				Operation: "undeploy.execute",
				Data:      json.RawMessage(`{"path":"/etc/cron.d/backup"}`),
			},
			setupMock: func(m *fileMocks.MockProvider) {
				m.EXPECT().
					Undeploy(gomock.Any(), fileProv.UndeployRequest{
						Path: "/etc/cron.d/backup",
					}).
					Return(&fileProv.UndeployResult{
						Changed: true,
						Path:    "/etc/cron.d/backup",
					}, nil)
			},
			validate: func(result json.RawMessage) {
				var r fileProv.UndeployResult
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.True(r.Changed)
				s.Equal("/etc/cron.d/backup", r.Path)
			},
		},
		{
			name: "undeploy with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "file",
				Operation: "undeploy.execute",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *fileMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "failed to parse file undeploy data",
		},
		{
			name: "undeploy provider error",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "file",
				Operation: "undeploy.execute",
				Data:      json.RawMessage(`{"path":"/etc/cron.d/backup"}`),
			},
			setupMock: func(m *fileMocks.MockProvider) {
				m.EXPECT().
					Undeploy(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("state KV unavailable"))
			},
			expectError: true,
			errorMsg:    "file undeploy failed",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			fMock := fileMocks.NewMockProvider(s.mockCtrl)
			tt.setupMock(fMock)

			a := s.newAgentWithFileMock(fMock)
			result, err := agent.ExportProcessFileOperation(a, tt.jobRequest)

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

func (s *ProcessorFilePublicTestSuite) TestProcessFileOperationNilProvider() {
	tests := []struct {
		name     string
		errorMsg string
	}{
		{
			name:     "returns error when file provider is nil",
			errorMsg: "file provider not configured",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := s.newAgentWithFileMock(nil)
			result, err := agent.ExportProcessFileOperation(a, job.Request{
				Type:      job.TypeModify,
				Category:  "file",
				Operation: "deploy.execute",
				Data:      json.RawMessage(`{}`),
			})

			s.Error(err)
			s.Contains(err.Error(), tt.errorMsg)
			s.Nil(result)
		})
	}
}

func (s *ProcessorFilePublicTestSuite) TestGetFileProvider() {
	tests := []struct {
		name string
	}{
		{
			name: "returns file provider",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			fMock := fileMocks.NewPlainMockProvider(s.mockCtrl)
			a := s.newAgentWithFileMock(fMock)

			provider := agent.ExportGetFileProvider(a)

			s.NotNil(provider)
		})
	}
}

func TestProcessorFilePublicTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorFilePublicTestSuite))
}
