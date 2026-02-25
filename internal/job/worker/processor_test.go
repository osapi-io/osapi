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

package worker

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
	"github.com/retr0h/osapi/internal/provider/network/dns"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/dns/mocks"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/system/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/system/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/system/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/system/mem/mocks"
)

type ProcessorTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	worker        *Worker
}

func (s *ProcessorTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)

	appFs := afero.NewMemMapFs()
	appConfig := config.Config{
		NATS: config.NATS{
			Stream: config.NATSStream{Name: "test-stream"},
		},
		Job: config.Job{
			Worker: config.JobWorker{
				Hostname:   "test-worker",
				QueueGroup: "test-queue",
				MaxJobs:    5,
			},
		},
	}

	// Create mock providers
	hostMock := hostMocks.NewDefaultMockProvider(s.mockCtrl)
	diskMock := diskMocks.NewDefaultMockProvider(s.mockCtrl)
	memMock := memMocks.NewDefaultMockProvider(s.mockCtrl)
	loadMock := loadMocks.NewDefaultMockProvider(s.mockCtrl)

	// Use plain DNS mock to avoid hardcoded interface expectations
	dnsMock := dnsMocks.NewPlainMockProvider(s.mockCtrl)
	// Set up expectations for eth0 interface used in tests
	dnsMock.EXPECT().GetResolvConfByInterface("eth0").Return(&dns.Config{
		DNSServers:    []string{"192.168.1.1", "8.8.8.8"},
		SearchDomains: []string{"example.com"},
	}, nil).AnyTimes()
	dnsMock.EXPECT().
		UpdateResolvConfByInterface(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&dns.Result{Changed: true}, nil).
		AnyTimes()

	// Use plain ping mock to avoid hardcoded address expectations
	pingMock := pingMocks.NewPlainMockProvider(s.mockCtrl)
	// Set up expectations for addresses used in tests
	pingMock.EXPECT().Do("127.0.0.1").Return(&ping.Result{
		PacketsSent:     3,
		PacketsReceived: 3,
		PacketLoss:      0,
	}, nil).AnyTimes()
	pingMock.EXPECT().Do("8.8.8.8").Return(&ping.Result{
		PacketsSent:     3,
		PacketsReceived: 3,
		PacketLoss:      0,
	}, nil).AnyTimes()

	commandMock := commandMocks.NewDefaultMockProvider(s.mockCtrl)

	s.worker = New(
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
	)
}

func (s *ProcessorTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ProcessorTestSuite) TestProcessJobOperation() {
	tests := []struct {
		name        string
		jobRequest  job.Request
		expectError bool
		errorMsg    string
		validate    func(json.RawMessage)
	}{
		{
			name: "successful system hostname operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "system",
				Operation: "hostname.get",
				Data:      json.RawMessage(`{}`),
			},
			expectError: false,
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
				s.Contains(response, "hostname")
				s.IsType("", response["hostname"])
			},
		},
		{
			name: "successful system status operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "system",
				Operation: "status.get",
				Data:      json.RawMessage(`{}`),
			},
			expectError: false,
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
				s.Contains(response, "hostname")
			},
		},
		{
			name: "successful system uptime operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "system",
				Operation: "uptime.get",
				Data:      json.RawMessage(`{}`),
			},
			expectError: false,
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
				s.Contains(response, "uptime_seconds")
				s.Contains(response, "uptime")
			},
		},
		{
			name: "successful system OS info operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "system",
				Operation: "osinfo.get",
				Data:      json.RawMessage(`{}`),
			},
			expectError: false,
			validate: func(result json.RawMessage) {
				// OS info should return a valid object
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
			},
		},
		{
			name: "successful system disk operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "system",
				Operation: "disk.get",
				Data:      json.RawMessage(`{}`),
			},
			expectError: false,
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
				s.Contains(response, "disks")
			},
		},
		{
			name: "successful system memory operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "system",
				Operation: "memory.get",
				Data:      json.RawMessage(`{}`),
			},
			expectError: false,
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
			},
		},
		{
			name: "successful system load operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "system",
				Operation: "load.get",
				Data:      json.RawMessage(`{}`),
			},
			expectError: false,
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
			},
		},
		{
			name: "successful network DNS query operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "network",
				Operation: "dns.get",
				Data:      json.RawMessage(`{"interface": "eth0"}`),
			},
			expectError: false,
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
				// DNS response should be a valid object
			},
		},
		{
			name: "successful network DNS modify operation",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "network",
				Operation: "dns.update",
				Data: json.RawMessage(
					`{"servers": ["8.8.8.8"], "search_domains": ["example.com"], "interface": "eth0"}`,
				),
			},
			expectError: false,
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
				s.Contains(response, "success")
				s.Contains(response, "message")
			},
		},
		{
			name: "successful network ping operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "network",
				Operation: "ping.do",
				Data:      json.RawMessage(`{"address": "8.8.8.8"}`),
			},
			expectError: false,
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
				// Ping response should be a valid object
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
			expectError: false,
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
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
			expectError: false,
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
				s.Contains(response, "stdout")
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
			expectError: true,
			errorMsg:    "unsupported job category: unsupported",
		},
		{
			name: "unsupported system operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "system",
				Operation: "unsupported.get",
				Data:      json.RawMessage(`{}`),
			},
			expectError: true,
			errorMsg:    "unsupported system operation: unsupported.get",
		},
		{
			name: "unsupported network operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "network",
				Operation: "unsupported.get",
				Data:      json.RawMessage(`{}`),
			},
			expectError: true,
			errorMsg:    "unsupported network operation: unsupported.get",
		},
		{
			name: "network ping missing address",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "network",
				Operation: "ping.do",
				Data:      json.RawMessage(`{}`),
			},
			expectError: true,
			errorMsg:    "missing ping address",
		},
		{
			name: "network ping invalid data format",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "network",
				Operation: "ping.do",
				Data:      json.RawMessage(`invalid json`),
			},
			expectError: true,
			errorMsg:    "failed to parse ping data",
		},
		{
			name: "network DNS invalid data format",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "network",
				Operation: "dns.get",
				Data:      json.RawMessage(`invalid json`),
			},
			expectError: true,
			errorMsg:    "failed to parse DNS data",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := s.worker.processJobOperation(tt.jobRequest)

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

