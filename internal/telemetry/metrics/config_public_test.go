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

package metrics_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/telemetry/metrics"
)

type ConfigPublicTestSuite struct {
	suite.Suite
}

// TestFromConfigNoCertificate: a metrics server whose self-signed
// certificate cannot be made does not start, and does not fall back to plain
// HTTP.
func (s *ConfigPublicTestSuite) TestFromConfigNoCertificate() {
	blocked := filepath.Join(s.T().TempDir(), "file")
	s.Require().NoError(os.WriteFile(blocked, nil, 0o600))
	got := metrics.FromConfig(config.MetricsServer{
		Enabled: true, Host: "127.0.0.1", TLS: config.ServerTLS{Enabled: true},
	}, filepath.Join(blocked, "pki"), "metrics", slog.New(slog.DiscardHandler))
	s.Nil(got, "no fall back to plain HTTP")

	got = metrics.FromConfig(config.MetricsServer{
		Enabled: true, Host: "127.0.0.1",
		TLS: config.ServerTLS{Enabled: true, CertFile: "/no/such.crt", KeyFile: "/no/such.key"},
	}, "", "metrics", slog.New(slog.DiscardHandler))
	s.Nil(got, "a certificate that does not load does not start it")
}

func (s *ConfigPublicTestSuite) TestFromConfig() {
	tests := []struct {
		name         string
		cfg          config.MetricsServer
		validateFunc func(*metrics.Server)
	}{
		{
			name: "when the config turns it off, there is no server, and no meter provider",
			cfg:  config.MetricsServer{},
			validateFunc: func(got *metrics.Server) {
				s.Nil(got)
				s.Nil(got.MeterProvider())
			},
		},
		{
			name: "when TLS is on with no certificate, it makes and serves a self-signed one",
			cfg: config.MetricsServer{
				Enabled: true, Host: "127.0.0.1",
				TLS: config.ServerTLS{Enabled: true},
			},
			validateFunc: func(got *metrics.Server) {
				s.NotNil(got)
			},
		},
		{
			name: "when TLS is on with a certificate, it is made",
			cfg: config.MetricsServer{
				Enabled: true, Host: "127.0.0.1",
				TLS: func() config.ServerTLS {
					cert, key := writeCert(s.T())

					return config.ServerTLS{Enabled: true, CertFile: cert, KeyFile: key}
				}(),
			},
			validateFunc: func(got *metrics.Server) {
				s.NotNil(got)
			},
		},
		{
			name: "when the config turns it on, it is made with a meter provider",
			cfg:  config.MetricsServer{Enabled: true, Host: "127.0.0.1", Port: 0},
			validateFunc: func(got *metrics.Server) {
				s.Require().NotNil(got)
				s.NotNil(got.MeterProvider())
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(
				metrics.FromConfig(
					tc.cfg,
					s.T().TempDir(),
					"metrics",
					slog.New(slog.DiscardHandler),
				),
			)
		})
	}
}

func TestConfigPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ConfigPublicTestSuite))
}
