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
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/controller/jointoken"
	"github.com/osapi-io/osapi/internal/pki"

	"github.com/osapi-io/osapi/internal/controller/endpoint"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/endpoints/agent"
	"github.com/osapi-io/osapi/internal/controller/endpoints/agent/mocks"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
)

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	jobs *mocks.MockJobs
	ctx  context.Context
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.jobs = mocks.NewMockJobs(s.ctrl)
	// List and Get read the controller's config of each agent.
	s.jobs.EXPECT().GetAgentConfig(gomock.Any(), gomock.Any()).
		Return(job.AgentConfig{}, nil).AnyTimes()
	s.ctx = context.Background()
}

func (s *MethodsPublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

func (s *MethodsPublicTestSuite) deps() agent.Deps {
	return agent.Deps{Jobs: s.jobs}
}

var errBoom = errors.New("nats: timeout")

// missing is the registry's error for an agent it does not have.
func missing() error {
	return fmt.Errorf("agent server9 not found: %w", job.ErrNotFound)
}

func (s *MethodsPublicTestSuite) TestFacts() {
	s.Run("each built-in fact key, with its description", func() {
		got, err := agent.Facts(s.ctx, s.deps(), agent.FactsInput{})

		s.Require().NoError(err)
		s.Equal(len(got.Items), got.TotalItems)
		s.NotEmpty(got.Items)
		for _, k := range got.Items {
			s.True(k.Builtin)
			s.NotEmpty(k.Description, k.Key)
		}
	})
}

func (s *MethodsPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		setup        func()
		validateFunc func(agent.ListAgentsResponse, error)
	}{
		{
			name: "when agents report, each is listed with the total",
			setup: func() {
				s.jobs.EXPECT().ListAgents(gomock.Any()).Return([]job.AgentInfo{
					{Hostname: "server1"}, {Hostname: "server2"},
				}, nil)
			},
			validateFunc: func(got agent.ListAgentsResponse, err error) {
				s.Require().NoError(err)
				s.Equal(2, got.TotalItems)
				s.Equal("server1", got.Items[0].Hostname)
				s.Equal(
					types.AgentReady,
					got.Items[1].State,
					"an agent that reported no state is Ready",
				)
			},
		},
		{
			name: "when no agents report, the list is empty, not null",
			setup: func() {
				s.jobs.EXPECT().ListAgents(gomock.Any()).Return(nil, nil)
			},
			validateFunc: func(got agent.ListAgentsResponse, err error) {
				s.Require().NoError(err)
				s.NotNil(got.Items)
				s.Zero(got.TotalItems)
			},
		},
		{
			name: "when the registry cannot be read, the error is returned",
			setup: func() {
				s.jobs.EXPECT().ListAgents(gomock.Any()).Return(nil, errBoom)
			},
			validateFunc: func(_ agent.ListAgentsResponse, err error) {
				s.ErrorIs(err, errBoom)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			tc.validateFunc(agent.List(s.ctx, s.deps(), agent.ListInput{}))
		})
	}
}

