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
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	prometheusExporter "go.opentelemetry.io/otel/exporters/prometheus"

	"github.com/osapi-io/osapi/internal/telemetry/instruments"
	"github.com/osapi-io/osapi/internal/telemetry/metrics"
)

type ServerPublicTestSuite struct {
	suite.Suite
}

func (s *ServerPublicTestSuite) getFreePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	defer func() { _ = l.Close() }()

	return l.Addr().(*net.TCPAddr).Port
}

func (s *ServerPublicTestSuite) TestNew() {
	tests := []struct {
		name         string
		validateFunc func(*metrics.Server)
	}{
		{
			name: "returns non-nil server",
			validateFunc: func(srv *metrics.Server) {
				s.NotNil(srv)
			},
		},
		{
			name: "has meter provider",
			validateFunc: func(srv *metrics.Server) {
				s.NotNil(srv.MeterProvider())
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			port := s.getFreePort()
			srv, err := metrics.New("127.0.0.1", port, slog.Default())
			s.Require().NoError(err)
			tc.validateFunc(srv)
		})
	}
}

func (s *ServerPublicTestSuite) TestStart() {
	tests := []struct {
		wantErr string
		name    string
		// listen returns the port to serve on and a function releasing
		// anything the row held open.
		listen       func() (int, func())
		validateFunc func(port int)
	}{
		{
			name: "serves metrics endpoint",
			listen: func() (int, func()) {
				return s.getFreePort(), func() {}
			},
			validateFunc: func(port int) {
				url := fmt.Sprintf("http://127.0.0.1:%d/metrics", port)

				resp, err := http.Get(url) //nolint:gosec
				s.Require().NoError(err)
				defer func() { _ = resp.Body.Close() }()

				s.Equal(200, resp.StatusCode)

				body, err := io.ReadAll(resp.Body)
				s.Require().NoError(err)
				s.Contains(string(body), "go_goroutines")
			},
		},
		{
			name: "returns 404 for unknown paths",
			listen: func() (int, func()) {
				return s.getFreePort(), func() {}
			},
			validateFunc: func(port int) {
				url := fmt.Sprintf("http://127.0.0.1:%d/unknown", port)

				resp, err := http.Get(url) //nolint:gosec
				s.Require().NoError(err)
				defer func() { _ = resp.Body.Close() }()

				s.Equal(404, resp.StatusCode)
			},
		},
		{
			name: "a port in use is the error of Start",
			listen: func() (int, func()) {
				// Occupy a port on the same interface the server will bind to.
				l, err := net.Listen("tcp", "127.0.0.1:0")
				s.Require().NoError(err)

				return l.Addr().(*net.TCPAddr).Port, func() { _ = l.Close() }
			},
			wantErr:      "metrics server: listen on 127.0.0.1:",
			validateFunc: func(_ int) {},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			port, release := tc.listen()
			defer release()

			srv, err := metrics.New("127.0.0.1", port, slog.Default())

			s.Require().NoError(err)
			s.Require().NotNil(srv)
			if tc.wantErr != "" {
				s.ErrorContains(srv.Start(), tc.wantErr)

				return
			}
			s.Require().NoError(srv.Start())

			time.Sleep(100 * time.Millisecond)

			tc.validateFunc(port)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			srv.Stop(ctx)
		})
	}
}

func (s *ServerPublicTestSuite) TestStop() {
	tests := []struct {
		name string
		// setup prepares the started server before Stop.
		setup func(srv *metrics.Server, port int)
		// stopCtx is the context Stop is called with.
		stopCtx func() (context.Context, context.CancelFunc)
		// stopTwice calls Stop a second time on the stopped server.
		stopTwice bool
	}{
		{
			name:  "stops a running server",
			setup: func(_ *metrics.Server, _ int) {},
			stopCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
		},
		{
			name:  "handles shutdown after already stopped",
			setup: func(_ *metrics.Server, _ int) {},
			stopCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			// The second stop exercises the error paths, since the server
			// and meter provider are already shut down.
			stopTwice: true,
		},
		{
			name: "logs error when shutdown cannot close active connections",
			setup: func(_ *metrics.Server, port int) {
				// Start a request that holds a connection open.
				go func() {
					//nolint:gosec
					resp, err := http.Get(
						fmt.Sprintf("http://127.0.0.1:%d/slow", port),
					)
					if err == nil {
						_ = resp.Body.Close()
					}
				}()

				time.Sleep(50 * time.Millisecond)
			},
			// An already-expired context, so Shutdown cannot wait for the
			// active connection to close.
			stopCtx: func() (context.Context, context.CancelFunc) {
				return context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			port := s.getFreePort()
			srv, err := metrics.New("127.0.0.1", port, slog.Default())
			s.Require().NoError(err)
			s.Require().NotNil(srv)

			// A slow handler that holds a connection open, for the row
			// that needs one.
			metrics.ExportServerAddRoute(
				srv,
				"/slow",
				func(c *echo.Context) error {
					c.Response().WriteHeader(http.StatusOK)
					// v5's Response() is a bare http.ResponseWriter, so
					// Flush comes from the interface rather than a method
					// on Echo's wrapper.
					if f, ok := c.Response().(http.Flusher); ok {
						f.Flush()
					}

					time.Sleep(5 * time.Second)

					return nil
				},
			)

			s.Require().NoError(srv.Start())
			time.Sleep(100 * time.Millisecond)

			tc.setup(srv, port)

			ctx, cancel := tc.stopCtx()
			defer cancel()

			srv.Stop(ctx)
			if tc.stopTwice {
				srv.Stop(ctx)
			}
		})
	}
}

