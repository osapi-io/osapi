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

package controller_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/controller/health"
	"github.com/osapi-io/osapi/internal/controller/mocks"
	"github.com/osapi-io/osapi/internal/job"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type MetricsPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	nc   *mocks.MockNATSStatus
	jobs *jobmocks.MockJobClient
}

func TestMetricsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MetricsPublicTestSuite))
}

func (s *MetricsPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.nc = mocks.NewMockNATSStatus(s.ctrl)
	s.jobs = jobmocks.NewMockJobClient(s.ctrl)
}

var errBoom = errors.New("boom")

func (s *MetricsPublicTestSuite) provider(
	registryBucket string,
) *health.ClosureMetricsProvider {
	return controller.ExportNewMetricsProvider(
		s.nc,
		"test-JOBS",
		[]string{"kv-missing", "kv-broken", "kv-ok"},
		[]string{"obj-missing", "obj-broken", "obj-ok"},
		s.jobs,
		registryBucket,
	)
}

func (s *MetricsPublicTestSuite) TestNATSInfo() {
	tests := []struct {
		name    string
		url     string
		want    *health.NATSMetrics
		wantErr string
	}{
		{
			name: "when connected returns the URL and the version",
			url:  "nats://127.0.0.1:4222",
			want: &health.NATSMetrics{URL: "nats://127.0.0.1:4222", Version: "2.11.0"},
		},
		{
			name:    "when not connected returns the error",
			wantErr: "NATS client unavailable",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.nc.EXPECT().ConnectedURL().Return(tc.url)
			if tc.url != "" {
				s.nc.EXPECT().ConnectedServerVersion().Return("2.11.0")
			}

			got, err := s.provider("").NATSInfoFn(context.Background())

			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)
				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *MetricsPublicTestSuite) TestStreamInfo() {
	tests := []struct {
		name    string
		info    *jetstream.StreamInfo
		err     error
		want    []health.StreamMetrics
		wantErr string
	}{
		{
			name: "when the stream answers returns its state",
			info: &jetstream.StreamInfo{
				State: jetstream.StreamState{Msgs: 3, Bytes: 120, Consumers: 1},
			},
			want: []health.StreamMetrics{
				{Name: "test-JOBS", Messages: 3, Bytes: 120, Consumers: 1},
			},
		},
		{
			name:    "when the stream fails returns the error",
			err:     errBoom,
			wantErr: "stream info: boom",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.nc.EXPECT().StreamInfo(gomock.Any(), "test-JOBS").Return(tc.info, tc.err)

			got, err := s.provider("").StreamInfoFn(context.Background())

			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)
				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *MetricsPublicTestSuite) TestKVInfo() {
	s.Run("when a bucket fails leaves it out", func() {
		status := mocks.NewMockKeyValueStatus(s.ctrl)
		status.EXPECT().Bucket().Return("kv-ok")
		status.EXPECT().Values().Return(uint64(4))
		status.EXPECT().Bytes().Return(uint64(256))

		s.nc.EXPECT().KVStatus(gomock.Any(), "kv-missing").Return(nil, errBoom)
		s.nc.EXPECT().KVStatus(gomock.Any(), "kv-broken").Return(nil, errBoom)
		s.nc.EXPECT().KVStatus(gomock.Any(), "kv-ok").Return(status, nil)

		got, err := s.provider("").KVInfoFn(context.Background())

		s.Require().NoError(err)
		s.Equal([]health.KVMetrics{{Name: "kv-ok", Keys: 4, Bytes: 256}}, got)
	})
}

func (s *MetricsPublicTestSuite) TestObjectStoreInfo() {
	s.Run("when a bucket fails leaves it out", func() {
		status := mocks.NewMockObjectStoreStatus(s.ctrl)
		status.EXPECT().Bucket().Return("obj-ok")
		status.EXPECT().Size().Return(uint64(1024))

		s.nc.EXPECT().ObjectStoreStatus(gomock.Any(), "obj-missing").Return(nil, errBoom)
		s.nc.EXPECT().ObjectStoreStatus(gomock.Any(), "obj-broken").Return(nil, errBoom)
		s.nc.EXPECT().ObjectStoreStatus(gomock.Any(), "obj-ok").Return(status, nil)

		got, err := s.provider("").ObjectStoreInfoFn(context.Background())

		s.Require().NoError(err)
		s.Equal([]health.ObjectStoreMetrics{{Name: "obj-ok", Size: 1024}}, got)
	})
}

func (s *MetricsPublicTestSuite) TestJobStats() {
	tests := []struct {
		name    string
		stats   *job.QueueStats
		err     error
		want    *health.JobMetrics
		wantErr string
	}{
		{
			name: "when the queue answers returns the counts",
			stats: &job.QueueStats{
				TotalJobs: 10,
				StatusCounts: map[types.JobStatus]int{
					"submitted":  1,
					"processing": 2,
					"completed":  5,
					"failed":     2,
				},
				DLQCount: 1,
			},
			want: &health.JobMetrics{
				Total:      10,
				Submitted:  1,
				Processing: 2,
				Completed:  5,
				Failed:     2,
				DLQ:        1,
			},
		},
		{
			name:    "when the queue fails returns the error",
			err:     errBoom,
			wantErr: "job stats: boom",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.jobs.EXPECT().GetQueueSummary(gomock.Any()).Return(tc.stats, tc.err)

			got, err := s.provider("").JobStatsFn(context.Background())

			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)
				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *MetricsPublicTestSuite) TestAgentStats() {
	tests := []struct {
		name       string
		agents     []job.AgentInfo
		err        error
		wantTotal  int
		wantLabels string
		wantErr    string
	}{
		{
			name: "when agents answer returns each one with its labels",
			agents: []job.AgentInfo{
				{
					Hostname:     "web-01",
					Labels:       map[string]string{"group": "web"},
					RegisteredAt: time.Now(),
				},
			},
			wantTotal:  1,
			wantLabels: "group=web",
		},
		{
			name:    "when the agents fail returns the error",
			err:     errBoom,
			wantErr: "agent stats: boom",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.jobs.EXPECT().ListAgents(gomock.Any()).Return(tc.agents, tc.err)

			got, err := s.provider("").AgentStatsFn(context.Background())

			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)
				return
			}
			s.Require().NoError(err)
			s.Equal(tc.wantTotal, got.Total)
			s.Equal(tc.wantTotal, got.Ready)
			s.Equal(tc.wantLabels, got.Agents[0].Labels)
			s.Equal("web-01", got.Agents[0].Hostname)
		})
	}

	s.Run("when an agent is drained counts it as registered, not ready", func() {
		s.jobs.EXPECT().ListAgents(gomock.Any()).Return([]job.AgentInfo{
			{Hostname: "web-01", State: string(job.AgentStateReady)},
			{Hostname: "web-02", State: string(job.AgentStateCordoned)},
			{Hostname: "web-03", State: string(job.AgentStateDraining)},
			{Hostname: "web-04"},
		}, nil)

		got, err := s.provider("").AgentStatsFn(context.Background())

		s.Require().NoError(err)
		s.Equal(4, got.Total)
		s.Equal(2, got.Ready, "Ready and no state are ready; Cordoned and Draining are not")
	})

	s.Run("when an agent has two labels joins them", func() {
		s.jobs.EXPECT().ListAgents(gomock.Any()).Return([]job.AgentInfo{{
			Hostname: "db-01",
			Labels:   map[string]string{"a": "1", "b": "2"},
		}}, nil)

		got, err := s.provider("").AgentStatsFn(context.Background())

		s.Require().NoError(err)
		s.Contains([]string{"a=1, b=2", "b=2, a=1"}, got.Agents[0].Labels)
	})
}