func (s *ProcessorTestSuite) TestSystemOperations() {
	tests := []struct {
		name        string
		operation   string
		labels      map[string]string
		expectError bool
		validate    func(json.RawMessage)
	}{
		{
			name:      "get hostname",
			operation: "hostname.get",
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
				s.Contains(response, "hostname")
			},
		},
		{
			name:      "get hostname with labels",
			operation: "hostname.get",
			labels:    map[string]string{"group": "web.dev.us-east"},
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
				s.Contains(response, "hostname")
				s.Contains(response, "labels")
				labels, ok := response["labels"].(map[string]interface{})
				s.True(ok)
				s.Equal("web.dev.us-east", labels["group"])
			},
		},
		{
			name:      "get system status",
			operation: "status.get",
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
				s.Contains(response, "hostname")
			},
		},
		{
			name:      "get uptime",
			operation: "uptime.get",
			validate: func(result json.RawMessage) {
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
				s.Contains(response, "uptime_seconds")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.worker.appConfig.Job.Worker.Labels = tt.labels

			request := job.Request{
				Type:      job.TypeQuery,
				Category:  "system",
				Operation: tt.operation,
				Data:      json.RawMessage(`{}`),
			}

			result, err := s.worker.processSystemOperation(request)

			if tt.expectError {
				s.Error(err)
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

func (s *ProcessorTestSuite) TestNetworkOperations() {
	tests := []struct {
		name        string
		operation   string
		data        string
		expectError bool
		errorMsg    string
		validate    func(json.RawMessage)
	}{
		{
			name:      "DNS query with interface",
			operation: "dns.get",
			data:      `{"interface": "eth0"}`,
			validate: func(result json.RawMessage) {
				// Should return valid DNS config
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
			},
		},
		{
			name:      "DNS query without interface",
			operation: "dns.get",
			data:      `{}`,
			validate: func(result json.RawMessage) {
				// Should return valid DNS config with default interface
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
			},
		},
		{
			name:      "ping with valid address",
			operation: "ping.do",
			data:      `{"address": "127.0.0.1"}`,
			validate: func(result json.RawMessage) {
				// Should return valid ping result
				var response map[string]interface{}
				err := json.Unmarshal(result, &response)
				s.NoError(err)
			},
		},
		{
			name:        "ping without address",
			operation:   "ping.execute",
			data:        `{}`,
			expectError: true,
			errorMsg:    "missing ping address",
		},
		{
			name:        "unsupported network operation",
			operation:   "unknown.get",
			data:        `{}`,
			expectError: true,
			errorMsg:    "unsupported network operation",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			request := job.Request{
				Type:      job.TypeQuery,
				Category:  "network",
				Operation: tt.operation,
				Data:      json.RawMessage(tt.data),
			}

			result, err := s.worker.processNetworkOperation(request)

			if tt.expectError {
				s.Error(err)
				s.Contains(err.Error(), tt.errorMsg)
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

func (s *ProcessorTestSuite) TestProviderFactoryMethods() {
	tests := []struct {
		name        string
		getProvider func() interface{}
	}{
		{
			name:        "getHostProvider",
			getProvider: func() interface{} { return s.worker.getHostProvider() },
		},
		{
			name:        "getDiskProvider",
			getProvider: func() interface{} { return s.worker.getDiskProvider() },
		},
		{
			name:        "getMemProvider",
			getProvider: func() interface{} { return s.worker.getMemProvider() },
		},
		{
			name:        "getLoadProvider",
			getProvider: func() interface{} { return s.worker.getLoadProvider() },
		},
		{
			name:        "getDNSProvider",
			getProvider: func() interface{} { return s.worker.getDNSProvider() },
		},
		{
			name:        "getPingProvider",
			getProvider: func() interface{} { return s.worker.getPingProvider() },
		},
		{
			name:        "getCommandProvider",
			getProvider: func() interface{} { return s.worker.getCommandProvider() },
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			provider := tt.getProvider()
			s.NotNil(provider)
		})
	}
}

func (s *ProcessorTestSuite) TestSystemOperationErrors() {
	tests := []struct {
		name         string
		operation    string
		errorMsg     string
		createWorker func() *Worker
	}{
		{
			name:      "hostname provider error",
			operation: "hostname.get",
			errorMsg:  "failed to get hostname",
			createWorker: func() *Worker {
				hostMock := hostMocks.NewPlainMockProvider(s.mockCtrl)
				hostMock.EXPECT().GetHostname().Return("", errors.New("hostname unavailable"))
				return New(
					afero.NewMemMapFs(),
					config.Config{},
					slog.Default(),
					s.mockJobClient,
					"test-stream",
					hostMock,
					diskMocks.NewPlainMockProvider(s.mockCtrl),
					memMocks.NewPlainMockProvider(s.mockCtrl),
					loadMocks.NewPlainMockProvider(s.mockCtrl),
					dnsMocks.NewPlainMockProvider(s.mockCtrl),
					pingMocks.NewPlainMockProvider(s.mockCtrl),
					commandMocks.NewPlainMockProvider(s.mockCtrl),
				)
			},
		},
		{
			name:      "uptime provider error",
			operation: "uptime.get",
			errorMsg:  "failed to get uptime",
			createWorker: func() *Worker {
				hostMock := hostMocks.NewPlainMockProvider(s.mockCtrl)
				hostMock.EXPECT().
					GetUptime().
					Return(time.Duration(0), errors.New("uptime unavailable"))
				return New(
					afero.NewMemMapFs(),
					config.Config{},
					slog.Default(),
					s.mockJobClient,
					"test-stream",
					hostMock,
					diskMocks.NewPlainMockProvider(s.mockCtrl),
					memMocks.NewPlainMockProvider(s.mockCtrl),
					loadMocks.NewPlainMockProvider(s.mockCtrl),
					dnsMocks.NewPlainMockProvider(s.mockCtrl),
					pingMocks.NewPlainMockProvider(s.mockCtrl),
					commandMocks.NewPlainMockProvider(s.mockCtrl),
				)
			},
		},
		{
			name:      "OS info provider error",
			operation: "os.get",
			errorMsg:  "failed to get OS info",
			createWorker: func() *Worker {
				hostMock := hostMocks.NewPlainMockProvider(s.mockCtrl)
				hostMock.EXPECT().GetOSInfo().Return(nil, errors.New("os info unavailable"))
				return New(
					afero.NewMemMapFs(),
					config.Config{},
					slog.Default(),
					s.mockJobClient,
					"test-stream",
					hostMock,
					diskMocks.NewPlainMockProvider(s.mockCtrl),
					memMocks.NewPlainMockProvider(s.mockCtrl),
					loadMocks.NewPlainMockProvider(s.mockCtrl),
					dnsMocks.NewPlainMockProvider(s.mockCtrl),
					pingMocks.NewPlainMockProvider(s.mockCtrl),
					commandMocks.NewPlainMockProvider(s.mockCtrl),
				)
			},
		},
		{
			name:      "disk provider error",
			operation: "disk.get",
			errorMsg:  "failed to get disk usage",
			createWorker: func() *Worker {
				diskMock := diskMocks.NewPlainMockProvider(s.mockCtrl)
				diskMock.EXPECT().GetLocalUsageStats().Return(nil, errors.New("disk unavailable"))
				return New(
					afero.NewMemMapFs(),
					config.Config{},
					slog.Default(),
					s.mockJobClient,
					"test-stream",
					hostMocks.NewPlainMockProvider(s.mockCtrl),
					diskMock,
					memMocks.NewPlainMockProvider(s.mockCtrl),
					loadMocks.NewPlainMockProvider(s.mockCtrl),
					dnsMocks.NewPlainMockProvider(s.mockCtrl),
					pingMocks.NewPlainMockProvider(s.mockCtrl),
					commandMocks.NewPlainMockProvider(s.mockCtrl),
				)
			},
		},
		{
			name:      "memory provider error",
			operation: "memory.get",
			errorMsg:  "failed to get memory stats",
			createWorker: func() *Worker {
				memMock := memMocks.NewPlainMockProvider(s.mockCtrl)
				memMock.EXPECT().GetStats().Return(nil, errors.New("memory unavailable"))
				return New(
					afero.NewMemMapFs(),
					config.Config{},
					slog.Default(),
					s.mockJobClient,
					"test-stream",
					hostMocks.NewPlainMockProvider(s.mockCtrl),
					diskMocks.NewPlainMockProvider(s.mockCtrl),
					memMock,
					loadMocks.NewPlainMockProvider(s.mockCtrl),
					dnsMocks.NewPlainMockProvider(s.mockCtrl),
					pingMocks.NewPlainMockProvider(s.mockCtrl),
					commandMocks.NewPlainMockProvider(s.mockCtrl),
				)
			},
		},
		{
			name:      "load provider error",
			operation: "load.get",
			errorMsg:  "failed to get load averages",
			createWorker: func() *Worker {
				loadMock := loadMocks.NewPlainMockProvider(s.mockCtrl)
				loadMock.EXPECT().GetAverageStats().Return(nil, errors.New("load unavailable"))
				return New(
					afero.NewMemMapFs(),
					config.Config{},
					slog.Default(),
					s.mockJobClient,
					"test-stream",
					hostMocks.NewPlainMockProvider(s.mockCtrl),
					diskMocks.NewPlainMockProvider(s.mockCtrl),
					memMocks.NewPlainMockProvider(s.mockCtrl),
					loadMock,
					dnsMocks.NewPlainMockProvider(s.mockCtrl),
					pingMocks.NewPlainMockProvider(s.mockCtrl),
					commandMocks.NewPlainMockProvider(s.mockCtrl),
				)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			w := tt.createWorker()
			request := job.Request{
				Type:      job.TypeQuery,
				Category:  "system",
				Operation: tt.operation,
				Data:      json.RawMessage(`{}`),
			}

			result, err := w.processSystemOperation(request)

			s.Error(err)
			s.Contains(err.Error(), tt.errorMsg)
			s.Nil(result)
		})
	}
}

func (s *ProcessorTestSuite) TestNetworkOperationErrors() {
	tests := []struct {
		name         string
		operation    string
		jobType      job.Type
		data         string
		errorMsg     string
		createWorker func() *Worker
	}{
		{
			name:      "DNS get error",
			operation: "dns.get",
			jobType:   job.TypeQuery,
			data:      `{"interface": "eth0"}`,
			errorMsg:  "failed to get DNS config",
			createWorker: func() *Worker {
				dnsMock := dnsMocks.NewPlainMockProvider(s.mockCtrl)
				dnsMock.EXPECT().
					GetResolvConfByInterface("eth0").
					Return(nil, errors.New("DNS lookup failed"))
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
					dnsMock,
					pingMocks.NewPlainMockProvider(s.mockCtrl),
					commandMocks.NewPlainMockProvider(s.mockCtrl),
				)
			},
		},
		{
			name:      "DNS update error",
			operation: "dns.update",
			jobType:   job.TypeModify,
			data:      `{"servers": ["8.8.8.8"], "search_domains": ["example.com"], "interface": "eth0"}`,
			errorMsg:  "failed to set DNS config",
			createWorker: func() *Worker {
				dnsMock := dnsMocks.NewPlainMockProvider(s.mockCtrl)
				dnsMock.EXPECT().
					UpdateResolvConfByInterface(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("DNS update failed"))
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
					dnsMock,
					pingMocks.NewPlainMockProvider(s.mockCtrl),
					commandMocks.NewPlainMockProvider(s.mockCtrl),
				)
			},
		},
		{
			name:      "ping provider error",
			operation: "ping.do",
			jobType:   job.TypeQuery,
			data:      `{"address": "8.8.8.8"}`,
			errorMsg:  "ping failed",
			createWorker: func() *Worker {
				pingMock := pingMocks.NewPlainMockProvider(s.mockCtrl)
				pingMock.EXPECT().Do("8.8.8.8").Return(nil, errors.New("ping timeout"))
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
					pingMock,
					commandMocks.NewPlainMockProvider(s.mockCtrl),
				)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			w := tt.createWorker()
			request := job.Request{
				Type:      tt.jobType,
				Category:  "network",
				Operation: tt.operation,
				Data:      json.RawMessage(tt.data),
			}

			result, err := w.processNetworkOperation(request)

			s.Error(err)
			s.Contains(err.Error(), tt.errorMsg)
			s.Nil(result)
		})
	}
}

func TestProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorTestSuite))
}
