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
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/endpoints/agent"
	"github.com/osapi-io/osapi/internal/controller/endpoints/agent/mocks"
	"github.com/osapi-io/osapi/internal/controller/enrollment"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
)

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl   *gomock.Controller
	jobs   *mocks.MockJobs
	enroll *mocks.MockEnrollment
	ctx    context.Context
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.jobs = mocks.NewMockJobs(s.ctrl)
	s.enroll = mocks.NewMockEnrollment(s.ctrl)
	s.ctx = context.Background()
}

func (s *MethodsPublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

func (s *MethodsPublicTestSuite) deps() agent.Deps {
	return agent.Deps{Jobs: s.jobs, Enrollment: s.enroll}
}

var errBoom = errors.New("nats: timeout")

// missing is the registry's error for an agent it does not have.
func missing() error {
	return fmt.Errorf("agent server9 not found: %w", job.ErrNotFound)
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
				s.Equal(2, got.Total)
				s.Equal("server1", got.Agents[0].Hostname)
				s.Equal("Ready", got.Agents[1].Status)
			},
		},
		{
			name: "when no agents report, the list is empty, not null",
			setup: func() {
				s.jobs.EXPECT().ListAgents(gomock.Any()).Return(nil, nil)
			},
			validateFunc: func(got agent.ListAgentsResponse, err error) {
				s.Require().NoError(err)
				s.NotNil(got.Agents)
				s.Zero(got.Total)
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
				OSInfo:       &host.Result{Distribution: "debian", Version: "13"},
				LoadAverages: &load.Result{Load1: 0.5, Load5: 0.25, Load15: 0.125},
				MemoryStats:  &memory.Result{Total: 8, Free: 2, Cached: 3},
				Architecture: "arm64", KernelVersion: "6.12", CPUCount: 4,
				FQDN: "server1.lab", ServiceMgr: "systemd", PackageMgr: "apt",
				Interfaces: []job.NetworkInterface{{
					Name: "eth0", IPv4: "10.0.0.2", IPv6: "fe80::1", MAC: "aa:bb", Family: "dual",
				}},
				PrimaryInterface: "eth0",
				Routes: []job.Route{{
					Destination: "0.0.0.0", Gateway: "10.0.0.1", Interface: "eth0",
					Mask: "/0", Metric: 100, Flags: "UG",
				}},
				Facts: map[string]any{"rack": "a1"},
				State: job.AgentStateDraining,
				Conditions: []job.Condition{{
					Type: "HighLoad", Status: true, Reason: "load", LastTransitionTime: at,
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
				s.Equal(&agent.OSInfoResponse{Distribution: "debian", Version: "13"}, got.OSInfo)
				s.Equal(&agent.LoadAverageResponse{OneMin: 0.5, FiveMin: 0.25, FifteenMin: 0.125},
					got.LoadAverage)
				s.Equal(&agent.MemoryResponse{Total: 8, Free: 2, Used: 3}, got.Memory)
				s.Equal(4, got.CPUCount)
				s.Equal([]agent.NetworkInterfaceResponse{{
					Name: "eth0", IPv4: "10.0.0.2", IPv6: "fe80::1", MAC: "aa:bb", Family: "dual",
				}}, got.Interfaces)
				s.Equal("eth0", got.PrimaryInterface)
				s.Equal(100, got.Routes[0].Metric)
				s.Equal(map[string]any{"rack": "a1"}, got.Facts)
				s.Equal("Draining", got.State)
				s.Equal("load", got.Conditions[0].Reason)
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

func (s *MethodsPublicTestSuite) TestPending() {
	at := time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		noEnrollment bool
		setup        func()
		validateFunc func(agent.ListPendingAgentsResponse, error)
	}{
		{
			name:         "when enrollment is not enabled, it is an error",
			noEnrollment: true,
			setup:        func() {},
			validateFunc: func(_ agent.ListPendingAgentsResponse, err error) {
				s.EqualError(err, "enrollment not enabled")
			},
		},
		{
			name: "when agents wait, each is listed with the total",
			setup: func() {
				s.enroll.EXPECT().ListPending(gomock.Any()).Return([]enrollment.PendingAgent{{
					MachineID: "m-1", Hostname: "web-01", Fingerprint: "SHA256:x", RequestedAt: at,
				}}, nil)
			},
			validateFunc: func(got agent.ListPendingAgentsResponse, err error) {
				s.Require().NoError(err)
				s.Equal(agent.ListPendingAgentsResponse{
					Agents: []agent.PendingAgentInfo{{
						MachineID: "m-1", Hostname: "web-01", Fingerprint: "SHA256:x", RequestedAt: at,
					}},
					Total: 1,
				}, got)
			},
		},
		{
			name: "when none wait, the list is empty, not null",
			setup: func() {
				s.enroll.EXPECT().ListPending(gomock.Any()).Return(nil, nil)
			},
			validateFunc: func(got agent.ListPendingAgentsResponse, err error) {
				s.Require().NoError(err)
				s.NotNil(got.Agents)
			},
		},
		{
			name: "when the list fails, the error is returned",
			setup: func() {
				s.enroll.EXPECT().ListPending(gomock.Any()).Return(nil, errBoom)
			},
			validateFunc: func(_ agent.ListPendingAgentsResponse, err error) {
				s.ErrorIs(err, errBoom)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			d := s.deps()
			if tc.noEnrollment {
				d.Enrollment = nil
			}
			tc.validateFunc(agent.Pending(s.ctx, d, agent.ListInput{}))
		})
	}
}

func (s *MethodsPublicTestSuite) TestRemove() {
	tests := []struct {
		name         string
		noEnrollment bool
		in           agent.RemoveInput
		setup        func()
		wantMsg      string
		wantErr      error
		wantErrMsg   string
	}{
		{
			name:         "when enrollment is not enabled, it is an error",
			noEnrollment: true,
			in:           agent.RemoveInput{Hostname: "web-01"},
			setup:        func() {},
			wantErrMsg:   "enrollment not enabled",
		},
		{
			name: "when the agent is removed, the message names it and its machine",
			in:   agent.RemoveInput{Hostname: "web-01"},
			setup: func() {
				s.enroll.EXPECT().RemoveByHostname(gomock.Any(), "web-01", "").Return("m-1", nil)
			},
			wantMsg: "agent web-01 (m-1) removed",
		},
		{
			name: "when no accepted agent has the hostname, it is ErrNotFound",
			in:   agent.RemoveInput{Hostname: "web-01"},
			setup: func() {
				s.enroll.EXPECT().RemoveByHostname(gomock.Any(), "web-01", "").
					Return("", enrollment.ErrNoAcceptedAgent)
			},
			wantErr:    agent.ErrNotFound,
			wantErrMsg: `no accepted agent with hostname "web-01"`,
		},
		{
			name: "when a machine ID was named and not found, the message names it",
			in:   agent.RemoveInput{Hostname: "web-01", MachineID: "m-9"},
			setup: func() {
				s.enroll.EXPECT().RemoveByHostname(gomock.Any(), "web-01", "m-9").
					Return("", enrollment.ErrNoAcceptedAgent)
			},
			wantErr:    agent.ErrNotFound,
			wantErrMsg: `no accepted agent with hostname "web-01" and machine ID "m-9"`,
		},
		{
			name: "when two accepted agents have the hostname, it is ErrAmbiguous",
			in:   agent.RemoveInput{Hostname: "web-01"},
			setup: func() {
				s.enroll.EXPECT().RemoveByHostname(gomock.Any(), "web-01", "").
					Return("", enrollment.ErrAmbiguousHostname)
			},
			wantErr:    agent.ErrAmbiguous,
			wantErrMsg: enrollment.ErrAmbiguousHostname.Error() + "; name one with machine_id",
		},
		{
			name: "when the store fails, it is that error",
			in:   agent.RemoveInput{Hostname: "web-01"},
			setup: func() {
				s.enroll.EXPECT().RemoveByHostname(gomock.Any(), "web-01", "").Return("", errBoom)
			},
			wantErr: errBoom,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			d := s.deps()
			if tc.noEnrollment {
				d.Enrollment = nil
			}

			got, err := agent.Remove(s.ctx, d, tc.in)

			s.check(got, err, tc.wantMsg, tc.wantErr, tc.wantErrMsg)
		})
	}
}

// check holds an action's answer to the message, or its error to the
// sentinel and the text.
func (s *MethodsPublicTestSuite) check(
	got agent.Message,
	err error,
	wantMsg string,
	wantErr error,
	wantErrMsg string,
) {
	if wantErr == nil && wantErrMsg == "" {
		s.Require().NoError(err)
		s.Equal(wantMsg, got.Message)

		return
	}
	if wantErr != nil {
		s.ErrorIs(err, wantErr)
	}
	if wantErrMsg != "" {
		s.EqualError(err, wantErrMsg)
	}
}

func (s *MethodsPublicTestSuite) TestAccept() {
	tests := []struct {
		name         string
		noEnrollment bool
		in           agent.AcceptInput
		setup        func()
		wantMsg      string
		wantErr      error
		wantErrMsg   string
	}{
		{
			name:         "when enrollment is not enabled, it is an error",
			noEnrollment: true,
			in:           agent.AcceptInput{Hostname: "web-01"},
			setup:        func() {},
			wantErrMsg:   "enrollment not enabled",
		},
		{
			name: "when accepted by hostname, the message names it",
			in:   agent.AcceptInput{Hostname: "web-01"},
			setup: func() {
				s.enroll.EXPECT().AcceptByHostname(gomock.Any(), "web-01").Return(nil)
			},
			wantMsg: "agent web-01 accepted",
		},
		{
			name: "when a fingerprint is given, it accepts by fingerprint",
			in:   agent.AcceptInput{Hostname: "web-01", Fingerprint: "SHA256:x"},
			setup: func() {
				s.enroll.EXPECT().AcceptByFingerprint(gomock.Any(), "SHA256:x").Return(nil)
			},
			wantMsg: "agent with fingerprint SHA256:x accepted",
		},
		{
			name: "when no pending agent has the hostname, it is ErrNotFound",
			in:   agent.AcceptInput{Hostname: "web-01"},
			setup: func() {
				s.enroll.EXPECT().AcceptByHostname(gomock.Any(), "web-01").
					Return(enrollment.ErrNoPendingAgent)
			},
			wantErr:    agent.ErrNotFound,
			wantErrMsg: `no pending agent with hostname "web-01"`,
		},
		{
			name: "when no pending agent has the fingerprint, it is ErrNotFound",
			in:   agent.AcceptInput{Hostname: "web-01", Fingerprint: "SHA256:x"},
			setup: func() {
				s.enroll.EXPECT().AcceptByFingerprint(gomock.Any(), "SHA256:x").
					Return(enrollment.ErrNoPendingAgent)
			},
			wantErr:    agent.ErrNotFound,
			wantErrMsg: `no pending agent with fingerprint "SHA256:x"`,
		},
		{
			name: "when accepting by hostname fails, it is that error",
			in:   agent.AcceptInput{Hostname: "web-01"},
			setup: func() {
				s.enroll.EXPECT().AcceptByHostname(gomock.Any(), "web-01").Return(errBoom)
			},
			wantErr: errBoom,
		},
		{
			name: "when accepting by fingerprint fails, it is that error",
			in:   agent.AcceptInput{Hostname: "web-01", Fingerprint: "SHA256:x"},
			setup: func() {
				s.enroll.EXPECT().AcceptByFingerprint(gomock.Any(), "SHA256:x").Return(errBoom)
			},
			wantErr: errBoom,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			d := s.deps()
			if tc.noEnrollment {
				d.Enrollment = nil
			}

			got, err := agent.Accept(s.ctx, d, tc.in)

			s.check(got, err, tc.wantMsg, tc.wantErr, tc.wantErrMsg)
		})
	}
}

func (s *MethodsPublicTestSuite) TestReject() {
	tests := []struct {
		name         string
		noEnrollment bool
		setup        func()
		wantMsg      string
		wantErr      error
		wantErrMsg   string
	}{
		{
			name:         "when enrollment is not enabled, it is an error",
			noEnrollment: true,
			setup:        func() {},
			wantErrMsg:   "enrollment not enabled",
		},
		{
			name: "when rejected, the message names it",
			setup: func() {
				s.enroll.EXPECT().
					RejectByHostname(gomock.Any(), "web-01", "rejected via API").
					Return(nil)
			},
			wantMsg: "agent web-01 rejected",
		},
		{
			name: "when no pending agent has the hostname, it is ErrNotFound",
			setup: func() {
				s.enroll.EXPECT().RejectByHostname(gomock.Any(), "web-01", "rejected via API").
					Return(enrollment.ErrNoPendingAgent)
			},
			wantErr:    agent.ErrNotFound,
			wantErrMsg: `no pending agent with hostname "web-01"`,
		},
		{
			name: "when the store fails, it is that error",
			setup: func() {
				s.enroll.EXPECT().RejectByHostname(gomock.Any(), "web-01", "rejected via API").
					Return(errBoom)
			},
			wantErr: errBoom,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			d := s.deps()
			if tc.noEnrollment {
				d.Enrollment = nil
			}

			got, err := agent.Reject(s.ctx, d, agent.RejectInput{Hostname: "web-01"})

			s.check(got, err, tc.wantMsg, tc.wantErr, tc.wantErrMsg)
		})
	}
}

