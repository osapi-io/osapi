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
	"errors"
	"log/slog"
	"net"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/bus/mocks"
	"github.com/osapi-io/osapi/internal/config"
)

type SetupPublicTestSuite struct {
	suite.Suite
}

func TestSetupPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SetupPublicTestSuite))
}

// fullNATS names every stream and bucket, so Setup makes each one.
func fullNATS() config.NATS {
	return config.NATS{
		Stream: config.NATSStream{
			Name:     "JOBS",
			Subjects: "jobs.>",
			MaxAge:   "1h",
			Storage:  "memory",
			Replicas: 1,
			Discard:  "new",
		},
		KV:        config.NATSKV{Bucket: "job-queue", ResponseBucket: "job-responses"},
		DLQ:       config.NATSDLQ{MaxAge: "24h", Storage: "memory", Replicas: 1},
		Audit:     config.NATSAudit{Stream: "AUDIT", Subject: "audit"},
		Registry:  config.NATSRegistry{Bucket: "agent-registry"},
		Facts:     config.NATSFacts{Bucket: "agent-facts"},
		State:     config.NATSState{Bucket: "agent-state"},
		Objects:   config.NATSObjects{Bucket: "file-objects"},
		FileState: config.NATSFileState{Bucket: "file-state"},
	}
}

func (s *SetupPublicTestSuite) TestSetup() {
	fail := errors.New("boom")

	tests := []struct {
		name        string
		cfg         func() config.NATS
		failStream  string
		failKV      string
		failObjects bool
		wantStreams []string
		wantKVs     []string
		wantObjects int
		wantErr     string
	}{
		{
			name:        "when each bucket has a name makes each stream and bucket",
			cfg:         fullNATS,
			wantStreams: []string{"ns-JOBS", "ns-AUDIT", "ns-JOBS-DLQ"},
			wantKVs: []string{
				"ns-job-queue", "ns-job-responses", "ns-agent-registry",
				"ns-agent-facts", "ns-agent-state", "ns-file-state",
			},
			wantObjects: 1,
		},
		{
			name: "when only the required buckets have a name makes only those",
			cfg: func() config.NATS {
				c := fullNATS()
				c.Audit.Stream = ""
				c.Registry.Bucket = ""
				c.Facts.Bucket = ""
				c.State.Bucket = ""
				c.Objects.Bucket = ""
				c.FileState.Bucket = ""
				c.Stream.Discard = "old"
				return c
			},
			wantStreams: []string{"ns-JOBS", "ns-JOBS-DLQ"},
			wantKVs:     []string{"ns-job-queue", "ns-job-responses"},
		},
		{
			name:       "when the JOBS stream fails returns the error",
			cfg:        fullNATS,
			failStream: "ns-JOBS",
			wantErr:    "create JOBS stream: boom",
		},
		{
			name:    "when a KV bucket fails returns the error",
			cfg:     fullNATS,
			failKV:  "ns-agent-facts",
			wantErr: "create KV bucket ns-agent-facts: boom",
		},
		{
			name:       "when the audit stream fails returns the error",
			cfg:        fullNATS,
			failStream: "ns-AUDIT",
			wantErr:    "create audit stream ns-AUDIT: boom",
		},
		{
			name:        "when the Object Store fails returns the error",
			cfg:         fullNATS,
			failObjects: true,
			wantErr:     "create Object Store bucket ns-file-objects: boom",
		},
		{
			name:       "when the DLQ stream fails returns the error",
			cfg:        fullNATS,
			failStream: "ns-JOBS-DLQ",
			wantErr:    "create DLQ stream: boom",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			js := mocks.NewMockJetStream(ctrl)

			var streams, kvs []string
			objects := 0
			js.EXPECT().
				CreateOrUpdateStreamWithConfig(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, c jetstream.StreamConfig) error {
					if c.Name == tc.failStream {
						return fail
					}
					streams = append(streams, c.Name)
					return nil
				}).
				AnyTimes()
			js.EXPECT().
				CreateOrUpdateKVBucketWithConfig(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, c jetstream.KeyValueConfig) (jetstream.KeyValue, error) {
					if c.Bucket == tc.failKV {
						return nil, fail
					}
					kvs = append(kvs, c.Bucket)
					return nil, nil
				}).
				AnyTimes()
			js.EXPECT().
				CreateOrUpdateObjectStore(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ jetstream.ObjectStoreConfig) (jetstream.ObjectStore, error) {
					if tc.failObjects {
						return nil, fail
					}
					objects++
					return nil, nil
				}).
				AnyTimes()

			_, err := bus.Setup(context.Background(), js, "ns", tc.cfg())

			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)
				return
			}
			s.Require().NoError(err)
			s.Equal(tc.wantStreams, streams)
			s.Equal(tc.wantKVs, kvs)
			s.Equal(tc.wantObjects, objects)
		})
	}
}

func (s *SetupPublicTestSuite) TestPrepare() {
	s.Run("when the server runs makes the streams and buckets", func() {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		s.Require().NoError(err)
		port := l.Addr().(*net.TCPAddr).Port
		s.Require().NoError(l.Close())

		ns, err := natsserver.NewServer(&natsserver.Options{
			Host:      "127.0.0.1",
			Port:      port,
			JetStream: true,
			StoreDir:  s.T().TempDir(),
		})
		s.Require().NoError(err)
		ns.Start()
		defer ns.Shutdown()
		s.Require().True(ns.ReadyForConnections(5 * time.Second))

		err = bus.Prepare(context.Background(), slog.Default(), &natsclient.Options{
			Host: "127.0.0.1",
			Port: port,
			Auth: natsclient.AuthOptions{AuthType: natsclient.NoAuth},
			Name: "test",
		}, "test", fullNATS())

		s.NoError(err)
	})

	s.Run("when nothing listens returns the error", func() {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		s.Require().NoError(err)
		port := l.Addr().(*net.TCPAddr).Port
		s.Require().NoError(l.Close())

		err = bus.Prepare(context.Background(), slog.Default(), &natsclient.Options{
			Host: "127.0.0.1",
			Port: port,
			Auth: natsclient.AuthOptions{AuthType: natsclient.NoAuth},
			Name: "test",
		}, "test", fullNATS())

		s.ErrorContains(err, "connect to NATS: ")
	})
}
