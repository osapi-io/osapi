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

package client_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus/server"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type ConnectPublicTestSuite struct {
	suite.Suite
}

func TestConnectPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ConnectPublicTestSuite))
}

// freePort returns a port that nothing listens on now.
func (s *ConnectPublicTestSuite) freePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	defer func() { _ = l.Close() }()

	return l.Addr().(*net.TCPAddr).Port
}

// natsConfig names each bucket that Connect opens.
func natsConfig(
	sc config.NATSServer,
) config.NATS {
	return config.NATS{
		Server: sc,
		Stream: config.NATSStream{
			Name:     "JOBS",
			Subjects: "jobs.>",
			Storage:  "memory",
			Replicas: 1,
		},
		KV:       config.NATSKV{Bucket: "job-queue"},
		DLQ:      config.NATSDLQ{Storage: "memory", Replicas: 1},
		Registry: config.NATSRegistry{Bucket: "agent-registry"},
		Facts:    config.NATSFacts{Bucket: "agent-facts"},
		State:    config.NATSState{Bucket: "agent-state"},
	}
}

// startServer runs an embedded server and returns how to connect to it.
func (s *ConnectPublicTestSuite) startServer() (config.NATS, config.NATSConnection) {
	cfg := natsConfig(config.NATSServer{
		Host:      "127.0.0.1",
		Port:      s.freePort(),
		StoreDir:  s.T().TempDir(),
		Namespace: "test",
		Auth:      config.NATSServerAuth{Type: "none"},
	})
	srv, err := server.New(context.Background(), slog.Default(), cfg, nil)
	s.Require().NoError(err)
	s.T().Cleanup(func() { srv.Stop(context.Background()) })

	return cfg, config.NATSConnection{
		Host:       "127.0.0.1",
		Port:       cfg.Server.Port,
		Namespace:  "test",
		ClientName: "test",
		Auth:       config.NATSAuth{Type: "none"},
	}
}

// TestConnectKeepsTheBucketsSettings: a process that connects with its own
// config does not change a bucket that exists. The agent connects with no
// TTL and the controller with one, and an update from each flipped the job
// queue's age, which removed every job older than the shorter one.
func (s *ConnectPublicTestSuite) TestConnectKeepsTheBucketsSettings() {
	// The server made the buckets from its config, with no TTL.
	cfg, conn := s.startServer()
	cfg.KV.TTL = "1h"
	cfg.Facts.Bucket = "new-facts"
	c, err := client.Connect(context.Background(), slog.Default(), conn, cfg, time.Second, nil, nil)
	s.Require().NoError(err)
	defer c.NATS.Close()

	jobs, err := c.NATS.KVStatus(context.Background(), c.JobsBucket)
	s.Require().NoError(err)
	s.Equal(time.Duration(0), jobs.TTL(), "a connection with another TTL did not change the bucket")
	_, err = c.NATS.KVStatus(context.Background(), "test-new-facts")
	s.NoError(err, "a bucket that was missing is created")
}

func (s *ConnectPublicTestSuite) TestConnect() {
	tests := []struct {
		name      string
		setup     func() (config.NATS, config.NATSConnection)
		wantFacts bool
		wantErr   string
	}{
		{
			name:      "when the server runs opens each bucket and makes a job client",
			setup:     s.startServer,
			wantFacts: true,
		},
		{
			name: "when the config names no facts or state bucket leaves them nil",
			setup: func() (config.NATS, config.NATSConnection) {
				cfg, conn := s.startServer()
				cfg.Facts.Bucket = ""
				cfg.State.Bucket = ""
				return cfg, conn
			},
		},
		{
			name: "when nothing listens returns the error",
			setup: func() (config.NATS, config.NATSConnection) {
				return natsConfig(config.NATSServer{}), config.NATSConnection{
					Host: "127.0.0.1",
					Port: s.freePort(),
					Auth: config.NATSAuth{Type: "none"},
				}
			},
			wantErr: "connect to NATS: ",
		},
		{
			name: "when the connection's TLS does not load returns the error",
			setup: func() (config.NATS, config.NATSConnection) {
				return natsConfig(config.NATSServer{}), config.NATSConnection{
					Host: "127.0.0.1",
					Port: s.freePort(),
					Auth: config.NATSAuth{Type: "none"},
					TLS:  config.ClientTLS{Enabled: true, CAFile: "/no/such/ca.crt"},
				}
			},
			wantErr: "nats tls: ",
		},
		{
			name: "when the job client fails returns the error",
			setup: func() (config.NATS, config.NATSConnection) {
				client.SetNewFn(
					func(*slog.Logger, client.NATSClient, *client.Options) (*client.Client, error) {
						return nil, errors.New("no client")
					},
				)
				s.T().Cleanup(client.ResetNewFn)
				return s.startServer()
			},
			wantErr: "create job client: no client",
		},
		{
			name: "when a bucket fails returns the error",
			setup: func() (config.NATS, config.NATSConnection) {
				cfg, conn := s.startServer()
				cfg.Registry.Bucket = "bad name"
				return cfg, conn
			},
			wantErr: "KV bucket test-bad name: ",
		},
		{
			name: "when a missing bucket cannot be created returns the error",
			setup: func() (config.NATS, config.NATSConnection) {
				cfg, conn := s.startServer()
				cfg.Facts.Bucket = "unmade"
				cfg.Facts.Replicas = 5
				return cfg, conn
			},
			wantErr: "create KV bucket test-unmade: ",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cfg, conn := tc.setup()

			c, err := client.Connect(
				context.Background(),
				slog.Default(),
				conn,
				cfg,
				time.Second,
				nil,
				nil,
			)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)
				s.Nil(c)
				return
			}
			s.Require().NoError(err)
			defer c.NATS.Close()
			s.NotNil(c.Jobs)
			s.NotEmpty(c.JobsBucket)
			s.NotEmpty(c.RegistryBucket)
			s.Equal(tc.wantFacts, c.FactsBucket != "")
			s.Equal(tc.wantFacts, c.StateBucket != "")
		})
	}
}

