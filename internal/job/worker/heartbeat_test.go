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

package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job/mocks"
	commandMocks "github.com/retr0h/osapi/internal/provider/command/mocks"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/dns/mocks"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/system/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/system/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/system/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/system/mem/mocks"
)

type HeartbeatTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	mockKV        *mocks.MockKeyValue
	worker        *Worker
}

func (s *HeartbeatTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.mockKV = mocks.NewMockKeyValue(s.mockCtrl)

	appConfig := config.Config{
		Job: config.Job{
			Worker: config.JobWorker{
				Labels: map[string]string{"group": "web"},
			},
		},
	}

	s.worker = New(
		afero.NewMemMapFs(),
		appConfig,
		slog.Default(),
		s.mockJobClient,
		"test-stream",
		hostMocks.NewPlainMockProvider(s.mockCtrl),
		diskMocks.NewPlainMockProvider(s.mockCtrl),
		memMocks.NewPlainMockProvider(s.mockCtrl),
		loadMocks.NewPlainMockProvider(s.mockCtrl),
		dnsMocks.NewPlainMockProvider(s.mockCtrl),
		pingMocks.NewPlainMockProvider(s.mockCtrl),
		commandMocks.NewPlainMockProvider(s.mockCtrl),
		s.mockKV,
	)
}

func (s *HeartbeatTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
	marshalJSON = json.Marshal
	heartbeatInterval = 10 * time.Second
}

func (s *HeartbeatTestSuite) TestWriteRegistration() {
	tests := []struct {
		name         string
		setupMock    func()
		teardownMock func()
	}{
		{
			name: "when marshal fails logs warning",
			setupMock: func() {
				marshalJSON = func(_ interface{}) ([]byte, error) {
					return nil, fmt.Errorf("marshal failure")
				}
			},
			teardownMock: func() {
				marshalJSON = json.Marshal
			},
		},
		{
			name: "when Put fails logs warning",
			setupMock: func() {
				s.mockKV.EXPECT().
					Put(gomock.Any(), "workers.test_worker", gomock.Any()).
					Return(uint64(0), errors.New("put failed"))
			},
		},
		{
			name: "when Put succeeds writes registration",
			setupMock: func() {
				s.mockKV.EXPECT().
					Put(gomock.Any(), "workers.test_worker", gomock.Any()).
					Return(uint64(1), nil)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			if tt.teardownMock != nil {
				defer tt.teardownMock()
			}
			s.worker.writeRegistration(context.Background(), "test-worker")
		})
	}
}

func (s *HeartbeatTestSuite) TestDeregister() {
	tests := []struct {
		name      string
		setupMock func()
	}{
		{
			name: "when Delete fails logs warning",
			setupMock: func() {
				s.mockKV.EXPECT().
					Delete(gomock.Any(), "workers.test_worker").
					Return(errors.New("delete failed"))
			},
		},
		{
			name: "when Delete succeeds logs deregistration",
			setupMock: func() {
				s.mockKV.EXPECT().
					Delete(gomock.Any(), "workers.test_worker").
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			s.worker.deregister("test-worker")
		})
	}
}

func (s *HeartbeatTestSuite) TestStartHeartbeatRefresh() {
	tests := []struct {
		name      string
		setupMock func()
	}{
		{
			name: "ticker fires and refreshes registration",
			setupMock: func() {
				// Initial write + at least 1 ticker refresh
				s.mockKV.EXPECT().
					Put(gomock.Any(), "workers.test_worker", gomock.Any()).
					Return(uint64(1), nil).
					MinTimes(2)

				// Deregister on cancel
				s.mockKV.EXPECT().
					Delete(gomock.Any(), "workers.test_worker").
					Return(nil).
					Times(1)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			heartbeatInterval = 10 * time.Millisecond

			ctx, cancel := context.WithCancel(context.Background())
			s.worker.startHeartbeat(ctx, "test-worker")

			// Wait for at least one ticker refresh
			time.Sleep(50 * time.Millisecond)
			cancel()

			// Wait for goroutine to finish
			s.worker.wg.Wait()
		})
	}
}

func (s *HeartbeatTestSuite) TestRegistryKey() {
	tests := []struct {
		name     string
		hostname string
		expected string
	}{
		{
			name:     "simple hostname",
			hostname: "web-01",
			expected: "workers.web_01",
		},
		{
			name:     "hostname with dots",
			hostname: "Johns-MacBook-Pro.local",
			expected: "workers.Johns_MacBook_Pro_local",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := registryKey(tt.hostname)
			s.Equal(tt.expected, result)
		})
	}
}

func TestHeartbeatTestSuite(t *testing.T) {
	suite.Run(t, new(HeartbeatTestSuite))
}
