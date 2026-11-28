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
	"bytes"
	"context"
	"log/slog"
	"testing"

	healthEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/health"
	pkiEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/pki"
	"github.com/osapi-io/osapi/internal/cryptoreport"
	"github.com/osapi-io/osapi/internal/telemetry/traces"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	auditmocks "github.com/osapi-io/osapi/internal/audit/mocks"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/controller/mocks"
	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type HandlersPublicTestSuite struct {
	suite.Suite
}

func TestHandlersPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HandlersPublicTestSuite))
}

func (s *HandlersPublicTestSuite) TestSecretsDeps() {
	s.Nil(controller.ExportSecretsDeps(nil).Vaults, "no vault set is no vaults")

	set, err := vault.Open(context.Background(), vault.Env{}, nil)
	s.Require().NoError(err)
	s.Equal(set, controller.ExportSecretsDeps(set).Vaults)
}

func (s *HandlersPublicTestSuite) TestSecretDeps() {
	s.Nil(controller.ExportSecretDeps(nil).Vaults, "no set is a nil interface")
	set, err := vault.Open(context.Background(), vault.Env{}, nil)
	s.Require().NoError(err)
	s.Equal(set, controller.ExportSecretDeps(set).Vaults)
}

func (s *HandlersPublicTestSuite) TestPKIDeps() {
	var cfg config.Config
	cfg.Controller.PKI.KeyDir = s.T().TempDir()
	cfg.NATS.Server.KeyDir = s.T().TempDir()
	report := &cryptoreport.Report{Mode: types.FIPSOff}
	d := controller.ExportPKIDeps(cfg, func(context.Context) healthEndpoint.StatusResponse {
		return healthEndpoint.StatusResponse{Registry: []healthEndpoint.ComponentEntry{
			{Type: "controller", Hostname: "hub", Crypto: report},
			{Type: "nats", Hostname: "hub"},
			{Type: "agent", Hostname: "web-01"},
		}}
	})
	s.Equal([]pkiEndpoint.Process{
		{Component: pkiEndpoint.ComponentController, Hostname: "hub", Crypto: report},
		{Component: pkiEndpoint.ComponentNATS, Hostname: "hub"},
	}, d.Processes(context.Background()))

	s.NotEmpty(d.Hostname)
	got := d.Local()
	names := make([]string, 0, len(got))
	for _, c := range got {
		names = append(names, c.Name)
		s.Equal("not made", c.State, "a key directory with no certificate in it")
	}
	s.Equal([]string{"api", "controller-metrics", "nats", "nats-metrics"}, names,
		"the controller's certificates and the NATS server's, not the agent's")
}

func (s *HandlersPublicTestSuite) TestTraceReader() {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, nil))
	s.Nil(controller.ExportTraceReader(log, config.TracingConfig{}), "no URL is no backend")
	s.IsType(
		&traces.Jaeger{},
		controller.ExportTraceReader(log, config.TracingConfig{URL: "http://jaeger:16686"}),
	)

	s.IsType(&traces.Jaeger{}, controller.ExportTraceReader(log, config.TracingConfig{
		URL: "https://jaeger:16686",
		TLS: config.ClientTLS{InsecureSkipVerify: true},
	}), "with TLS of its own")

	s.Nil(controller.ExportTraceReader(log, config.TracingConfig{
		URL: "https://jaeger:16686",
		TLS: config.ClientTLS{CAFile: "/no/such/ca.crt"},
	}), "TLS that does not load is no backend")
	s.Contains(buf.String(), "read traces: telemetry.tracing.tls")
}

func (s *HandlersPublicTestSuite) TestHandlers() {
	ctrl := gomock.NewController(s.T())
	var cfg config.Config
	cfg.Controller.UI.Enabled = true
	cfg.Controller.API.Security.SigningKeys = []config.SigningKey{
		{ID: "test", Key: "test-key-of-at-least-32-characters"},
	}
	cfg.Controller.API.Security.Roles = map[types.Role]config.CustomRole{
		"reader": {Permissions: []types.Permission{"status:read"}},
	}
	base := len(controller.ExportHandlers(slog.Default(), cfg, nil, nil, ""))

	tests := []struct {
		name  string
		cfg   func() config.Config
		audit bool
		files bool
		want  int
	}{
		{
			name:  "when the audit store is there the endpoint groups serve it, with no handler of its own",
			cfg:   func() config.Config { return cfg },
			audit: true,
			want:  base,
		},
		{
			name:  "when the Object Store is there the endpoint groups serve it, with no handler of its own",
			cfg:   func() config.Config { return cfg },
			files: true,
			want:  base,
		},
		{
			name: "when the UI is off leaves the UI handler out",
			cfg: func() config.Config {
				c := cfg
				c.Controller.UI.Enabled = false
				return c
			},
			want: base - 1,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var store *auditmocks.MockStore
			if tc.audit {
				store = auditmocks.NewMockStore(ctrl)
			}
			var objStore *mocks.MockFileBuckets
			if tc.files {
				objStore = mocks.NewMockFileBuckets(ctrl)
			}

			var got int
			switch {
			case tc.audit:
				got = len(controller.ExportHandlers(slog.Default(), tc.cfg(), store, nil, ""))
			case tc.files:
				got = len(
					controller.ExportHandlers(slog.Default(), tc.cfg(), nil, objStore, "objects"),
				)
			default:
				got = len(controller.ExportHandlers(slog.Default(), tc.cfg(), nil, nil, ""))
			}

			s.Equal(tc.want, got)
		})
	}
}
