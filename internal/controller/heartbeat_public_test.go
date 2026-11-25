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
	"encoding/json"
	"errors"
	"log/slog"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
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
	s.Run("when the hostname fails registers the controller as unknown", func() {
		controller.SetHostnameFn(func() (string, error) { return "", errors.New("no hostname") })
		defer controller.ResetHostnameFn()
		cfg := startServer(s.T())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c, err := controller.New(ctx, slog.Default(), cfg, controller.WithVersion("v1.2.3-test"))
		s.Require().NoError(err)
		defer c.Stop(context.Background())

		nc, err := bus.Dial(slog.Default(), natsOptions(cfg))
		s.Require().NoError(err)
		defer nc.Close()
		bucket := bus.RegistryKV("test", cfg.NATS.Registry).Bucket
		s.Eventually(func() bool {
			keys, _ := nc.KVKeys(context.Background(), bucket)
			return slices.Contains(keys, "controller.unknown")
		}, 5*time.Second, 50*time.Millisecond)

		// The registration has the build version that WithVersion gave.
		entry, err := nc.KVGet(context.Background(), bucket, "controller.unknown")
		s.Require().NoError(err)
		var reg struct {
			Version string `json:"version"`
		}
		s.Require().NoError(json.Unmarshal(entry.Value(), &reg))
		s.Equal("v1.2.3-test", reg.Version)
	})
}

