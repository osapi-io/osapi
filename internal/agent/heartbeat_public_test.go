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
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	agentMocks "github.com/osapi-io/osapi/internal/agent/mocks"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/agent/identity"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/pki"
	diskMocks "github.com/osapi-io/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/osapi-io/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/osapi-io/osapi/internal/provider/node/load/mocks"
	memoryMocks "github.com/osapi-io/osapi/internal/provider/node/memory/mocks"
	processMocks "github.com/osapi-io/osapi/internal/telemetry/process/mocks"
)

type HeartbeatPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	mockKV        *agentMocks.MockNATSClient
	appFs         avfs.VFS
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *HeartbeatPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	// The heartbeat reads the controller's config of the agent.
	s.mockJobClient.EXPECT().
		GetAgentConfig(gomock.Any(), gomock.Any()).
		Return(job.AgentConfig{}, nil).
		AnyTimes()
	// The heartbeat follows the rotation of the controller's key.
	s.mockJobClient.EXPECT().
		ControllerKey(gomock.Any()).
		Return(pki.ControllerKeyRecord{}, false, nil).
		AnyTimes()
	// The last-seen write is checked in TestWriteSeen; the rest allow it.
	s.mockJobClient.EXPECT().WriteAgentSeen(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	s.mockKV = agentMocks.NewMockNATSClient(s.mockCtrl)
	// The agent reads its previous registration and facts once, for
	// the timeline; there is none.
	s.mockKV.EXPECT().KVGet(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, jetstream.ErrKeyNotFound).AnyTimes()
	s.appFs = memfs.New()
	s.logger = slog.Default()

	// Mock identity so tests don't depend on /etc/machine-id.
	agent.SetGetIdentityFn(func(_ avfs.VFS, _ string) (*identity.Identity, error) {
		return &identity.Identity{
			MachineID: "test-machine-id",
			Hostname:  "test-agent",
		}, nil
	})

	s.appConfig = config.Config{
		NATS: config.NATS{
			Stream: config.NATSStream{Name: "test-stream"},
			Registry: config.NATSRegistry{
				Bucket:   "agent-registry",
				TTL:      "30s",
				Storage:  "file",
				Replicas: 1,
			},
		},
		Agent: config.AgentConfig{
			Hostname: "test-agent",
			MaxJobs:  5,
			Labels:   map[string]string{"group": "web"},
			PKI:      config.AgentPKI{KeyDir: keyDir},
			Consumer: config.AgentConsumer{
				AckWait:       "30s",
				BackOff:       []string{"1s", "2s", "5s"},
				MaxDeliver:    3,
				MaxAckPending: 10,
				ReplayPolicy:  "instant",
			},
		},
	}
	enroll(s.appFs, s.mockJobClient)
}

func (s *HeartbeatPublicTestSuite) TearDownTest() {
	agent.ResetGetIdentityFn()
	s.mockCtrl.Finish()
}

