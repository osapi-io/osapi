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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/agent"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/mocks"
	commandMocks "github.com/retr0h/osapi/internal/provider/command/mocks"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/dns/mocks"
	netinfoMocks "github.com/retr0h/osapi/internal/provider/network/netinfo/mocks"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/node/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/node/mem/mocks"
	processMocks "github.com/retr0h/osapi/internal/provider/process/mocks"
)

type FactsPublicTestSuite struct {
	suite.Suite

	mockCtrl         *gomock.Controller
	mockJobClient    *mocks.MockJobClient
	mockFactsKV      *mocks.MockKeyValue
	mockHostProvider *hostMocks.MockProvider
	mockNetinfo      *netinfoMocks.MockProvider
	testAgent        *agent.Agent
}

func (s *FactsPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.mockFactsKV = mocks.NewMockKeyValue(s.mockCtrl)
	s.mockHostProvider = hostMocks.NewDefaultMockProvider(s.mockCtrl)
	s.mockNetinfo = netinfoMocks.NewDefaultMockProvider(s.mockCtrl)

	appConfig := config.Config{
		Agent: config.AgentConfig{
			Labels: map[string]string{"group": "web"},
		},
	}

	s.testAgent = agent.New(
		memfs.New(),
		appConfig,
		slog.Default(),
		s.mockJobClient,
		"test-stream",
		s.mockHostProvider,
		diskMocks.NewDefaultMockProvider(s.mockCtrl),
		memMocks.NewDefaultMockProvider(s.mockCtrl),
		loadMocks.NewDefaultMockProvider(s.mockCtrl),
		dnsMocks.NewDefaultMockProvider(s.mockCtrl),
		pingMocks.NewDefaultMockProvider(s.mockCtrl),
		s.mockNetinfo,
		commandMocks.NewDefaultMockProvider(s.mockCtrl),
		nil,
		nil,
		nil,
		processMocks.NewDefaultMockProvider(s.mockCtrl),
		nil,
		s.mockFactsKV,
	)
}

func (s *FactsPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
	agent.ResetMarshalJSON()
	agent.ResetUnmarshalJSON()
	agent.ResetFactsInterval()
}

func (s *FactsPublicTestSuite) TestWriteFacts() {
	tests := []struct {
		name         string
		setupMock    func()
		teardownMock func()
		validateFunc func()
	}{
		{
			name: "when Put succeeds writes facts",
			setupMock: func() {
				s.mockFactsKV.EXPECT().
					Put(gomock.Any(), "facts.test_agent", gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						_ string,
						data []byte,
					) (uint64, error) {
						var reg job.FactsRegistration
						err := json.Unmarshal(data, &reg)
						s.NoError(err)
						s.Equal("amd64", reg.Architecture)
						s.Equal(4, reg.CPUCount)
						s.Equal("default-hostname.local", reg.FQDN)
						s.Equal("5.15.0-91-generic", reg.KernelVersion)
						s.Equal("systemd", reg.ServiceMgr)
						s.Equal("apt", reg.PackageMgr)
						s.Len(reg.Interfaces, 1)
						s.Equal("eth0", reg.Interfaces[0].Name)
						return uint64(1), nil
					})
			},
		},
		{
			name: "when Put fails logs warning",
			setupMock: func() {
				s.mockFactsKV.EXPECT().
					Put(gomock.Any(), "facts.test_agent", gomock.Any()).
					Return(uint64(0), errors.New("put failed"))
			},
		},
		{
			name: "when marshal fails logs warning",
			setupMock: func() {
				agent.SetMarshalJSON(func(_ interface{}) ([]byte, error) {
					return nil, fmt.Errorf("marshal failure")
				})
			},
			teardownMock: func() {
				agent.ResetMarshalJSON()
			},
		},
		{
			name: "when provider errors partial data still written",
			setupMock: func() {
				// Override the default mock expectations with error-returning ones.
				errHostProvider := func() *hostMocks.MockProvider {
					m := hostMocks.NewPlainMockProvider(s.mockCtrl)
					m.EXPECT().GetArchitecture().Return("", errors.New("arch fail")).AnyTimes()
					m.EXPECT().GetKernelVersion().Return("", errors.New("kernel fail")).AnyTimes()
					m.EXPECT().GetFQDN().Return("", errors.New("fqdn fail")).AnyTimes()
					m.EXPECT().GetCPUCount().Return(0, errors.New("cpu fail")).AnyTimes()
					m.EXPECT().GetServiceManager().Return("", errors.New("svc fail")).AnyTimes()
					m.EXPECT().GetPackageManager().Return("", errors.New("pkg fail")).AnyTimes()
					return m
				}()
				agent.SetAgentHostProvider(s.testAgent, errHostProvider)

				errNetinfoProvider := func() *netinfoMocks.MockProvider {
					m := netinfoMocks.NewPlainMockProvider(s.mockCtrl)
					m.EXPECT().GetInterfaces().Return(nil, errors.New("net fail")).AnyTimes()
					m.EXPECT().GetRoutes().Return(nil, errors.New("routes fail")).AnyTimes()
					m.EXPECT().
						GetPrimaryInterface().
						Return("", errors.New("primary fail")).
						AnyTimes()
					return m
				}()
				agent.SetAgentNetinfoProvider(s.testAgent, errNetinfoProvider)

				s.mockFactsKV.EXPECT().
					Put(gomock.Any(), "facts.test_agent", gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						_ string,
						data []byte,
					) (uint64, error) {
						var reg job.FactsRegistration
						err := json.Unmarshal(data, &reg)
						s.NoError(err)
						// All fields should be zero values since providers failed.
						s.Empty(reg.Architecture)
						s.Empty(reg.KernelVersion)
						s.Empty(reg.FQDN)
						s.Zero(reg.CPUCount)
						s.Empty(reg.ServiceMgr)
						s.Empty(reg.PackageMgr)
						s.Nil(reg.Interfaces)
						return uint64(1), nil
					})
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			if tt.teardownMock != nil {
				defer tt.teardownMock()
			}
			agent.ExportWriteFacts(context.Background(), s.testAgent, "test-agent")
		})
	}
}

