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

package bus_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
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

func (s *HeartbeatPublicTestSuite) TearDownTest() {
	bus.ResetHostnameFn()
}

// startServer runs an embedded server for one test and returns its config.
func (s *HeartbeatPublicTestSuite) startServer() config.NATS {
	cfg := serverNATS(s.T())
	srv, err := bus.NewServer(context.Background(), slog.Default(), cfg)
	s.Require().NoError(err)
	s.T().Cleanup(func() { srv.Stop(context.Background()) })

	return cfg
}

// registryKeys returns the keys in the registry bucket of cfg.
func (s *HeartbeatPublicTestSuite) registryKeys(
	cfg config.NATS,
) func() []string {
	nc := natsclient.New(slog.Default(), &natsclient.Options{
		Host: cfg.Server.Host,
		Port: cfg.Server.Port,
		Auth: natsclient.AuthOptions{AuthType: natsclient.NoAuth},
		Name: "test",
	})
	s.Require().NoError(nc.Connect())
	s.T().Cleanup(nc.Close)
	bucket := bus.RegistryKV(cfg.Server.Namespace, cfg.Registry).Bucket

	return func() []string {
		keys, _ := nc.KVKeys(bucket)
		return keys
	}
}

func (s *HeartbeatPublicTestSuite) TestStartHeartbeat() {
	tests := []struct {
		name     string
		cfg      func() config.NATS
		hostname func() (string, error)
		wantKey  string
	}{
		{
			name: "when the server runs writes the registration",
			cfg:  s.startServer,
			hostname: func() (string, error) {
				return "nats-host", nil
			},
			wantKey: "nats.nats_host",
		},
		{
			name: "when the hostname fails writes the registration as unknown",
			cfg:  s.startServer,
			hostname: func() (string, error) {
				return "", errors.New("no hostname")
			},
			wantKey: "nats.unknown",
		},
		{
			name: "when nothing listens returns without a heartbeat",
			cfg:  func() config.NATS { return serverNATS(s.T()) },
		},
		{
			name: "when the registry bucket fails returns without a heartbeat",
			cfg: func() config.NATS {
				c := s.startServer()
				c.Registry.Bucket = "bad name"
				return c
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.hostname != nil {
				bus.SetHostnameFn(tc.hostname)
			}
			cfg := tc.cfg()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			bus.StartHeartbeat(ctx, slog.Default(), cfg, config.ProcessConditions{})

			if tc.wantKey == "" {
				return
			}
			keys := s.registryKeys(cfg)
			s.Eventually(func() bool {
				for _, k := range keys() {
					if k == tc.wantKey {
						return true
					}
				}
				return false
			}, 5*time.Second, 50*time.Millisecond)
		})
	}
}

func (s *HeartbeatPublicTestSuite) TestStartHeartbeatOn() {
	s.Run("when given a registry bucket writes the registration", func() {
		bus.SetHostnameFn(func() (string, error) { return "combined", nil })
		cfg := s.startServer()
		nc := natsclient.New(slog.Default(), &natsclient.Options{
			Host: cfg.Server.Host,
			Port: cfg.Server.Port,
			Auth: natsclient.AuthOptions{AuthType: natsclient.NoAuth},
			Name: "test",
		})
		s.Require().NoError(nc.Connect())
		defer nc.Close()
		kv, err := nc.KeyValue(
			context.Background(),
			bus.RegistryKV(cfg.Server.Namespace, cfg.Registry).Bucket,
		)
		s.Require().NoError(err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		bus.StartHeartbeatOn(ctx, slog.Default(), kv, cfg.Server, config.ProcessConditions{})

		s.Eventually(func() bool {
			_, err := kv.Get(context.Background(), "nats.combined")
			return err == nil
		}, 5*time.Second, 50*time.Millisecond)
	})
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
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, bus.SubComponents(tc.sc))
		})
	}
}
