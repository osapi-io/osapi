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
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/telemetry/metrics"
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
			srv := metrics.New("127.0.0.1", port, slog.Default())
			tc.validateFunc(srv)
		})
	}
}

func (s *ServerPublicTestSuite) TestStartAndStop() {
	tests := []struct {
		name         string
		validateFunc func(port int)
	}{
		{
			name: "serves metrics endpoint",
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
			validateFunc: func(port int) {
				url := fmt.Sprintf("http://127.0.0.1:%d/unknown", port)

				resp, err := http.Get(url) //nolint:gosec
				s.Require().NoError(err)
				defer func() { _ = resp.Body.Close() }()

				s.Equal(404, resp.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			port := s.getFreePort()
			srv := metrics.New("127.0.0.1", port, slog.Default())
			srv.Start()

			time.Sleep(100 * time.Millisecond)

			tc.validateFunc(port)

			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			srv.Stop(ctx)
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
		wantContains  []string
	}{
		{
			name:          "reports 0 when no readiness func set",
			readinessFunc: nil,
			wantContains:  []string{"osapi_component_up", "} 0"},
		},
		{
			name:          "reports 1 when readiness func returns nil",
			readinessFunc: func() error { return nil },
			wantContains:  []string{"osapi_component_up", "} 1"},
		},
		{
			name:          "reports 0 when readiness func returns error",
			readinessFunc: func() error { return errors.New("fail") },
			wantContains:  []string{"osapi_component_up", "} 0"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			port := s.getFreePort()
			srv := metrics.New("127.0.0.1", port, slog.Default())

			if tc.readinessFunc != nil {
				srv.SetReadinessFunc(tc.readinessFunc)
			}

			srv.Start()
			time.Sleep(100 * time.Millisecond)

			body := scrapeMetrics(port)
			for _, want := range tc.wantContains {
				s.Contains(body, want)
			}

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
		wantContains []string
	}{
		{
			name: "registers gauges for each subsystem",
			subsystems: []metrics.SubsystemStatus{
				{Name: "api", StatusFn: func() string { return "ok" }},
				{Name: "heartbeat", StatusFn: func() string { return "ok" }},
				{Name: "notifier", StatusFn: func() string { return "disabled" }},
			},
			wantContains: []string{
				`subsystem="api"} 1`,
				`subsystem="heartbeat"} 1`,
				`subsystem="notifier"} 0`,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			port := s.getFreePort()
			srv := metrics.New("127.0.0.1", port, slog.Default())
			srv.RegisterSubsystems(tc.subsystems)

			srv.Start()
			time.Sleep(100 * time.Millisecond)

			body := scrapeMetrics(port)
			for _, want := range tc.wantContains {
				s.Contains(body, want)
			}

			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			srv.Stop(ctx)
		})
	}
}

func (s *ServerPublicTestSuite) TestStartListenError() {
	tests := []struct {
		name         string
		validateFunc func()
	}{
		{
			name: "logs error when port is already in use",
			validateFunc: func() {
				// Occupy a port on the same interface the server will bind to.
				l, err := net.Listen("tcp", "127.0.0.1:0")
				s.Require().NoError(err)

				port := l.Addr().(*net.TCPAddr).Port

				srv := metrics.New("127.0.0.1", port, slog.Default())
				s.Require().NotNil(srv)

				srv.Start()
				time.Sleep(100 * time.Millisecond)

				_ = l.Close()
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc()
		})
	}
}

func TestServerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ServerPublicTestSuite))
}
