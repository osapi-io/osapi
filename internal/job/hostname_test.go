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

package job

import (
	"errors"
	"testing"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/stretchr/testify/suite"
)

// mockHostnameProvider for testing
type mockHostnameProvider struct {
	hostname string
	err      error
}

func (m *mockHostnameProvider) Hostname() (string, error) {
	return m.hostname, m.err
}

type HostnameTestSuite struct {
	suite.Suite
}

func (s *HostnameTestSuite) SetupTest() {
	// Setup for each test
}

func (s *HostnameTestSuite) TearDownTest() {
	// Cleanup after each test
}

func (s *HostnameTestSuite) TestGopsutilHostnameProvider() {
	tests := []struct {
		name         string
		expectError  bool
		validateFunc func(string, error)
	}{
		{
			name:        "should return non-empty hostname without error",
			expectError: false,
			validateFunc: func(hostname string, err error) {
				s.NoError(err)
				s.NotEmpty(hostname)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			provider := gopsutilHostnameProvider{}
			hostname, err := provider.Hostname()

			tt.validateFunc(hostname, err)
		})
	}
}

func (s *HostnameTestSuite) TestHostnameProviderInterface() {
	tests := []struct {
		name     string
		provider HostnameProvider
	}{
		{
			name:     "gopsutilHostnameProvider implements interface",
			provider: gopsutilHostnameProvider{},
		},
		{
			name:     "mockHostnameProvider implements interface",
			provider: &mockHostnameProvider{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Test that providers implement the interface
			_ = tt.provider
		})
	}
}

func (s *HostnameTestSuite) TestGopsutilHostnameProviderError() {
	tests := []struct {
		name string
	}{
		{
			name: "returns error when host.Info fails",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			original := hostInfoFn
			defer func() { hostInfoFn = original }()

			hostInfoFn = func() (*host.InfoStat, error) {
				return nil, errors.New("host info failed")
			}

			provider := gopsutilHostnameProvider{}
			hostname, err := provider.Hostname()

			s.Error(err)
			s.Empty(hostname)
		})
	}
}

func (s *HostnameTestSuite) TestGetWorkerHostnameProviderError() {
	tests := []struct {
		name string
	}{
		{
			name: "falls back to unknown when provider errors",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			original := defaultHostnameProvider
			defer func() { defaultHostnameProvider = original }()

			defaultHostnameProvider = &mockHostnameProvider{
				hostname: "",
				err:      errors.New("provider error"),
			}

			hostname, err := GetWorkerHostname("")

			s.NoError(err)
			s.Equal("unknown", hostname)
		})
	}
}

func TestHostnameTestSuite(t *testing.T) {
	suite.Run(t, new(HostnameTestSuite))
}