func (s *MethodsPublicTestSuite) TestBuildAgentInfo() {
	at := time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		info         job.AgentInfo
		validateFunc func(agent.AgentInfo)
	}{
		{
			name: "when the registry has every field, each one is in the answer",
			info: job.AgentInfo{
				Hostname: "server1", MachineID: "m-1", Fingerprint: "SHA256:x",
				AgentVersion: "v1.4.0", Labels: map[string]string{"role": "web"},
				RegisteredAt: at, StartedAt: at.Add(-time.Hour),
				Uptime:       26*time.Hour + 1*time.Minute,
				OSInfo:       &host.Result{Distribution: "debian", Version: "13", Family: "debian"},
				LoadAverages: &load.Result{Load1: 0.5, Load5: 0.25, Load15: 0.125},
				MemoryStats:  &memory.Result{Total: 8, Available: 5, Free: 2, Cached: 1},
				Architecture: "arm64", KernelVersion: "6.12", CPUCount: 4,
				FQDN: "server1.lab", ServiceMgr: "systemd", PackageMgr: "apt",
				Interfaces: []job.NetworkInterface{{
					Name: "eth0", IPv4: "10.0.0.2", IPv6: "fe80::1", MAC: "aa:bb", Family: "dual",
				}},
				PrimaryInterface: "eth0",
				Routes: []job.Route{{
					Destination: "0.0.0.0", Gateway: "10.0.0.1", Interface: "eth0",
					Mask: "/0", Metric: 100,
				}},
				Facts:           map[string]any{"rack": "a1"},
				MissingPrograms: []string{"chronyc"},
				State:           types.AgentDraining,
				Conditions: []job.Condition{{
					Type: "HighLoad", Status: true, Reason: "load", LastTransitionTime: at,
				}, {
					Type: "DiskPressure", Status: false, LastTransitionTime: at,
				}},
				Timeline: []job.TimelineEvent{{
					Timestamp: at, Event: "drain", Hostname: "server1", Message: "m", Error: "e",
				}},
				KeyStored: true,
				Verified:  true,
			},
			validateFunc: func(got agent.AgentInfo) {
				s.Equal("m-1", got.MachineID)
				s.Equal("v1.4.0", got.Version)
				s.Equal(map[string]string{"role": "web"}, got.Labels)
				s.Equal(at, got.RegisteredAt)
				s.Equal("1 day, 2 hours, 1 minute", got.Uptime)
				s.Equal(
					&agent.OSInfoResponse{Distribution: "debian", Version: "13", Family: "debian"},
					got.OSInfo,
				)
				s.Equal(&agent.LoadAverageResponse{OneMin: 0.5, FiveMin: 0.25, FifteenMin: 0.125},
					got.LoadAverage)
				s.Equal(&agent.MemoryResponse{Total: 8, Free: 2, Used: 3}, got.Memory,
					"used is the total less the available, not the cached")
				s.Equal(4, got.CPUCount)
				s.Equal([]agent.NetworkInterfaceResponse{{
					Name: "eth0", IPv4: "10.0.0.2", IPv6: "fe80::1", MAC: "aa:bb", Family: "dual",
				}}, got.Interfaces)
				s.Equal("eth0", got.PrimaryInterface)
				s.Equal(100, got.Routes[0].Metric)
				s.Equal(map[string]any{"rack": "a1"}, got.Facts)
				s.Equal([]string{"chronyc"}, got.MissingPrograms)
				s.Equal(types.AgentDraining, got.State)
				s.Equal(
					[]agent.NodeCondition{
						{Type: "HighLoad", Reason: "load", LastTransitionTime: at},
					},
					got.Conditions,
					"a condition that does not hold is left out",
				)
				s.Equal("e", got.Timeline[0].Error)
				s.True(*got.KeyStored)
				s.True(*got.Verified)
			},
		},
		{
			name: "when the registry has only the hostname, the rest is left out, but key_stored and verified are false",
			info: job.AgentInfo{Hostname: "server1"},
			validateFunc: func(got agent.AgentInfo) {
				s.Empty(got.MachineID)
				s.Nil(got.Labels)
				s.Nil(got.Facts)
				s.Nil(got.MissingPrograms)
				s.Empty(got.Uptime)
				s.Nil(got.OSInfo)
				s.Nil(got.LoadAverage)
				s.Nil(got.Memory)
				s.Nil(got.Interfaces)
				s.Nil(got.Routes)
				s.Nil(got.Conditions)
				s.Nil(got.Timeline)
				s.False(*got.KeyStored)
				s.False(*got.Verified)
			},
		},
		{
			name: "when memory is past the largest int, it is the largest int",
			info: job.AgentInfo{
				Hostname:    "server1",
				MemoryStats: &memory.Result{Total: ^uint64(0)},
			},
			validateFunc: func(got agent.AgentInfo) {
				s.Equal(int(^uint(0)>>1), got.Memory.Total)
			},
		},
		{
			name: "when the uptime is one of each, the words are singular",
			info: job.AgentInfo{Hostname: "server1", Uptime: 25*time.Hour + time.Minute},
			validateFunc: func(got agent.AgentInfo) {
				s.Equal("1 day, 1 hour, 1 minute", got.Uptime)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			info := tc.info
			s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(&info, nil)

			got, err := agent.Get(s.ctx, s.deps(), agent.GetInput{Hostname: "server1"})

			s.Require().NoError(err)
			tc.validateFunc(got)
		})
	}
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name    string
		err     error
		wantErr error
		wantMsg string
	}{
		{
			name:    "when the registry has no such agent, it is ErrNotFound",
			err:     missing(),
			wantErr: agent.ErrNotFound,
			wantMsg: "agent server9 not found: not found",
		},
		{
			name:    "when the registry cannot be read, it is that error, a 500",
			err:     errBoom,
			wantErr: errBoom,
			wantMsg: "nats: timeout",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.jobs.EXPECT().GetAgent(gomock.Any(), "server9").Return(nil, tc.err)

			_, err := agent.Get(s.ctx, s.deps(), agent.GetInput{Hostname: "server9"})

			s.ErrorIs(err, tc.wantErr)
			s.EqualError(err, tc.wantMsg)
		})
	}
}

