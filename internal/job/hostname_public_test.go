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

package job_test

import (
	"errors"
	"testing"

	"github.com/osapi-io/osapi/internal/provider/gohaitest"

	"github.com/avfs/avfs/vfs/memfs"

	"github.com/osapi-io/gohai/pkg/gohai/source"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job"
	hostnamemocks "github.com/osapi-io/osapi/internal/job/mocks/hostname"
)

type HostnamePublicTestSuite struct {
	suite.Suite

	mockCtrl *gomock.Controller
}

func (s *HostnamePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
}

func (s *HostnamePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *HostnamePublicTestSuite) TearDownSubTest() {
	job.ResetDefaultHostnameProvider()
}

func (s *HostnamePublicTestSuite) TestGetAgentHostname() {
	tests := []struct {
		name               string
		configuredHostname string
		setup              func()
		validateFunc       func(string, error)
	}{
		{
			name:               "configured hostname takes precedence",
			configuredHostname: "configured-agent",
			setup:              func() {},
			validateFunc: func(hostname string, err error) {
				s.NoError(err)
				s.Equal("configured-agent", hostname)
			},
		},
		{
			name:               "empty configured hostname falls back to system",
			configuredHostname: "",
			setup:              func() {},
			validateFunc: func(hostname string, err error) {
				s.NoError(err)
				s.NotEmpty(hostname, "System hostname should not be empty")
				s.NotEqual(
					"unknown",
					hostname,
					"Should get actual system hostname, not unknown fallback",
				)
			},
		},
		{
			name:               "falls back to unknown when the provider gives no name",
			configuredHostname: "",
			setup: func() {
				mockProvider := hostnamemocks.NewMockHostnameProvider(s.mockCtrl)
				mockProvider.EXPECT().Hostname().Return("", nil)
				job.SetDefaultHostnameProvider(mockProvider)
			},
			validateFunc: func(hostname string, err error) {
				s.ErrorContains(err, "it is empty")
				s.Equal("unknown", hostname)
			},
		},
		{
			name:               "falls back to unknown when provider errors",
			configuredHostname: "",
			setup: func() {
				mockProvider := hostnamemocks.NewMockHostnameProvider(s.mockCtrl)
				mockProvider.EXPECT().Hostname().Return("", errors.New("provider error"))
				job.SetDefaultHostnameProvider(mockProvider)
			},
			validateFunc: func(hostname string, err error) {
				s.ErrorContains(err, "provider error")
				s.Equal("unknown", hostname)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setup()

			tt.validateFunc(job.GetAgentHostname(tt.configuredHostname))
		})
	}
}

func (s *HostnamePublicTestSuite) TestGohaiHostnameProvider() {
	hostname, err := job.ExportNewGohaiHostnameProvider().Hostname()
	s.NoError(err)
	s.NotEmpty(hostname)
}

func (s *HostnamePublicTestSuite) TestMachineName() {
	_, err := job.ExportMachineName(source.Sources{
		FS: memfs.New(), Exec: gohaitest.Commands(s.T(), nil), Family: "debian",
	})
	s.Error(err, "no hostname command and no /proc/sys/kernel/hostname")
}

func TestHostnamePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HostnamePublicTestSuite))
}