func (s *ServerPublicTestSuite) TestComponentUpGauge() {
	scrapeMetrics := func(port int) string {
		url := fmt.Sprintf("http://127.0.0.1:%d/metrics", port)

		resp, err := http.Get(url) //nolint:gosec
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)

		return string(body)
	}

	tests := []struct {
		name          string
		readinessFunc func() error
		validateFunc  func(string)
	}{
		{
			name:          "reports 0 when no readiness func set",
			readinessFunc: nil,
			validateFunc: func(body string) {
				for _, want := range []string{"osapi_component_up", "} 0"} {
					s.Contains(body, want)
				}
			},
		},
		{
			name:          "reports 1 when readiness func returns nil",
			readinessFunc: func() error { return nil },
			validateFunc: func(body string) {
				for _, want := range []string{"osapi_component_up", "} 1"} {
					s.Contains(body, want)
				}
			},
		},
		{
			name:          "reports 0 when readiness func returns error",
			readinessFunc: func() error { return errors.New("fail") },
			validateFunc: func(body string) {
				for _, want := range []string{"osapi_component_up", "} 0"} {
					s.Contains(body, want)
				}
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			port := s.getFreePort()
			srv, err := metrics.New("127.0.0.1", port, slog.Default())
			s.Require().NoError(err)

			if tc.readinessFunc != nil {
				srv.SetReadinessFunc(tc.readinessFunc)
			}

			s.Require().NoError(srv.Start())
			time.Sleep(100 * time.Millisecond)

			tc.validateFunc(scrapeMetrics(port))

			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			srv.Stop(ctx)
		})
	}
}

func (s *ServerPublicTestSuite) TestRegisterSubsystems() {
	scrapeMetrics := func(port int) string {
		url := fmt.Sprintf("http://127.0.0.1:%d/metrics", port)

		resp, err := http.Get(url) //nolint:gosec
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)

		return string(body)
	}

	tests := []struct {
		name         string
		subsystems   []metrics.SubsystemStatus
		validateFunc func(string)
	}{
		{
			name: "registers gauges for each subsystem",
			subsystems: []metrics.SubsystemStatus{
				{Name: "api", StatusFn: func() string { return "ok" }},
				{Name: "heartbeat", StatusFn: func() string { return "ok" }},
				{Name: "notifier", StatusFn: func() string { return "disabled" }},
			},
			validateFunc: func(body string) {
				for _, want := range []string{
					`subsystem="api"} 1`,
					`subsystem="heartbeat"} 1`,
					`subsystem="notifier"} 0`,
				} {
					s.Contains(body, want)
				}
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			port := s.getFreePort()
			srv, err := metrics.New("127.0.0.1", port, slog.Default())
			s.Require().NoError(err)
			srv.RegisterSubsystems(tc.subsystems)

			s.Require().NoError(srv.Start())
			time.Sleep(100 * time.Millisecond)

			tc.validateFunc(scrapeMetrics(port))

			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			srv.Stop(ctx)
		})
	}
}

func (s *ServerPublicTestSuite) TestRegisterHeartbeatAge() {
	scrapeMetrics := func(port int) string {
		url := fmt.Sprintf("http://127.0.0.1:%d/metrics", port)

		resp, err := http.Get(url) //nolint:gosec
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)

		return string(body)
	}

	tests := []struct {
		name         string
		timeFn       func() time.Time
		validateFunc func(string)
	}{
		{
			name:   "reports 0 when heartbeat time is zero",
			timeFn: func() time.Time { return time.Time{} },
			validateFunc: func(body string) {
				for _, want := range []string{"osapi_heartbeat_age_seconds", "} 0"} {
					s.Contains(body, want)
				}
			},
		},
		{
			name: "reports positive age when heartbeat time is in the past",
			timeFn: func() time.Time {
				return time.Now().Add(-5 * time.Second)
			},
			validateFunc: func(body string) {
				for _, want := range []string{"osapi_heartbeat_age_seconds"} {
					s.Contains(body, want)
				}
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			port := s.getFreePort()
			srv, err := metrics.New("127.0.0.1", port, slog.Default())
			s.Require().NoError(err)
			srv.RegisterHeartbeatAge(tc.timeFn)

			s.Require().NoError(srv.Start())
			time.Sleep(100 * time.Millisecond)

			tc.validateFunc(scrapeMetrics(port))

			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			srv.Stop(ctx)
		})
	}
}

