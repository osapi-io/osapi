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
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nats-io/nkeys"

	"github.com/avfs/avfs/vfs/osfs"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/bus/server"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/job"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
	"github.com/osapi-io/osapi/internal/validation"
)

type ControllerPublicTestSuite struct {
	suite.Suite
}

func TestControllerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ControllerPublicTestSuite))
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

// natsConfig names each stream and bucket that the controller uses.
func natsConfig() config.NATS {
	return config.NATS{
		Stream: config.NATSStream{
			Name:     "JOBS",
			Subjects: "jobs.>",
			Storage:  "memory",
			Replicas: 1,
		},
		KV:  config.NATSKV{Bucket: "job-queue"},
		DLQ: config.NATSDLQ{Storage: "memory", Replicas: 1},
		Audit: config.NATSAudit{
			Stream:   "AUDIT",
			Subject:  "audit",
			Storage:  "memory",
			Replicas: 1,
		},
		Registry:   config.NATSRegistry{Bucket: "agent-registry"},
		Facts:      config.NATSFacts{Bucket: "agent-facts"},
		State:      config.NATSState{Bucket: "agent-state"},
		Objects:    config.NATSObjects{Bucket: "file-objects"},
		FileState:  config.NATSFileState{Bucket: "file-state"},
		Enrollment: config.NATSEnrollment{Bucket: "agent-enrollment"},
		Secrets:    config.NATSSecrets{Bucket: "secrets", History: 5},
	}
}

// withVault is cfg with a nats vault default, whose key is in a new file.
func withVault(
	t *testing.T,
	cfg config.Config,
) config.Config {
	file := filepath.Join(t.TempDir(), "secrets.key")
	_, err := natsvault.CreateKey(context.Background(), osfs.NewWithNoIdm(), file)
	require.NoError(t, err)
	cfg.Controller.Secrets.Vaults = []vault.Spec{{Name: "default", Type: "nats", KeyFile: file}}

	return cfg
}

// withSecret puts value at key in the vault default of cfg, as the
// controller reads it.
func withSecret(
	t *testing.T,
	cfg config.Config,
	key string,
	value string,
) config.Config {
	nc := natsclient.New(slog.Default(), natsOptions(cfg))
	require.NoError(t, nc.Connect())
	defer nc.Close()
	spec := cfg.Controller.Secrets.Vaults[0]
	v, err := natsvault.Open(context.Background(), nc,
		job.ApplyNamespaceToInfraName(cfg.Controller.NATS.Namespace, cfg.NATS.Secrets.Bucket),
		spec.Name, natsvault.Keys{FS: osfs.NewWithNoIdm(), File: spec.KeyFile})
	require.NoError(t, err)
	_, err = v.Put(context.Background(), key, []byte(value))
	require.NoError(t, err)

	return cfg
}

// startServer runs an embedded server and returns a controller config that
// connects to it.
func startServer(
	t *testing.T,
) config.Config {
	n := natsConfig()
	n.Server = config.NATSServer{
		Host:      "127.0.0.1",
		Port:      freePort(t),
		StoreDir:  t.TempDir(),
		Namespace: "test",
		Auth:      config.NATSServerAuth{Type: "none"},
	}
	srv, err := server.New(context.Background(), slog.Default(), n, nil)
	require.NoError(t, err)
	t.Cleanup(func() { srv.Stop(context.Background()) })

	var cfg config.Config
	cfg.NATS = n
	cfg.Controller.NATS = config.NATSConnection{
		Host:       "127.0.0.1",
		Port:       n.Server.Port,
		Namespace:  "test",
		ClientName: "test-controller",
		Auth:       config.NATSAuth{Type: "none"},
	}
	cfg.Controller.API.Port = freePort(t)
	cfg.Controller.API.JobTimeout = "5s"
	cfg.Controller.API.Security.SigningKeys = []config.SigningKey{{ID: "test", Key: "test-key-of-at-least-32-characters"}}
	cfg.Controller.PKI.KeyDir = t.TempDir()

	return cfg
}

// natsOptions are the client options for the NATS that cfg names.
func natsOptions(
	cfg config.Config,
) *natsclient.Options {
	return &natsclient.Options{
		Host: cfg.Controller.NATS.Host,
		Port: cfg.Controller.NATS.Port,
		Auth: natsclient.AuthOptions{AuthType: natsclient.NoAuth},
		Name: "test",
	}
}

