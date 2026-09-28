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

package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
	aptMocks "github.com/osapi-io/osapi/internal/provider/node/apt/mocks"
	diskMocks "github.com/osapi-io/osapi/internal/provider/node/disk/mocks"
	nodeHost "github.com/osapi-io/osapi/internal/provider/node/host"
	hostMocks "github.com/osapi-io/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/osapi-io/osapi/internal/provider/node/load/mocks"
	logMocks "github.com/osapi-io/osapi/internal/provider/node/log/mocks"
	memMocks "github.com/osapi-io/osapi/internal/provider/node/mem/mocks"
	ntpMocks "github.com/osapi-io/osapi/internal/provider/node/ntp/mocks"
	powerMocks "github.com/osapi-io/osapi/internal/provider/node/power/mocks"
	processMocks "github.com/osapi-io/osapi/internal/provider/node/process/mocks"
	serviceMocks "github.com/osapi-io/osapi/internal/provider/node/service/mocks"
	sysctlMocks "github.com/osapi-io/osapi/internal/provider/node/sysctl/mocks"
	timezoneMocks "github.com/osapi-io/osapi/internal/provider/node/timezone/mocks"
	userMocks "github.com/osapi-io/osapi/internal/provider/node/user/mocks"
)

// NodeStatusPublicTestSuite covers a status assembled from six independent reads,
// where one failing must not read as a host with nothing to report.
type NodeStatusPublicTestSuite struct {
	suite.Suite

	mockCtrl *gomock.Controller
	logger   *slog.Logger
}

func (s *NodeStatusPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

// processorWith builds a node processor whose host, disk, memory and load reads
// behave as the test describes.
func (s *NodeStatusPublicTestSuite) processorWith(
	setup func(*hostMocks.MockProvider, *diskMocks.MockProvider, *memMocks.MockProvider, *loadMocks.MockProvider),
) agent.ProcessorFunc {
	hostMock := hostMocks.NewMockProvider(s.mockCtrl)
	diskMock := diskMocks.NewMockProvider(s.mockCtrl)
	memMock := memMocks.NewMockProvider(s.mockCtrl)
	loadMock := loadMocks.NewMockProvider(s.mockCtrl)

	setup(hostMock, diskMock, memMock, loadMock)

	return agent.NewNodeProcessor(
		hostMock,
		diskMock,
		memMock,
		loadMock,
		sysctlMocks.NewMockProvider(s.mockCtrl),
		ntpMocks.NewMockProvider(s.mockCtrl),
		timezoneMocks.NewMockProvider(s.mockCtrl),
		powerMocks.NewMockProvider(s.mockCtrl),
		processMocks.NewMockProvider(s.mockCtrl),
		userMocks.NewMockProvider(s.mockCtrl),
		aptMocks.NewMockProvider(s.mockCtrl),
		logMocks.NewMockProvider(s.mockCtrl),
		serviceMocks.NewMockProvider(s.mockCtrl),
		config.Config{},
		s.logger,
	)
}

func (s *NodeStatusPublicTestSuite) TestGetNodeStatus() {
	statusRequest := job.Request{
		Type:      job.TypeQuery,
		Category:  "node",
		Operation: "status.get",
		Data:      json.RawMessage(`{}`),
	}

	tests := []struct {
		name         string
		setup        func(*hostMocks.MockProvider, *diskMocks.MockProvider, *memMocks.MockProvider, *loadMocks.MockProvider)
		validateFunc func(map[string]any, error)
	}{
		{
			name: "every read succeeds",
			setup: func(
				h *hostMocks.MockProvider,
				d *diskMocks.MockProvider,
				m *memMocks.MockProvider,
				l *loadMocks.MockProvider,
			) {
				h.EXPECT().GetHostname().Return("web-01", nil)
				h.EXPECT().GetOSInfo().Return(&nodeHost.Result{Distribution: "Debian"}, nil)
				h.EXPECT().GetUptime().Return(time.Hour, nil)
				d.EXPECT().GetLocalUsageStats().Return(nil, nil)
				m.EXPECT().GetStats().Return(nil, nil)
				l.EXPECT().GetAverageStats().Return(nil, nil)
			},
			validateFunc: func(response map[string]any, err error) {
				s.Require().NoError(err)
				s.Equal("web-01", response["hostname"])
				s.NotContains(response, "field_errors")
			},
		},
		{
			name: "one read fails and is named",
			setup: func(
				h *hostMocks.MockProvider,
				d *diskMocks.MockProvider,
				m *memMocks.MockProvider,
				l *loadMocks.MockProvider,
			) {
				h.EXPECT().GetHostname().Return("web-01", nil)
				h.EXPECT().GetOSInfo().Return(&nodeHost.Result{Distribution: "Debian"}, nil)
				h.EXPECT().GetUptime().Return(time.Hour, nil)
				d.EXPECT().GetLocalUsageStats().Return(nil, nil)
				m.EXPECT().
					GetStats().
					Return(nil, errors.New("open /proc/meminfo: permission denied"))
				l.EXPECT().GetAverageStats().Return(nil, nil)
			},
			validateFunc: func(response map[string]any, err error) {
				s.Require().NoError(err)

				// The rest of the status still answers, which is why the errors
				// were dropped in the first place — but now it says which.
				s.Equal("web-01", response["hostname"])

				fieldErrors, ok := response["field_errors"].(map[string]any)
				s.Require().True(ok)
				s.Len(fieldErrors, 1)
				s.Contains(fieldErrors["memory_stats"], "permission denied")
			},
		},
		{
			name: "every read fails and every one is named",
			setup: func(
				h *hostMocks.MockProvider,
				d *diskMocks.MockProvider,
				m *memMocks.MockProvider,
				l *loadMocks.MockProvider,
			) {
				h.EXPECT().GetHostname().Return("", errors.New("no hostname"))
				h.EXPECT().GetOSInfo().Return(nil, errors.New("no os-release"))
				h.EXPECT().GetUptime().Return(time.Duration(0), errors.New("no uptime"))
				d.EXPECT().GetLocalUsageStats().Return(nil, errors.New("no disks"))
				m.EXPECT().GetStats().Return(nil, errors.New("no memory"))
				l.EXPECT().GetAverageStats().Return(nil, errors.New("no load"))
			},
			validateFunc: func(response map[string]any, err error) {
				// Still not an error: a host that can report nothing is a
				// finding, and a caller learns more from six named failures than
				// from one rejected request.
				s.Require().NoError(err)

				fieldErrors, ok := response["field_errors"].(map[string]any)
				s.Require().True(ok)
				s.Len(fieldErrors, 6)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			processor := s.processorWith(tt.setup)

			result, err := processor(context.Background(), statusRequest)

			response := map[string]any{}
			if result != nil {
				s.Require().NoError(json.Unmarshal(result, &response))
			}

			tt.validateFunc(response, err)
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestNodeStatusPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(NodeStatusPublicTestSuite))
}