func (s *HeartbeatPublicTestSuite) TestStartWithHeartbeat() {
	tests := []struct {
		name      string
		setupFunc func() *agent.Agent
		stopFunc  func(a *agent.Agent)
	}{
		{
			name: "when registryKV is set registers and deregisters",
			setupFunc: func() *agent.Agent {
				// Drain check on each heartbeat tick (no drain flag present).
				// machineID is resolved from the real system at Start(), so use Any().
				s.mockJobClient.EXPECT().
					CheckDrainFlag(gomock.Any(), gomock.Any()).
					Return(false).
					AnyTimes()

				// Heartbeat initial write (machineID-based key from real system)
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", gomock.Any(), gomock.Any()).
					Return(uint64(1), nil).
					MinTimes(1)

				// Deregister on stop
				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "registry", gomock.Any()).
					Return(nil).
					Times(1)

					// 3 base + 1 label = 4 consumers per job type, x2 (query+modify) = 8
				// Start and Stop write started and stopped to the timeline.
				s.mockJobClient.EXPECT().
					WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(8)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(8)

				return newTestAgent(newTestAgentParams{
					appFs:           s.appFs,
					appConfig:       s.appConfig,
					logger:          s.logger,
					jobClient:       s.mockJobClient,
					streamName:      "test-stream",
					hostProvider:    hostMocks.NewDefaultMockProvider(s.mockCtrl),
					diskProvider:    diskMocks.NewDefaultMockProvider(s.mockCtrl),
					memoryProvider:  memoryMocks.NewDefaultMockProvider(s.mockCtrl),
					loadProvider:    loadMocks.NewDefaultMockProvider(s.mockCtrl),
					processProvider: processMocks.NewDefaultMockProvider(s.mockCtrl),
					registryBucket:  "registry",
					natsClient:      s.mockKV,
				})
			},
			stopFunc: func(a *agent.Agent) {
				stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				a.Stop(stopCtx)
			},
		},
		{
			name: "when registryKV is nil skips heartbeat",
			setupFunc: func() *agent.Agent {
				// 3 base + 1 label = 4 consumers per job type, x2 (query+modify) = 8
				// Start and Stop write started and stopped to the timeline.
				s.mockJobClient.EXPECT().
					WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(8)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(8)

				return newTestAgent(newTestAgentParams{
					appFs:          s.appFs,
					appConfig:      s.appConfig,
					logger:         s.logger,
					jobClient:      s.mockJobClient,
					streamName:     "test-stream",
					hostProvider:   hostMocks.NewDefaultMockProvider(s.mockCtrl),
					diskProvider:   diskMocks.NewDefaultMockProvider(s.mockCtrl),
					memoryProvider: memoryMocks.NewDefaultMockProvider(s.mockCtrl),
					loadProvider:   loadMocks.NewDefaultMockProvider(s.mockCtrl),
				})
			},
			stopFunc: func(a *agent.Agent) {
				stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				a.Stop(stopCtx)
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := tt.setupFunc()
			a.Start()
			tt.stopFunc(a)
		})
	}
}

// TestStartOrder checks that Start creates the job consumers before the
// first heartbeat makes the agent Ready: the consumers deliver only the jobs
// sent after they exist (osapi-io/osapi#1560).
func (s *HeartbeatPublicTestSuite) TestStartOrder() {
	var (
		mu    sync.Mutex
		calls []string
	)
	record := func(call string) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, call)
	}

	tests := []struct {
		name         string
		setupFunc    func() *agent.Agent
		stopFunc     func(a *agent.Agent)
		validateFunc func()
	}{
		{
			name: "creates each job consumer before the first heartbeat",
			setupFunc: func() *agent.Agent {
				calls = nil
				s.mockJobClient.EXPECT().
					CheckDrainFlag(gomock.Any(), gomock.Any()).
					Return(false).
					AnyTimes()
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", gomock.Any(), gomock.Any()).
					DoAndReturn(func(context.Context, string, string, []byte) (uint64, error) {
						record("heartbeat")

						return 1, nil
					}).
					MinTimes(1)
				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "registry", gomock.Any()).
					Return(nil).
					Times(1)
				s.mockJobClient.EXPECT().
					WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					DoAndReturn(func(context.Context, string, jetstream.ConsumerConfig) error {
						record("consumer")

						return nil
					}).
					Times(8)
				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(8)

				return newTestAgent(newTestAgentParams{
					appFs:           s.appFs,
					appConfig:       s.appConfig,
					logger:          s.logger,
					jobClient:       s.mockJobClient,
					streamName:      "test-stream",
					hostProvider:    hostMocks.NewDefaultMockProvider(s.mockCtrl),
					diskProvider:    diskMocks.NewDefaultMockProvider(s.mockCtrl),
					memoryProvider:  memoryMocks.NewDefaultMockProvider(s.mockCtrl),
					loadProvider:    loadMocks.NewDefaultMockProvider(s.mockCtrl),
					processProvider: processMocks.NewDefaultMockProvider(s.mockCtrl),
					registryBucket:  "registry",
					natsClient:      s.mockKV,
				})
			},
			stopFunc: func(a *agent.Agent) {
				stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				a.Stop(stopCtx)
			},
			validateFunc: func() {
				mu.Lock()
				defer mu.Unlock()
				first := slices.Index(calls, "heartbeat")
				s.Require().NotEqual(-1, first)
				s.Equal(8, strings.Count(strings.Join(calls[:first], ","), "consumer"),
					"the first heartbeat came before each consumer: %v", calls)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := tt.setupFunc()
			a.Start()
			tt.stopFunc(a)
			tt.validateFunc()
		})
	}
}

func TestHeartbeatPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HeartbeatPublicTestSuite))
}

// HeartbeatLowLevelPublicTestSuite tests the lower-level heartbeat methods.
type HeartbeatLowLevelPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	mockKV        *agentMocks.MockNATSClient
	testAgent     *agent.Agent
}

func (s *HeartbeatLowLevelPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	// The heartbeat reads the controller's config of the agent.
	s.mockJobClient.EXPECT().
		GetAgentConfig(gomock.Any(), gomock.Any()).
		Return(job.AgentConfig{}, nil).
		AnyTimes()
	// The heartbeat follows the rotation of the controller's key.
	s.mockJobClient.EXPECT().
		ControllerKey(gomock.Any()).
		Return(pki.ControllerKeyRecord{}, false, nil).
		AnyTimes()
	// The last-seen write is checked in TestWriteSeen; the rest allow it.
	s.mockJobClient.EXPECT().WriteAgentSeen(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	s.mockKV = agentMocks.NewMockNATSClient(s.mockCtrl)
	// The agent reads its previous registration and facts once, for
	// the timeline; there is none.
	s.mockKV.EXPECT().KVGet(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, jetstream.ErrKeyNotFound).AnyTimes()

	appConfig := config.Config{
		Agent: config.AgentConfig{
			Labels: map[string]string{"group": "web"},
		},
	}

	// Use DefaultMockProviders so provider calls during writeRegistration are satisfied.
	s.testAgent = newTestAgent(newTestAgentParams{
		appConfig:       appConfig,
		jobClient:       s.mockJobClient,
		streamName:      "test-stream",
		hostProvider:    hostMocks.NewDefaultMockProvider(s.mockCtrl),
		diskProvider:    diskMocks.NewDefaultMockProvider(s.mockCtrl),
		memoryProvider:  memoryMocks.NewDefaultMockProvider(s.mockCtrl),
		loadProvider:    loadMocks.NewDefaultMockProvider(s.mockCtrl),
		processProvider: processMocks.NewDefaultMockProvider(s.mockCtrl),
		registryBucket:  "registry",
		natsClient:      s.mockKV,
	})
	agent.SetAgentState(s.testAgent, types.AgentReady)
	agent.SetAgentMachineID(s.testAgent, "test-machine-id")

	// writeRegistration now calls handleDrainDetection which checks drain flag.
	// Default: no drain flag present.
	s.mockJobClient.EXPECT().
		CheckDrainFlag(gomock.Any(), "test-machine-id").
		Return(false).
		AnyTimes()
}

func (s *HeartbeatLowLevelPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
	agent.ResetMarshalJSON()
	agent.ResetHeartbeatInterval()
}

func (s *HeartbeatLowLevelPublicTestSuite) TestWriteRegistration() {
	tests := []struct {
		name         string
		setupMock    func()
		teardownMock func()
		validateFunc func(before time.Time)
	}{
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
			validateFunc: func(_ time.Time) {
				// Marshal failure occurs before Put, so heartbeat time
				// should NOT be updated.
				got := s.testAgent.LastHeartbeatTime()
				s.True(got.IsZero(), "expected zero heartbeat time after marshal failure")
			},
		},
		{
			name: "when Put fails logs warning",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "agents.test_machine_id", gomock.Any()).
					Return(uint64(0), errors.New("put failed"))
			},
			validateFunc: func(_ time.Time) {
				// Put failure means heartbeat time should NOT be updated.
				got := s.testAgent.LastHeartbeatTime()
				s.True(got.IsZero(), "expected zero heartbeat time after Put failure")
			},
		},
		{
			name: "when Put succeeds writes registration",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "agents.test_machine_id", gomock.Any()).
					Return(uint64(1), nil)
			},
			validateFunc: func(before time.Time) {
				got := s.testAgent.LastHeartbeatTime()
				s.False(got.IsZero(), "expected non-zero heartbeat time after successful Put")
				s.True(
					!got.Before(before),
					"heartbeat time should be at or after test start",
				)
			},
		},
		{
			name: "when PKI enabled includes fingerprint in registration",
			setupMock: func() {
				m := pki.New(memfs.New(), "/keys", "agent")
				s.Require().NoError(m.LoadOrGenerate())
				agent.SetAgentPKIManager(s.testAgent, m)

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "agents.test_machine_id", gomock.Any()).
					DoAndReturn(func(
						_ interface{},
						_ string,
						_ string,
						data []byte,
					) (uint64, error) {
						s.Contains(string(data), "fingerprint")
						s.Contains(string(data), "SHA256:")
						return uint64(1), nil
					})
			},
			teardownMock: func() {
				agent.SetAgentPKIManager(s.testAgent, nil)
			},
			validateFunc: func(_ time.Time) {},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			if tt.teardownMock != nil {
				defer tt.teardownMock()
			}
			before := time.Now()
			agent.ExportWriteRegistration(
				context.Background(),
				s.testAgent,
				"test-machine-id",
				"test-agent",
			)
			tt.validateFunc(before)
		})
	}
}

