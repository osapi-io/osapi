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
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

// registrationJSON returns a minimal agent registration JSON for the given hostname.
func registrationJSON(hostname string) []byte {
	return []byte(fmt.Sprintf(
		`{"hostname":%q,"registered_at":"2026-01-01T00:00:00Z"}`,
		hostname,
	))
}

// setupRegistryKV configures mockRegistryKV to return the provided hostnames
// from Keys() and then return registration data for each.
func setupRegistryKV(
	ctrl *gomock.Controller,
	hostnames []string,
) *jobmocks.MockKeyValue {
	mockRegistryKV := jobmocks.NewMockKeyValue(ctrl)

	keys := make([]string, len(hostnames))
	for i, h := range hostnames {
		keys[i] = "agents." + job.SanitizeHostname(h)
	}

	mockRegistryKV.EXPECT().
		Keys(gomock.Any()).
		Return(keys, nil)

	for _, h := range hostnames {
		key := "agents." + job.SanitizeHostname(h)
		entry := jobmocks.NewMockKeyValueEntry(ctrl)
		entry.EXPECT().Value().Return(registrationJSON(h))
		mockRegistryKV.EXPECT().
			Get(gomock.Any(), key).
			Return(entry, nil)
	}

	return mockRegistryKV
}

type ClientPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	mockKV         *jobmocks.MockKeyValue
	jobsClient     *client.Client
	ctx            context.Context
}

func (s *ClientPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNATSClient = jobmocks.NewMockNATSClient(s.mockCtrl)
	s.mockKV = jobmocks.NewMockKeyValue(s.mockCtrl)
	s.ctx = context.Background()

	opts := &client.Options{
		Timeout:    30 * time.Second,
		KVBucket:   s.mockKV,
		StreamName: "JOBS",
	}
	var err error
	s.jobsClient, err = client.New(slog.Default(), s.mockNATSClient, opts)
	s.Require().NoError(err)
}

func (s *ClientPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ClientPublicTestSuite) TestQuery() {
	const (
		target    = "server1"
		category  = "node"
		operation = job.OperationType("node.hostname.get")
		subject   = "jobs.query.host.server1"
	)

	successResp := `{"status":"completed","hostname":"server1"}`
	failedResp := `{"status":"failed","error":"provider error"}`
	skippedResp := `{"status":"skipped","error":"unsupported"}`

	tests := []struct {
		name         string
		setupMocks   func()
		expectedErr  string
		validateFunc func(jobID string, resp *job.Response)
	}{
		{
			name: "when succeeds",
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					successResp,
					nil,
				)
			},
			validateFunc: func(jobID string, resp *job.Response) {
				s.NotEmpty(jobID)
				s.NotNil(resp)
				s.Equal(job.StatusCompleted, resp.Status)
			},
		},
		{
			name: "when job failed",
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					failedResp,
					nil,
				)
			},
			expectedErr: "job failed: provider error",
		},
		{
			name: "when job skipped",
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					skippedResp,
					nil,
				)
			},
			expectedErr: "job failed: unsupported",
		},
		{
			name: "when publish error",
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					"",
					errors.New("kv put failed"),
				)
			},
			expectedErr: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			jobID, resp, err := s.jobsClient.Query(
				s.ctx,
				target,
				category,
				operation,
				nil,
			)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				s.Empty(jobID)
				s.Nil(resp)
			} else {
				s.NoError(err)
				if tt.validateFunc != nil {
					tt.validateFunc(jobID, resp)
				}
			}
		})
	}
}

func (s *ClientPublicTestSuite) TestQueryBroadcast() {
	const (
		target    = "_all"
		category  = "node"
		operation = job.OperationType("node.hostname.get")
		subject   = "jobs.query._all"
	)

	host1Resp := `{"status":"completed","hostname":"server1"}`
	host2FailResp := `{"status":"failed","error":"provider error","hostname":"server2"}`

	tests := []struct {
		name         string
		setupMocks   func() *client.Client
		expectedErr  string
		validateFunc func(jobID string, results map[string]*job.Response, errs map[string]string)
	}{
		{
			name: "when succeeds",
			setupMocks: func() *client.Client {
				registryKV := setupRegistryKV(s.mockCtrl, []string{"server1"})
				c := s.newClientWithRegistry(registryKV)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response, errs map[string]string) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
				s.Empty(errs)
				s.Contains(results, "server1")
			},
		},
		{
			name: "when job failed",
			setupMocks: func() *client.Client {
				registryKV := setupRegistryKV(s.mockCtrl, []string{"server1"})
				c := s.newClientWithRegistry(registryKV)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{
							`{"status":"failed","error":"provider error","hostname":"server1"}`,
						},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response, errs map[string]string) {
				s.NotEmpty(jobID)
				s.Empty(results)
				s.Len(errs, 1)
				s.Equal("provider error", errs["server1"])
			},
		},
		{
			name: "when job skipped",
			setupMocks: func() *client.Client {
				registryKV := setupRegistryKV(s.mockCtrl, []string{"server1"})
				c := s.newClientWithRegistry(registryKV)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{
							`{"status":"skipped","hostname":"server1"}`,
						},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response, errs map[string]string) {
				s.NotEmpty(jobID)
				s.Empty(results)
				s.Len(errs, 1)
				s.Equal("skipped", errs["server1"])
			},
		},
		{
			name: "when publish error",
			setupMocks: func() *client.Client {
				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						mockError: errors.New("kv put failed"),
						errorMode: errorOnKVPut,
					},
				)
				return s.jobsClient
			},
			expectedErr: "failed to collect broadcast responses",
		},
		{
			name: "when partial failure",
			setupMocks: func() *client.Client {
				registryKV := setupRegistryKV(s.mockCtrl, []string{"server1", "server2"})
				c := s.newClientWithRegistry(registryKV)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp, host2FailResp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response, errs map[string]string) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
				s.Contains(results, "server1")
				s.Len(errs, 1)
				s.Equal("provider error", errs["server2"])
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			c := tt.setupMocks()

			jobID, results, errs, err := c.QueryBroadcast(
				s.ctx,
				target,
				category,
				operation,
				nil,
			)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				s.Empty(jobID)
				s.Nil(results)
				s.Nil(errs)
			} else {
				s.NoError(err)
				if tt.validateFunc != nil {
					tt.validateFunc(jobID, results, errs)
				}
			}
		})
	}
}

