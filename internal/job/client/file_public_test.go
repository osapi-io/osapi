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
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type FilePublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	mockKV         *jobmocks.MockKeyValue
	jobsClient     *client.Client
	ctx            context.Context
}

func (s *FilePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNATSClient = jobmocks.NewMockNATSClient(s.mockCtrl)
	s.mockKV = jobmocks.NewMockKeyValue(s.mockCtrl)
	s.ctx = context.Background()

	opts := &client.Options{
		Timeout:  30 * time.Second,
		KVBucket: s.mockKV,
	}
	var err error
	s.jobsClient, err = client.New(slog.Default(), s.mockNATSClient, opts)
	s.Require().NoError(err)

	s.jobsClient.SetMeterProvider(sdkmetric.NewMeterProvider())
}

func (s *FilePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *FilePublicTestSuite) TestModifyFileDeploy() {
	tests := []struct {
		name          string
		hostname      string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
		expectChanged bool
	}{
		{
			name:     "when deploy succeeds",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"hostname": "server1",
				"changed": true,
				"data": {"changed":true,"sha256":"abc123","path":"/etc/app.conf"}
			}`,
			expectChanged: true,
		},
		{
			name:     "when deploy succeeds unchanged",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"hostname": "server1",
				"changed": false,
				"data": {"changed":false,"sha256":"abc123","path":"/etc/app.conf"}
			}`,
			expectChanged: false,
		},
		{
			name:     "when job failed",
			hostname: "server1",
			responseData: `{
				"status": "failed",
				"error": "failed to get object: not found",
				"data": {}
			}`,
			expectError:   true,
			errorContains: "job failed",
		},
		{
			name:          "when publish fails",
			hostname:      "server1",
			mockError:     errors.New("connection failed"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.modify.host.server1",
				tt.responseData,
				tt.mockError,
			)

			jobID, hostname, changed, err := s.jobsClient.ModifyFileDeploy(
				s.ctx,
				tt.hostname,
				"app.conf",
				"/etc/app.conf",
				"raw",
				"0644",
				"root",
				"root",
				nil,
			)

			if tt.expectError {
				s.Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotEmpty(jobID)
				s.Equal("server1", hostname)
				s.Equal(tt.expectChanged, changed)
			}
		})
	}
}

func (s *FilePublicTestSuite) TestQueryFileStatus() {
	tests := []struct {
		name          string
		hostname      string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "when status succeeds",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"hostname": "server1",
				"data": {"path":"/etc/app.conf","status":"in-sync","sha256":"abc123"}
			}`,
		},
		{
			name:          "when publish fails",
			hostname:      "server1",
			mockError:     errors.New("connection failed"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
		{
			name:     "when job failed",
			hostname: "server1",
			responseData: `{
				"status": "failed",
				"error": "file not found",
				"data": {}
			}`,
			expectError:   true,
			errorContains: "job failed: file not found",
		},
		{
			name:     "when unmarshal fails",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": "not valid json object"
			}`,
			expectError:   true,
			errorContains: "failed to unmarshal file status response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query.host.server1",
				tt.responseData,
				tt.mockError,
			)

			jobID, result, hostname, err := s.jobsClient.QueryFileStatus(
				s.ctx,
				tt.hostname,
				"/etc/app.conf",
			)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotEmpty(jobID)
				s.NotNil(result)
				s.Equal("server1", hostname)
			}
		})
	}
}

func TestFilePublicTestSuite(t *testing.T) {
	suite.Run(t, new(FilePublicTestSuite))
}