func (s *HeartbeatLowLevelPublicTestSuite) TestWriteRegistrationStoresHeartbeatTime() {
	tests := []struct {
		name         string
		validateFunc func(time.Time, time.Time, time.Time)
	}{
		{
			name: "when Put succeeds stores last heartbeat time",
			validateFunc: func(got, before, after time.Time) {
				s.False(got.IsZero(), "expected non-zero heartbeat time after successful Put")
				s.True(
					!got.Before(before) && !got.After(after),
					"heartbeat time should be between before and after write",
				)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockKV.EXPECT().
				KVPut(gomock.Any(), "registry", "agents.test_machine_id", gomock.Any()).
				Return(uint64(1), nil)

			before := time.Now()
			agent.ExportWriteRegistration(
				context.Background(),
				s.testAgent,
				"test-machine-id",
				"test-agent",
			)
			after := time.Now()

			got := s.testAgent.LastHeartbeatTime()
			tt.validateFunc(got, before, after)
		})
	}
}

func (s *HeartbeatLowLevelPublicTestSuite) TestWriteRegistrationHasVersion() {
	s.Run("when the agent has a version its registration has it", func() {
		s.testAgent.SetVersion("v1.4.0")

		var got job.AgentRegistration
		s.mockKV.EXPECT().
			KVPut(gomock.Any(), "registry", "agents.test_machine_id", gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ string, data []byte) (uint64, error) {
				s.Require().NoError(json.Unmarshal(data, &got))

				return 1, nil
			})

		agent.ExportWriteRegistration(
			context.Background(),
			s.testAgent,
			"test-machine-id",
			"test-agent",
		)

		s.Equal("v1.4.0", got.AgentVersion)
	})
}

func (s *HeartbeatLowLevelPublicTestSuite) TestWriteRegistrationHasFamily() {
	s.Run("its OS info has the family that the agent picks providers by", func() {
		agent.SetPlatformFns(func() string { return "suse" }, func() bool { return false })
		defer agent.ResetPlatformFns()

		var got job.AgentRegistration
		s.mockKV.EXPECT().
			KVPut(gomock.Any(), "registry", "agents.test_machine_id", gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ string, data []byte) (uint64, error) {
				s.Require().NoError(json.Unmarshal(data, &got))

				return 1, nil
			})

		agent.ExportWriteRegistration(
			context.Background(),
			s.testAgent,
			"test-machine-id",
			"test-agent",
		)

		s.Require().NotNil(got.OSInfo)
		s.Equal("suse", got.OSInfo.Family)
	})
}