func (s *HeartbeatPublicTestSuite) TestSubComponents() {
	s.Run("when each part is on reports each one as ok", func() {
		var cfg config.Config
		cfg.Controller.API.Port = 8080
		cfg.Controller.Metrics = config.MetricsServer{Enabled: true, Host: "127.0.0.1", Port: 9090}
		cfg.Controller.UI.Enabled = true
		cfg.Controller.Notifications.Enabled = true
		cfg.Telemetry.Tracing.Enabled = true

		got := controller.ExportSubComponents(cfg)

		s.Equal("http://0.0.0.0:8080", got["controller.api"].Address)
		s.Equal("http://127.0.0.1:9090", got["controller.metrics"].Address)
		for _, name := range []string{"controller.metrics", "controller.pki", "controller.notifier", "controller.tracing", "controller.ui"} {
			s.Equal("ok", got[name].Status, name)
		}
	})

	s.Run("when the API serves HTTPS reports an https address", func() {
		var cfg config.Config
		cfg.Controller.API.Port = 8443
		cfg.Controller.API.TLS.Enabled = true

		s.Equal(
			"https://0.0.0.0:8443",
			controller.ExportSubComponents(cfg)["controller.api"].Address,
		)
	})

	s.Run("when url keys and an api host are set reports them", func() {
		var cfg config.Config
		cfg.Controller.API = config.APIServer{
			Host: "127.0.0.1",
			Port: 8080,
			URL:  "https://hub.example.ts.net/",
		}
		cfg.Controller.UI.URL = "https://hub.example.ts.net/ui"
		cfg.Controller.Metrics = config.MetricsServer{
			Enabled: true,
			Host:    "127.0.0.1",
			Port:    9090,
			URL:     "http://hub.example.ts.net:9090/metrics",
		}
		cfg.Telemetry.Tracing = config.TracingConfig{
			Enabled: true,
			URL:     "http://hub.example.ts.net:16686/",
		}

		got := controller.ExportSubComponents(cfg)

		s.Equal("http://127.0.0.1:8080", got["controller.api"].Address)
		s.Equal("http://127.0.0.1:8080", got["controller.ui"].Address)
		s.Equal("https://hub.example.ts.net/", got["controller.api"].URL)
		s.Equal("https://hub.example.ts.net/ui", got["controller.ui"].URL)
		s.Equal("http://hub.example.ts.net:9090/metrics", got["controller.metrics"].URL)
		s.Equal("http://hub.example.ts.net:16686/", got["controller.tracing"].URL)
		s.Empty(got["controller.tracing"].Address)
		s.Empty(got["controller.pki"].URL)
	})

	s.Run("when converted for the health handler keeps the url", func() {
		got := controller.ExportConvertSubComponents(map[string]job.SubComponentInfo{
			"agent.metrics": {
				Status:  "ok",
				Address: "http://0.0.0.0:9091",
				URL:     "http://vm.example:9091/metrics",
			},
		})

		s.Equal("ok", got["agent.metrics"].Status)
		s.Equal("http://0.0.0.0:9091", got["agent.metrics"].Address)
		s.Equal("http://vm.example:9091/metrics", got["agent.metrics"].URL)
		s.Nil(controller.ExportConvertSubComponents(nil))
	})

	s.Run("when tracing exports over otlp reports the endpoint as its address", func() {
		var cfg config.Config
		cfg.Telemetry.Tracing = config.TracingConfig{
			Enabled:      true,
			Exporter:     "otlp",
			OTLPEndpoint: "127.0.0.1:4317",
		}

		got := controller.ExportSubComponents(cfg)

		s.Equal("grpc://127.0.0.1:4317", got["controller.tracing"].Address)
	})

	s.Run("when tracing exports over otlp with TLS reports grpcs://", func() {
		var cfg config.Config
		cfg.Telemetry.Tracing = config.TracingConfig{
			Enabled:      true,
			Exporter:     "otlp",
			OTLPEndpoint: "127.0.0.1:4317",
			TLS:          config.ClientTLS{Enabled: true},
		}

		got := controller.ExportSubComponents(cfg)

		s.Equal("grpcs://127.0.0.1:4317", got["controller.tracing"].Address)
	})

	s.Run("when tracing exports to stdout or is off reports no address", func() {
		for _, tc := range []config.TracingConfig{
			{Enabled: true, Exporter: "stdout"},
			{Enabled: false, Exporter: "otlp", OTLPEndpoint: "127.0.0.1:4317"},
			{Enabled: true, Exporter: "otlp"},
		} {
			var cfg config.Config
			cfg.Telemetry.Tracing = tc

			s.Empty(controller.ExportSubComponents(cfg)["controller.tracing"].Address, tc)
		}
	})

	s.Run("when the api host is IPv6 brackets it in the address", func() {
		var cfg config.Config
		cfg.Controller.API = config.APIServer{Host: "::1", Port: 8080}

		got := controller.ExportSubComponents(cfg)

		s.Equal("http://[::1]:8080", got["controller.api"].Address)
	})

	s.Run("when the parts are off reports them as disabled", func() {
		var cfg config.Config

		got := controller.ExportSubComponents(cfg)

		for _, name := range []string{"controller.metrics", "controller.notifier", "controller.tracing", "controller.ui"} {
			s.Equal("disabled", got[name].Status, name)
		}
		s.Equal("ok", got["controller.pki"].Status, "PKI is always on")
	})
}

func (s *HeartbeatPublicTestSuite) TestTLSOptions() {
	s.Run("when TLS is off there is no option", func() {
		opts, err := controller.ExportTLSOptions(config.Config{})
		s.Require().NoError(err)
		s.Empty(opts)
	})

	s.Run("when TLS is on the API gets the self-signed certificate", func() {
		var cfg config.Config
		cfg.Controller.API.TLS.Enabled = true
		cfg.Controller.PKI.KeyDir = s.T().TempDir()
		opts, err := controller.ExportTLSOptions(cfg)
		s.Require().NoError(err)
		s.Len(opts, 1)
	})

	s.Run("when the certificate does not load it is an error", func() {
		var cfg config.Config
		cfg.Controller.API.TLS = config.ServerTLS{
			Enabled:  true,
			CertFile: "/none.crt",
			KeyFile:  "/none.key",
		}
		_, err := controller.ExportTLSOptions(cfg)
		s.Error(err)
	})
}