// check holds an action's answer to the message, or its error to the
// sentinel and the text.
func (s *MethodsPublicTestSuite) check(
	err error,
	wantErr error,
	wantErrMsg string,
) {
	if wantErr == nil && wantErrMsg == "" {
		s.Require().NoError(err)

		return
	}
	if wantErr != nil {
		s.ErrorIs(err, wantErr)
	}
	if wantErrMsg != "" {
		s.EqualError(err, wantErrMsg)
	}
}

func (s *MethodsPublicTestSuite) TestDrainAndUndrain() {
	ready := &job.AgentInfo{Hostname: "server1", MachineID: "m-1", State: types.AgentReady}
	draining := &job.AgentInfo{Hostname: "server1", MachineID: "m-1", State: types.AgentDraining}
	cordoned := &job.AgentInfo{Hostname: "server1", MachineID: "m-1", State: types.AgentCordoned}

	type call func(context.Context, agent.Deps, string) (struct{}, error)
	drain := func(ctx context.Context, d agent.Deps, h string) (struct{}, error) {
		return agent.Drain(ctx, d, agent.DrainInput{Hostname: h})
	}
	undrain := func(ctx context.Context, d agent.Deps, h string) (struct{}, error) {
		return agent.Undrain(ctx, d, agent.UndrainInput{Hostname: h})
	}

	tests := []struct {
		name       string
		call       call
		setup      func()
		wantErr    error
		wantErrMsg string
	}{
		{
			name: "when a ready agent is drained, its flag and timeline are written",
			call: drain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(ready, nil)
				s.jobs.EXPECT().SetDrainFlag(gomock.Any(), "m-1").Return(nil)
				s.jobs.EXPECT().
					WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), "server1", "drain",
						"Drain initiated via API").
					Return(nil)
			},
		},
		{
			name: "when the agent is not in the registry, it is ErrNotFound",
			call: drain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(nil, missing())
			},
			wantErr:    agent.ErrNotFound,
			wantErrMsg: "agent not found: server1",
		},
		{
			name: "when the registry cannot be read, it is that error, a 500, not a 404 or a 409",
			call: drain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(nil, errBoom)
			},
			wantErr: errBoom,
		},
		{
			name: "when the agent is already draining, it is ErrState",
			call: drain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(draining, nil)
			},
			wantErr:    agent.ErrState,
			wantErrMsg: "agent server1 is already in Draining state",
		},
		{
			name: "when the agent is cordoned, draining it is ErrState",
			call: drain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(cordoned, nil)
			},
			wantErr: agent.ErrState,
		},
		{
			name: "when the drain flag cannot be written, it is that error, a 500",
			call: drain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(ready, nil)
				s.jobs.EXPECT().SetDrainFlag(gomock.Any(), "m-1").Return(errBoom)
			},
			wantErr:    errBoom,
			wantErrMsg: "failed to set drain flag: nats: timeout",
		},
		{
			name: "when the timeline cannot be written, it is that error, a 500",
			call: drain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(ready, nil)
				s.jobs.EXPECT().SetDrainFlag(gomock.Any(), "m-1").Return(nil)
				s.jobs.EXPECT().
					WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), "server1", "drain",
						"Drain initiated via API").
					Return(errBoom)
			},
			wantErr: errBoom,
		},
		{
			name: "when a draining agent is undrained, its flag goes and the timeline is written",
			call: undrain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(draining, nil)
				s.jobs.EXPECT().DeleteDrainFlag(gomock.Any(), "m-1").Return(nil)
				s.jobs.EXPECT().
					WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), "server1", "undrain",
						"Undrain initiated via API").
					Return(nil)
			},
		},
		{
			name: "when a cordoned agent is undrained, it is allowed",
			call: undrain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(cordoned, nil)
				s.jobs.EXPECT().DeleteDrainFlag(gomock.Any(), "m-1").Return(nil)
				s.jobs.EXPECT().
					WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), "server1", "undrain",
						"Undrain initiated via API").
					Return(nil)
			},
		},
		{
			name: "when undrain finds no agent, it is ErrNotFound",
			call: undrain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(nil, missing())
			},
			wantErr: agent.ErrNotFound,
		},
		{
			name: "when a ready agent is undrained, it is ErrState",
			call: undrain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(ready, nil)
			},
			wantErr:    agent.ErrState,
			wantErrMsg: "agent server1 is not in draining or cordoned state (current: Ready)",
		},
		{
			name: "when the drain flag cannot be deleted, it is that error, a 500",
			call: undrain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(draining, nil)
				s.jobs.EXPECT().DeleteDrainFlag(gomock.Any(), "m-1").Return(errBoom)
			},
			wantErr:    errBoom,
			wantErrMsg: "failed to delete drain flag: nats: timeout",
		},
		{
			name: "when the undrain timeline cannot be written, it is that error, a 500",
			call: undrain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(draining, nil)
				s.jobs.EXPECT().DeleteDrainFlag(gomock.Any(), "m-1").Return(nil)
				s.jobs.EXPECT().
					WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), "server1", "undrain",
						"Undrain initiated via API").
					Return(errBoom)
			},
			wantErr: errBoom,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			_, err := tc.call(s.ctx, s.deps(), "server1")

			s.check(err, tc.wantErr, tc.wantErrMsg)
		})
	}
}