func (s *HeartbeatLowLevelPublicTestSuite) TestWriteRegistrationFindsProgramsAgain() {
	s.Run("a program installed after the start leaves the list at the next heartbeat", func() {
		agent.SetPlatformFns(func() string { return "debian" }, func() bool { return false })
		defer agent.ResetPlatformFns()
		behavior.Register(behavior.Define("ntp", "NTP", behavior.Node, behavior.Runs("chronyc")))
		defer behavior.Reset()
		installed := false
		agent.SetLookPath(func(p string) (string, error) {
			if installed || !strings.HasSuffix(p, "chronyc") {
				return "/usr/bin/" + p, nil
			}

			return "", errors.New("not found")
		})
		defer agent.ResetLookPath()
		agent.SetAppFs(s.testAgent, memfs.New())

		var got []job.AgentRegistration
		s.mockKV.EXPECT().
			KVPut(gomock.Any(), "registry", "agents.test_machine_id", gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ string, data []byte) (uint64, error) {
				var reg job.AgentRegistration
				s.Require().NoError(json.Unmarshal(data, &reg))
				got = append(got, reg)

				return 1, nil
			}).
			Times(2)

		agent.ExportWriteRegistration(
			context.Background(),
			s.testAgent,
			"test-machine-id",
			"test-agent",
		)
		installed = true
		agent.ExportWriteRegistration(
			context.Background(),
			s.testAgent,
			"test-machine-id",
			"test-agent",
		)

		s.Require().Len(got, 2)
		s.Equal([]string{"chronyc"}, got[0].MissingPrograms)
		s.Empty(got[1].MissingPrograms)
		// The registration has the certificate of the agent's metrics
		// server, which this agent has not made.
		s.Require().Len(got[0].Certificates, 1)
		s.Equal("agent-metrics", got[0].Certificates[0].Name)
		s.Equal(pki.CertNotMade, got[0].Certificates[0].State)
	})
}

func (s *HeartbeatLowLevelPublicTestSuite) TestDeregister() {
	tests := []struct {
		name         string
		setupMock    func()
		validateFunc func(assert.PanicTestFunc)
	}{
		{
			name: "when Delete fails logs warning",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "registry", "agents.test_machine_id").
					Return(errors.New("delete failed"))
			},
			validateFunc: func(deregister assert.PanicTestFunc) {
				s.NotPanics(deregister)
			},
		},
		{
			name: "when Delete succeeds logs deregistration",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "registry", "agents.test_machine_id").
					Return(nil)
			},
			validateFunc: func(deregister assert.PanicTestFunc) {
				s.NotPanics(deregister)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			// Deregister is best-effort (fire-and-forget). It deletes a
			// KV key and logs the outcome but does not mutate agent state.
			// The gomock expectations above verify the correct KV call was
			// made; beyond that, we only verify the function completes
			// without panicking.
			tt.validateFunc(func() {
				agent.ExportDeregister(s.testAgent, "test-machine-id")
			})
		})
	}
}

func (s *HeartbeatLowLevelPublicTestSuite) TestStartHeartbeatRefresh() {
	tests := []struct {
		name      string
		setupMock func()
	}{
		{
			name: "ticker fires and refreshes registration",
			setupMock: func() {
				// Initial write + at least 1 ticker refresh
				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "agents.test_machine_id", gomock.Any()).
					Return(uint64(1), nil).
					MinTimes(2)

				// Deregister on cancel
				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "registry", "agents.test_machine_id").
					Return(nil).
					Times(1)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			// Return the same hostname to prevent hostname change detection.
			agent.SetGetAgentHostnameFn(func(_ string) (string, error) {
				return "test-agent", nil
			})
			defer agent.ResetGetAgentHostnameFn()

			agent.SetHeartbeatInterval(10 * time.Millisecond)

			ctx, cancel := context.WithCancel(context.Background())
			agent.ExportStartHeartbeat(ctx, s.testAgent, "test-machine-id", "test-agent")

			// Wait for at least one ticker refresh
			time.Sleep(50 * time.Millisecond)
			cancel()

			// Wait for goroutine to finish
			agent.WaitAgentWG(s.testAgent)
		})
	}
}