// entry returns a registry entry whose value is v as JSON, or raw when v is
// a string.
func (s *MetricsPublicTestSuite) entry(
	v any,
) jetstream.KeyValueEntry {
	data, ok := v.(string)
	if !ok {
		b, err := json.Marshal(v)
		s.Require().NoError(err)
		data = string(b)
	}
	e := jobmocks.NewMockKeyValueEntry(s.ctrl)
	e.EXPECT().Value().Return([]byte(data)).AnyTimes()

	return e
}

func (s *MetricsPublicTestSuite) TestComponentRegistry() {
	started := time.Now().Add(-time.Hour)
	conditions := []job.Condition{
		{Type: "HighCPU", Status: true},
		{Type: "MemoryPressure", Status: false},
	}
	process := &job.ProcessMetrics{CPUPercent: 1.5, RSSBytes: 2048}
	subs := map[string]job.SubComponentInfo{"agent.heartbeat": {Status: "ok"}}

	s.Run("when the registry has components returns each one, sorted", func() {
		s.nc.EXPECT().KVKeys(gomock.Any(), "registry").Return([]string{
			"agents.web-01", "agents.bad-json", "controller.ctl-01",
			"nats.bad-json", "nodot", "unknown.x", "agents.gone", "agents.app-01",
			"agents.drain-01",
		}, nil)
		// Only drain-01 has its drain flag set, and it has not applied it.
		s.jobs.EXPECT().CheckDrainFlag(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, machineID string) bool {
				return machineID == "m-drain"
			}).
			AnyTimes()
		s.nc.EXPECT().
			KVGet(gomock.Any(), "registry", "agents.drain-01").
			Return(s.entry(job.AgentRegistration{
				Hostname:  "drain-01",
				MachineID: "m-drain",
				State:     string(job.AgentStateReady),
			}), nil)
		s.nc.EXPECT().
			KVGet(gomock.Any(), "registry", "agents.web-01").
			Return(s.entry(job.AgentRegistration{
				Hostname:      "web-01",
				AgentVersion:  "v1.2.3-agent",
				StartedAt:     started,
				Conditions:    conditions,
				Process:       process,
				SubComponents: subs,
			}), nil)
		s.nc.EXPECT().
			KVGet(gomock.Any(), "registry", "agents.app-01").
			Return(s.entry(job.AgentRegistration{
				Hostname: "app-01",
				State:    string(job.AgentStateCordoned),
			}), nil)
		s.nc.EXPECT().KVGet(gomock.Any(), "registry", "agents.bad-json").Return(s.entry("{"), nil)
		s.nc.EXPECT().
			KVGet(gomock.Any(), "registry", "controller.ctl-01").
			Return(s.entry(job.ComponentRegistration{
				Type:          "controller",
				Hostname:      "ctl-01",
				Version:       "v1.2.3-ctl",
				StartedAt:     started,
				Conditions:    conditions,
				Process:       process,
				SubComponents: map[string]job.SubComponentInfo{"controller.api": {Status: "ok"}},
			}), nil)
		s.nc.EXPECT().KVGet(gomock.Any(), "registry", "nats.bad-json").Return(s.entry("{"), nil)
		s.nc.EXPECT().KVGet(gomock.Any(), "registry", "unknown.x").Return(s.entry("{}"), nil)
		s.nc.EXPECT().KVGet(gomock.Any(), "registry", "agents.gone").Return(nil, errBoom)

		got, err := s.provider("registry").ComponentRegistryFn(context.Background())

		s.Require().NoError(err)
		s.Require().Len(got, 4)
		s.Equal("agent", got[0].Type)
		s.Equal("app-01", got[0].Hostname)
		s.Equal("Cordoned", got[0].Status, "a drained agent shows its state")
		s.Nil(got[0].SubComponents)
		s.Equal("drain-01", got[1].Hostname)
		s.Equal("Draining", got[1].Status, "a drain the agent has not applied is Draining")
		got = append(got[:1], got[2:]...)
		s.Equal("web-01", got[1].Hostname)
		s.Equal("Ready", got[1].Status, "an agent with no state is Ready")
		s.Equal([]string{"HighCPU"}, got[1].Conditions)
		s.Equal("v1.2.3-agent", got[1].Version, "an agent's entry has its binary's version")
		s.Equal(1.5, got[1].CPUPercent)
		s.Equal(int64(2048), got[1].MemBytes)
		s.Equal(
			map[string]health.SubComponentInfo{"agent.heartbeat": {Status: "ok"}},
			got[1].SubComponents,
		)
		s.Equal("controller", got[2].Type)
		s.Equal("ctl-01", got[2].Hostname)
		s.Equal("Ready", got[2].Status)
		s.Equal([]string{"HighCPU"}, got[2].Conditions)
		s.Equal("v1.2.3-ctl", got[2].Version, "a controller's entry has its binary's version")
	})

	s.Run("when the registry is empty returns no components", func() {
		s.nc.EXPECT().KVKeys(gomock.Any(), "registry").Return([]string{}, nil)

		got, err := s.provider("registry").ComponentRegistryFn(context.Background())

		s.Require().NoError(err)
		s.Empty(got)
	})

	s.Run("when the registry fails returns the error", func() {
		s.nc.EXPECT().KVKeys(gomock.Any(), "registry").Return(nil, errBoom)

		_, err := s.provider("registry").ComponentRegistryFn(context.Background())

		s.EqualError(err, "list registry keys: boom")
	})

	s.Run("when there is no registry returns the error", func() {
		_, err := s.provider("").ComponentRegistryFn(context.Background())

		s.EqualError(err, "registry KV not configured")
	})
}

