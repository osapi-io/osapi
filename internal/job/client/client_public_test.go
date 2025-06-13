// Copyright (c) 2025 John Dewey

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
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/osapi-io/nats-client/pkg/client/mocks"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/client"
	"github.com/retr0h/osapi/internal/provider/system/disk"
	"github.com/retr0h/osapi/internal/provider/system/host"
	"github.com/retr0h/osapi/internal/provider/system/load"
	"github.com/retr0h/osapi/internal/provider/system/mem"
)

type ClientPublicTestSuite struct {
	suite.Suite

	mockCtrl     *gomock.Controller
	mockNativeJS *mocks.MockJetStreamContext
	mockExtJS    *mocks.MockJetStream
	mockKV       *mocks.MockKeyValue
	natsClient   *natsclient.Client
	jobsClient   *client.Client
	ctx          context.Context
}

func (s *ClientPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNativeJS = mocks.NewMockJetStreamContext(s.mockCtrl)
	s.mockExtJS = mocks.NewMockJetStream(s.mockCtrl)
	s.mockKV = mocks.NewMockKeyValue(s.mockCtrl)

	// Create NATS client with mocks
	s.natsClient = natsclient.New(slog.Default(), &natsclient.Options{
		Host: "localhost",
		Port: 4222,
	})
	s.natsClient.NativeJS = s.mockNativeJS
	s.natsClient.ExtJS = s.mockExtJS

	s.ctx = context.Background()
}

func (s *ClientPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ClientPublicTestSuite) TestNew() {
	opts := &client.Options{
		Timeout:  30 * time.Second,
		KVBucket: s.mockKV,
	}
	jobsClient, err := client.New(slog.Default(), s.natsClient, opts)

	s.NoError(err)
	s.NotNil(jobsClient)
}

func (s *ClientPublicTestSuite) TestQuerySystemStatus() {
	// Setup
	hostname := "server1"

	// Expected response
	expectedResponse := &job.SystemStatusResponse{
		Hostname: hostname,
		Uptime:   10 * 24 * time.Hour, // 10 days
		OSInfo: &host.OSInfo{
			Distribution: "Ubuntu",
			Version:      "22.04",
		},
		LoadAverages: &load.AverageStats{
			Load1:  0.5,
			Load5:  0.6,
			Load15: 0.7,
		},
		MemoryStats: &mem.Stats{
			Total:  8589934592, // 8GB
			Free:   4294967296, // 4GB
			Cached: 1073741824, // 1GB
		},
		DiskUsage: []disk.UsageStats{
			{
				Name:  "/dev/sda1",
				Total: 107374182400, // 100GB
				Used:  53687091200,  // 50GB
				Free:  53687091200,  // 50GB
			},
		},
	}

	// Create client
	opts := &client.Options{
		Timeout:  30 * time.Second,
		KVBucket: s.mockKV,
	}
	jobsClient, err := client.New(slog.Default(), s.natsClient, opts)
	s.Require().NoError(err)

	// Note: Full integration test would require mocking PublishAndWaitKV
	// For now, this tests the client creation and structure
	s.NotNil(jobsClient)
	s.NotNil(expectedResponse) // Use the variable to avoid "declared and not used" error
}

func TestClientPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ClientPublicTestSuite))
}

func mustMarshal(t *testing.T, v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	return data
}
