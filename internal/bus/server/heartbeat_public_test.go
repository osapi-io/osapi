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

package server_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus/server"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/cryptoreport"
	"github.com/osapi-io/osapi/internal/job"
)

type HeartbeatPublicTestSuite struct {
	suite.Suite
}

func TestHeartbeatPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HeartbeatPublicTestSuite))
}

func (s *HeartbeatPublicTestSuite) TestStartHeartbeat() {
	tests := []struct {
		name     string
		registry string
		hostname func() (string, error)
		wantKey  string
	}{
		{
			name:     "when the server runs writes the registration",
			registry: "agent-registry",
			hostname: func() (string, error) { return "nats-host", nil },
			wantKey:  "nats.nats_host",
		},
		{
			name:     "when the hostname fails writes the registration as unknown",
			registry: "agent-registry",
			hostname: func() (string, error) { return "", errors.New("no hostname") },
			wantKey:  "nats.unknown",
		},
		{
			name: "when the config names no registry bucket writes nothing",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.hostname != nil {
				server.SetHostnameFn(tc.hostname)
				defer server.ResetHostnameFn()
			}
			cfg := serverNATS(s.T())
			cfg.Registry.Bucket = tc.registry
			srv := startServer(s.T(), cfg)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			srv.StartHeartbeat(ctx, "v1.2.3-test", config.ProcessConditions{}, &cryptoreport.Report{Go: "go1.27.3"})

			if tc.wantKey == "" {
				return
			}
			keys := registryKeys(s.T(), cfg)
			s.Eventually(func() bool {
				return slices.Contains(keys(), tc.wantKey)
			}, 5*time.Second, 50*time.Millisecond)
			s.Equal("v1.2.3-test", registryVersion(s.T(), cfg, tc.wantKey),
				"the registration has the build version it was given")
		})
	}
}

func (s *HeartbeatPublicTestSuite) TestSubComponents() {
	tests := []struct {
		name string
		sc   config.NATSServer
		want map[string]job.SubComponentInfo
	}{
		{
			name: "when metrics are off reports them as disabled",
			sc:   config.NATSServer{Host: "0.0.0.0", Port: 4222},
			want: map[string]job.SubComponentInfo{
				"nats.heartbeat": {Status: "ok"},
				"nats.server":    {Status: "ok", Address: "nats://0.0.0.0:4222"},
				"nats.metrics":   {Status: "disabled"},
			},
		},
		{
			name: "when the server serves TLS reports a tls:// address",
			sc: config.NATSServer{
				Host: "0.0.0.0",
				Port: 4222,
				TLS:  config.ServerTLS{Enabled: true},
			},
			want: map[string]job.SubComponentInfo{
				"nats.heartbeat": {Status: "ok"},
				"nats.server":    {Status: "ok", Address: "tls://0.0.0.0:4222"},
				"nats.metrics":   {Status: "disabled"},
			},
		},
		{
			name: "when metrics are on reports their address",
			sc: config.NATSServer{
				Host: "0.0.0.0",
				Port: 4222,
				Metrics: config.MetricsServer{
					Enabled: true,
					Host:    "127.0.0.1",
					Port:    9090,
				},
			},
			want: map[string]job.SubComponentInfo{
				"nats.heartbeat": {Status: "ok"},
				"nats.server":    {Status: "ok", Address: "nats://0.0.0.0:4222"},
				"nats.metrics":   {Status: "ok", Address: "http://127.0.0.1:9090"},
			},
		},
		{
			name: "when url keys are set reports them beside the address",
			sc: config.NATSServer{
				Host: "0.0.0.0",
				Port: 4222,
				URL:  "nats://hub.example.ts.net:4222",
				Metrics: config.MetricsServer{
					Enabled: true,
					Host:    "127.0.0.1",
					Port:    9092,
					URL:     "http://hub.example.ts.net:9092/",
				},
			},
			want: map[string]job.SubComponentInfo{
				"nats.heartbeat": {Status: "ok"},
				"nats.server": {
					Status:  "ok",
					Address: "nats://0.0.0.0:4222",
					URL:     "nats://hub.example.ts.net:4222",
				},
				"nats.metrics": {
					Status:  "ok",
					Address: "http://127.0.0.1:9092",
					URL:     "http://hub.example.ts.net:9092/",
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, server.SubComponents(tc.sc))
		})
	}
}