func (s *ClientPublicTestSuite) TestModify() {
	const (
		target    = "server1"
		category  = "node"
		operation = job.OperationType("node.hostname.set")
		subject   = "jobs.modify.host.server1"
	)

	successResp := `{"status":"completed","hostname":"server1","changed":true}`
	failedResp := `{"status":"failed","error":"permission denied"}`
	skippedResp := `{"status":"skipped","error":"unsupported"}`

	tests := []struct {
		name         string
		setupMocks   func()
		expectedErr  string
		validateFunc func(jobID string, resp *job.Response)
	}{
		{
			name: "when succeeds",
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					successResp,
					nil,
				)
			},
			validateFunc: func(jobID string, resp *job.Response) {
				s.NotEmpty(jobID)
				s.NotNil(resp)
				s.Equal(job.StatusCompleted, resp.Status)
			},
		},
		{
			name: "when job failed",
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					failedResp,
					nil,
				)
			},
			expectedErr: "job failed: permission denied",
		},
		{
			name: "when job skipped",
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					skippedResp,
					nil,
				)
			},
			expectedErr: "job failed: unsupported",
		},
		{
			name: "when publish error",
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					"",
					errors.New("kv put failed"),
				)
			},
			expectedErr: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			jobID, resp, err := s.jobsClient.Modify(
				s.ctx,
				target,
				category,
				operation,
				nil,
			)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				s.Empty(jobID)
				s.Nil(resp)
			} else {
				s.NoError(err)
				if tt.validateFunc != nil {
					tt.validateFunc(jobID, resp)
				}
			}
		})
	}
}

func (s *ClientPublicTestSuite) TestModifyBroadcast() {
	const (
		target    = "_all"
		category  = "node"
		operation = job.OperationType("node.hostname.set")
		subject   = "jobs.modify._all"
	)

	host1Resp := `{"status":"completed","hostname":"server1","changed":true}`
	host2FailResp := `{"status":"failed","error":"permission denied","hostname":"server2"}`

	tests := []struct {
		name         string
		setupMocks   func() *client.Client
		expectedErr  string
		validateFunc func(jobID string, results map[string]*job.Response, errs map[string]string)
	}{
		{
			name: "when succeeds",
			setupMocks: func() *client.Client {
				registryKV := setupRegistryKV(s.mockCtrl, []string{"server1"})
				c := s.newClientWithRegistry(registryKV)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response, errs map[string]string) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
				s.Empty(errs)
				s.Contains(results, "server1")
			},
		},
		{
			name: "when job failed",
			setupMocks: func() *client.Client {
				registryKV := setupRegistryKV(s.mockCtrl, []string{"server1"})
				c := s.newClientWithRegistry(registryKV)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{
							`{"status":"failed","error":"permission denied","hostname":"server1"}`,
						},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response, errs map[string]string) {
				s.NotEmpty(jobID)
				s.Empty(results)
				s.Len(errs, 1)
				s.Equal("permission denied", errs["server1"])
			},
		},
		{
			name: "when job skipped",
			setupMocks: func() *client.Client {
				registryKV := setupRegistryKV(s.mockCtrl, []string{"server1"})
				c := s.newClientWithRegistry(registryKV)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{
							`{"status":"skipped","hostname":"server1"}`,
						},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response, errs map[string]string) {
				s.NotEmpty(jobID)
				s.Empty(results)
				s.Len(errs, 1)
				s.Equal("skipped", errs["server1"])
			},
		},
		{
			name: "when publish error",
			setupMocks: func() *client.Client {
				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						mockError: errors.New("kv put failed"),
						errorMode: errorOnKVPut,
					},
				)
				return s.jobsClient
			},
			expectedErr: "failed to collect broadcast responses",
		},
		{
			name: "when partial failure",
			setupMocks: func() *client.Client {
				registryKV := setupRegistryKV(s.mockCtrl, []string{"server1", "server2"})
				c := s.newClientWithRegistry(registryKV)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockKV,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp, host2FailResp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response, errs map[string]string) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
				s.Contains(results, "server1")
				s.Len(errs, 1)
				s.Equal("permission denied", errs["server2"])
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			c := tt.setupMocks()

			jobID, results, errs, err := c.ModifyBroadcast(
				s.ctx,
				target,
				category,
				operation,
				nil,
			)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				s.Empty(jobID)
				s.Nil(results)
				s.Nil(errs)
			} else {
				s.NoError(err)
				if tt.validateFunc != nil {
					tt.validateFunc(jobID, results, errs)
				}
			}
		})
	}
}

// newClientWithRegistry creates a new client using the suite's existing mocks
// plus the provided registryKV.
func (s *ClientPublicTestSuite) newClientWithRegistry(
	registryKV *jobmocks.MockKeyValue,
) *client.Client {
	opts := &client.Options{
		Timeout:    30 * time.Second,
		KVBucket:   s.mockKV,
		StreamName: "JOBS",
		RegistryKV: registryKV,
	}
	c, err := client.New(slog.Default(), s.mockNATSClient, opts)
	s.Require().NoError(err)
	return c
}

func TestClientPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ClientPublicTestSuite))
}
