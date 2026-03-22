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

package ops

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	prometheusExporter "go.opentelemetry.io/otel/exporters/prometheus"
)

type ServerTestSuite struct {
	suite.Suite
}

func (s *ServerTestSuite) getFreePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

func (s *ServerTestSuite) TestNew() {
	tests := []struct {
		name         string
		setup        func()
		teardown     func()
		validateFunc func(*Server)
	}{
		{
			name: "when prometheus exporter fails returns nil",
			setup: func() {
				prometheusNewFn = func(
					_ ...prometheusExporter.Option,
				) (*prometheusExporter.Exporter, error) {
					return nil, errors.New("exporter error")
				}
			},
			teardown: func() {
				prometheusNewFn = prometheusExporter.New
			},
			validateFunc: func(srv *Server) {
				s.Nil(srv)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.setup != nil {
				tc.setup()
			}
			if tc.teardown != nil {
				defer tc.teardown()
			}

			port := s.getFreePort()
			srv := New(port, slog.Default())
			tc.validateFunc(srv)
		})
	}
}

func (s *ServerTestSuite) TestStartListenError() {
	tests := []struct {
		name         string
		validateFunc func()
	}{
		{
			name: "logs error when port is already in use",
			validateFunc: func() {
				port := s.getFreePort()

				// Occupy the port.
				l, err := net.Listen("tcp", ":0")
				s.Require().NoError(err)

				port = l.Addr().(*net.TCPAddr).Port

				srv := New(port, slog.Default())
				s.Require().NotNil(srv)

				srv.Start()
				// Give the goroutine time to fail.
				time.Sleep(100 * time.Millisecond)

				l.Close()
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc()
		})
	}
}

func (s *ServerTestSuite) TestStopErrors() {
	tests := []struct {
		name         string
		validateFunc func()
	}{
		{
			name: "handles shutdown after already stopped",
			validateFunc: func() {
				port := s.getFreePort()
				srv := New(port, slog.Default())
				s.Require().NotNil(srv)

				srv.Start()
				time.Sleep(100 * time.Millisecond)

				ctx, cancel := context.WithTimeout(
					context.Background(),
					5*time.Second,
				)
				defer cancel()

				// First stop succeeds.
				srv.Stop(ctx)

				// Second stop exercises the error paths since
				// the server and meter provider are already shut down.
				srv.Stop(ctx)
			},
		},
		{
			name: "logs error when shutdown cannot close active connections",
			validateFunc: func() {
				port := s.getFreePort()

				// Add a slow handler that holds the connection open.
				srv := New(port, slog.Default())
				s.Require().NotNil(srv)

				srv.httpServer.Handler.(*http.ServeMux).HandleFunc(
					"/slow",
					func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusOK)

						if f, ok := w.(http.Flusher); ok {
							f.Flush()
						}

						time.Sleep(5 * time.Second)
					},
				)

				srv.Start()
				time.Sleep(100 * time.Millisecond)

				// Start a request that holds a connection open.
				go func() {
					//nolint:gosec
					resp, err := http.Get(
						fmt.Sprintf("http://127.0.0.1:%d/slow", port),
					)
					if err == nil {
						resp.Body.Close()
					}
				}()

				time.Sleep(50 * time.Millisecond)

				// Stop with an already-expired context so Shutdown
				// can't wait for the active connection to close.
				ctx, cancel := context.WithDeadline(
					context.Background(),
					time.Now().Add(-time.Second),
				)
				defer cancel()

				srv.Stop(ctx)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc()
		})
	}
}

func TestServerTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ServerTestSuite))
}
