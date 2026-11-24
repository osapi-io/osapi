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

package heartbeat_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	heartbeatmocks "github.com/osapi-io/osapi/internal/heartbeat/mocks"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/heartbeat"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/process"
	processMocks "github.com/osapi-io/osapi/internal/telemetry/process/mocks"
)

type HeartbeatPublicTestSuite struct {
	suite.Suite

	mockCtrl    *gomock.Controller
	mockKV      *heartbeatmocks.MockRegistryStore
	mockProcess *processMocks.MockProvider
	heartbeat   *heartbeat.Heartbeat
}

func (s *HeartbeatPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockKV = heartbeatmocks.NewMockRegistryStore(s.mockCtrl)
	s.mockProcess = processMocks.NewMockProvider(s.mockCtrl)

	s.heartbeat = heartbeat.New(slog.Default(), heartbeat.Options{
		KV:             s.mockKV,
		RegistryBucket: "registry",
		Hostname:       "web-server-01",
		Version:        "0.1.0",
		Component:      "api",
		Process:        s.mockProcess,
		Interval:       10 * time.Second,
		Thresholds:     process.ConditionThresholds{},
		SubComponents:  nil,
	})
}

func (s *HeartbeatPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
	heartbeat.ResetHeartbeatMarshalFn()
}

func (s *HeartbeatPublicTestSuite) TestWriteRegistration() {
	tests := []struct {
		name      string
		setupMock func()
	}{
		{
			name: "writes registration to KV with correct key and fields",
			setupMock: func() {
				s.mockProcess.EXPECT().
					GetMetrics().
					Return(nil, errors.New("no metrics"))

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "api.web_server_01", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, _ string, data []byte) (uint64, error) {
						var reg job.ComponentRegistration
						s.Require().NoError(json.Unmarshal(data, &reg))
						s.Equal("api", reg.Type)
						s.Equal("web-server-01", reg.Hostname)
						s.Equal("0.1.0", reg.Version)
						return uint64(1), nil
					})
			},
		},
		{
			name: "includes process metrics when provider succeeds",
			setupMock: func() {
				s.mockProcess.EXPECT().
					GetMetrics().
					Return(&process.Metrics{
						CPUPercent: 2.5,
						RSSBytes:   1024 * 1024 * 100,
						Goroutines: 20,
					}, nil)

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "api.web_server_01", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, _ string, data []byte) (uint64, error) {
						var reg job.ComponentRegistration
						s.Require().NoError(json.Unmarshal(data, &reg))
						s.Require().NotNil(reg.Process)
						s.InDelta(2.5, reg.Process.CPUPercent, 0.001)
						s.Equal(int64(1024*1024*100), reg.Process.RSSBytes)
						s.Equal(20, reg.Process.Goroutines)
						return uint64(1), nil
					})
			},
		},
		{
			name: "evaluates process conditions when thresholds are set",
			setupMock: func() {
				heartbeat.SetHeartbeatThresholds(s.heartbeat, process.ConditionThresholds{
					MemoryPressureBytes: 1, // 1 byte threshold — RSS will exceed it
					HighCPUPercent:      0, // disabled
				})
				s.T().Cleanup(func() {
					heartbeat.SetHeartbeatThresholds(s.heartbeat, process.ConditionThresholds{})
				})

				s.mockProcess.EXPECT().
					GetMetrics().
					Return(&process.Metrics{
						CPUPercent: 0,
						RSSBytes:   1024 * 1024 * 100,
						Goroutines: 10,
					}, nil)

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "api.web_server_01", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, _ string, data []byte) (uint64, error) {
						var reg job.ComponentRegistration
						s.Require().NoError(json.Unmarshal(data, &reg))
						s.Require().Len(reg.Conditions, 1)
						s.Equal(types.ConditionProcessMemoryPressure, reg.Conditions[0].Type)
						s.True(reg.Conditions[0].Status)
						return uint64(1), nil
					})
			},
		},
		{
			name: "emits no conditions when thresholds are zero",
			setupMock: func() {
				s.mockProcess.EXPECT().
					GetMetrics().
					Return(&process.Metrics{
						CPUPercent: 99.9,
						RSSBytes:   1024 * 1024 * 1024,
						Goroutines: 100,
					}, nil)

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "api.web_server_01", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, _ string, data []byte) (uint64, error) {
						var reg job.ComponentRegistration
						s.Require().NoError(json.Unmarshal(data, &reg))
						s.Empty(reg.Conditions)
						return uint64(1), nil
					})
			},
		},
		{
			name: "omits process metrics when provider fails",
			setupMock: func() {
				s.mockProcess.EXPECT().
					GetMetrics().
					Return(nil, errors.New("process unavailable"))

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "api.web_server_01", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, _ string, data []byte) (uint64, error) {
						var reg job.ComponentRegistration
						s.Require().NoError(json.Unmarshal(data, &reg))
						s.Nil(reg.Process)
						return uint64(1), nil
					})
			},
		},
		{
			name: "includes sub-components when configured",
			setupMock: func() {
				heartbeat.SetHeartbeatSubComponents(s.heartbeat, map[string]job.SubComponentInfo{
					"api.server":    {Status: "ok", Address: "http://0.0.0.0:8080"},
					"api.heartbeat": {Status: "ok"},
				})
				s.T().Cleanup(func() {
					heartbeat.SetHeartbeatSubComponents(s.heartbeat, nil)
				})

				s.mockProcess.EXPECT().
					GetMetrics().
					Return(nil, errors.New("no metrics"))

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "api.web_server_01", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, _ string, data []byte) (uint64, error) {
						var reg job.ComponentRegistration
						s.Require().NoError(json.Unmarshal(data, &reg))
						s.Require().Len(reg.SubComponents, 2)
						s.Equal("ok", reg.SubComponents["api.server"].Status)
						s.Equal("http://0.0.0.0:8080", reg.SubComponents["api.server"].Address)
						s.Equal("ok", reg.SubComponents["api.heartbeat"].Status)
						s.Empty(reg.SubComponents["api.heartbeat"].Address)
						return uint64(1), nil
					})
			},
		},
		{
			name: "when Put fails logs warning",
			setupMock: func() {
				s.mockProcess.EXPECT().
					GetMetrics().
					Return(nil, errors.New("no metrics"))

				s.mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "api.web_server_01", gomock.Any()).
					Return(uint64(0), errors.New("put failed"))
			},
		},
		{
			name: "when marshal fails logs warning",
			setupMock: func() {
				heartbeat.SetHeartbeatMarshalFn(func(_ any) ([]byte, error) {
					return nil, errors.New("marshal failed")
				})

				s.mockProcess.EXPECT().
					GetMetrics().
					Return(nil, errors.New("no metrics"))
				// Put should not be called when marshal fails.
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			heartbeat.ExportWriteRegistration(context.Background(), s.heartbeat)
		})
	}
}