// TestBehaviorMetricNames scrapes the behavior metrics as an agent exports
// them: the exporter adds the osapi namespace once, so a name that has it
// already gets it twice.
func (s *ServerPublicTestSuite) TestBehaviorMetricNames() {
	port := s.getFreePort()
	srv, err := metrics.New("127.0.0.1", port, slog.Default())
	s.Require().NoError(err)
	meter := srv.MeterProvider().Meter("osapi-behavior")
	for _, name := range []string{instruments.BehaviorCalls.Name, instruments.BehaviorChanges.Name} {
		c, err := meter.Int64Counter(name)
		s.Require().NoError(err)
		c.Add(context.Background(), 1)
	}
	h, err := meter.Float64Histogram(instruments.BehaviorDuration.Name)
	s.Require().NoError(err)
	h.Record(context.Background(), 0.1)
	s.Require().NoError(srv.Start())
	time.Sleep(100 * time.Millisecond)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Stop(ctx)
	}()

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", port)) //nolint:gosec
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	for _, want := range []string{
		"\nosapi_behavior_calls_total{",
		"\nosapi_behavior_changes_total{",
		"\nosapi_behavior_duration_seconds_bucket{",
	} {
		s.Contains(string(body), want)
	}
	s.NotContains(string(body), "osapi_osapi_")
}

func (s *ServerPublicTestSuite) TestNewPrometheusError() {
	tests := []struct {
		name         string
		validateFunc func(*metrics.Server)
	}{
		{
			name: "when the prometheus exporter fails, it is the error",
			validateFunc: func(srv *metrics.Server) {
				s.Nil(srv)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			metrics.SetPrometheusNewFn(func(
				_ ...prometheusExporter.Option,
			) (*prometheusExporter.Exporter, error) {
				return nil, errors.New("exporter error")
			})
			defer metrics.ResetPrometheusNewFn()

			port := s.getFreePort()
			srv, err := metrics.New("127.0.0.1", port, slog.Default())
			s.ErrorContains(err, "create the prometheus exporter: exporter error")
			tc.validateFunc(srv)
		})
	}
}

func TestServerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ServerPublicTestSuite))
}

// lockedBuffer is a bytes.Buffer that the goroutine of the server and the
// test can use at the same time.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(
	p []byte,
) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

func (s *ServerPublicTestSuite) TestStartLogsThroughLogger() {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "writes the server started line through the logger of the server",
			want: []string{
				"msg=\"http(s) server started\"",
				"subsystem=metrics",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			out := &lockedBuffer{}
			logger := slog.New(slog.NewTextHandler(out, nil))

			srv, err := metrics.New("127.0.0.1", s.getFreePort(), logger)

			s.Require().NoError(err)
			s.Require().NotNil(srv)
			s.Require().NoError(srv.Start())

			s.Eventually(func() bool {
				return strings.Contains(out.String(), "http(s) server started")
			}, 2*time.Second, 10*time.Millisecond)

			for _, w := range tt.want {
				s.Contains(out.String(), w)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			srv.Stop(ctx)
		})
	}
}

func (s *ServerPublicTestSuite) TestStartServes() {
	tests := []struct {
		name    string
		tls     *tls.Config
		closeLn bool
		wantLog string
	}{
		{
			name: "with TLS, it serves through a TLS listener",
			tls:  &tls.Config{MinVersion: tls.VersionTLS12},
		},
		{
			name:    "a listener that fails while it serves is logged",
			closeLn: true,
			wantLog: "metrics server error",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var ln net.Listener
			metrics.SetListenFn(func(ctx context.Context, network, _ string) (net.Listener, error) {
				l, err := (&net.ListenConfig{}).Listen(ctx, network, "127.0.0.1:0")
				ln = l

				return l, err
			})
			defer metrics.ResetListenFn()

			var buf lockedBuffer
			log := slog.New(slog.NewTextHandler(&buf, nil))
			logged := func(msg string) func() bool {
				return func() bool { return strings.Contains(buf.String(), msg) }
			}

			srv, err := metrics.New("127.0.0.1", 0, log)

			s.Require().NoError(err)
			s.Require().NotNil(srv)
			if tt.tls != nil {
				metrics.SetServerTLS(srv, tt.tls)
			}
			s.Require().NoError(srv.Start())
			s.Eventually(logged("metrics server started"), 2*time.Second, 10*time.Millisecond)
			if tt.closeLn {
				_ = ln.Close()
			}
			if tt.wantLog != "" {
				s.Eventually(logged(tt.wantLog), 2*time.Second, 10*time.Millisecond)
			}
			srv.Stop(context.Background())
		})
	}
}

// writeCert writes a self-signed certificate for localhost and its key, and
// returns their paths.
func writeCert(
	t *testing.T,
) (string, string) {
	t.Helper()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost"},
	}
	der, _ := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	keyDER, _ := x509.MarshalPKCS8PrivateKey(key)
	dir := t.TempDir()
	cert, keyFile := filepath.Join(dir, "tls.crt"), filepath.Join(dir, "tls.key")
	if err := os.WriteFile(
		cert,
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		keyFile,
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	return cert, keyFile
}
