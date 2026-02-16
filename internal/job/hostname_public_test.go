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

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
)

// mockHostnameProvider for testing
type mockHostnameProvider struct {
	hostname string
	err      error
}

func (m *mockHostnameProvider) Hostname() (string, error) {
	return m.hostname, m.err
}

type HostnamePublicTestSuite struct {
	suite.Suite
}

func (s *HostnamePublicTestSuite) SetupTest() {
	// Setup for each test
}

func (s *HostnamePublicTestSuite) TearDownTest() {
	// Cleanup after each test
}

func (s *HostnamePublicTestSuite) TestGetWorkerHostname() {
	tests := []struct {
		name               string
		configuredHostname string
		expectError        bool
		expectedResult     string
		validateFunc       func(string)
	}{
		{
			name:               "configured hostname takes precedence",
			configuredHostname: "configured-worker",
			expectError:        false,
			expectedResult:     "configured-worker",
			validateFunc: func(hostname string) {
				s.Equal("configured-worker", hostname)
			},
		},
		{
			name:               "empty configured hostname falls back to system",
			configuredHostname: "",
			expectError:        false,
			validateFunc: func(hostname string) {
				s.NotEmpty(hostname, "System hostname should not be empty")
				s.NotEqual(
					"unknown",
					hostname,
					"Should get actual system hostname, not unknown fallback",
				)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			hostname, err := job.GetWorkerHostname(tt.configuredHostname)

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				if tt.validateFunc != nil {
					tt.validateFunc(hostname)
				}
				if tt.expectedResult != "" {
					s.Equal(tt.expectedResult, hostname)
				}
			}
		})
	}
}

func (s *HostnamePublicTestSuite) TestGetLocalHostname() {
	tests := []struct {
		name         string
		expectError  bool
		validateFunc func(string)
	}{
		{
			name:        "returns system hostname",
			expectError: false,
			validateFunc: func(hostname string) {
				s.NotEmpty(hostname, "Local hostname should not be empty")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			hostname, err := job.GetLocalHostname()

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				if tt.validateFunc != nil {
					tt.validateFunc(hostname)
				}
			}
		})
	}
}

func (s *HostnamePublicTestSuite) TestGetWorkerHostnameWithProvider() {
	tests := []struct {
		name               string
		configuredHostname string
		provider           job.HostnameProvider
		expectedHostname   string
		expectError        bool
	}{
		{
			name:               "configured hostname bypasses provider",
			configuredHostname: "configured-worker",
			provider:           &mockHostnameProvider{hostname: "system-host", err: nil},
			expectedHostname:   "configured-worker",
			expectError:        false,
		},
		{
			name:               "empty config uses provider successfully",
			configuredHostname: "",
			provider:           &mockHostnameProvider{hostname: "system-host", err: nil},
			expectedHostname:   "system-host",
			expectError:        false,
		},
		{
			name:               "empty config with provider error returns unknown",
			configuredHostname: "",
			provider: &mockHostnameProvider{
				hostname: "",
				err:      errors.New("provider error"),
			},
			expectedHostname: "unknown",
			expectError:      false,
		},
		{
			name:               "empty config with empty hostname returns unknown",
			configuredHostname: "",
			provider:           &mockHostnameProvider{hostname: "", err: nil},
			expectedHostname:   "unknown",
			expectError:        false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			hostname, err := job.GetWorkerHostnameWithProvider(tt.configuredHostname, tt.provider)

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.expectedHostname, hostname)
			}
		})
	}
}

func (s *HostnamePublicTestSuite) TestGetLocalHostnameWithProvider() {
	tests := []struct {
		name             string
		provider         job.HostnameProvider
		expectedHostname string
		expectError      bool
	}{
		{
			name:             "successful hostname retrieval",
			provider:         &mockHostnameProvider{hostname: "test-host", err: nil},
			expectedHostname: "test-host",
			expectError:      false,
		},
		{
			name:        "provider error",
			provider:    &mockHostnameProvider{hostname: "", err: errors.New("hostname error")},
			expectError: true,
		},
		{
			name:             "empty hostname from provider",
			provider:         &mockHostnameProvider{hostname: "", err: nil},
			expectedHostname: "",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			hostname, err := job.GetLocalHostnameWithProvider(tt.provider)

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.expectedHostname, hostname)
			}
		})
	}
}

func TestHostnamePublicTestSuite(t *testing.T) {
	suite.Run(t, new(HostnamePublicTestSuite))
}
