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
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/gohai/pkg/gohai"

	agentMocks "github.com/osapi-io/osapi/internal/agent/mocks"

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

type FactsPublicTestSuite struct {
	suite.Suite

	mockCtrl         *gomock.Controller
	mockJobClient    *mocks.MockJobClient
	mockFactsKV      *agentMocks.MockNATSClient
	mockHostProvider *hostMocks.MockProvider
	testAgent        *agent.Agent
}

func (s *FactsPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.mockFactsKV = agentMocks.NewMockNATSClient(s.mockCtrl)
	// The agent reads its previous registration and facts once, for
	// the timeline; there is none.
	s.mockFactsKV.EXPECT().KVGet(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, jetstream.ErrKeyNotFound).AnyTimes()
	s.mockHostProvider = hostMocks.NewDefaultMockProvider(s.mockCtrl)

	appConfig := config.Config{
		Agent: config.AgentConfig{
			Labels: map[string]string{"group": "web"},
		},
	}

	s.testAgent = newTestAgent(newTestAgentParams{
		appConfig:       appConfig,
		jobClient:       s.mockJobClient,
		streamName:      "test-stream",
		hostProvider:    s.mockHostProvider,
		diskProvider:    diskMocks.NewDefaultMockProvider(s.mockCtrl),
		memoryProvider:  memoryMocks.NewDefaultMockProvider(s.mockCtrl),
		loadProvider:    loadMocks.NewDefaultMockProvider(s.mockCtrl),
		processProvider: processMocks.NewDefaultMockProvider(s.mockCtrl),
		factsBucket:     "facts",
		natsClient:      s.mockFactsKV,
	})
}

func (s *FactsPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
	agent.ResetMarshalJSON()
	agent.ResetUnmarshalJSON()
	agent.ResetDefaultFactsInterval()
	agent.ResetCollectFactsFn()
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
					KVPut(gomock.Any(), "facts", "facts.test_machine_id", gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						_ string,
						_ string,
						data []byte,
					) (uint64, error) {
						var reg job.FactsRegistration
						err := json.Unmarshal(data, &reg)
						s.NoError(err)
						s.Equal("x86_64", reg.Architecture)
						s.Equal(4, reg.CPUCount)
						s.Equal("default-hostname.local", reg.FQDN)
						s.Equal("5.15.0-91-generic", reg.KernelVersion)
						s.Equal("systemd", reg.ServiceMgr)
						s.Equal("apt", reg.PackageMgr)
						s.True(reg.Containerized)
						s.Equal("eth0", reg.PrimaryInterface)
						s.Equal([]job.NetworkInterface{
							{
								Name:   "eth0",
								IPv4:   "10.0.0.5",
								IPv6:   "fe80::1",
								MAC:    "02:42:ac:11:00:02",
								Family: "dual",
							},
							{Name: "wg0", IPv6: "fd00::2", Family: "inet6"},
							{Name: "eth2", IPv4: "192.168.1.2", Family: "inet"},
						}, reg.Interfaces, "up, not loopback; the first address of each family")
						s.Equal([]job.Route{
							{
								Destination: "0.0.0.0",
								Gateway:     "10.0.0.1",
								Interface:   "eth0",
								Mask:        "/0",
								Metric:      100,
							},
							{Destination: "10.0.0.0", Interface: "eth0", Mask: "/24"},
							{Destination: "127", Gateway: "127.0.0.1", Interface: "lo"},
						}, reg.Routes, "IPv4 only")
						return uint64(1), nil
					})
			},
		},
		{
			name: "when Put fails logs warning",
			setupMock: func() {
				s.mockFactsKV.EXPECT().
					KVPut(gomock.Any(), "facts", "facts.test_machine_id", gomock.Any()).
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
			name: "when gohai fails the entry is still written, empty",
			setupMock: func() {
				agent.SetCollectFactsFn(func() (*gohai.Facts, error) {
					return nil, errors.New("collect failed")
				})
				s.mockFactsKV.EXPECT().
					KVPut(gomock.Any(), "facts", "facts.test_machine_id", gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						_ string,
						_ string,
						data []byte,
					) (uint64, error) {
						var reg job.FactsRegistration
						s.NoError(json.Unmarshal(data, &reg))
						s.Equal(job.FactsRegistration{}, reg)
						return uint64(1), nil
					})
			},
		},
		{
			name: "when a collector failed its fields stay empty",
			setupMock: func() {
				agent.SetCollectFactsFn(func() (*gohai.Facts, error) {
					return &gohai.Facts{}, nil
				})
				s.mockFactsKV.EXPECT().
					KVPut(gomock.Any(), "facts", "facts.test_machine_id", gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						_ string,
						_ string,
						data []byte,
					) (uint64, error) {
						var reg job.FactsRegistration
						s.NoError(json.Unmarshal(data, &reg))
						s.Equal(job.FactsRegistration{}, reg)
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
			agent.ExportWriteFacts(
				context.Background(),
				s.testAgent,
				"test-machine-id",
				"test-agent",
			)
		})
	}
}