func (s *HeartbeatLowLevelPublicTestSuite) TestStartHeartbeatHostnameChange() {
	tests := []struct {
		name            string
		initialHostname string
		hostnameReply   string
		expectChanged   bool
		validateFunc    func(any)
	}{
		{
			name:            "when hostname changes updates cached hostname",
			initialHostname: "old-host",
			hostnameReply:   "new-host",
			expectChanged:   true,
			validateFunc: func(got any) {
				s.Equal("new-host", got)
			},
		},
		{
			name:            "when hostname unchanged does not resubscribe",
			initialHostname: "same-host",
			hostnameReply:   "same-host",
			expectChanged:   false,
			validateFunc: func(got any) {
				s.Equal("same-host", got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Create fresh mocks per subtest to avoid gomock expectation leaks.
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()

			mockKV := agentMocks.NewMockNATSClient(ctrl)
			// The agent reads its previous registration and facts once, for
			// the timeline; there is none.
			mockKV.EXPECT().KVGet(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, jetstream.ErrKeyNotFound).AnyTimes()
			mockJobClient := mocks.NewMockJobClient(ctrl)
			mockJobClient.EXPECT().
				GetAgentConfig(gomock.Any(), gomock.Any()).
				Return(job.AgentConfig{}, nil).
				AnyTimes()
			// The heartbeat follows the rotation of the controller's key.
			mockJobClient.EXPECT().
				ControllerKey(gomock.Any()).
				Return(pki.ControllerKeyRecord{}, false, nil).
				AnyTimes()
			mockJobClient.EXPECT().WriteAgentSeen(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			appConfig := config.Config{
				Agent: config.AgentConfig{
					Labels: map[string]string{"group": "web"},
				},
			}

			testAgent := newTestAgent(newTestAgentParams{
				appConfig:       appConfig,
				jobClient:       mockJobClient,
				streamName:      "test-stream",
				hostProvider:    hostMocks.NewDefaultMockProvider(ctrl),
				diskProvider:    diskMocks.NewDefaultMockProvider(ctrl),
				memoryProvider:  memoryMocks.NewDefaultMockProvider(ctrl),
				loadProvider:    loadMocks.NewDefaultMockProvider(ctrl),
				processProvider: processMocks.NewDefaultMockProvider(ctrl),
				registryBucket:  "registry",
				natsClient:      mockKV,
			})
			agent.SetAgentState(testAgent, types.AgentReady)
			agent.SetAgentMachineID(testAgent, "test-machine-id")
			agent.SetAgentHostname(testAgent, tt.initialHostname)

			// Drain check — no drain flag present.
			mockJobClient.EXPECT().
				CheckDrainFlag(gomock.Any(), "test-machine-id").
				Return(false).
				AnyTimes()

			reply := tt.hostnameReply
			agent.SetGetAgentHostnameFn(func(_ string) (string, error) {
				return reply, nil
			})
			defer agent.ResetGetAgentHostnameFn()

			// Consumers use machine ID (permanent), so hostname changes
			// do NOT trigger consumer stop/start — only the cached
			// hostname is updated for registry entries and logging.

			// Initial write + at least 1 ticker refresh
			mockKV.EXPECT().
				KVPut(gomock.Any(), "registry", "agents.test_machine_id", gomock.Any()).
				Return(uint64(1), nil).
				MinTimes(2)

			// Deregister on cancel
			mockKV.EXPECT().
				KVDelete(gomock.Any(), "registry", "agents.test_machine_id").
				Return(nil).
				Times(1)

			agent.SetHeartbeatInterval(10 * time.Millisecond)

			ctx, cancel := context.WithCancel(context.Background())
			agent.ExportStartHeartbeat(ctx, testAgent, "test-machine-id", tt.initialHostname)

			// Wait for at least one ticker refresh
			time.Sleep(50 * time.Millisecond)
			cancel()

			// Wait for goroutine to finish
			agent.WaitAgentWG(testAgent)

			tt.validateFunc(agent.GetAgentHostname(testAgent))
		})
	}
}

// TestStartHeartbeatRenamed registers with the new hostname as soon as a
// hostname update completes, not at the next tick.
func (s *HeartbeatLowLevelPublicTestSuite) TestStartHeartbeatRenamed() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	mockKV := agentMocks.NewMockNATSClient(ctrl)
	// The agent reads its previous registration and facts once, for
	// the timeline; there is none.
	mockKV.EXPECT().KVGet(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, jetstream.ErrKeyNotFound).AnyTimes()
	mockJobClient := mocks.NewMockJobClient(ctrl)
	mockJobClient.EXPECT().
		GetAgentConfig(gomock.Any(), gomock.Any()).
		Return(job.AgentConfig{}, nil).
		AnyTimes()
	// The heartbeat follows the rotation of the controller's key.
	mockJobClient.EXPECT().
		ControllerKey(gomock.Any()).
		Return(pki.ControllerKeyRecord{}, false, nil).
		AnyTimes()
	mockJobClient.EXPECT().WriteAgentSeen(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockJobClient.EXPECT().CheckDrainFlag(gomock.Any(), "test-machine-id").Return(false).AnyTimes()
	testAgent := newTestAgent(newTestAgentParams{
		jobClient:       mockJobClient,
		streamName:      "test-stream",
		hostProvider:    hostMocks.NewDefaultMockProvider(ctrl),
		diskProvider:    diskMocks.NewDefaultMockProvider(ctrl),
		memoryProvider:  memoryMocks.NewDefaultMockProvider(ctrl),
		loadProvider:    loadMocks.NewDefaultMockProvider(ctrl),
		processProvider: processMocks.NewDefaultMockProvider(ctrl),
		registryBucket:  "registry",
		natsClient:      mockKV,
	})
	agent.SetAgentState(testAgent, types.AgentReady)
	agent.SetAgentMachineID(testAgent, "test-machine-id")
	agent.SetAgentHostname(testAgent, "web-01")
	agent.SetGetAgentHostnameFn(func(_ string) (string, error) { return "db-01", nil })
	defer agent.ResetGetAgentHostnameFn()

	registered := make(chan struct{}, 2)
	mockKV.EXPECT().
		KVPut(gomock.Any(), "registry", "agents.test_machine_id", gomock.Any()).
		DoAndReturn(func(context.Context, string, string, []byte) (uint64, error) {
			registered <- struct{}{}

			return 1, nil
		}).
		Times(2)
	mockKV.EXPECT().KVDelete(gomock.Any(), "registry", "agents.test_machine_id").Return(nil)

	// A tick never comes in this test, so only the rename registers again.
	agent.SetHeartbeatInterval(time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	agent.ExportStartHeartbeat(ctx, testAgent, "test-machine-id", "web-01")
	<-registered
	agent.SignalRenamed(testAgent)
	<-registered
	cancel()
	agent.WaitAgentWG(testAgent)

	s.Equal("db-01", agent.GetAgentHostname(testAgent))
}

// TestNoteRename: only a hostname update tells the heartbeat, and a second
// one before the heartbeat reads the first is the same news.
func (s *HeartbeatLowLevelPublicTestSuite) TestNoteRename() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	testAgent := newTestAgent(newTestAgentParams{
		jobClient:       mocks.NewMockJobClient(ctrl),
		hostProvider:    hostMocks.NewDefaultMockProvider(ctrl),
		diskProvider:    diskMocks.NewDefaultMockProvider(ctrl),
		memoryProvider:  memoryMocks.NewDefaultMockProvider(ctrl),
		loadProvider:    loadMocks.NewDefaultMockProvider(ctrl),
		processProvider: processMocks.NewDefaultMockProvider(ctrl),
	})
	get := job.Request{Category: job.CategoryBehavior, Operation: "hostname.get"}
	update := job.Request{Category: job.CategoryBehavior, Operation: "hostname.update"}

	defer agent.SetRenameWait(time.Millisecond)()
	s.False(agent.ExportNoteRename(testAgent, get), "a read is not a rename")
	s.True(agent.ExportNoteRename(testAgent, update))
	s.True(agent.ExportNoteRename(testAgent, update), "a second rename waits with the first")
	agent.AnswerRename(testAgent)

	// The update waits for the registration, not for the whole wait.
	defer agent.SetRenameWait(time.Hour)()
	go agent.AnswerRename(testAgent)
	done := make(chan struct{})
	go func() {
		agent.NoteRename(testAgent, update)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		s.Fail("the update did not return when the agent registered")
	}
}

func (s *HeartbeatLowLevelPublicTestSuite) TestRegistryKey() {
	tests := []struct {
		name         string
		machineID    string
		validateFunc func(string)
	}{
		{
			name:      "simple machine ID",
			machineID: "abc-123-def",
			validateFunc: func(got string) {
				s.Equal("agents.abc_123_def", got)
			},
		},
		{
			name:      "machine ID with dots",
			machineID: "A1B2C3D4-E5F6.7890",
			validateFunc: func(got string) {
				s.Equal("agents.A1B2C3D4_E5F6_7890", got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := agent.ExportRegistryKey(tt.machineID)
			tt.validateFunc(result)
		})
	}
}

func TestHeartbeatLowLevelPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HeartbeatLowLevelPublicTestSuite))
}

