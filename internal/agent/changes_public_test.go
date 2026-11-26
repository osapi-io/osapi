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

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	agentMocks "github.com/osapi-io/osapi/internal/agent/mocks"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// events records the timeline events that the agent writes.
func (s *AgentPublicTestSuite) events() *[]string {
	var got []string
	s.mockJobClient.EXPECT().
		WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _, event, msg string) error {
			got = append(got, event+": "+msg)
			return nil
		}).
		AnyTimes()

	return &got
}

// previous is a NATS client whose bucket holds v under each key, or none.
func (s *AgentPublicTestSuite) previous(v any) *agentMocks.MockNATSClient {
	nc := agentMocks.NewMockNATSClient(s.mockCtrl)
	if v == nil {
		nc.EXPECT().KVGet(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, jetstream.ErrKeyNotFound).AnyTimes()

		return nc
	}
	data, err := json.Marshal(v)
	s.Require().NoError(err)
	entry := mocks.NewMockKeyValueEntry(s.mockCtrl)
	entry.EXPECT().Value().Return(data).AnyTimes()
	nc.EXPECT().KVGet(gomock.Any(), gomock.Any(), gomock.Any()).Return(entry, nil).AnyTimes()

	return nc
}

func reg(os string, labels map[string]string, conds ...job.Condition) job.AgentRegistration {
	return job.AgentRegistration{
		MachineID: "m-1", Labels: labels, Conditions: conds,
		OSInfo: &host.Result{Distribution: "ubuntu", Version: os},
	}
}

func (s *AgentPublicTestSuite) TestNoteRegistration() {
	memory := job.Condition{Type: types.ConditionMemoryPressure, Status: true, Reason: "92% used"}
	tests := []struct {
		name     string
		previous any
		version  int
		regs     []job.AgentRegistration
		want     []string
	}{
		{
			name: "the first report with none before is the baseline",
			regs: []job.AgentRegistration{reg("24.04", nil, memory)},
		},
		{
			name:     "an OS upgrade against the report before the start",
			previous: reg("22.04", nil),
			regs:     []job.AgentRegistration{reg("24.04", nil)},
			want:     []string{"os_changed: ubuntu 22.04 → ubuntu 24.04"},
		},
		{
			name: "new labels from the agent's own config",
			regs: []job.AgentRegistration{
				reg("24.04", map[string]string{"group": "db"}),
				reg("24.04", map[string]string{"group": "web"}),
			},
			want: []string{"labels_changed: group=db → group=web"},
		},
		{
			name:     "new labels from the controller's config are config_applied's",
			version:  2,
			previous: reg("24.04", map[string]string{"group": "db"}),
			regs:     []job.AgentRegistration{reg("24.04", map[string]string{"group": "web"})},
		},
		{
			name:     "a condition starts and clears",
			previous: reg("24.04", nil),
			regs: []job.AgentRegistration{
				reg("24.04", nil, memory),
				reg("24.04", nil, job.Condition{Type: types.ConditionMemoryPressure}),
			},
			want: []string{
				"condition_started: MemoryPressure: 92% used",
				"condition_cleared: MemoryPressure",
			},
		},
		{
			name:     "a report with no OS keeps the OS",
			previous: reg("24.04", nil),
			regs:     []job.AgentRegistration{{MachineID: "m-1"}, reg("24.04", nil)},
		},
		{
			name:     "a previous report that is not JSON is no baseline",
			previous: "not a registration",
			regs:     []job.AgentRegistration{reg("24.04", nil, memory)},
		},
	}
	got := s.events()
	for _, tt := range tests {
		s.Run(tt.name, func() {
			*got = nil
			a := s.buildAgent()
			nc := s.previous(tt.previous)
			for _, r := range tt.regs {
				agent.ExportNoteRegistration(a, nc, "registry", tt.version, r)
			}
			s.Equal(tt.want, *got)
		})
	}

	s.Run("an agent with no registry bucket starts from its first report", func() {
		*got = nil
		a := s.buildAgent()
		agent.ExportNoteRegistration(a, nil, "", 0, reg("22.04", nil))
		agent.ExportNoteRegistration(a, nil, "", 0, reg("24.04", nil))
		s.Equal([]string{"os_changed: ubuntu 22.04 → ubuntu 24.04"}, *got)
	})
}

func facts(iface, ip string) job.FactsRegistration {
	return job.FactsRegistration{
		PrimaryInterface: iface,
		Interfaces: []job.NetworkInterface{
			{Name: "lo", IPv4: "127.0.0.1"},
			{Name: iface, IPv4: ip},
		},
	}
}

func (s *AgentPublicTestSuite) TestNoteFacts() {
	tests := []struct {
		name     string
		previous any
		bucket   string
		facts    []job.FactsRegistration
		want     []string
	}{
		{
			name:     "a new address against the facts before the start",
			previous: facts("eth0", "10.0.0.5"),
			bucket:   "facts",
			facts:    []job.FactsRegistration{facts("eth0", "10.0.0.9")},
			want:     []string{"address_changed: eth0 10.0.0.5 → eth0 10.0.0.9"},
		},
		{
			name:   "the same address writes nothing",
			bucket: "facts",
			facts:  []job.FactsRegistration{facts("eth0", "10.0.0.5"), facts("eth0", "10.0.0.5")},
		},
		{
			name:   "facts with no primary address keep the address",
			bucket: "facts",
			facts: []job.FactsRegistration{
				facts("eth0", "10.0.0.5"), {PrimaryInterface: "eth0"}, facts("eth0", "10.0.0.5"),
			},
		},
		{
			name:     "previous facts that are not JSON are no baseline",
			previous: "not facts",
			bucket:   "facts",
			facts:    []job.FactsRegistration{facts("eth0", "10.0.0.5")},
		},
		{
			name:  "an agent with no facts bucket starts from its first facts",
			facts: []job.FactsRegistration{facts("eth0", "10.0.0.5"), facts("eth1", "10.0.0.6")},
			want:  []string{"address_changed: eth0 10.0.0.5 → eth1 10.0.0.6"},
		},
	}
	got := s.events()
	for _, tt := range tests {
		s.Run(tt.name, func() {
			*got = nil
			a := s.buildAgent()
			nc := s.previous(tt.previous)
			for _, f := range tt.facts {
				agent.ExportNoteFacts(a, nc, tt.bucket, "m-1", f)
			}
			s.Equal(tt.want, *got)
		})
	}
}
