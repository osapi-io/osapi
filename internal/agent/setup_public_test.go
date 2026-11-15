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
	"log/slog"
	"net"
	"testing"
	"testing/fstest"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/bus/server"
	"github.com/osapi-io/osapi/internal/config"
	dockerProv "github.com/osapi-io/osapi/internal/provider/container/docker"
	dockermocks "github.com/osapi-io/osapi/internal/provider/container/docker/mocks"
)

type SetupPublicTestSuite struct {
	suite.Suite
}

func TestSetupPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SetupPublicTestSuite))
}

func (s *SetupPublicTestSuite) SetupSubTest() {
	agent.SetDockerNewFn(func() (dockerProv.Provider, error) {
		return nil, errors.New("no docker")
	})
	agent.SetPlatformFns(func() string { return "rhel" }, func() bool { return false })
}

func (s *SetupPublicTestSuite) TearDownSubTest() {
	agent.ResetDockerNewFn()
	agent.ResetPlatformFns()
	agent.ResetEmbeddedFS()
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

// agentConfig starts an embedded server and returns an agent config that
// connects to it.
func agentConfig(
	t *testing.T,
) config.Config {
	var cfg config.Config
	cfg.NATS = config.NATS{
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
			Storage:  "memory",
			Replicas: 1,
		},
		KV:        config.NATSKV{Bucket: "job-queue", ResponseBucket: "job-responses"},
		DLQ:       config.NATSDLQ{Storage: "memory", Replicas: 1},
		Registry:  config.NATSRegistry{Bucket: "agent-registry"},
		Facts:     config.NATSFacts{Bucket: "agent-facts"},
		State:     config.NATSState{Bucket: "agent-state"},
		Objects:   config.NATSObjects{Bucket: "file-objects"},
		FileState: config.NATSFileState{Bucket: "file-state"},
	}
	srv, err := server.New(context.Background(), slog.Default(), cfg.NATS)
	require.NoError(t, err)
	t.Cleanup(func() { srv.Stop(context.Background()) })

	cfg.Agent.NATS = config.NATSConnection{
		Host:       "127.0.0.1",
		Port:       cfg.NATS.Server.Port,
		Namespace:  "test",
		ClientName: "test-agent",
		Auth:       config.NATSAuth{Type: "none"},
	}
	cfg.Agent.Hostname = "test-host"
	cfg.Controller.API.JobTimeout = "5s"

	return cfg
}

func (s *SetupPublicTestSuite) TestSetup() {
	tests := []struct {
		name    string
		cfg     func() config.Config
		setup   func()
		wantErr string
	}{
		{
			name: "when NATS runs makes the agent with its file operations",
			cfg:  func() config.Config { return agentConfig(s.T()) },
		},
		{
			name: "when the host is macOS makes the agent",
			cfg:  func() config.Config { return agentConfig(s.T()) },
			setup: func() {
				agent.SetPlatformFns(
					func() string { return "darwin" },
					func() bool { return false },
				)
			},
		},
		{
			name: "when the templates do not seed makes the agent",
			cfg:  func() config.Config { return agentConfig(s.T()) },
			setup: func() {
				agent.SetEmbeddedFS(fstest.MapFS{})
			},
		},
		{
			name: "when the Object Store does not open makes the agent without it",
			cfg: func() config.Config {
				cfg := agentConfig(s.T())
				cfg.NATS.Objects.Bucket = "bad name"
				return cfg
			},
		},
		{
			name: "when the file-state bucket does not open makes the agent without it",
			cfg: func() config.Config {
				cfg := agentConfig(s.T())
				cfg.NATS.FileState.Bucket = "bad name"
				return cfg
			},
		},
		{
			name: "when the config names no file buckets makes the agent without them",
			cfg: func() config.Config {
				cfg := agentConfig(s.T())
				cfg.NATS.Objects.Bucket = ""
				cfg.NATS.FileState.Bucket = ""
				return cfg
			},
		},
		{
			name: "when nothing listens returns the error",
			cfg: func() config.Config {
				var cfg config.Config
				cfg.Agent.NATS = config.NATSConnection{
					Host: "127.0.0.1",
					Port: freePort(s.T()),
					Auth: config.NATSAuth{Type: "none"},
				}
				return cfg
			},
			wantErr: "connect to NATS: ",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.setup != nil {
				tc.setup()
			}

			a, conn, err := agent.Setup(context.Background(), slog.Default(), memfs.New(), tc.cfg())

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)
				s.Nil(a)
				return
			}
			s.Require().NoError(err)
			defer bus.Close(conn.NATS)
			s.NotNil(a)
			s.Contains(a.SubComponents(), "agent.heartbeat")
		})
	}
}

func (s *SetupPublicTestSuite) TestSubComponents() {
	s.Run("when PKI and metrics are on reports them as ok", func() {
		cfg := agentConfig(s.T())
		cfg.Agent.PKI.Enabled = true
		cfg.Agent.Metrics = config.MetricsServer{Enabled: true, Host: "127.0.0.1", Port: 9091}

		a, conn, err := agent.Setup(context.Background(), slog.Default(), memfs.New(), cfg)
		s.Require().NoError(err)
		defer bus.Close(conn.NATS)

		subs := a.SubComponents()
		s.Equal("ok", subs["agent.pki"].Status)
		s.Equal("ok", subs["agent.metrics"].Status)
		s.Equal("http://127.0.0.1:9091", subs["agent.metrics"].Address)
	})
}

func (s *SetupPublicTestSuite) TestDocker() {
	tests := []struct {
		name     string
		newFn    func(*gomock.Controller) func() (dockerProv.Provider, error)
		wantNone bool
	}{
		{
			name: "when Docker answers returns the client",
			newFn: func(ctrl *gomock.Controller) func() (dockerProv.Provider, error) {
				return func() (dockerProv.Provider, error) {
					d := dockermocks.NewMockProvider(ctrl)
					d.EXPECT().Ping(gomock.Any()).Return(nil)
					return d, nil
				}
			},
		},
		{
			name: "when Docker does not answer returns none",
			newFn: func(ctrl *gomock.Controller) func() (dockerProv.Provider, error) {
				return func() (dockerProv.Provider, error) {
					d := dockermocks.NewMockProvider(ctrl)
					d.EXPECT().Ping(gomock.Any()).Return(errors.New("refused"))
					return d, nil
				}
			},
			wantNone: true,
		},
		{
			name: "when the client is not made returns none",
			newFn: func(*gomock.Controller) func() (dockerProv.Provider, error) {
				return func() (dockerProv.Provider, error) {
					return nil, errors.New("no docker")
				}
			},
			wantNone: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			agent.SetDockerNewFn(tc.newFn(gomock.NewController(s.T())))

			got := agent.ExportDocker(context.Background(), slog.Default())

			s.Equal(tc.wantNone, got == nil)
		})
	}

	s.Run("when the default constructor runs returns a client or none", func() {
		agent.ResetDockerNewFn()

		_ = agent.ExportDocker(context.Background(), slog.Default())
	})
}