// TestWriteRegistrationSigning covers what the controller relies on to tell
// this agent's registration from one written by anything else able to reach the
// registry bucket.
func (s *HeartbeatLowLevelPublicTestSuite) TestWriteRegistrationSigning() {
	tests := []struct {
		name         string
		withPKI      bool
		validateFunc func(job.AgentRegistration, *pki.Manager)
	}{
		{
			name:    "with PKI the registration is signed over its routing fields",
			withPKI: true,
			validateFunc: func(reg job.AgentRegistration, m *pki.Manager) {
				s.NotEmpty(reg.Signature)
				s.True(
					job.VerifyRegistration(&reg, m.PublicKey()),
					"the controller must be able to verify this with the stored key",
				)
			},
		},
		{
			name:    "a signature does not cover a hostname it did not claim",
			withPKI: true,
			validateFunc: func(reg job.AgentRegistration, m *pki.Manager) {
				// Whoever rewrites the hostname has to re-sign it, and cannot.
				reg.Hostname = "web-99"
				s.False(job.VerifyRegistration(&reg, m.PublicKey()))
			},
		},
		{
			name:    "without PKI the registration is unsigned",
			withPKI: false,
			validateFunc: func(reg job.AgentRegistration, _ *pki.Manager) {
				s.Empty(reg.Signature)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var m *pki.Manager

			if tt.withPKI {
				m = pki.New(memfs.New(), "/keys", "agent")
				s.Require().NoError(m.LoadOrGenerate())
				agent.SetAgentPKIManager(s.testAgent, m)

				defer agent.SetAgentPKIManager(s.testAgent, nil)
			}

			var written job.AgentRegistration

			s.mockKV.EXPECT().
				KVPut(gomock.Any(), "registry", "agents.test_machine_id", gomock.Any()).
				DoAndReturn(func(
					_ interface{},
					_ string,
					_ string,
					data []byte,
				) (uint64, error) {
					s.Require().NoError(json.Unmarshal(data, &written))

					return uint64(1), nil
				})

			agent.ExportWriteRegistration(
				context.Background(),
				s.testAgent,
				"test-machine-id",
				"test-agent",
			)

			tt.validateFunc(written, m)
		})
	}
}