func (s *MethodsPublicTestSuite) TestDrainAndUndrain() {
	ready := &job.AgentInfo{Hostname: "server1", MachineID: "m-1", State: job.AgentStateReady}
	draining := &job.AgentInfo{Hostname: "server1", MachineID: "m-1", State: job.AgentStateDraining}
	cordoned := &job.AgentInfo{Hostname: "server1", MachineID: "m-1", State: job.AgentStateCordoned}

	type call func(context.Context, agent.Deps, string) (agent.Message, error)
	drain := func(ctx context.Context, d agent.Deps, h string) (agent.Message, error) {
		return agent.Drain(ctx, d, agent.DrainInput{Hostname: h})
	}
	undrain := func(ctx context.Context, d agent.Deps, h string) (agent.Message, error) {
		return agent.Undrain(ctx, d, agent.UndrainInput{Hostname: h})
	}

	tests := []struct {
		name       string
		call       call
		setup      func()
		wantMsg    string
		wantErr    error
		wantErrMsg string
	}{
		{
			name: "when a ready agent is drained, its flag and timeline are written",
			call: drain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(ready, nil)
				s.jobs.EXPECT().SetDrainFlag(gomock.Any(), "m-1").Return(nil)
				s.jobs.EXPECT().WriteAgentTimelineEvent(gomock.Any(), "server1", "drain",
					"Drain initiated via API").Return(nil)
			},
			wantMsg: "drain initiated for agent server1",
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
				s.jobs.EXPECT().WriteAgentTimelineEvent(gomock.Any(), "server1", "drain",
					"Drain initiated via API").Return(errBoom)
			},
			wantErr: errBoom,
		},
		{
			name: "when a draining agent is undrained, its flag goes and the timeline is written",
			call: undrain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(draining, nil)
				s.jobs.EXPECT().DeleteDrainFlag(gomock.Any(), "m-1").Return(nil)
				s.jobs.EXPECT().WriteAgentTimelineEvent(gomock.Any(), "server1", "undrain",
					"Undrain initiated via API").Return(nil)
			},
			wantMsg: "undrain initiated for agent server1",
		},
		{
			name: "when a cordoned agent is undrained, it is allowed",
			call: undrain,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").Return(cordoned, nil)
				s.jobs.EXPECT().DeleteDrainFlag(gomock.Any(), "m-1").Return(nil)
				s.jobs.EXPECT().WriteAgentTimelineEvent(gomock.Any(), "server1", "undrain",
					"Undrain initiated via API").Return(nil)
			},
			wantMsg: "undrain initiated for agent server1",
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
				s.jobs.EXPECT().WriteAgentTimelineEvent(gomock.Any(), "server1", "undrain",
					"Undrain initiated via API").Return(errBoom)
			},
			wantErr: errBoom,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			got, err := tc.call(s.ctx, s.deps(), "server1")

			s.check(got, err, tc.wantMsg, tc.wantErr, tc.wantErrMsg)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