func (s *FactsPublicTestSuite) TestStartFactsRefresh() {
	tests := []struct {
		name      string
		setupMock func()
	}{
		{
			name: "ticker fires and refreshes facts",
			setupMock: func() {
				// Initial write + at least 1 ticker refresh
				s.mockFactsKV.EXPECT().
					Put(gomock.Any(), "facts.test_agent", gomock.Any()).
					Return(uint64(1), nil).
					MinTimes(2)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			agent.SetFactsInterval(10 * time.Millisecond)

			ctx, cancel := context.WithCancel(context.Background())
			agent.ExportStartFacts(ctx, s.testAgent, "test-agent")

			// Wait for at least one ticker refresh
			time.Sleep(50 * time.Millisecond)
			cancel()

			// Wait for goroutine to finish
			agent.WaitAgentWG(s.testAgent)
		})
	}
}

func (s *FactsPublicTestSuite) TestGetFacts() {
	tests := []struct {
		name         string
		setupFunc    func()
		teardownFunc func()
		validateFunc func(result map[string]any)
	}{
		{
			name:      "when cachedFacts is nil returns nil",
			setupFunc: func() {},
			validateFunc: func(result map[string]any) {
				s.Nil(result)
			},
		},
		{
			name: "when cachedFacts populated returns fact map",
			setupFunc: func() {
				agent.SetAgentCachedFacts(s.testAgent, &job.FactsRegistration{
					Architecture: "amd64",
					CPUCount:     4,
					FQDN:         "test.local",
				})
			},
			validateFunc: func(result map[string]any) {
				s.Require().NotNil(result)
				s.Equal("amd64", result["architecture"])
				s.Equal(float64(4), result["cpu_count"])
				s.Equal("test.local", result["fqdn"])
			},
		},
		{
			name: "when marshal fails returns nil",
			setupFunc: func() {
				agent.SetAgentCachedFacts(s.testAgent, &job.FactsRegistration{
					Architecture: "amd64",
				})
				agent.SetMarshalJSON(func(_ interface{}) ([]byte, error) {
					return nil, fmt.Errorf("marshal failure")
				})
			},
			teardownFunc: func() {
				agent.ResetMarshalJSON()
			},
			validateFunc: func(result map[string]any) {
				s.Nil(result)
			},
		},
		{
			name: "when unmarshal fails returns nil",
			setupFunc: func() {
				agent.SetAgentCachedFacts(s.testAgent, &job.FactsRegistration{
					Architecture: "amd64",
				})
				agent.SetUnmarshalJSON(func(_ []byte, _ interface{}) error {
					return fmt.Errorf("unmarshal failure")
				})
			},
			teardownFunc: func() {
				agent.ResetUnmarshalJSON()
			},
			validateFunc: func(result map[string]any) {
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			agent.SetAgentCachedFacts(s.testAgent, nil)
			tt.setupFunc()
			if tt.teardownFunc != nil {
				defer tt.teardownFunc()
			}
			result := s.testAgent.GetFacts()
			tt.validateFunc(result)
		})
	}
}

func (s *FactsPublicTestSuite) TestFactsKey() {
	tests := []struct {
		name     string
		hostname string
		expected string
	}{
		{
			name:     "simple hostname",
			hostname: "web-01",
			expected: "facts.web_01",
		},
		{
			name:     "hostname with dots",
			hostname: "Johns-MacBook-Pro.local",
			expected: "facts.Johns_MacBook_Pro_local",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := agent.ExportFactsKey(tt.hostname)
			s.Equal(tt.expected, result)
		})
	}
}

func TestFactsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(FactsPublicTestSuite))
}
