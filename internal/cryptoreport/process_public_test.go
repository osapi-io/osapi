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

package cryptoreport_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/cryptoreport"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type ProcessPublicTestSuite struct {
	suite.Suite
}

func (s *ProcessPublicTestSuite) SetupTest() {
	cryptoreport.SetMode(types.FIPSOn)
}

func (s *ProcessPublicTestSuite) TearDownTest() {
	cryptoreport.ResetMode()
}

// names are the names of the uses and of the listeners of r.
func names(
	r cryptoreport.Report,
) ([]string, []string) {
	uses := make([]string, 0, len(r.Algorithms))
	listeners := make([]string, 0, len(r.Listeners))
	for _, a := range r.Algorithms {
		uses = append(uses, a.Name)
	}
	for _, l := range r.Listeners {
		listeners = append(listeners, l.Name)
	}

	return uses, listeners
}

func (s *ProcessPublicTestSuite) TestProcesses() {
	yes := true
	cert := []pki.Cert{
		{Name: "api", State: pki.CertOK, Key: "ECDSA P-256", Signature: "ECDSA-SHA256", FIPS: &yes},
	}
	on := config.ServerTLS{Enabled: true}
	tests := []struct {
		name          string
		report        func() cryptoreport.Report
		wantUses      []string
		wantListeners []string
		wantTLS       []bool
	}{
		{
			name: "when the controller has a nats vault, a metrics server and TLS",
			report: func() cryptoreport.Report {
				var cfg config.Config
				cfg.Controller.API.TLS = on
				cfg.Controller.Metrics = config.MetricsServer{Enabled: true}
				cfg.Controller.Secrets.Vaults = []vault.Spec{
					{Name: "other", Type: "file"},
					{Name: "default", Type: "nats"},
				}

				return cryptoreport.Controller(cfg, cert)
			},
			wantUses: []string{
				"API tokens", "job, join and admission signatures", "secrets sealed to agents",
				"controller vault", "NATS operator, account and user JWTs", "certificate api",
			},
			wantListeners: []string{"api", "metrics"},
			wantTLS:       []bool{true, false},
		},
		{
			name: "when the controller has no vault and no metrics server",
			report: func() cryptoreport.Report {
				return cryptoreport.Controller(config.Config{}, nil)
			},
			wantUses: []string{
				"API tokens", "job, join and admission signatures", "secrets sealed to agents",
				"NATS operator, account and user JWTs",
			},
			wantListeners: []string{"api"},
			wantTLS:       []bool{false},
		},
		{
			name: "when a key encrypts the NATS store",
			report: func() cryptoreport.Report {
				return cryptoreport.NATS(config.NATSServer{
					TLS:       on,
					JetStream: config.NATSServerJetStream{KeyFile: "/etc/osapi/jetstream.key"},
				}, nil)
			},
			wantUses:      []string{"NATS operator, account and user JWTs", "JetStream store"},
			wantListeners: []string{"nats"},
			wantTLS:       []bool{true},
		},
		{
			name: "when no key encrypts the NATS store, it is no use",
			report: func() cryptoreport.Report {
				return cryptoreport.NATS(config.NATSServer{}, nil)
			},
			wantUses:      []string{"NATS operator, account and user JWTs"},
			wantListeners: []string{"nats"},
			wantTLS:       []bool{false},
		},
		{
			name: "when the agent serves metrics with TLS",
			report: func() cryptoreport.Report {
				var cfg config.Config
				cfg.Agent.Metrics = config.MetricsServer{Enabled: true, TLS: on}

				return cryptoreport.Agent(cfg, nil)
			},
			wantUses: []string{
				"job, join and admission signatures", "secrets sealed to agents",
				"NATS operator, account and user JWTs",
			},
			wantListeners: []string{"metrics"},
			wantTLS:       []bool{true},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			r := tc.report()
			uses, listeners := names(r)
			s.Equal(tc.wantUses, uses)
			s.Equal(tc.wantListeners, listeners)
			for i, want := range tc.wantTLS {
				s.Equal(want, r.Listeners[i].TLS, r.Listeners[i].Name)
			}
		})
	}
}

func TestProcessPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProcessPublicTestSuite))
}