func (s *FactsPublicTestSuite) TestCollectFacts() {
	f, err := agent.ExportCollectFacts()
	s.Require().NoError(err, "gohai collects this host's facts")
	s.NotNil(f.Kernel)
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
					KVPut(gomock.Any(), "facts", "facts.test_machine_id", gomock.Any()).
					Return(uint64(1), nil).
					MinTimes(2)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			agent.SetDefaultFactsInterval(10 * time.Millisecond)

			ctx, cancel := context.WithCancel(context.Background())
			agent.ExportStartFacts(ctx, s.testAgent, "test-machine-id", "test-agent")

			// Wait for at least one ticker refresh
			time.Sleep(50 * time.Millisecond)
			cancel()

			// Wait for goroutine to finish
			agent.WaitAgentWG(s.testAgent)
		})
	}
}

func (s *FactsPublicTestSuite) TestStartFactsInterval() {
	tests := []struct {
		name         string
		interval     string
		setupMock    func()
		validateFunc func()
	}{
		{
			name:     "when facts interval is configured uses configured value",
			interval: "20ms",
			setupMock: func() {
				s.mockFactsKV.EXPECT().
					KVPut(gomock.Any(), "facts", "facts.test_machine_id", gomock.Any()).
					Return(uint64(1), nil).
					AnyTimes()
			},
			validateFunc: func() {},
		},
		{
			name:     "when facts interval is empty uses default",
			interval: "",
			setupMock: func() {
				s.mockFactsKV.EXPECT().
					KVPut(gomock.Any(), "facts", "facts.test_machine_id", gomock.Any()).
					Return(uint64(1), nil).
					AnyTimes()
			},
			validateFunc: func() {},
		},
		{
			name:     "when facts interval is invalid uses default",
			interval: "7d",
			setupMock: func() {
				s.mockFactsKV.EXPECT().
					KVPut(gomock.Any(), "facts", "facts.test_machine_id", gomock.Any()).
					Return(uint64(1), nil).
					AnyTimes()
			},
			validateFunc: func() {},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			// Override the default to be very short so "empty" and "invalid"
			// cases also tick quickly.
			agent.SetDefaultFactsInterval(10 * time.Millisecond)

			cfg := agent.GetAgentAppConfig(s.testAgent)
			cfg.Agent.Facts.Interval = tt.interval
			agent.SetAgentAppConfig(s.testAgent, cfg)

			ctx, cancel := context.WithCancel(context.Background())
			agent.ExportStartFacts(ctx, s.testAgent, "test-machine-id", "test-agent")

			// Wait for at least one ticker refresh
			time.Sleep(50 * time.Millisecond)
			cancel()

			// Wait for goroutine to finish
			agent.WaitAgentWG(s.testAgent)

			tt.validateFunc()
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

func (s *FactsPublicTestSuite) TestAgentFacts() {
	tests := []struct {
		name         string
		agent        *agent.Agent
		validateFunc func(result map[string]any)
	}{
		{
			name:  "when the agent is not made yet returns nil",
			agent: nil,
			validateFunc: func(result map[string]any) {
				s.Nil(result)
			},
		},
		{
			name:  "when the agent is made returns its facts",
			agent: s.testAgent,
			validateFunc: func(result map[string]any) {
				s.Require().NotNil(result)
				s.Equal("amd64", result["architecture"])
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			agent.SetAgentCachedFacts(s.testAgent, &job.FactsRegistration{
				Architecture: "amd64",
			})
			defer agent.SetAgentCachedFacts(s.testAgent, nil)

			tt.validateFunc(agent.ExportAgentFacts(tt.agent)())
		})
	}
}

func (s *FactsPublicTestSuite) TestFactsKey() {
	tests := []struct {
		name         string
		machineID    string
		validateFunc func(string)
	}{
		{
			name:      "simple machine ID",
			machineID: "abc-123-def",
			validateFunc: func(got string) {
				s.Equal("facts.abc_123_def", got)
			},
		},
		{
			name:      "machine ID with dots",
			machineID: "A1B2C3D4-E5F6.7890",
			validateFunc: func(got string) {
				s.Equal("facts.A1B2C3D4_E5F6_7890", got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := agent.ExportFactsKey(tt.machineID)
			tt.validateFunc(result)
		})
	}
}

func TestFactsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(FactsPublicTestSuite))
}