// TestWriteSeen records when the agent last reported at most once each
// seenInterval, and tries again at the next heartbeat when a write fails.
func (s *HeartbeatLowLevelPublicTestSuite) TestWriteSeen() {
	s.Run("when the interval has passed writes, then not again until it passes", func() {
		ctrl := gomock.NewController(s.T())
		jobs := mocks.NewMockJobClient(ctrl)
		jobs.EXPECT().
			GetAgentConfig(gomock.Any(), gomock.Any()).
			Return(job.AgentConfig{}, nil).
			AnyTimes()
		// The heartbeat follows the rotation of the controller's key.
		jobs.EXPECT().
			ControllerKey(gomock.Any()).
			Return(pki.ControllerKeyRecord{}, false, nil).
			AnyTimes()
		a := newTestAgent(newTestAgentParams{jobClient: jobs})
		jobs.EXPECT().WriteAgentSeen(gomock.Any(), "m-1").Return(nil).Times(1)

		agent.ExportWriteSeen(context.Background(), a, "m-1")
		agent.ExportWriteSeen(context.Background(), a, "m-1")
	})

	s.Run("when the write fails tries again at the next heartbeat", func() {
		ctrl := gomock.NewController(s.T())
		jobs := mocks.NewMockJobClient(ctrl)
		jobs.EXPECT().
			GetAgentConfig(gomock.Any(), gomock.Any()).
			Return(job.AgentConfig{}, nil).
			AnyTimes()
		// The heartbeat follows the rotation of the controller's key.
		jobs.EXPECT().
			ControllerKey(gomock.Any()).
			Return(pki.ControllerKeyRecord{}, false, nil).
			AnyTimes()
		a := newTestAgent(newTestAgentParams{jobClient: jobs})
		gomock.InOrder(
			jobs.EXPECT().WriteAgentSeen(gomock.Any(), "m-1").Return(errors.New("kv down")),
			jobs.EXPECT().WriteAgentSeen(gomock.Any(), "m-1").Return(nil),
		)

		agent.ExportWriteSeen(context.Background(), a, "m-1")
		agent.ExportWriteSeen(context.Background(), a, "m-1")
	})
}