// TestAFullJobStoreRemovesTheOldestFinishedJobs: a jobs bucket at its
// max_bytes refuses each new write, until the 24h TTL removes old jobs.
// The controller now purges the oldest finished jobs and writes again, so
// it keeps taking jobs. Before, the job after the bucket filled failed.
func (s *ConnectPublicTestSuite) TestAFullJobStoreRemovesTheOldestFinishedJobs() {
	cfg := natsConfig(config.NATSServer{
		Host:     "127.0.0.1",
		Port:     s.freePort(),
		StoreDir: s.T().TempDir(),
		Auth:     config.NATSServerAuth{Type: "none"},
	})
	cfg.KV.MaxBytes = 16 * 1024
	srv, err := server.New(context.Background(), slog.Default(), cfg, nil)
	s.Require().NoError(err)
	s.T().Cleanup(func() { srv.Stop(context.Background()) })
	conn := config.NATSConnection{
		Host:       "127.0.0.1",
		Port:       cfg.Server.Port,
		ClientName: "test",
		Auth:       config.NATSAuth{Type: "none"},
	}
	c, err := client.Connect(context.Background(), slog.Default(), conn, cfg, time.Second, nil, nil)
	s.Require().NoError(err)
	defer c.NATS.Close()
	trust(c.Jobs, nil)

	ctx := context.Background()
	done := map[string]bool{}
	for i := range 200 {
		_, err := c.Jobs.CreateJob(ctx, types.JobTypeQuery, map[string]interface{}{
			"type": "node.hostname.get",
			"data": map[string]interface{}{"padding": strings.Repeat("x", 256)},
		}, "web-01")
		s.Require().NoError(err, "job %d", i)
		// The agent finished each job.
		keys, err := c.NATS.KVKeys(ctx, c.JobsBucket, "jobs.>")
		s.Require().NoError(err)
		for _, key := range keys {
			id := strings.TrimPrefix(key, "jobs.")
			if done[id] {
				continue
			}
			done[id] = true
			_, err = c.NATS.KVPut(ctx, c.JobsBucket,
				fmt.Sprintf("status.%s.completed.web-01.%d", id, i), []byte("{}"))
			if err != nil {
				s.Require().ErrorContains(err, "maximum bytes exceeded")
			}
		}
	}

	status, err := c.NATS.KVStatus(ctx, c.JobsBucket)
	s.Require().NoError(err)
	s.LessOrEqual(status.Bytes(), uint64(cfg.KV.MaxBytes))
}

// TestPurgeOfAMissingStream: a purge names the stream it could not find.
func (s *ConnectPublicTestSuite) TestPurgeOfAMissingStream() {
	cfg, conn := s.startServer()
	c, err := client.Connect(context.Background(), slog.Default(), conn, cfg, time.Second, nil, nil)
	s.Require().NoError(err)
	defer c.NATS.Close()

	js, err := c.NATS.JetStream()
	s.Require().NoError(err)
	err = client.PurgeSubject(context.Background(), js, "KV_missing", "$KV.missing.>")
	s.ErrorContains(err, "stream KV_missing")
}