// TestSpecEnums: the spec's AgentState and ConditionType have each value
// of their type, so a client checks a value when it compiles.
func (s *MethodsPublicTestSuite) TestJoin() {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	in := agent.JoinInput{Body: agent.JoinRequest{
		Token:         "osapi_jt_t1.secret",
		MachineID:     "m-1",
		Hostname:      "web-01",
		PublicKey:     base64.StdEncoding.EncodeToString(pub),
		EncryptionKey: "XKEY",
		Signature:     base64.StdEncoding.EncodeToString(ed25519.Sign(priv, []byte("x"))),
	}}

	s.Run("when the token admits, the answer is the controller's key", func() {
		joiner := mocks.NewMockJoiner(s.ctrl)
		ctrlKey := make(ed25519.PublicKey, ed25519.PublicKeySize)
		joiner.EXPECT().
			Join(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req pki.JoinRequest) (ed25519.PublicKey, error) {
				s.Equal("osapi_jt_t1.secret", req.Token)
				s.Equal(pub, req.PublicKey, "the key, decoded from base64")
				s.Equal("XKEY", req.EncryptionKey)

				return ctrlKey, nil
			})

		got, err := agent.Join(s.ctx, agent.Deps{Joiner: joiner}, in)

		s.Require().NoError(err)
		s.Equal(base64.StdEncoding.EncodeToString(ctrlKey), got.ControllerPublicKey)
	})

	s.Run("when the token does not admit, it is the error", func() {
		joiner := mocks.NewMockJoiner(s.ctrl)
		joiner.EXPECT().Join(gomock.Any(), gomock.Any()).Return(nil, jointoken.ErrExpired)

		_, err := agent.Join(s.ctx, agent.Deps{Joiner: joiner}, in)

		s.ErrorIs(err, jointoken.ErrExpired)
	})

	s.Run("when the controller takes no join, it says so", func() {
		_, err := agent.Join(s.ctx, agent.Deps{}, in)

		s.ErrorIs(err, admission.ErrNoJoinTokens)
	})
}

func (s *MethodsPublicTestSuite) TestSpecEnums() {
	doc, err := endpoint.OpenAPI(agent.Group)
	s.Require().NoError(err)

	states := make([]any, 0, len(types.AgentState("").Values()))
	for _, v := range types.AgentState("").Values() {
		states = append(states, string(v))
	}
	s.Equal(states, doc.Components.Schemas["AgentState"].Value.Enum)
	conditions := make([]any, 0, len(types.ConditionType("").Values()))
	for _, v := range types.ConditionType("").Values() {
		conditions = append(conditions, string(v))
	}
	s.Equal(conditions, doc.Components.Schemas["ConditionType"].Value.Enum)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