func (s *HeartbeatPublicTestSuite) TestDeregister() {
	tests := []struct {
		name      string
		setupMock func()
	}{
		{
			name: "when Delete succeeds logs deregistration",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "registry", "api.web_server_01").
					Return(nil)
			},
		},
		{
			name: "when Delete fails logs warning",
			setupMock: func() {
				s.mockKV.EXPECT().
					KVDelete(gomock.Any(), "registry", "api.web_server_01").
					Return(errors.New("delete failed"))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			heartbeat.ExportDeregister(s.heartbeat)
		})
	}
}

func (s *HeartbeatPublicTestSuite) TestStart() {
	tests := []struct {
		name      string
		sleep     time.Duration
		setupMock func(mockKV *heartbeatmocks.MockRegistryStore, mockProcess *processMocks.MockProvider)
	}{
		{
			name:  "deletes key on context cancel",
			sleep: 5 * time.Millisecond,
			setupMock: func(mockKV *heartbeatmocks.MockRegistryStore, mockProcess *processMocks.MockProvider) {
				mockProcess.EXPECT().
					GetMetrics().
					Return(nil, errors.New("no metrics")).
					AnyTimes()

				mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "api.web_server_01", gomock.Any()).
					Return(uint64(1), nil).
					AnyTimes()

				mockKV.EXPECT().
					KVDelete(gomock.Any(), "registry", "api.web_server_01").
					Return(nil).
					Times(1)
			},
		},
		{
			name:  "ticker fires and refreshes registration",
			sleep: 50 * time.Millisecond,
			setupMock: func(mockKV *heartbeatmocks.MockRegistryStore, mockProcess *processMocks.MockProvider) {
				mockProcess.EXPECT().
					GetMetrics().
					Return(nil, errors.New("no metrics")).
					AnyTimes()

				// Initial write + at least 1 ticker refresh.
				mockKV.EXPECT().
					KVPut(gomock.Any(), "registry", "api.web_server_01", gomock.Any()).
					Return(uint64(1), nil).
					MinTimes(2)

				mockKV.EXPECT().
					KVDelete(gomock.Any(), "registry", "api.web_server_01").
					Return(nil).
					Times(1)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()

			mockKV := heartbeatmocks.NewMockRegistryStore(ctrl)
			mockProcess := processMocks.NewMockProvider(ctrl)
			tt.setupMock(mockKV, mockProcess)

			hb := heartbeat.New(slog.Default(), heartbeat.Options{
				KV:             mockKV,
				RegistryBucket: "registry",
				Hostname:       "web-server-01",
				Version:        "0.1.0",
				Component:      "api",
				Process:        mockProcess,
				Interval:       10 * time.Millisecond,
				Thresholds:     process.ConditionThresholds{},
				SubComponents:  nil,
			})

			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() {
				hb.Start(ctx)
				close(done)
			}()

			time.Sleep(tt.sleep)
			cancel()
			<-done
		})
	}
}

func TestHeartbeatPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HeartbeatPublicTestSuite))
}
