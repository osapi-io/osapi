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
	"path/filepath"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/osapi-io/gohai/pkg/gohai"
	"github.com/osapi-io/gohai/pkg/gohai/source"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	agentMocks "github.com/osapi-io/osapi/internal/agent/mocks"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/bus/server"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/provider/gohaitest"
	"github.com/osapi-io/osapi/pkg/sdk/types"
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
	agent.SetPlatformFns(func() string { return "rhel" }, func() bool { return false })
}

func (s *SetupPublicTestSuite) TearDownSubTest() {
	agent.ResetPlatformFns()
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
		KV:        config.NATSKV{Bucket: "job-queue"},
		DLQ:       config.NATSDLQ{Storage: "memory", Replicas: 1},
		Registry:  config.NATSRegistry{Bucket: "agent-registry"},
		Facts:     config.NATSFacts{Bucket: "agent-facts"},
		State:     config.NATSState{Bucket: "agent-state"},
		Objects:   config.NATSObjects{Bucket: "file-objects"},
		FileState: config.NATSFileState{Bucket: "file-state"},
	}
	srv, err := server.New(context.Background(), slog.Default(), cfg.NATS, nil)
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
			name: "when the join over NATS fails returns the error",
			cfg: func() config.Config {
				cfg := agentConfig(s.T())
				cfg.Agent.Join = "nats://hub?token=t"
				return cfg
			},
			wantErr: "join over NATS: ",
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

			a, conn, err := agent.Setup(
				context.Background(),
				slog.Default(),
				memfs.New(),
				tc.cfg(),
				nil,
			)

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

func (s *SetupPublicTestSuite) TestInContainer() {
	tests := []struct {
		name  string
		files map[string]string
		want  bool
	}{
		{
			name:  "docker, from /.dockerenv",
			files: map[string]string{"/.dockerenv": ""},
			want:  true,
		},
		{
			name:  "podman, which has no /.dockerenv",
			files: map[string]string{"/proc/1/environ": "PATH=/usr/bin\x00container=podman\x00"},
			want:  true,
		},
		{
			name:  "lxc",
			files: map[string]string{"/proc/1/environ": "container=lxc\x00"},
			want:  true,
		},
		{
			name:  "a host that is not a container",
			files: map[string]string{"/proc/cpuinfo": "processor\t: 0\n"},
			want:  false,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			fs := memfs.New()
			for path, content := range tt.files {
				s.Require().NoError(fs.MkdirAll(filepath.Dir(path), 0o755))
				s.Require().NoError(fs.WriteFile(path, []byte(content), 0o644))
			}
			got := agent.ExportInContainer(source.Sources{
				FS: fs, Exec: gohaitest.Commands(s.T(), nil), Family: "debian",
			})
			s.Equal(tt.want, got)
		})
	}
}

func (s *SetupPublicTestSuite) TestDefaultContainerFn() {
	tests := []struct {
		name string
	}{
		{name: "when it reads the real host it agrees with inContainer"},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			want := agent.ExportInContainer(source.Sources{})
			s.Equal(want, agent.ExportDefaultContainerFn())
		})
	}
}

func (s *SetupPublicTestSuite) TestGohaiFacts() {
	_, err := agent.ExportGohaiFacts(gohai.WithCollectors("no-such-collector"))
	s.Error(err)
}

func (s *SetupPublicTestSuite) TestSubComponents() {
	s.Run("when PKI and metrics are on reports them as ok", func() {
		cfg := agentConfig(s.T())
		cfg.Agent.Metrics = config.MetricsServer{
			Enabled: true,
			Host:    "127.0.0.1",
			Port:    9091,
			URL:     "http://vm.example:9091/metrics",
		}

		a, conn, err := agent.Setup(context.Background(), slog.Default(), memfs.New(), cfg, nil)
		s.Require().NoError(err)
		defer bus.Close(conn.NATS)

		subs := a.SubComponents()
		s.Equal("ok", subs["agent.pki"].Status)
		s.Equal("ok", subs["agent.metrics"].Status)
		s.Equal("http://127.0.0.1:9091", subs["agent.metrics"].Address)
		s.Equal("http://vm.example:9091/metrics", subs["agent.metrics"].URL)
	})
}

func (s *SetupPublicTestSuite) TestObject() {
	tests := []struct {
		name      string
		get       error
		info      *jetstream.ObjectInfo
		infoErr   error
		wantType  types.ContentType
		wantErr   error
		wantError string
	}{
		{
			name: "when the upload gave a content type returns it",
			info: &jetstream.ObjectInfo{ObjectMeta: jetstream.ObjectMeta{
				Headers: nats.Header{"Osapi-Content-Type": []string{"template"}},
			}},
			wantType: types.ContentTemplate,
		},
		{
			name:     "when the object has no headers is raw",
			info:     &jetstream.ObjectInfo{},
			wantType: types.ContentRaw,
		},
		{
			name:     "when the info does not read is raw",
			infoErr:  errors.New("boom"),
			wantType: types.ContentRaw,
		},
		{
			name:    "when the store does not have the object is not found",
			get:     jetstream.ErrObjectNotFound,
			wantErr: behavior.ErrNotFound,
		},
		{
			name:      "when the store fails returns its error",
			get:       errors.New("store down"),
			wantError: "store down",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			store := agentMocks.NewMockObjectReader(gomock.NewController(s.T()))
			store.EXPECT().ObjectGetBytes(gomock.Any(), "b", "a").Return([]byte("x"), tc.get)
			if tc.get == nil {
				store.EXPECT().ObjectInfo(gomock.Any(), "b", "a").Return(tc.info, tc.infoErr)
			}

			content, contentType, err := agent.ExportStoreObjects(store, "b").
				Object(context.Background(), "a")

			switch {
			case tc.wantErr != nil:
				s.ErrorIs(err, tc.wantErr)
			case tc.wantError != "":
				s.ErrorContains(err, tc.wantError)
			default:
				s.Require().NoError(err)
				s.Equal("x", string(content))
				s.Equal(tc.wantType, contentType)
			}
		})
	}
}