func (s *MetricsPublicTestSuite) TestHealthChecker() {
	tests := []struct {
		name    string
		url     string
		kvErr   error
		wantErr string
	}{
		{name: "when NATS and the bucket answer reports healthy", url: "nats://x"},
		{name: "when NATS is not connected reports it", wantErr: "nats not connected"},
		{
			name:    "when the bucket fails reports it",
			url:     "nats://x",
			kvErr:   errBoom,
			wantErr: "kv bucket not accessible: boom",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.nc.EXPECT().ConnectedURL().Return(tc.url)
			s.nc.EXPECT().KVStatus(gomock.Any(), "jobs").Return(nil, tc.kvErr).AnyTimes()
			checker := controller.ExportNewHealthChecker(s.nc, "jobs")

			err := checker.NATSCheck()
			if err == nil {
				err = checker.KVCheck()
			}

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)
				return
			}
			s.NoError(err)
		})
	}
}

func (s *MetricsPublicTestSuite) TestConfiguredBuckets() {
	s.Run("when buckets have names returns them with the namespace", func() {
		cfg := natsConfig()
		cfg.Facts.Bucket = ""

		kvs := controller.ExportConfiguredKVBuckets("ns", cfg)
		objs := controller.ExportConfiguredObjectBuckets("ns", cfg)

		s.Contains(kvs, "ns-job-queue")
		s.NotContains(kvs, "ns-")
		s.Equal([]string{"ns-file-objects"}, objs)
	})

	s.Run("when no bucket has a name returns none", func() {
		s.Empty(controller.ExportConfiguredObjectBuckets("ns", config.NATS{}))
	})
}
