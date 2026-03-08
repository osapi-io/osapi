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

package osapi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/osapi/gen"
)

type HealthTypesTestSuite struct {
	suite.Suite
}

func (suite *HealthTypesTestSuite) TestHealthStatusFromGen() {
	tests := []struct {
		name         string
		input        *gen.HealthResponse
		validateFunc func(HealthStatus)
	}{
		{
			name: "when status is ok",
			input: &gen.HealthResponse{
				Status: "ok",
			},
			validateFunc: func(h HealthStatus) {
				suite.Equal("ok", h.Status)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := healthStatusFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *HealthTypesTestSuite) TestReadyStatusFromGen() {
	tests := []struct {
		name               string
		input              *gen.ReadyResponse
		serviceUnavailable bool
		validateFunc       func(ReadyStatus)
	}{
		{
			name: "when ready with no error",
			input: &gen.ReadyResponse{
				Status: "ready",
			},
			serviceUnavailable: false,
			validateFunc: func(r ReadyStatus) {
				suite.Equal("ready", r.Status)
				suite.Empty(r.Error)
				suite.False(r.ServiceUnavailable)
			},
		},
		{
			name: "when not ready with error",
			input: func() *gen.ReadyResponse {
				errMsg := "NATS connection failed"

				return &gen.ReadyResponse{
					Status: "not_ready",
					Error:  &errMsg,
				}
			}(),
			serviceUnavailable: true,
			validateFunc: func(r ReadyStatus) {
				suite.Equal("not_ready", r.Status)
				suite.Equal("NATS connection failed", r.Error)
				suite.True(r.ServiceUnavailable)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := readyStatusFromGen(tc.input, tc.serviceUnavailable)
			tc.validateFunc(result)
		})
	}
}

func (suite *HealthTypesTestSuite) TestSystemStatusFromGen() {
	tests := []struct {
		name               string
		input              *gen.StatusResponse
		serviceUnavailable bool
		validateFunc       func(SystemStatus)
	}{
		{
			name: "when all fields are populated",
			input: func() *gen.StatusResponse {
				errMsg := "connection timeout"
				labels := "group=web"

				return &gen.StatusResponse{
					Status:  "degraded",
					Version: "1.2.3",
					Uptime:  "5d 3h",
					Components: map[string]gen.ComponentHealth{
						"nats": {
							Status: "healthy",
						},
						"store": {
							Status: "unhealthy",
							Error:  &errMsg,
						},
					},
					Nats: &gen.NATSInfo{
						Url:     "nats://localhost:4222",
						Version: "2.10.0",
					},
					Agents: &gen.AgentStats{
						Total: 3,
						Ready: 2,
						Agents: &[]gen.AgentDetail{
							{
								Hostname:   "web-01",
								Labels:     &labels,
								Registered: "5m ago",
							},
							{
								Hostname:   "web-02",
								Registered: "10m ago",
							},
						},
					},
					Jobs: &gen.JobStats{
						Total:       100,
						Completed:   80,
						Failed:      5,
						Processing:  10,
						Unprocessed: 3,
						Dlq:         2,
					},
					Consumers: &gen.ConsumerStats{
						Total: 2,
						Consumers: &[]gen.ConsumerDetail{
							{
								Name:        "jobs-agent",
								Pending:     5,
								AckPending:  2,
								Redelivered: 1,
							},
						},
					},
					Streams: &[]gen.StreamInfo{
						{
							Name:      "JOBS",
							Messages:  500,
							Bytes:     1048576,
							Consumers: 2,
						},
					},
					KvBuckets: &[]gen.KVBucketInfo{
						{
							Name:  "job-queue",
							Keys:  50,
							Bytes: 524288,
						},
						{
							Name:  "audit-log",
							Keys:  200,
							Bytes: 2097152,
						},
					},
					ObjectStores: &[]gen.ObjectStoreInfo{
						{
							Name: "file-objects",
							Size: 5242880,
						},
					},
				}
			}(),
			serviceUnavailable: false,
			validateFunc: func(s SystemStatus) {
				suite.Equal("degraded", s.Status)
				suite.Equal("1.2.3", s.Version)
				suite.Equal("5d 3h", s.Uptime)
				suite.False(s.ServiceUnavailable)

				suite.Require().Len(s.Components, 2)
				suite.Equal("healthy", s.Components["nats"].Status)
				suite.Empty(s.Components["nats"].Error)
				suite.Equal("unhealthy", s.Components["store"].Status)
				suite.Equal("connection timeout", s.Components["store"].Error)

				suite.Require().NotNil(s.NATS)
				suite.Equal("nats://localhost:4222", s.NATS.URL)
				suite.Equal("2.10.0", s.NATS.Version)

				suite.Require().NotNil(s.Agents)
				suite.Equal(3, s.Agents.Total)
				suite.Equal(2, s.Agents.Ready)
				suite.Require().Len(s.Agents.Agents, 2)
				suite.Equal("web-01", s.Agents.Agents[0].Hostname)
				suite.Equal("group=web", s.Agents.Agents[0].Labels)
				suite.Equal("5m ago", s.Agents.Agents[0].Registered)
				suite.Equal("web-02", s.Agents.Agents[1].Hostname)
				suite.Empty(s.Agents.Agents[1].Labels)
				suite.Equal("10m ago", s.Agents.Agents[1].Registered)

				suite.Require().NotNil(s.Jobs)
				suite.Equal(100, s.Jobs.Total)
				suite.Equal(80, s.Jobs.Completed)
				suite.Equal(5, s.Jobs.Failed)
				suite.Equal(10, s.Jobs.Processing)
				suite.Equal(3, s.Jobs.Unprocessed)
				suite.Equal(2, s.Jobs.Dlq)

				suite.Require().NotNil(s.Consumers)
				suite.Equal(2, s.Consumers.Total)
				suite.Require().Len(s.Consumers.Consumers, 1)
				suite.Equal("jobs-agent", s.Consumers.Consumers[0].Name)
				suite.Equal(5, s.Consumers.Consumers[0].Pending)
				suite.Equal(2, s.Consumers.Consumers[0].AckPending)
				suite.Equal(1, s.Consumers.Consumers[0].Redelivered)

				suite.Require().Len(s.Streams, 1)
				suite.Equal("JOBS", s.Streams[0].Name)
				suite.Equal(500, s.Streams[0].Messages)
				suite.Equal(1048576, s.Streams[0].Bytes)
				suite.Equal(2, s.Streams[0].Consumers)

				suite.Require().Len(s.KVBuckets, 2)
				suite.Equal("job-queue", s.KVBuckets[0].Name)
				suite.Equal(50, s.KVBuckets[0].Keys)
				suite.Equal(524288, s.KVBuckets[0].Bytes)
				suite.Equal("audit-log", s.KVBuckets[1].Name)
				suite.Equal(200, s.KVBuckets[1].Keys)
				suite.Equal(2097152, s.KVBuckets[1].Bytes)

				suite.Require().Len(s.ObjectStores, 1)
				suite.Equal("file-objects", s.ObjectStores[0].Name)
				suite.Equal(5242880, s.ObjectStores[0].Size)
			},
		},
		{
			name: "when only required fields are set",
			input: &gen.StatusResponse{
				Status:     "ok",
				Version:    "1.0.0",
				Uptime:     "1h",
				Components: map[string]gen.ComponentHealth{},
			},
			serviceUnavailable: false,
			validateFunc: func(s SystemStatus) {
				suite.Equal("ok", s.Status)
				suite.Equal("1.0.0", s.Version)
				suite.Equal("1h", s.Uptime)
				suite.False(s.ServiceUnavailable)
				suite.Empty(s.Components)
				suite.Nil(s.NATS)
				suite.Nil(s.Agents)
				suite.Nil(s.Jobs)
				suite.Nil(s.Consumers)
				suite.Nil(s.Streams)
				suite.Nil(s.KVBuckets)
				suite.Nil(s.ObjectStores)
			},
		},
		{
			name: "when service unavailable is true",
			input: &gen.StatusResponse{
				Status:     "degraded",
				Version:    "1.0.0",
				Uptime:     "30m",
				Components: map[string]gen.ComponentHealth{},
			},
			serviceUnavailable: true,
			validateFunc: func(s SystemStatus) {
				suite.Equal("degraded", s.Status)
				suite.True(s.ServiceUnavailable)
			},
		},
		{
			name: "when agents has nil agents list",
			input: &gen.StatusResponse{
				Status:  "ok",
				Version: "1.0.0",
				Uptime:  "1h",
				Components: map[string]gen.ComponentHealth{
					"nats": {Status: "healthy"},
				},
				Agents: &gen.AgentStats{
					Total: 0,
					Ready: 0,
				},
			},
			serviceUnavailable: false,
			validateFunc: func(s SystemStatus) {
				suite.Require().NotNil(s.Agents)
				suite.Equal(0, s.Agents.Total)
				suite.Equal(0, s.Agents.Ready)
				suite.Nil(s.Agents.Agents)
			},
		},
		{
			name: "when consumers has nil consumers list",
			input: &gen.StatusResponse{
				Status:     "ok",
				Version:    "1.0.0",
				Uptime:     "1h",
				Components: map[string]gen.ComponentHealth{},
				Consumers: &gen.ConsumerStats{
					Total: 0,
				},
			},
			serviceUnavailable: false,
			validateFunc: func(s SystemStatus) {
				suite.Require().NotNil(s.Consumers)
				suite.Equal(0, s.Consumers.Total)
				suite.Nil(s.Consumers.Consumers)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := systemStatusFromGen(tc.input, tc.serviceUnavailable)
			tc.validateFunc(result)
		})
	}
}

func TestHealthTypesTestSuite(t *testing.T) {
	suite.Run(t, new(HealthTypesTestSuite))
}
