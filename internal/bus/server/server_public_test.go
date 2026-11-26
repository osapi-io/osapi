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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/avfs/avfs"
	"github.com/osapi-io/osapi/internal/natsauth"

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

// TestNewOperatorFails: a server whose operator cannot be made does not
// start.
func (s *ServerPublicTestSuite) TestNewOperatorFails() {
	defer server.SetNatsauthOperator(func(
		context.Context, avfs.VFS, string, string, config.NATS,
	) (natsauth.Result, error) {
		return natsauth.Result{}, errors.New("no entropy")
	})()
	c := serverNATS(s.T())
	c.Server.Auth = config.NATSServerAuth{Type: "operator"}
	c.Server.KeyDir = s.T().TempDir()
	_, err := server.New(context.Background(), slog.Default(), c, nil)
	s.ErrorContains(err, "make the operator: no entropy")
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
		KV:        config.NATSKV{Bucket: "job-queue"},
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
				c.Server.Auth = config.NATSServerAuth{
					Type: "operator", ResolverFile: "/no/such/resolver.conf",
				}
				return c
			},
			wantErr: "NATS server: ",
		},
		{
			name: "when operator mode has no resolver_file, it makes the operator at its first start",
			cfg: func() config.NATS {
				c := serverNATS(s.T())
				c.Server.Auth = config.NATSServerAuth{Type: "operator"}
				c.Server.KeyDir = s.T().TempDir()
				return c
			},
		},
		{
			name: "when the operator cannot be made returns the error",
			cfg: func() config.NATS {
				c := serverNATS(s.T())
				blocked := filepath.Join(s.T().TempDir(), "file")
				s.Require().NoError(os.WriteFile(blocked, nil, 0o600))
				c.Server.Auth = config.NATSServerAuth{Type: "operator"}
				c.Server.KeyDir = blocked
				return c
			},
			wantErr: "NATS server: make",
		},
		{
			name: "when a user has a role that osapi does not have returns the error",
			cfg: func() config.NATS {
				c := serverNATS(s.T())
				c.Server.Auth = config.NATSServerAuth{
					Type:  "user_pass",
					Users: []config.NATSServerUser{{Username: "x", Password: "p", Role: "admin"}},
				}
				return c
			},
			wantErr: "NATS server: user x: no such role",
		},
		{
			name: "when TLS is on with no certificate, it makes nats.crt in key_dir and serves it",
			cfg: func() config.NATS {
				c := serverNATS(s.T())
				c.Server.TLS = config.ServerTLS{Enabled: true}
				c.Server.KeyDir = s.T().TempDir()
				return c
			},
		},
		{
			name: "when TLS is on and key_dir cannot hold a certificate returns the error",
			cfg: func() config.NATS {
				c := serverNATS(s.T())
				blocked := filepath.Join(s.T().TempDir(), "file")
				s.Require().NoError(os.WriteFile(blocked, nil, 0o600))
				c.Server.TLS = config.ServerTLS{Enabled: true}
				c.Server.KeyDir = filepath.Join(blocked, "pki")
				return c
			},
			wantErr: "make the key directory",
		},
		{
			name: "when TLS is on with a certificate, it sets itself up over TLS",
			cfg: func() config.NATS {
				c := serverNATS(s.T())
				cert, key := writeCert(s.T())
				c.Server.TLS = config.ServerTLS{Enabled: true, CertFile: cert, KeyFile: key}
				return c
			},
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
			name: "when the key file cannot be read returns the error",
			cfg: func() config.NATS {
				c := serverNATS(s.T())
				c.Server.JetStream.KeyFile = filepath.Join(s.T().TempDir(), "missing.key")
				return c
			},
			wantErr: "read the JetStream key",
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

// fileNATS is serverNATS with each stream and bucket in file storage, in
// storeDir, so that the key applies to them.
func fileNATS(
	t *testing.T,
	storeDir string,
) config.NATS {
	c := serverNATS(t)
	c.Server.StoreDir = storeDir
	c.Stream.Storage = "file"
	c.DLQ.Storage = "file"

	return c
}

// putAndGet writes value to key in the registry bucket of a server with cfg,
// when value is not "", then reads key back. The server stops at the end.
func putAndGet(
	t *testing.T,
	cfg config.NATS,
	key string,
	value string,
) (string, error) {
	srv, err := server.New(context.Background(), slog.Default(), cfg, nil)
	if err != nil {
		return "", err
	}
	defer srv.Stop(context.Background())

	nc, err := bus.Dial(slog.Default(), &natsclient.Options{
		Host: cfg.Server.Host,
		Port: cfg.Server.Port,
		Auth: natsclient.AuthOptions{AuthType: natsclient.NoAuth},
		Name: "test",
	})
	require.NoError(t, err)
	defer nc.Close()
	bucket := bus.RegistryKV(cfg.Server.Namespace, cfg.Registry).Bucket
	if value != "" {
		_, err = nc.KVPut(context.Background(), bucket, key, []byte(value))
		require.NoError(t, err)
	}
	entry, err := nc.KVGet(context.Background(), bucket, key)
	require.NoError(t, err)

	return string(entry.Value()), nil
}

// storeHas reports whether a file under dir has text in it.
func storeHas(
	t *testing.T,
	dir string,
	text string,
) bool {
	found := false
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(b), text) {
			found = true
		}

		return nil
	})
	require.NoError(t, err)

	return found
}

