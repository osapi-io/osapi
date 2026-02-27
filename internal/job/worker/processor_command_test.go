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

package worker

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
	"github.com/retr0h/osapi/internal/provider/command"
	commandMocks "github.com/retr0h/osapi/internal/provider/command/mocks"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/dns/mocks"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/node/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/node/mem/mocks"
)

type ProcessorCommandTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
}

func (s *ProcessorCommandTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
}

func (s *ProcessorCommandTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ProcessorCommandTestSuite) newWorkerWithCommandMock(
	cmdMock command.Provider,
) *Worker {
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
		cmdMock,
		nil,
	)
}

func (s *ProcessorCommandTestSuite) TestProcessCommandOperation() {
	tests := []struct {
		name        string
		jobRequest  job.Request
		setupMock   func(*commandMocks.MockProvider)
		expectError bool
		errorMsg    string
		validate    func(json.RawMessage)
	}{
		{
			name: "successful exec operation",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "command",
				Operation: "exec.execute",
				Data: json.RawMessage(
					`{"command":"ls","args":["-la"],"cwd":"/tmp","timeout":30}`,
				),
			},
			setupMock: func(m *commandMocks.MockProvider) {
				m.EXPECT().
					Exec(command.ExecParams{
						Command: "ls",
						Args:    []string{"-la"},
						Cwd:     "/tmp",
						Timeout: 30,
					}).
					Return(&command.Result{
						Stdout:     "total 0\n",
						Stderr:     "",
						ExitCode:   0,
						DurationMs: 12,
					}, nil)
			},
			validate: func(result json.RawMessage) {
				var r command.Result
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Equal("total 0\n", r.Stdout)
				s.Equal(0, r.ExitCode)
			},
		},
		{
			name: "successful shell operation",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "command",
				Operation: "shell.execute",
				Data: json.RawMessage(
					`{"command":"echo hello | tr a-z A-Z","cwd":"/tmp","timeout":30}`,
				),
			},
			setupMock: func(m *commandMocks.MockProvider) {
				m.EXPECT().
					Shell(command.ShellParams{
						Command: "echo hello | tr a-z A-Z",
						Cwd:     "/tmp",
						Timeout: 30,
					}).
					Return(&command.Result{
						Stdout:     "HELLO\n",
						Stderr:     "",
						ExitCode:   0,
						DurationMs: 5,
					}, nil)
			},
			validate: func(result json.RawMessage) {
				var r command.Result
				err := json.Unmarshal(result, &r)
				s.NoError(err)
				s.Equal("HELLO\n", r.Stdout)
			},
		},
		{
			name: "unsupported command operation",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "command",
				Operation: "unknown.execute",
				Data:      json.RawMessage(`{}`),
			},
			setupMock:   func(_ *commandMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "unsupported command operation",
		},
		{
			name: "exec with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "command",
				Operation: "exec.execute",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *commandMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "failed to parse command exec data",
		},
		{
			name: "shell with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "command",
				Operation: "shell.execute",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock:   func(_ *commandMocks.MockProvider) {},
			expectError: true,
			errorMsg:    "failed to parse command shell data",
		},
		{
			name: "exec provider error",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "command",
				Operation: "exec.execute",
				Data:      json.RawMessage(`{"command":"fail"}`),
			},
			setupMock: func(m *commandMocks.MockProvider) {
				m.EXPECT().
					Exec(gomock.Any()).
					Return(nil, errors.New("execution failed"))
			},
			expectError: true,
			errorMsg:    "command exec failed",
		},
		{
			name: "shell provider error",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "command",
				Operation: "shell.execute",
				Data:      json.RawMessage(`{"command":"fail"}`),
			},
			setupMock: func(m *commandMocks.MockProvider) {
				m.EXPECT().
					Shell(gomock.Any()).
					Return(nil, errors.New("shell failed"))
			},
			expectError: true,
			errorMsg:    "command shell failed",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cmdMock := commandMocks.NewMockProvider(s.mockCtrl)
			tt.setupMock(cmdMock)

			w := s.newWorkerWithCommandMock(cmdMock)
			result, err := w.processCommandOperation(tt.jobRequest)

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

func (s *ProcessorCommandTestSuite) TestGetCommandProvider() {
	tests := []struct {
		name string
	}{
		{
			name: "returns command provider",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cmdMock := commandMocks.NewPlainMockProvider(s.mockCtrl)
			w := s.newWorkerWithCommandMock(cmdMock)

			provider := w.getCommandProvider()

			s.NotNil(provider)
		})
	}
}

func TestProcessorCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorCommandTestSuite))
}
