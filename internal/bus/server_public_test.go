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

package bus_test

import (
	"context"
	"log/slog"
	"net"
	"testing"

	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
)

type ServerPublicTestSuite struct {
	suite.Suite
}

func TestServerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ServerPublicTestSuite))
}

// freePort returns a port that nothing listens on now.
func (s *ServerPublicTestSuite) freePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	defer func() { _ = l.Close() }()

	return l.Addr().(*net.TCPAddr).Port
}

func (s *ServerPublicTestSuite) serverNATS() config.NATS {
	c := fullNATS()
	c.Server = config.NATSServer{
		Host:      "127.0.0.1",
		Port:      s.freePort(),
		StoreDir:  s.T().TempDir(),
		Namespace: "test",
		Auth:      config.NATSServerAuth{Type: "none"},
	}

	return c
}

func (s *ServerPublicTestSuite) TestNewServer() {
	tests := []struct {
		name    string
		cfg     func() config.NATS
		wantErr string
	}{
		{
			name: "when the server starts makes the streams and buckets",
			cfg:  s.serverNATS,
		},
		{
			name: "when the auth is not valid returns the error",
			cfg: func() config.NATS {
				c := s.serverNATS()
				c.Server.Auth = config.NATSServerAuth{Type: "operator"}
				return c
			},
			wantErr: "NATS server auth: ",
		},
		{
			name: "when the port is in use returns the error",
			cfg: func() config.NATS {
				c := s.serverNATS()
				l, err := net.Listen("tcp", "127.0.0.1:0")
				s.Require().NoError(err)
				s.T().Cleanup(func() { _ = l.Close() })
				c.Server.Port = l.Addr().(*net.TCPAddr).Port
				return c
			},
			wantErr: "start embedded NATS server: ",
		},
		{
			name: "when the setup fails stops the server and returns the error",
			cfg: func() config.NATS {
				c := s.serverNATS()
				c.Stream.Name = "bad name"
				return c
			},
			wantErr: "set up JetStream: ",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			srv, err := bus.NewServer(context.Background(), slog.Default(), tc.cfg())

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)
				s.Nil(srv)
				return
			}
			s.Require().NoError(err)
			srv.Start()
			srv.Stop(context.Background())
		})
	}
}

func (s *ServerPublicTestSuite) TestPrepare() {
	s.Run("when nothing listens returns the error", func() {
		err := bus.Prepare(context.Background(), slog.Default(), &natsclient.Options{
			Host: "127.0.0.1",
			Port: s.freePort(),
			Auth: natsclient.AuthOptions{AuthType: natsclient.NoAuth},
			Name: "test",
		}, "test", fullNATS())

		s.ErrorContains(err, "connect to NATS: ")
	})
}