func (s *ControllerPublicTestSuite) TestNew() {
	tests := []struct {
		name       string
		cfg        func() config.Config
		opts       []controller.Option
		wantErr    string
		wantVaults []string
	}{
		{
			name: "when NATS runs serves the API and answers health checks",
			cfg:  func() config.Config { return startServer(s.T()) },
		},
		{
			name: "when PKI, stale cleanup, notifications and metrics are on starts each of them",
			cfg: func() config.Config {
				cfg := startServer(s.T())
				cfg.Controller.PKI.RotationGracePeriod = "1h"
				cfg.Controller.PKI.DeleteAfter = "720h"
				cfg.Controller.Notifications.Enabled = true
				return cfg
			},
			opts: []controller.Option{
				controller.WithMeterProvider(sdkmetric.NewMeterProvider()),
			},
		},
		{
			name: "when the join is in operator mode serves renewals",
			cfg: func() config.Config {
				cfg := startServer(s.T())
				acct, _ := nkeys.CreateAccount()
				seed, _ := acct.Seed()
				cfg.Controller.Join.Account, _ = acct.PublicKey()
				cfg.Controller.Join.SigningKeyFile = filepath.Join(s.T().TempDir(), "signing.nk")
				s.Require().NoError(os.WriteFile(cfg.Controller.Join.SigningKeyFile, seed, 0o600))
				return cfg
			},
		},
		{
			name: "when the join is through a server's auth callout serves it",
			cfg: func() config.Config {
				cfg := startServer(s.T())
				acct, _ := nkeys.CreateAccount()
				seed, _ := acct.Seed()
				cfg.Controller.Join.Callout.IssuerKeyFile = filepath.Join(
					s.T().TempDir(),
					"issuer.nk",
				)
				cfg.Controller.Join.Callout.Account = "OSAPI"
				cfg.Controller.Join.Callout.Auth = config.NATSAuth{
					Type: "user_pass", Username: "callout", Password: "pw",
				}
				s.Require().
					NoError(os.WriteFile(cfg.Controller.Join.Callout.IssuerKeyFile, seed, 0o600))
				return cfg
			},
		},
		{
			name: "when the issuer key of the auth callout does not load does not start",
			cfg: func() config.Config {
				cfg := startServer(s.T())
				cfg.Controller.Join.Callout.IssuerKeyFile = "/no/such/key"
				return cfg
			},
			wantErr: "auth callout: read the issuer key",
		},
		{
			name: "when the join's signing key does not load does not start",
			cfg: func() config.Config {
				cfg := startServer(s.T())
				cfg.Controller.Join.SigningKeyFile = "/no/such/key"
				return cfg
			},
			wantErr: "join over NATS: read the signing key",
		},
		{
			name: "when the enrollment bucket fails does not start",
			cfg: func() config.Config {
				cfg := startServer(s.T())
				cfg.NATS.Enrollment.Bucket = "bad name"
				return cfg
			},
			wantErr: "create the enrollment KV bucket",
		},
		{
			name:       "when a vault is configured opens it",
			cfg:        func() config.Config { return withVault(s.T(), startServer(s.T())) },
			wantVaults: []string{"default"},
		},
		{
			name: "when a signing key is a reference resolves it from the vault",
			cfg: func() config.Config {
				cfg := withSecret(s.T(), withVault(s.T(), startServer(s.T())), "jwt",
					"0123456789abcdef0123456789abcdef")
				cfg.Controller.API.Security.SigningKeys[0].Key = "secret://default/jwt"
				return cfg
			},
			wantVaults: []string{"default"},
		},
		{
			name: "when a signing key is a reference of a key that is not there returns the error",
			cfg: func() config.Config {
				cfg := withVault(s.T(), startServer(s.T()))
				cfg.Controller.API.Security.SigningKeys[0].Key = "secret://default/none"
				return cfg
			},
			wantErr: `resolve the references of the config: controller.api.security.signing_keys[0].key: secret://default/none: the vault "default" has no key "none"`,
		},
		{
			name: "when there is no signing key and its file cannot be read returns the error",
			cfg: func() config.Config {
				cfg := withVault(s.T(), startServer(s.T()))
				cfg.Controller.API.Security.SigningKeys = nil
				cfg.Controller.PKI.KeyDir = s.T().TempDir()
				s.Require().
					NoError(os.Mkdir(filepath.Join(cfg.Controller.PKI.KeyDir, config.SigningKeyFile), 0o700))
				return cfg
			},
			wantErr: "read the signing key",
		},
		{
			name: "when the value of a signing key breaks its rule returns the error",
			cfg: func() config.Config {
				cfg := withSecret(s.T(), withVault(s.T(), startServer(s.T())), "jwt", "short")
				cfg.Controller.API.Security.SigningKeys[0].Key = "secret://default/jwt"
				return cfg
			},
			wantErr: `the value of secret://default/jwt breaks its rule "signing_key"`,
		},
		{
			name: "when the key of a vault is not there makes it, and opens the vault",
			cfg: func() config.Config {
				cfg := withVault(s.T(), startServer(s.T()))
				cfg.Controller.Secrets.Vaults[0].KeyFile = filepath.Join(s.T().TempDir(), "none")
				return cfg
			},
			wantVaults: []string{"default"},
		},
		{
			name: "when the key of a vault cannot be made returns the error",
			cfg: func() config.Config {
				cfg := withVault(s.T(), startServer(s.T()))
				blocked := filepath.Join(s.T().TempDir(), "file")
				s.Require().NoError(os.WriteFile(blocked, nil, 0o600))
				cfg.Controller.Secrets.Vaults[0].KeyFile = filepath.Join(blocked, "secrets.key")
				return cfg
			},
			wantErr: `vault "default": make its key: `,
		},
		{
			name: "when no secrets bucket is named and a vault needs it returns the error",
			cfg: func() config.Config {
				cfg := withVault(s.T(), startServer(s.T()))
				cfg.NATS.Secrets.Bucket = ""
				return cfg
			},
			wantErr: `vault "default": nats.secrets.bucket is empty`,
		},
		{
			name: "when nothing listens returns the error",
			cfg: func() config.Config {
				var cfg config.Config
				cfg.Controller.NATS = config.NATSConnection{
					Host: "127.0.0.1",
					Port: freePort(s.T()),
					Auth: config.NATSAuth{Type: "none"},
				}
				return cfg
			},
			wantErr: "connect to NATS: ",
		},
		{
			name: "when the audit stream fails returns the error",
			cfg: func() config.Config {
				cfg := startServer(s.T())
				cfg.NATS.Audit.Stream = "bad name"
				return cfg
			},
			wantErr: "create audit stream: ",
		},
		{
			name: "when the API's certificate does not load returns the error",
			cfg: func() config.Config {
				cfg := startServer(s.T())
				cfg.Controller.API.TLS = config.ServerTLS{
					Enabled: true, CertFile: "/none.crt", KeyFile: "/none.key",
				}
				return cfg
			},
			wantErr: "load the API's certificate: ",
		},
		{
			name: "when the keypair does not load returns the error",
			cfg: func() config.Config {
				cfg := startServer(s.T())
				blocked := filepath.Join(s.T().TempDir(), "file")
				s.Require().NoError(os.WriteFile(blocked, nil, 0o600))
				cfg.Controller.PKI.KeyDir = filepath.Join(blocked, "keys")
				return cfg
			},
			wantErr: "load or generate controller keypair: ",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			c, err := controller.New(ctx, slog.Default(), tc.cfg(), tc.opts...)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)
				s.Nil(c)
				return
			}
			s.Require().NoError(err)
			s.Require().NoError(c.Start())
			s.NoError(c.CheckHealth(context.Background()))
			s.Contains(c.SubComponents(), "controller.api")
			s.Equal(tc.wantVaults, c.Vaults().Names())
			c.Stop(context.Background())
		})
	}
}

func (s *ControllerPublicTestSuite) TestRegisterTargetValidator() {
	tests := []struct {
		name   string
		agents []job.AgentInfo
		err    error
		want   string
	}{
		{
			name: "when the agents answer resolves a hostname to its machine ID",
			agents: []job.AgentInfo{
				{MachineID: "machine-1", Hostname: "web-01", State: "Ready", Verified: true},
			},
			want: "machine-1",
		},
		{
			name: "when the agents fail leaves the target as it is",
			err:  errBoom,
			want: "web-01",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobs := jobmocks.NewMockJobClient(gomock.NewController(s.T()))
			jobs.EXPECT().ListAgents(gomock.Any()).Return(tc.agents, tc.err)

			controller.ExportRegisterTargetValidator(jobs)

			s.Equal(tc.want, validation.ResolveTarget(context.Background(), "web-01"))
		})
	}
}

func (s *ControllerPublicTestSuite) TestBackground() {
	s.Run("when the function fails logs the error", func() {
		done := make(chan struct{})

		controller.ExportBackground(
			context.Background(),
			slog.Default(),
			"test watcher",
			func(context.Context) error {
				defer close(done)
				return errBoom
			},
		)

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			s.Fail("the function did not run")
		}
	})
}
