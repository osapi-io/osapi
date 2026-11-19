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
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/bus/server"
	"github.com/osapi-io/osapi/internal/config"
)

type ServerPublicTestSuite struct {
	suite.Suite
}

func TestServerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ServerPublicTestSuite))
}

// freePort returns a port that nothing listens on now.
func freePort(
	t *testing.T,
) int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = l.Close() }()

	return l.Addr().(*net.TCPAddr).Port
}

// serverNATS is a config for an embedded server on a free port, with each
// stream and bucket named.
func serverNATS(
	t *testing.T,
) config.NATS {
	return config.NATS{
		Server: config.NATSServer{
			Host:      "127.0.0.1",
			Port:      freePort(t),
			StoreDir:  t.TempDir(),
			Namespace: "test",
			Auth:      config.NATSServerAuth{Type: "none"},
		},
		Stream: config.NATSStream{
			Name:     "JOBS",
			Subjects: "jobs.>",
			MaxAge:   "1h",
			Storage:  "memory",
			Replicas: 1,
		},
		KV:        config.NATSKV{Bucket: "job-queue", ResponseBucket: "job-responses"},
		DLQ:       config.NATSDLQ{MaxAge: "24h", Storage: "memory", Replicas: 1},
		Audit:     config.NATSAudit{Stream: "AUDIT", Subject: "audit"},
		Registry:  config.NATSRegistry{Bucket: "agent-registry"},
		Facts:     config.NATSFacts{Bucket: "agent-facts"},
		State:     config.NATSState{Bucket: "agent-state"},
		Objects:   config.NATSObjects{Bucket: "file-objects"},
		FileState: config.NATSFileState{Bucket: "file-state"},
	}
}

func (s *ServerPublicTestSuite) TestNew() {
	tests := []struct {
		name    string
		cfg     func() config.NATS
		dial    func(*slog.Logger, *natsclient.Options) (*natsclient.Client, error)
		wantErr string
	}{
		{
			name: "when the server starts makes the streams and buckets",
			cfg:  func() config.NATS { return serverNATS(s.T()) },
		},
		{
			name: "when the auth is not valid returns the error",
			cfg: func() config.NATS {
				c := serverNATS(s.T())
				c.Server.Auth = config.NATSServerAuth{Type: "operator"}
				return c
			},
			wantErr: "NATS server auth: ",
		},
		{
			name: "when the port is in use returns the error",
			cfg: func() config.NATS {
				c := serverNATS(s.T())
				l, err := net.Listen("tcp", "127.0.0.1:0")
				s.Require().NoError(err)
				s.T().Cleanup(func() { _ = l.Close() })
				c.Server.Port = l.Addr().(*net.TCPAddr).Port
				return c
			},
			wantErr: "start embedded NATS server: ",
		},
		{
			name: "when the connection fails stops the server and returns the error",
			cfg:  func() config.NATS { return serverNATS(s.T()) },
			dial: func(*slog.Logger, *natsclient.Options) (*natsclient.Client, error) {
				return nil, errors.New("connect to NATS: refused")
			},
			wantErr: "connect to NATS: refused",
		},
		{
			name: "when the setup fails stops the server and returns the error",
			cfg: func() config.NATS {
				c := serverNATS(s.T())
				c.Stream.Name = "bad name"
				return c
			},
			wantErr: "set up JetStream: ",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.dial != nil {
				server.SetDialFn(tc.dial)
				defer server.ResetDialFn()
			}

			srv, err := server.New(context.Background(), slog.Default(), tc.cfg(), nil)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)
				s.Nil(srv)
				return
			}
			s.Require().NoError(err)
			srv.Start()
			srv.Stop(context.Background())
		})
	}
}

// startServer runs an embedded server for one test.
func startServer(
	t *testing.T,
	cfg config.NATS,
) *server.Server {
	srv, err := server.New(context.Background(), slog.Default(), cfg, sdkmetric.NewMeterProvider())
	require.NoError(t, err)
	t.Cleanup(func() { srv.Stop(context.Background()) })

	return srv
}

// registryKeys returns a function that lists the keys in the registry bucket.
func registryKeys(
	t *testing.T,
	cfg config.NATS,
) func() []string {
	nc, err := bus.Dial(slog.Default(), &natsclient.Options{
		Host: cfg.Server.Host,
		Port: cfg.Server.Port,
		Auth: natsclient.AuthOptions{AuthType: natsclient.NoAuth},
		Name: "test",
	})
	require.NoError(t, err)
	t.Cleanup(nc.Close)
	bucket := bus.RegistryKV(cfg.Server.Namespace, cfg.Registry).Bucket

	return func() []string {
		keys, _ := nc.KVKeys(context.Background(), bucket)
		return keys
	}
}

// registryVersion is the version in the registration under key.
func registryVersion(
	t *testing.T,
	cfg config.NATS,
	key string,
) string {
	nc, err := bus.Dial(slog.Default(), &natsclient.Options{
		Host: cfg.Server.Host,
		Port: cfg.Server.Port,
		Auth: natsclient.AuthOptions{AuthType: natsclient.NoAuth},
		Name: "test",
	})
	require.NoError(t, err)
	defer nc.Close()
	entry, err := nc.KVGet(
		context.Background(),
		bus.RegistryKV(cfg.Server.Namespace, cfg.Registry).Bucket,
		key,
	)
	require.NoError(t, err)
	var reg struct {
		Version string `json:"version"`
	}
	require.NoError(t, json.Unmarshal(entry.Value(), &reg))

	return reg.Version
}
