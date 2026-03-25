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
	fileMocks "github.com/retr0h/osapi/internal/provider/file/mocks"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/dns/mocks"
	netinfoMocks "github.com/retr0h/osapi/internal/provider/network/netinfo/mocks"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/node/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/node/mem/mocks"
	"github.com/retr0h/osapi/internal/provider/scheduled/cron"
	cronMocks "github.com/retr0h/osapi/internal/provider/scheduled/cron/mocks"
)

type ProcessorSchedulePublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
}

func (s *ProcessorSchedulePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
}

func (s *ProcessorSchedulePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ProcessorSchedulePublicTestSuite) newAgentWithCronMock(
	cronProvider cron.Provider,
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
		fileMocks.NewPlainMockProvider(s.mockCtrl),
		nil,
		cronProvider,
		nil,
		nil,
		nil,
	)
}

func (s *ProcessorSchedulePublicTestSuite) TestProcessScheduleOperation() {
	tests := []struct {
		name        string
		jobRequest  job.Request
		setupMock   func() cron.Provider
		expectError bool
		errorMsg    string
		validate    func(json.RawMessage)
	}{
		{
			name: "nil provider returns error",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "schedule",
				Operation: "cron.list",
				Data:      json.RawMessage(`{}`),
			},
			setupMock:   nil,
			expectError: true,
			errorMsg:    "cron provider not available",
		},
		{
			name: "dispatches to cron operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "schedule",
				Operation: "cron.list",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: func() cron.Provider {
				m := cronMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().List(gomock.Any()).Return([]cron.Entry{}, nil)
				return m
			},
			validate: func(result json.RawMessage) {
				var entries []cron.Entry
				err := json.Unmarshal(result, &entries)
				s.NoError(err)
				s.Empty(entries)
			},
		},
		{
			name: "unsupported schedule operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "schedule",
				Operation: "unknown.list",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: func() cron.Provider {
				return cronMocks.NewMockProvider(s.mockCtrl)
			},
			expectError: true,
			errorMsg:    "unsupported schedule operation: unknown.list",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var a *agent.Agent
			if tt.setupMock != nil {
				a = s.newAgentWithCronMock(tt.setupMock())
			} else {
				a = s.newAgentWithCronMock(nil)
			}

			result, err := agent.ExportProcessScheduleOperation(a, tt.jobRequest)

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

func (s *ProcessorSchedulePublicTestSuite) TestProcessCronOperation() {
	tests := []struct {
		name        string
		jobRequest  job.Request
		setupMock   func() cron.Provider
		expectError bool
		errorMsg    string
		validate    func(json.RawMessage)
	}{
		{
			name: "invalid cron operation missing sub-operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "schedule",
				Operation: "cron",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: func() cron.Provider {
				return cronMocks.NewMockProvider(s.mockCtrl)
			},
			expectError: true,
			errorMsg:    "invalid cron operation: cron",
		},
		{
			name: "successful cron list",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "schedule",
				Operation: "cron.list",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: func() cron.Provider {
				m := cronMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().List(gomock.Any()).Return([]cron.Entry{
					{
						Name:     "backup",
						Schedule: "0 2 * * *",
						User:     "root",
						Object:   "/usr/local/bin/backup.sh",
					},
				}, nil)
				return m
			},
			validate: func(result json.RawMessage) {
				var entries []cron.Entry
				err := json.Unmarshal(result, &entries)
				s.NoError(err)
				s.Len(entries, 1)
				s.Equal("backup", entries[0].Name)
				s.Equal("0 2 * * *", entries[0].Schedule)
			},
		},
		{
			name: "cron list provider error",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "schedule",
				Operation: "cron.list",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: func() cron.Provider {
				m := cronMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().List(gomock.Any()).Return(nil, errors.New("permission denied"))
				return m
			},
			expectError: true,
			errorMsg:    "failed to list cron entries",
		},
		{
			name: "successful cron get",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "schedule",
				Operation: "cron.get",
				Data:      json.RawMessage(`{"name":"backup"}`),
			},
			setupMock: func() cron.Provider {
				m := cronMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().Get(gomock.Any(), "backup").Return(&cron.Entry{
					Name:     "backup",
					Schedule: "0 2 * * *",
					User:     "root",
					Object:   "/usr/local/bin/backup.sh",
				}, nil)
				return m
			},
			validate: func(result json.RawMessage) {
				var entry cron.Entry
				err := json.Unmarshal(result, &entry)
				s.NoError(err)
				s.Equal("backup", entry.Name)
			},
		},
		{
			name: "cron get with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "schedule",
				Operation: "cron.get",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock: func() cron.Provider {
				return cronMocks.NewMockProvider(s.mockCtrl)
			},
			expectError: true,
			errorMsg:    "unmarshal cron get data",
		},
		{
			name: "cron get provider error",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "schedule",
				Operation: "cron.get",
				Data:      json.RawMessage(`{"name":"missing"}`),
			},
			setupMock: func() cron.Provider {
				m := cronMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().Get(gomock.Any(), "missing").Return(nil, errors.New("not found"))
				return m
			},
			expectError: true,
			errorMsg:    "failed to get cron entry",
		},
		{
			name: "successful cron create",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "schedule",
				Operation: "cron.create",
				Data: json.RawMessage(
					`{"name":"logrotate","schedule":"0 0 * * *","user":"root","command":"/usr/sbin/logrotate"}`,
				),
			},
			setupMock: func() cron.Provider {
				m := cronMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ interface{}, entry cron.Entry) (*cron.CreateResult, error) {
						s.Equal("logrotate", entry.Name)
						return &cron.CreateResult{
							Name:    "logrotate",
							Changed: true,
						}, nil
					},
				)
				return m
			},
			validate: func(result json.RawMessage) {
				var r cron.CreateResult
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Equal("logrotate", r.Name)
				s.True(r.Changed)
			},
		},
		{
			name: "cron create with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "schedule",
				Operation: "cron.create",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock: func() cron.Provider {
				return cronMocks.NewMockProvider(s.mockCtrl)
			},
			expectError: true,
			errorMsg:    "unmarshal cron create data",
		},
		{
			name: "cron create provider error",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "schedule",
				Operation: "cron.create",
				Data: json.RawMessage(
					`{"name":"dup","schedule":"* * * * *","user":"root","command":"echo"}`,
				),
			},
			setupMock: func() cron.Provider {
				m := cronMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("already exists"))
				return m
			},
			expectError: true,
			errorMsg:    "failed to create cron entry",
		},
		{
			name: "successful cron update",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "schedule",
				Operation: "cron.update",
				Data: json.RawMessage(
					`{"name":"backup","schedule":"0 3 * * *","user":"root","command":"/usr/local/bin/backup.sh"}`,
				),
			},
			setupMock: func() cron.Provider {
				m := cronMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ interface{}, entry cron.Entry) (*cron.UpdateResult, error) {
						s.Equal("backup", entry.Name)
						s.Equal("0 3 * * *", entry.Schedule)
						return &cron.UpdateResult{
							Name:    "backup",
							Changed: true,
						}, nil
					},
				)
				return m
			},
			validate: func(result json.RawMessage) {
				var r cron.UpdateResult
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Equal("backup", r.Name)
				s.True(r.Changed)
			},
		},
		{
			name: "cron update with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "schedule",
				Operation: "cron.update",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock: func() cron.Provider {
				return cronMocks.NewMockProvider(s.mockCtrl)
			},
			expectError: true,
			errorMsg:    "unmarshal cron update data",
		},
		{
			name: "cron update provider error",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "schedule",
				Operation: "cron.update",
				Data: json.RawMessage(
					`{"name":"missing","schedule":"* * * * *","user":"root","command":"echo"}`,
				),
			},
			setupMock: func() cron.Provider {
				m := cronMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))
				return m
			},
			expectError: true,
			errorMsg:    "failed to update cron entry",
		},
		{
			name: "successful cron delete",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "schedule",
				Operation: "cron.delete",
				Data:      json.RawMessage(`{"name":"backup"}`),
			},
			setupMock: func() cron.Provider {
				m := cronMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().Delete(gomock.Any(), "backup").Return(&cron.DeleteResult{
					Name:    "backup",
					Changed: true,
				}, nil)
				return m
			},
			validate: func(result json.RawMessage) {
				var r cron.DeleteResult
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Equal("backup", r.Name)
				s.True(r.Changed)
			},
		},
		{
			name: "cron delete with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "schedule",
				Operation: "cron.delete",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock: func() cron.Provider {
				return cronMocks.NewMockProvider(s.mockCtrl)
			},
			expectError: true,
			errorMsg:    "unmarshal cron delete data",
		},
		{
			name: "cron delete provider error",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "schedule",
				Operation: "cron.delete",
				Data:      json.RawMessage(`{"name":"missing"}`),
			},
			setupMock: func() cron.Provider {
				m := cronMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().Delete(gomock.Any(), "missing").Return(nil, errors.New("not found"))
				return m
			},
			expectError: true,
			errorMsg:    "failed to delete cron entry",
		},
		{
			name: "unsupported cron sub-operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "schedule",
				Operation: "cron.unknown",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: func() cron.Provider {
				return cronMocks.NewMockProvider(s.mockCtrl)
			},
			expectError: true,
			errorMsg:    "unsupported cron operation: cron.unknown",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := s.newAgentWithCronMock(tt.setupMock())

			result, err := agent.ExportProcessCronOperation(a, tt.jobRequest)

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

func (s *ProcessorSchedulePublicTestSuite) TestGetCronProvider() {
	tests := []struct {
		name      string
		setupMock func() cron.Provider
		expectNil bool
	}{
		{
			name: "returns injected provider",
			setupMock: func() cron.Provider {
				return cronMocks.NewMockProvider(s.mockCtrl)
			},
			expectNil: false,
		},
		{
			name: "returns nil when no provider",
			setupMock: func() cron.Provider {
				return nil
			},
			expectNil: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := s.newAgentWithCronMock(tt.setupMock())

			provider := agent.ExportGetCronProvider(a)

			if tt.expectNil {
				s.Nil(provider)
			} else {
				s.NotNil(provider)
			}
		})
	}
}

func (s *ProcessorSchedulePublicTestSuite) TestProcessJobOperationScheduleCategory() {
	tests := []struct {
		name        string
		jobRequest  job.Request
		setupMock   func() cron.Provider
		expectError bool
		errorMsg    string
		validate    func(json.RawMessage)
	}{
		{
			name: "schedule category dispatches correctly",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "schedule",
				Operation: "cron.list",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: func() cron.Provider {
				m := cronMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().List(gomock.Any()).Return([]cron.Entry{}, nil)
				return m
			},
			validate: func(result json.RawMessage) {
				var entries []cron.Entry
				err := json.Unmarshal(result, &entries)
				s.NoError(err)
				s.Empty(entries)
			},
		},
		{
			name: "schedule category with nil provider",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "schedule",
				Operation: "cron.list",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: func() cron.Provider {
				return nil
			},
			expectError: true,
			errorMsg:    "cron provider not available",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := s.newAgentWithCronMock(tt.setupMock())

			result, err := agent.ExportProcessJobOperation(a, tt.jobRequest)

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

func TestProcessorSchedulePublicTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorSchedulePublicTestSuite))
}