func (s *ServerPublicTestSuite) TestJetStreamEncryption() {
	keyDir := s.T().TempDir()
	keyFile := func(name, key string) string {
		path := filepath.Join(keyDir, name)
		s.Require().NoError(os.WriteFile(path, []byte(key+"\n"), 0o600))

		return path
	}
	key := keyFile("js.key", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	wrong := keyFile("wrong.key", "cccccccccccccccccccccccccccccccc")
	store := s.T().TempDir()
	cfg := func(keyFile string) config.NATS {
		c := fileNATS(s.T(), store)
		c.Server.JetStream = config.NATSServerJetStream{KeyFile: keyFile}
		return c
	}

	s.Run("with no key, the store has the plain value", func() {
		got, err := putAndGet(s.T(), cfg(""), "plain", "PLAIN-MARK-1")
		s.Require().NoError(err)
		s.Equal("PLAIN-MARK-1", got)
		s.True(storeHas(s.T(), store, "PLAIN-MARK-1"))
	})

	s.Run("with a key, the plain store becomes encrypted and stays readable", func() {
		got, err := putAndGet(s.T(), cfg(key), "plain", "")
		s.Require().NoError(err)
		s.Equal("PLAIN-MARK-1", got)

		got, err = putAndGet(s.T(), cfg(key), "secret", "SECRET-MARK-2")
		s.Require().NoError(err)
		s.Equal("SECRET-MARK-2", got)
		s.False(storeHas(s.T(), store, "PLAIN-MARK-1"))
		s.False(storeHas(s.T(), store, "SECRET-MARK-2"))
	})

	s.Run("with a wrong key, the server does not start", func() {
		_, err := putAndGet(s.T(), cfg(wrong), "secret", "")
		s.ErrorContains(err, "set up JetStream: ")
	})

	s.Run("with no key, the server does not start", func() {
		_, err := putAndGet(s.T(), cfg(""), "secret", "")
		s.ErrorContains(err, "set up JetStream: ")
	})

	s.Run("with the key again, the store is readable", func() {
		got, err := putAndGet(s.T(), cfg(key), "secret", "")
		s.Require().NoError(err)
		s.Equal("SECRET-MARK-2", got)
	})
}

// TestTLS is a client of a server that serves TLS: one that trusts the
// certificate connects, and one without TLS is refused.
func (s *ServerPublicTestSuite) TestTLS() {
	cfg := serverNATS(s.T())
	cert, key := writeCert(s.T())
	cfg.Server.TLS = config.ServerTLS{Enabled: true, CertFile: cert, KeyFile: key}
	srv, err := server.New(context.Background(), slog.Default(), cfg, nil)
	s.Require().NoError(err)
	defer srv.Stop(context.Background())

	conn := config.NATSConnection{
		Host: "127.0.0.1",
		Port: cfg.Server.Port,
		Auth: config.NATSAuth{Type: "none"},
		TLS:  config.ClientTLS{Enabled: true, CAFile: cert},
	}
	opts, err := bus.ConnectionOptions(conn, "test")
	s.Require().NoError(err)
	nc, err := bus.Dial(slog.Default(), opts)
	s.Require().NoError(err)
	bus.Close(nc)

	conn.TLS = config.ClientTLS{}
	opts, err = bus.ConnectionOptions(conn, "test")
	s.Require().NoError(err)
	_, err = bus.Dial(slog.Default(), opts)
	s.Error(err, "plain text is refused")
}
