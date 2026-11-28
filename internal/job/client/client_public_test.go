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
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/secret/seal"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// registrationJSON returns a minimal agent registration JSON for the given hostname.
func registrationJSON(
	hostname string,
) []byte {
	return []byte(fmt.Sprintf(
		`{"hostname":%q,"registered_at":"2026-01-01T00:00:00Z"}`,
		hostname,
	))
}

// setupRegistryKV makes the registry bucket of nc return the provided
// hostnames from KVKeys and then registration data for each.
func setupRegistryKV(
	ctrl *gomock.Controller,
	nc *jobmocks.MockNATSClient,
	hostnames []string,
) {
	keys := make([]string, len(hostnames))
	for i, h := range hostnames {
		keys[i] = "agents." + job.SanitizeHostname(h)
	}

	nc.EXPECT().
		KVKeys(gomock.Any(), "registry").
		Return(keys, nil)

	for _, h := range hostnames {
		key := "agents." + job.SanitizeHostname(h)
		entry := jobmocks.NewMockKeyValueEntry(ctrl)
		entry.EXPECT().Value().Return(registrationJSON(h))
		nc.EXPECT().
			KVGet(gomock.Any(), "registry", key).
			Return(entry, nil)
	}
}

type ClientPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	jobsClient     *client.Client
	ctx            context.Context
	mockNC         *jobmocks.MockNATSClient
	web            *ecdh.PrivateKey
	db             *ecdh.PrivateKey
}

func (s *ClientPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNATSClient = jobmocks.NewMockNATSClient(s.mockCtrl)
	s.ctx = context.Background()

	opts := &client.Options{
		Timeout:    30 * time.Second,
		JobsBucket: "jobs",
		StreamName: "JOBS",
	}
	var err error
	s.jobsClient, err = newTrusted(s.mockNATSClient, opts)
	s.Require().NoError(err)
	s.mockNC = jobmocks.NewMockNATSClient(s.mockCtrl)
	s.web, err = ecdh.P256().GenerateKey(rand.Reader)
	s.Require().NoError(err)
	s.db, err = ecdh.P256().GenerateKey(rand.Reader)
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
		name      string
		data      any
		withMeter bool
		// timeout, when set, is the client's timeout, so a case that waits
		// for no answer does not wait the suite's 30 seconds.
		timeout      time.Duration
		setupMocks   func()
		expectedErr  string
		validateFunc func(jobID string, resp *job.Response)
	}{
		{
			name:      "when succeeds with meter provider",
			data:      nil,
			withMeter: true,
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					successResp,
					nil,
				)
			},
			validateFunc: func(jobID string, resp *job.Response) {
				s.NotEmpty(jobID)
				s.NotNil(resp)
			},
		},
		{
			name: "when succeeds",
			data: nil,
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					successResp,
					nil,
				)
			},
			validateFunc: func(jobID string, resp *job.Response) {
				s.NotEmpty(jobID)
				s.NotNil(resp)
				s.Equal(types.AgentJobStatusCompleted, resp.Status)
			},
		},
		{
			name: "when job failed",
			data: nil,
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockNATSClient,
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
			data: nil,
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					skippedResp,
					nil,
				)
			},
			validateFunc: func(jobID string, resp *job.Response) {
				s.NotEmpty(jobID)
				s.NotNil(resp)
				s.Equal(types.AgentJobStatusSkipped, resp.Status)
				s.Equal("unsupported", resp.Error)
			},
		},
		{
			name: "when data marshal fails",
			// A channel cannot be marshaled to JSON.
			data:        make(chan int),
			setupMocks:  func() {},
			expectedErr: "marshal data",
		},
		{
			name: "when publish error",
			data: nil,
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					"",
					errors.New("kv put failed"),
				)
			},
			expectedErr: "failed to publish and wait",
		},
		{
			name: "when nats publish error",
			data: nil,
			setupMocks: func() {
				setupPublishAndWaitMocksWithOpts(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndWaitMockOpts{
						mockError: errors.New("nats publish failed"),
						errorMode: errorOnPublish,
					},
				)
			},
			expectedErr: "failed to publish and wait",
		},
		{
			name: "when watch error",
			data: nil,
			setupMocks: func() {
				setupPublishAndWaitMocksWithOpts(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndWaitMockOpts{
						mockError: errors.New("watch failed"),
						errorMode: errorOnWatch,
					},
				)
			},
			expectedErr: "failed to publish and wait",
		},
		{
			name:    "when timeout waiting for response",
			data:    nil,
			timeout: 50 * time.Millisecond,
			setupMocks: func() {
				setupPublishAndWaitMocksWithOpts(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndWaitMockOpts{
						mockError: errors.New("timeout"),
						errorMode: errorOnTimeout,
					},
				)
			},
			expectedErr: "failed to publish and wait",
		},
		{
			name: "when unmarshal error on response",
			data: nil,
			setupMocks: func() {
				setupPublishAndWaitMocksWithOpts(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndWaitMockOpts{
						responseData: "not-valid-json",
					},
				)
			},
			expectedErr: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.withMeter {
				s.jobsClient.SetMeterProvider(sdkmetric.NewMeterProvider())
			}

			tt.setupMocks()

			jobs := s.jobsClient
			if tt.timeout != 0 {
				jobs = s.newClient(tt.timeout, "")
			}
			jobID, resp, err := jobs.Query(
				s.ctx,
				target,
				category,
				operation,
				tt.data,
			)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				if strings.HasPrefix(tt.expectedErr, "job failed") {
					s.NotEmpty(jobID, "a job that ran keeps its ID")
				} else {
					s.Empty(jobID)
				}
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
		data         any
		withMeter    bool
		setupMocks   func() *client.Client
		expectedErr  string
		validateFunc func(jobID string, results map[string]*job.Response)
	}{
		{
			name:      "when succeeds with meter provider",
			data:      nil,
			withMeter: true,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1"})
				c := s.newClientWithRegistry()

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
			},
		},
		{
			name: "when succeeds",
			data: nil,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1"})
				c := s.newClientWithRegistry()

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
				s.Contains(results, "server1")
				s.Equal(types.AgentJobStatusCompleted, results["server1"].Status)
			},
		},
		{
			name: "when data marshal fails in broadcast",
			// A channel cannot be marshaled to JSON.
			data: make(chan int),
			setupMocks: func() *client.Client {
				return s.jobsClient
			},
			expectedErr: "marshal data",
		},
		{
			name: "when job failed",
			data: nil,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1"})
				c := s.newClientWithRegistry()

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
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
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
				s.Equal(types.AgentJobStatusFailed, results["server1"].Status)
				s.Equal("provider error", results["server1"].Error)
			},
		},
		{
			name: "when job skipped",
			data: nil,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1"})
				c := s.newClientWithRegistry()

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
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
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
				s.Equal(types.AgentJobStatusSkipped, results["server1"].Status)
			},
		},
		{
			name: "when publish error",
			data: nil,
			setupMocks: func() *client.Client {
				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
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
			data: nil,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1", "server2"})
				c := s.newClientWithRegistry()

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp, host2FailResp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 2)
				s.Contains(results, "server1")
				s.Equal(types.AgentJobStatusCompleted, results["server1"].Status)
				s.Contains(results, "server2")
				s.Equal(types.AgentJobStatusFailed, results["server2"].Status)
				s.Equal("provider error", results["server2"].Error)
			},
		},
		{
			name: "when ListAgents fails falls back to full timeout and collects responses",
			data: nil,
			setupMocks: func() *client.Client {
				// The registry returns an error so ListAgents fails — warn path is exercised.
				s.mockNATSClient.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return(nil, errors.New("registry unavailable"))
				// It waits the client's whole timeout, so the timeout is short.
				c := s.newClient(50*time.Millisecond, "registry")

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
			},
		},
		{
			name: "when nats publish error in broadcast",
			data: nil,
			setupMocks: func() *client.Client {
				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						mockError: errors.New("nats publish failed"),
						errorMode: errorOnPublish,
					},
				)
				return s.jobsClient
			},
			expectedErr: "failed to collect broadcast responses",
		},
		{
			name: "when watch error in broadcast",
			data: nil,
			setupMocks: func() *client.Client {
				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						mockError: errors.New("watch failed"),
						errorMode: errorOnWatch,
					},
				)
				return s.jobsClient
			},
			expectedErr: "failed to collect broadcast responses",
		},
		{
			name: "when the watch ends early waits for the timeout",
			data: nil,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1", "server2"})
				opts := &client.Options{
					Timeout:        50 * time.Millisecond,
					JobsBucket:     "jobs",
					StreamName:     "JOBS",
					RegistryBucket: "registry",
				}
				c, err := newTrusted(s.mockNATSClient, opts)
				s.Require().NoError(err)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp},
						closeWatch:      true,
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 2)
				s.Equal(types.AgentJobStatusTimeout, results["server2"].Status)
			},
		},
		{
			name: "when an unexpected host answers waits for each expected host",
			data: nil,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1", "server2"})
				opts := &client.Options{
					Timeout:        50 * time.Millisecond,
					JobsBucket:     "jobs",
					StreamName:     "JOBS",
					RegistryBucket: "registry",
				}
				c, err := newTrusted(s.mockNATSClient, opts)
				s.Require().NoError(err)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{
							`{"status":"completed","hostname":"server3"}`,
							host1Resp,
						},
						closeWatch: true,
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 3)
				s.Equal(types.AgentJobStatusCompleted, results["server1"].Status)
				s.Equal(types.AgentJobStatusTimeout, results["server2"].Status)
			},
		},
		{
			name: "when broadcast response has invalid JSON skips entry then collects valid",
			data: nil,
			setupMocks: func() *client.Client {
				// Two agents in registry: expectedCount=2. Send bad JSON first,
				// then one valid response; the bad one is skipped, the valid one
				// counts as 1 of 2 and the second slot stays open until timeout.
				// Use a tiny timeout so this doesn't wait 30s.
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1"})
				opts := &client.Options{
					Timeout:        50 * time.Millisecond,
					JobsBucket:     "jobs",
					StreamName:     "JOBS",
					RegistryBucket: "registry",
				}
				c, err := newTrusted(s.mockNATSClient, opts)
				s.Require().NoError(err)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						// bad JSON followed by valid JSON; bad is skipped,
						// valid gives us 1 response which satisfies expectedCount=1.
						responseEntries: []string{"not-valid-json", host1Resp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				// The valid response was collected after skipping the bad one.
				s.Len(results, 1)
			},
		},
		{
			name: "when broadcast times out after partial responses returns what was collected",
			data: nil,
			setupMocks: func() *client.Client {
				// No registry so expectedCount=0 (full timeout path).
				// Send one valid response then block. Timeout fires with
				// len(responses)>0 so the non-error timeout branch executes.
				opts := &client.Options{
					Timeout:    50 * time.Millisecond,
					JobsBucket: "jobs",
					StreamName: "JOBS",
				}
				c, err := newTrusted(s.mockNATSClient, opts)
				s.Require().NoError(err)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp},
						mockError:       errors.New("partial"),
						errorMode:       errorOnTimeoutWithPartialResponse,
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
			},
		},
		{
			name: "when broadcast response has empty hostname uses unknown",
			data: nil,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1"})
				c := s.newClientWithRegistry()

				// Response has no hostname field — falls back to "unknown" key.
				noHostnameResp := `{"status":"completed"}`
				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{noHostnameResp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Contains(results, "unknown")
			},
		},
		{
			name: "when broadcast times out injects timeout entries for missing agents",
			data: nil,
			setupMocks: func() *client.Client {
				// Registry reports 2 agents but only server1 responds.
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1", "server2"})
				opts := &client.Options{
					Timeout:        50 * time.Millisecond,
					JobsBucket:     "jobs",
					StreamName:     "JOBS",
					RegistryBucket: "registry",
				}
				c, err := newTrusted(s.mockNATSClient, opts)
				s.Require().NoError(err)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp},
						mockError:       errors.New("partial"),
						errorMode:       errorOnTimeoutWithPartialResponse,
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 2)
				s.Contains(results, "server1")
				s.Equal(types.AgentJobStatusCompleted, results["server1"].Status)
				s.Contains(results, "server2")
				// An agent that never answered is a timeout, not a failure:
				// nobody knows whether the operation ran.
				s.Equal(types.AgentJobStatusTimeout, results["server2"].Status)
				s.Equal(job.ErrorCodeTimeout, results["server2"].ErrorCode)
				s.Equal("timeout: agent did not respond", results["server2"].Error)
			},
		},
		{
			name: "when broadcast times out with no responses returns error",
			data: nil,
			setupMocks: func() *client.Client {
				opts := &client.Options{
					Timeout:    50 * time.Millisecond,
					JobsBucket: "jobs",
					StreamName: "JOBS",
				}
				c, err := newTrusted(s.mockNATSClient, opts)
				s.Require().NoError(err)

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						mockError: errors.New("timeout"),
						errorMode: errorOnTimeout,
					},
				)
				return c
			},
			expectedErr: "timeout waiting for broadcast responses: no agents responded",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			c := tt.setupMocks()
			if tt.withMeter {
				c.SetMeterProvider(sdkmetric.NewMeterProvider())
			}

			jobID, results, err := c.QueryBroadcast(
				s.ctx,
				target,
				category,
				operation,
				tt.data,
			)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				s.Empty(jobID)
				s.Nil(results)
			} else {
				s.NoError(err)
				if tt.validateFunc != nil {
					tt.validateFunc(jobID, results)
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
		data         any
		withMeter    bool
		setupMocks   func()
		expectedErr  string
		validateFunc func(jobID string, resp *job.Response)
	}{
		{
			name:      "when succeeds with meter provider",
			data:      nil,
			withMeter: true,
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					successResp,
					nil,
				)
			},
			validateFunc: func(jobID string, resp *job.Response) {
				s.NotEmpty(jobID)
				s.NotNil(resp)
			},
		},
		{
			name: "when succeeds",
			data: nil,
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					successResp,
					nil,
				)
			},
			validateFunc: func(jobID string, resp *job.Response) {
				s.NotEmpty(jobID)
				s.NotNil(resp)
				s.Equal(types.AgentJobStatusCompleted, resp.Status)
			},
		},
		{
			name: "when job failed",
			data: nil,
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockNATSClient,
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
			data: nil,
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					skippedResp,
					nil,
				)
			},
			validateFunc: func(jobID string, resp *job.Response) {
				s.NotEmpty(jobID)
				s.NotNil(resp)
				s.Equal(types.AgentJobStatusSkipped, resp.Status)
				s.Equal("unsupported", resp.Error)
			},
		},
		{
			name:        "when data marshal fails",
			data:        make(chan int),
			setupMocks:  func() {},
			expectedErr: "marshal data",
		},
		{
			name: "when publish error",
			data: nil,
			setupMocks: func() {
				setupPublishAndWaitMocks(
					s.mockCtrl,
					s.mockNATSClient,
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
			if tt.withMeter {
				s.jobsClient.SetMeterProvider(sdkmetric.NewMeterProvider())
			}

			tt.setupMocks()

			jobID, resp, err := s.jobsClient.Modify(
				s.ctx,
				target,
				category,
				operation,
				tt.data,
			)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				if strings.HasPrefix(tt.expectedErr, "job failed") {
					s.NotEmpty(jobID, "a job that ran keeps its ID")
				} else {
					s.Empty(jobID)
				}
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
		data         any
		withMeter    bool
		setupMocks   func() *client.Client
		expectedErr  string
		validateFunc func(jobID string, results map[string]*job.Response)
	}{
		{
			name:      "when succeeds with meter provider",
			data:      nil,
			withMeter: true,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1"})
				c := s.newClientWithRegistry()

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
			},
		},
		{
			name: "when succeeds",
			data: nil,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1"})
				c := s.newClientWithRegistry()

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
				s.Contains(results, "server1")
				s.Equal(types.AgentJobStatusCompleted, results["server1"].Status)
			},
		},
		{
			name: "when data marshal fails in modify broadcast",
			// A channel cannot be marshaled to JSON.
			data: make(chan int),
			setupMocks: func() *client.Client {
				return s.jobsClient
			},
			expectedErr: "marshal data",
		},
		{
			name: "when job failed",
			data: nil,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1"})
				c := s.newClientWithRegistry()

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
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
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
				s.Equal(types.AgentJobStatusFailed, results["server1"].Status)
				s.Equal("permission denied", results["server1"].Error)
			},
		},
		{
			name: "when job skipped",
			data: nil,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1"})
				c := s.newClientWithRegistry()

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
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
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 1)
				s.Equal(types.AgentJobStatusSkipped, results["server1"].Status)
			},
		},
		{
			name: "when publish error",
			data: nil,
			setupMocks: func() *client.Client {
				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
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
			data: nil,
			setupMocks: func() *client.Client {
				setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1", "server2"})
				c := s.newClientWithRegistry()

				setupPublishAndCollectMocks(
					s.mockCtrl,
					s.mockNATSClient,
					s.mockNATSClient,
					subject,
					&publishAndCollectMockOpts{
						responseEntries: []string{host1Resp, host2FailResp},
					},
				)
				return c
			},
			validateFunc: func(jobID string, results map[string]*job.Response) {
				s.NotEmpty(jobID)
				s.Len(results, 2)
				s.Contains(results, "server1")
				s.Equal(types.AgentJobStatusCompleted, results["server1"].Status)
				s.Contains(results, "server2")
				s.Equal(types.AgentJobStatusFailed, results["server2"].Status)
				s.Equal("permission denied", results["server2"].Error)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			c := tt.setupMocks()
			if tt.withMeter {
				c.SetMeterProvider(sdkmetric.NewMeterProvider())
			}

			jobID, results, err := c.ModifyBroadcast(
				s.ctx,
				target,
				category,
				operation,
				tt.data,
			)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				s.Empty(jobID)
				s.Nil(results)
			} else {
				s.NoError(err)
				if tt.validateFunc != nil {
					tt.validateFunc(jobID, results)
				}
			}
		})
	}
}

// newClientWithRegistry creates a new client using the suite's existing mocks
// and the registry bucket.
func (s *ClientPublicTestSuite) newClientWithRegistry() *client.Client {
	return s.newClient(30*time.Second, "registry")
}

// newClient is a client with timeout, and with registry as its registry
// bucket when it is not empty.
func (s *ClientPublicTestSuite) newClient(
	timeout time.Duration,
	registry string,
) *client.Client {
	opts := &client.Options{
		Timeout:        timeout,
		JobsBucket:     "jobs",
		StreamName:     "JOBS",
		RegistryBucket: registry,
	}
	c, err := newTrusted(s.mockNATSClient, opts)
	s.Require().NoError(err)

	return c
}

func (s *ClientPublicTestSuite) TestQueryWithPKISigner() {
	const (
		target    = "server1"
		category  = "node"
		operation = job.OperationType("node.hostname.get")
		subject   = "jobs.query.host.server1"
	)

	signer, _ := newSigner(gomock.NewController(s.T()))

	tests := []struct {
		name         string
		setupMocks   func()
		expectedErr  string
		validateFunc func(jobID string, resp *job.Response)
	}{
		{
			name: "when PKI signs job data and unwraps response",
			setupMocks: func() {
				// KV Put should receive a signed envelope.
				s.mockNATSClient.EXPECT().
					KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, _ string, _ string, data []byte) (uint64, error) {
						// Verify the stored data is a valid signed envelope.
						var envelope job.SignedEnvelope
						s.NoError(json.Unmarshal(data, &envelope))
						s.NotEmpty(envelope.Payload)
						s.NotEmpty(envelope.Signature)
						s.Equal("SHA256:test-fingerprint", envelope.Fingerprint)
						return uint64(1), nil
					})

				// Publish succeeds.
				s.mockNATSClient.EXPECT().
					Publish(gomock.Any(), subject, gomock.Any()).
					Return(nil)

				// Watch returns a response.
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return(
					[]byte(`{"status":"completed","hostname":"server1"}`),
				)
				ch := make(chan jetstream.KeyValueEntry, 1)
				ch <- mockEntry

				s.mockNATSClient.EXPECT().
					KVWatch(gomock.Any(), "jobs", gomock.Any()).
					Return(ch, nil)
			},
			validateFunc: func(jobID string, resp *job.Response) {
				s.NotEmpty(jobID)
				s.NotNil(resp)
				s.Equal(types.AgentJobStatusCompleted, resp.Status)
			},
		},
		{
			name: "when signing marshal fails returns sign error",
			setupMocks: func() {
				client.SetSigningMarshalFn(func(_ any) ([]byte, error) {
					return nil, errors.New("marshal boom")
				})
			},
			expectedErr: "failed to sign job data",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Create a client with PKI signer.
			opts := &client.Options{
				Timeout:    30 * time.Second,
				JobsBucket: "jobs",
				StreamName: "JOBS",
				PKISigner:  signer,
			}
			c, err := newTrusted(s.mockNATSClient, opts)
			s.Require().NoError(err)

			tt.setupMocks()

			jobID, resp, err := c.Query(
				s.ctx,
				target,
				category,
				operation,
				nil,
			)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				return
			}

			s.NoError(err)
			if tt.validateFunc != nil {
				tt.validateFunc(jobID, resp)
			}
		})
	}
}

func (s *ClientPublicTestSuite) TearDownSubTest() {
	client.ResetSigningMarshalFn()
}

func (s *ClientPublicTestSuite) TestModifyBroadcastWithPKISigner() {
	const (
		target    = "_all"
		category  = "node"
		operation = job.OperationType("node.hostname.get")
		subject   = "jobs.modify._all"
	)

	signer, _ := newSigner(gomock.NewController(s.T()))

	tests := []struct {
		name         string
		setupMocks   func()
		expectedErr  string
		validateFunc func(responses map[string]*job.Response)
	}{
		{
			name: "when PKI signing marshal fails returns error",
			setupMocks: func() {
				client.SetSigningMarshalFn(func(_ any) ([]byte, error) {
					return nil, errors.New("marshal boom")
				})

				opts := &client.Options{
					Timeout:    30 * time.Second,
					JobsBucket: "jobs",
					StreamName: "JOBS",
					PKISigner:  signer,
				}
				c, err := newTrusted(s.mockNATSClient, opts)
				s.Require().NoError(err)
				s.jobsClient = c
			},
			expectedErr: "failed to sign job data",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			_, responses, err := s.jobsClient.ModifyBroadcast(
				s.ctx, target, category, operation, nil,
			)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				return
			}

			s.NoError(err)
			if tt.validateFunc != nil {
				tt.validateFunc(responses)
			}
		})
	}
}

func (s *ClientPublicTestSuite) TestQueryWatchEndsEarly() {
	s.Run("when the watch ends before a response waits for the timeout", func() {
		c, err := newTrusted(s.mockNATSClient, &client.Options{
			Timeout:    50 * time.Millisecond,
			JobsBucket: "jobs",
			StreamName: "JOBS",
		})
		s.Require().NoError(err)

		setupPublishAndWaitMocksWithOpts(
			s.mockCtrl,
			s.mockNATSClient,
			s.mockNATSClient,
			"jobs.query.host.server1",
			&publishAndWaitMockOpts{closeWatch: true},
		)

		_, _, err = c.Query(
			s.ctx,
			"server1",
			"node",
			job.OperationType("node.hostname.get"),
			nil,
		)
		s.ErrorContains(err, "timeout waiting for job response")
	})
}

func (s *ClientPublicTestSuite) TestQueryWithTargetResolver() {
	const (
		hostname  = "web-01"
		machineID = "abc123def456"
		category  = "node"
		operation = job.OperationType("node.hostname.get")
	)

	// The resolver translates hostname → machineID.
	resolver := func(_ context.Context, target string) string {
		if target == hostname {
			return machineID
		}
		return target
	}

	// Expect subject to use the resolved machineID, not the hostname.
	expectedSubject := "jobs.query.host." + machineID

	s.mockNATSClient.EXPECT().
		KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
		Return(uint64(1), nil)
	s.mockNATSClient.EXPECT().
		Publish(gomock.Any(), expectedSubject, gomock.Any()).
		Return(nil)

	mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
	mockEntry.EXPECT().Value().Return(
		[]byte(`{"status":"completed","hostname":"web-01"}`),
	)
	ch := make(chan jetstream.KeyValueEntry, 1)
	ch <- mockEntry

	s.mockNATSClient.EXPECT().
		KVWatch(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
		Return(ch, nil)

	opts := &client.Options{
		Timeout:        30 * time.Second,
		JobsBucket:     "jobs",
		StreamName:     "JOBS",
		TargetResolver: resolver,
	}
	c, err := newTrusted(s.mockNATSClient, opts)
	s.Require().NoError(err)

	_, resp, err := c.Query(s.ctx, hostname, category, operation, nil)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(types.AgentJobStatusCompleted, resp.Status)
}

func TestClientPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ClientPublicTestSuite))
}

// publishAndWaitErrorMode controls which step of the publishAndWait flow fails.
type publishAndWaitErrorMode int

const (
	// errorOnKVPut makes kv.Put return the error.
	errorOnKVPut publishAndWaitErrorMode = iota
	// errorOnPublish makes natsClient.Publish return the error.
	errorOnPublish
	// errorOnWatch makes kv.Watch return the error.
	errorOnWatch
	// errorOnTimeout simulates a timeout by providing a channel that never sends.
	errorOnTimeout
	// errorOnTimeoutWithPartialResponse sends one valid response then blocks,
	// simulating a timeout after partial collection. Only valid for
	// publishAndCollect (broadcast) flows.
	errorOnTimeoutWithPartialResponse
)

// publishAndWaitMockOpts configures the mock behavior for publishAndWait tests.
type publishAndWaitMockOpts struct {
	// responseData is the JSON response to return from the watcher entry.
	responseData string
	// mockError is the error to inject (used with errorMode).
	mockError error
	// errorMode controls which step fails when mockError is set.
	errorMode publishAndWaitErrorMode
	// closeWatch ends the watch with no response, so the client waits for
	// its timeout.
	closeWatch bool
}

// setupPublishAndWaitMocks configures mocks for the publishAndWait flow.
func setupPublishAndWaitMocks(
	ctrl *gomock.Controller,
	mockKV *jobmocks.MockNATSClient,
	mockNATSClient *jobmocks.MockNATSClient,
	subject string,
	responseData string,
	mockError error,
) {
	setupPublishAndWaitMocksWithOpts(ctrl, mockKV, mockNATSClient, subject, &publishAndWaitMockOpts{
		responseData: responseData,
		mockError:    mockError,
		errorMode:    errorOnKVPut,
	})
}

// publishAndCollectMockOpts configures mock behavior for publishAndCollect tests.
type publishAndCollectMockOpts struct {
	// responseEntries is a list of JSON response strings to send on the watcher.
	responseEntries []string
	// mockError is the error to inject (used with errorMode).
	mockError error
	// errorMode controls which step fails when mockError is set.
	errorMode publishAndWaitErrorMode
	// closeWatch ends the watch after the response entries.
	closeWatch bool
}

// setupPublishAndCollectMocks configures mocks for the publishAndCollect flow.
func setupPublishAndCollectMocks(
	ctrl *gomock.Controller,
	mockKV *jobmocks.MockNATSClient,
	mockNATSClient *jobmocks.MockNATSClient,
	subject string,
	opts *publishAndCollectMockOpts,
) {
	if opts.mockError != nil && opts.errorMode == errorOnKVPut {
		mockKV.EXPECT().
			KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
			Return(uint64(0), opts.mockError)
		return
	}

	// kv.Put succeeds
	mockKV.EXPECT().
		KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
		Return(uint64(1), nil)

	if opts.mockError != nil && opts.errorMode == errorOnPublish {
		mockNATSClient.EXPECT().
			Publish(gomock.Any(), subject, gomock.Any()).
			Return(opts.mockError)
		return
	}

	// natsClient.Publish succeeds
	mockNATSClient.EXPECT().
		Publish(gomock.Any(), subject, gomock.Any()).
		Return(nil)

	if opts.mockError != nil && opts.errorMode == errorOnWatch {
		mockKV.EXPECT().
			KVWatch(gomock.Any(), "jobs", gomock.Any()).
			Return(nil, opts.mockError)
		return
	}

	if opts.mockError != nil && opts.errorMode == errorOnTimeout {
		// Return a channel that never sends anything, causing timeout with 0 responses
		ch := make(chan jetstream.KeyValueEntry)

		mockKV.EXPECT().
			KVWatch(gomock.Any(), "jobs", gomock.Any()).
			Return(ch, nil)
		return
	}

	if opts.mockError != nil && opts.errorMode == errorOnTimeoutWithPartialResponse {
		// Send one valid response then block so the timeout fires with partial results.
		// The first responseEntry provides the partial response data.
		ch := make(chan jetstream.KeyValueEntry, 1)
		mockEntry := jobmocks.NewMockKeyValueEntry(ctrl)
		responseData := ""
		if len(opts.responseEntries) > 0 {
			responseData = opts.responseEntries[0]
		}
		mockEntry.EXPECT().Value().Return([]byte(responseData))
		ch <- mockEntry

		mockKV.EXPECT().
			KVWatch(gomock.Any(), "jobs", gomock.Any()).
			Return(ch, nil)
		return
	}

	// Create buffered channel with all response entries
	ch := make(chan jetstream.KeyValueEntry, len(opts.responseEntries))
	for _, data := range opts.responseEntries {
		if data == "" {
			continue
		}
		mockEntry := jobmocks.NewMockKeyValueEntry(ctrl)
		mockEntry.EXPECT().Value().Return([]byte(data))
		ch <- mockEntry
	}
	if opts.closeWatch {
		close(ch)
	}

	mockKV.EXPECT().
		KVWatch(gomock.Any(), "jobs", gomock.Any()).
		Return(ch, nil)
}

// setupPublishAndWaitMocksWithOpts configures mocks with fine-grained control.
func setupPublishAndWaitMocksWithOpts(
	ctrl *gomock.Controller,
	mockKV *jobmocks.MockNATSClient,
	mockNATSClient *jobmocks.MockNATSClient,
	subject string,
	opts *publishAndWaitMockOpts,
) {
	if opts.mockError != nil && opts.errorMode == errorOnKVPut {
		mockKV.EXPECT().
			KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
			Return(uint64(0), opts.mockError)
		return
	}

	// kv.Put succeeds
	mockKV.EXPECT().
		KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
		Return(uint64(1), nil)

	if opts.mockError != nil && opts.errorMode == errorOnPublish {
		mockNATSClient.EXPECT().
			Publish(gomock.Any(), subject, gomock.Any()).
			Return(opts.mockError)
		return
	}

	// natsClient.Publish succeeds
	mockNATSClient.EXPECT().
		Publish(gomock.Any(), subject, gomock.Any()).
		Return(nil)

	if opts.mockError != nil && opts.errorMode == errorOnWatch {
		mockKV.EXPECT().
			KVWatch(gomock.Any(), "jobs", gomock.Any()).
			Return(nil, opts.mockError)
		return
	}

	if opts.mockError != nil && opts.errorMode == errorOnTimeout {
		// Return a channel that never sends anything, causing timeout
		ch := make(chan jetstream.KeyValueEntry)

		mockKV.EXPECT().
			KVWatch(gomock.Any(), "jobs", gomock.Any()).
			Return(ch, nil)
		return
	}

	if opts.closeWatch {
		ch := make(chan jetstream.KeyValueEntry)
		close(ch)

		mockKV.EXPECT().
			KVWatch(gomock.Any(), "jobs", gomock.Any()).
			Return(ch, nil)
		return
	}

	// Create mock entry with response data
	mockEntry := jobmocks.NewMockKeyValueEntry(ctrl)
	mockEntry.EXPECT().Value().Return([]byte(opts.responseData))

	ch := make(chan jetstream.KeyValueEntry, 1)
	ch <- mockEntry

	mockKV.EXPECT().
		KVWatch(gomock.Any(), "jobs", gomock.Any()).
		Return(ch, nil)
}

// TestQueryWithAgentKeyStoreEnforcing covers the single-target response path
// once the controller holds agents' keys: a response is a result only when it
// verifies against the key stored for the agent that claims to have sent it.
func (s *ClientPublicTestSuite) TestQueryWithAgentKeyStoreEnforcing() {
	const (
		target    = "server1"
		category  = "node"
		operation = job.OperationType("node.hostname.get")
		subject   = "jobs.query.host.server1"
	)

	tests := []struct {
		name         string
		responseData func(store *jobmocks.MockAgentKeyStore) []byte
		wantErr      string
		validateFunc func(resp *job.Response)
	}{
		{
			name: "a response signed by the stored key is a result",
			responseData: func(store *jobmocks.MockAgentKeyStore) []byte {
				signer, pub := newSigner(s.mockCtrl)
				store.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "server1",
						PublicKey: pub,
					}, nil)

				inner := []byte(`{"status":"completed","hostname":"server1"}`)
				wrapped, _ := client.ExportWrapInSignedEnvelope(
					signer, "machine-001", inner,
				)

				return wrapped
			},
			validateFunc: func(resp *job.Response) {
				s.Require().NotNil(resp)
				s.Equal(types.AgentJobStatusCompleted, resp.Status)
			},
		},
		{
			name: "a forged response fails the job rather than returning data",
			responseData: func(store *jobmocks.MockAgentKeyStore) []byte {
				signer, _ := newSigner(s.mockCtrl)

				// The stored record holds a different key, so the signature
				// does not verify.
				otherPub, _, _ := ed25519.GenerateKey(rand.Reader)
				store.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "server1",
						PublicKey: otherPub,
					}, nil)

				inner := []byte(`{"status":"completed","hostname":"server1"}`)
				wrapped, _ := client.ExportWrapInSignedEnvelope(
					signer, "machine-001", inner,
				)

				return wrapped
			},
			wantErr: "response verification failed",
		},
		{
			name: "an agent with no stored key is refused while enforcing",
			responseData: func(store *jobmocks.MockAgentKeyStore) []byte {
				signer, _ := newSigner(s.mockCtrl)
				store.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(nil, client.ErrResponseKeyUnknown)

				inner := []byte(`{"status":"completed","hostname":"server1"}`)
				wrapped, _ := client.ExportWrapInSignedEnvelope(
					signer, "machine-001", inner,
				)

				return wrapped
			},
			wantErr: "no stored key for responding agent",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
			data := tt.responseData(store)

			opts := &client.Options{
				Timeout:    30 * time.Second,
				JobsBucket: "jobs",
				StreamName: "JOBS",
			}
			c, err := newTrusted(s.mockNATSClient, opts)
			s.Require().NoError(err)
			c.SetAgentKeyStore(store)

			s.mockNATSClient.EXPECT().
				KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
				Return(uint64(1), nil)
			// A refused response is recorded as the host's state of the job.
			s.mockNATSClient.EXPECT().
				KVPut(gomock.Any(), "jobs", gomock.Cond(func(k any) bool {
					return strings.HasPrefix(k.(string), "status.") &&
						strings.Contains(k.(string), ".refused.")
				}), gomock.Any()).
				Return(uint64(1), nil).
				AnyTimes()
			s.mockNATSClient.EXPECT().
				Publish(gomock.Any(), subject, gomock.Any()).
				Return(nil)

			mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
			mockEntry.EXPECT().Value().Return(data)
			ch := make(chan jetstream.KeyValueEntry, 1)
			ch <- mockEntry

			s.mockNATSClient.EXPECT().
				KVWatch(gomock.Any(), "jobs", gomock.Any()).
				Return(ch, nil)

			_, resp, err := c.Query(s.ctx, target, category, operation, nil)
			if tt.wantErr != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.wantErr)

				return
			}

			s.Require().NoError(err)
			tt.validateFunc(resp)
		})
	}
}

// TestRefused holds that a refusal is the host's row also when its status
// event cannot be stored, and that an error that is not a refusal is none.
func (s *ClientPublicTestSuite) TestRefused() {
	c, err := newTrusted(s.mockNATSClient, &client.Options{JobsBucket: "jobs", StreamName: "JOBS"})
	s.Require().NoError(err)

	s.Run("a status event that cannot be stored", func() {
		s.mockNATSClient.EXPECT().
			KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
			Return(uint64(0), errors.New("nats: timeout"))
		row, ok := client.ExportRefused(s.ctx, c, "j-1",
			&client.RefusedError{MachineID: "m-1", Hostname: "web-01"})
		s.Require().True(ok)
		s.Equal(types.AgentJobStatusRefused, row.Status)
		s.Equal("web-01", row.Hostname)
	})
	s.Run("the first refusal of a machine is on its timeline and in the audit log, once", func() {
		c, err := newTrusted(s.mockNATSClient, &client.Options{
			JobsBucket: "jobs", StateBucket: "state", StreamName: "JOBS",
		})
		s.Require().NoError(err)
		log := &auditLog{}
		c.SetAudit(log)
		s.mockNATSClient.EXPECT().
			KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
			Return(uint64(1), nil).Times(3)
		s.mockNATSClient.EXPECT().
			KVPut(gomock.Any(), "state", gomock.Cond(func(k any) bool {
				return strings.HasPrefix(k.(string), "timeline.m_1.refused.")
			}), gomock.Any()).
			Return(uint64(1), nil)
		// A timeline that cannot be written still leaves the audit entry.
		s.mockNATSClient.EXPECT().
			KVPut(gomock.Any(), "state", gomock.Cond(func(k any) bool {
				return strings.HasPrefix(k.(string), "timeline.m_2.refused.")
			}), gomock.Any()).
			Return(uint64(0), errors.New("nats: timeout"))

		for _, m := range []string{"m-1", "m-1", "m-2"} {
			_, ok := client.ExportRefused(s.ctx, c, "j-1",
				&client.RefusedError{MachineID: m, Hostname: "web-01"})
			s.True(ok)
		}

		s.Require().Len(log.entries, 2)
		s.Equal(audit.ActorSystem, log.entries[0].Actor)
		s.Equal("agent.refuse", log.entries[0].Action)
		s.Equal("m-1", log.entries[0].Target)
		s.Contains(log.entries[0].RequestSummary, "not the key that the controller admitted")
		s.Equal("m-2", log.entries[1].Target)
	})
	s.Run("an audit entry that cannot be written", func() {
		c, err := newTrusted(
			s.mockNATSClient,
			&client.Options{JobsBucket: "jobs", StreamName: "JOBS"},
		)
		s.Require().NoError(err)
		log := &auditLog{err: errors.New("stream down")}
		c.SetAudit(log)
		s.mockNATSClient.EXPECT().
			KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
			Return(uint64(1), nil)
		_, ok := client.ExportRefused(s.ctx, c, "j-1",
			&client.RefusedError{MachineID: "m-1", Hostname: "web-01"})
		s.True(ok)
		s.Len(log.entries, 1)
	})
	s.Run("an error that is not a refusal", func() {
		_, ok := client.ExportRefused(s.ctx, c, "j-1", client.ErrResponseNotSigned)
		s.False(ok)
	})
	s.Run("the error says the cause", func() {
		err := &client.RefusedError{MachineID: "m-1", Hostname: "web-01"}
		s.ErrorIs(err, client.ErrResponseSignatureInvalid)
		s.Contains(err.Error(), "machine m-1")
	})
}

// TestModifyBroadcastWithAgentKeyStoreEnforcing covers the broadcast path: a
// response that does not verify is not that agent's reply, so the agent shows
// as not having answered rather than as answered by whoever sent it.
func (s *ClientPublicTestSuite) TestModifyBroadcastWithAgentKeyStoreEnforcing() {
	const (
		target    = "_all"
		category  = "node"
		operation = job.OperationType("node.hostname.get")
		subject   = "jobs.modify._all"
	)

	tests := []struct {
		name         string
		responseData func(store *jobmocks.MockAgentKeyStore) []byte
		expectedErr  string
		validateFunc func(responses map[string]*job.Response)
	}{
		{
			name: "a verified broadcast response counts as the agent's reply",
			responseData: func(store *jobmocks.MockAgentKeyStore) []byte {
				signer, pub := newSigner(s.mockCtrl)
				store.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "server1",
						PublicKey: pub,
					}, nil)

				inner := []byte(`{"status":"completed","hostname":"server1"}`)
				wrapped, _ := client.ExportWrapInSignedEnvelope(
					signer, "machine-001", inner,
				)

				return wrapped
			},
			validateFunc: func(responses map[string]*job.Response) {
				s.Len(responses, 1)
				s.Equal(types.AgentJobStatusCompleted, responses["server1"].Status)
			},
		},
		{
			name: "a response that the admitted key did not sign is a refused row",
			responseData: func(store *jobmocks.MockAgentKeyStore) []byte {
				signer, _ := newSigner(s.mockCtrl)
				otherPub, _, _ := ed25519.GenerateKey(rand.Reader)
				store.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-001").
					Return(&client.AgentKey{
						MachineID: "machine-001",
						Hostname:  "server1",
						PublicKey: otherPub,
					}, nil)
				inner := []byte(`{"status":"completed","hostname":"server1"}`)
				wrapped, _ := client.ExportWrapInSignedEnvelope(signer, "machine-001", inner)

				return wrapped
			},
			validateFunc: func(responses map[string]*job.Response) {
				s.Require().Contains(responses, "server1")
				s.Equal(types.AgentJobStatusRefused, responses["server1"].Status,
					"not completed, and not a timeout later")
				s.Contains(responses["server1"].Error, "refused the result")
			},
		},
		{
			name: "a response from an agent that claims another host's name is its own",
			responseData: func(store *jobmocks.MockAgentKeyStore) []byte {
				signer, pub := newSigner(s.mockCtrl)

				// Signature verifies, but this agent enrolled as evil-01.
				store.EXPECT().
					LookupAgentKey(gomock.Any(), "machine-evil").
					Return(&client.AgentKey{
						MachineID: "machine-evil",
						Hostname:  "evil-01",
						PublicKey: pub,
					}, nil)

				inner := []byte(`{"status":"completed","hostname":"server1"}`)
				wrapped, _ := client.ExportWrapInSignedEnvelope(
					signer, "machine-evil", inner,
				)

				return wrapped
			},
			validateFunc: func(responses map[string]*job.Response) {
				// It cannot answer for server1: server1 never answered, so
				// its row is a timeout.
				s.Require().Contains(responses, "server1")
				s.Equal(types.AgentJobStatusTimeout, responses["server1"].Status)
				s.Contains(responses, "evil-01", "its answer is under its own name")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			store := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
			data := tt.responseData(store)

			// The broadcast path also lists agents to learn who is expected to
			// answer, and that now verifies each registration. Those lookups
			// are not what this test asserts, and the registry agent was never
			// accepted, so they report no stored key.
			store.EXPECT().
				LookupAgentKey(gomock.Any(), gomock.Any()).
				Return(nil, client.ErrResponseKeyUnknown).
				AnyTimes()

			setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1"})

			opts := &client.Options{
				Timeout:        1 * time.Second,
				JobsBucket:     "jobs",
				StreamName:     "JOBS",
				RegistryBucket: "registry",
			}
			c, err := newTrusted(s.mockNATSClient, opts)
			s.Require().NoError(err)
			c.SetAgentKeyStore(store)

			s.mockNATSClient.EXPECT().
				KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
				Return(uint64(1), nil)
			// A refused response is recorded as the host's state of the job.
			s.mockNATSClient.EXPECT().
				KVPut(gomock.Any(), "jobs", gomock.Cond(func(k any) bool {
					return strings.HasPrefix(k.(string), "status.") &&
						strings.Contains(k.(string), ".refused.")
				}), gomock.Any()).
				Return(uint64(1), nil).
				AnyTimes()
			s.mockNATSClient.EXPECT().
				Publish(gomock.Any(), subject, gomock.Any()).
				Return(nil)

			mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
			mockEntry.EXPECT().Value().Return(data)
			ch := make(chan jetstream.KeyValueEntry, 1)
			ch <- mockEntry

			s.mockNATSClient.EXPECT().
				KVWatch(gomock.Any(), "jobs", gomock.Any()).
				Return(ch, nil)

			_, responses, err := c.ModifyBroadcast(
				s.ctx, target, category, operation, nil,
			)
			if tt.expectedErr != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.expectedErr)

				return
			}

			s.Require().NoError(err)
			tt.validateFunc(responses)
		})
	}
}

// TestAuditRecordsSteps covers the steps that a dispatch gives the audit
// entry of its request: the steps of each host that answers, also when the job
// failed on the host.
func (s *ClientPublicTestSuite) TestAuditRecordsSteps() {
	failedResp := `{"status":"failed","error":"permission denied","hostname":"server1",` +
		`"steps":[{"behavior":"managedfile","method":"update","error":"denied","duration_ms":3}]}`
	host1Resp := `{"status":"completed","hostname":"server1",` +
		`"steps":[{"behavior":"managedfile","method":"update","duration_ms":3}]}`
	host2Resp := `{"status":"completed","hostname":"server2"}`

	s.Run("a job that failed on its host", func() {
		setupPublishAndWaitMocks(
			s.mockCtrl,
			s.mockNATSClient,
			s.mockNATSClient,
			"jobs.modify.host.server1",
			failedResp,
			nil,
		)
		ctx := audit.WithJobIDSink(s.ctx)

		_, _, err := s.jobsClient.Modify(
			ctx, "server1", "node", job.OperationType("node.sysctl.update"), nil,
		)

		s.Error(err)
		s.Equal([]audit.HostSteps{{
			Hostname: "server1",
			Steps: []types.Step{
				{Behavior: "managedfile", Method: "update", Error: "denied", DurationMS: 3},
			},
		}}, audit.StepsFrom(ctx))
	})

	s.Run("a broadcast records the steps of each host that made a call", func() {
		setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1", "server2"})
		c := s.newClientWithRegistry()
		setupPublishAndCollectMocks(
			s.mockCtrl,
			s.mockNATSClient,
			s.mockNATSClient,
			"jobs.modify._all",
			&publishAndCollectMockOpts{responseEntries: []string{host1Resp, host2Resp}},
		)
		ctx := audit.WithJobIDSink(s.ctx)

		_, _, err := c.ModifyBroadcast(
			ctx, "_all", "node", job.OperationType("node.sysctl.update"), nil,
		)

		s.NoError(err)
		s.Equal([]audit.HostSteps{{
			Hostname: "server1",
			Steps: []types.Step{
				{Behavior: "managedfile", Method: "update", DurationMS: 3},
			},
		}}, audit.StepsFrom(ctx))
	})
}

// TestCancelIsNotATimeout: a caller whose ctx ends first stopped waiting.
// No agent timed out, so the error says the wait ended, and a broadcast
// fills no timeout rows. Before, both said timeout.
func (s *ClientPublicTestSuite) TestCancelIsNotATimeout() {
	opts := &client.Options{
		Timeout:        5 * time.Second,
		JobsBucket:     "jobs",
		StreamName:     "JOBS",
		RegistryBucket: "registry",
	}
	cancelSoon := func() context.Context {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(20*time.Millisecond, cancel)

		return ctx
	}

	s.Run("a query", func() {
		c, err := newTrusted(s.mockNATSClient, opts)
		s.Require().NoError(err)
		setupPublishAndWaitMocksWithOpts(
			s.mockCtrl, s.mockNATSClient, s.mockNATSClient,
			"jobs.query.host.server1",
			&publishAndWaitMockOpts{closeWatch: true},
		)

		_, _, err = c.Query(cancelSoon(), "server1", "node",
			job.OperationType("node.hostname.get"), nil)

		s.ErrorIs(err, context.Canceled)
		s.ErrorContains(err, "wait for job response")
		s.NotContains(err.Error(), "timeout")
	})

	s.Run("a broadcast with one answer", func() {
		setupRegistryKV(s.mockCtrl, s.mockNATSClient, []string{"server1", "server2"})
		c, err := newTrusted(s.mockNATSClient, opts)
		s.Require().NoError(err)
		setupPublishAndCollectMocks(
			s.mockCtrl, s.mockNATSClient, s.mockNATSClient,
			"jobs.query._all",
			&publishAndCollectMockOpts{
				responseEntries: []string{`{"status":"completed","hostname":"server1"}`},
				mockError:       errors.New("partial"),
				errorMode:       errorOnTimeoutWithPartialResponse,
			},
		)

		_, results, err := c.QueryBroadcast(cancelSoon(), "_all", "node",
			job.OperationType("node.hostname.get"), nil)

		s.ErrorIs(err, context.Canceled)
		s.ErrorContains(err, "wait for broadcast responses")
		s.Nil(results, "no timeout rows")
	})
}

type createInput struct {
	Name string                  `json:"name"`
	Env  map[string]secret.Value `json:"env,omitempty"`
}

// registry makes the registry hold an agent for each registration.
func (s *ClientPublicTestSuite) registry(
	regs ...job.AgentRegistration,
) {
	keys := make([]string, 0, len(regs))
	for _, reg := range regs {
		keys = append(keys, "agents."+reg.MachineID)
		data, _ := json.Marshal(reg)
		entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
		entry.EXPECT().Value().Return(data)
		s.mockNC.EXPECT().
			KVGet(gomock.Any(), "registry", "agents."+reg.MachineID).
			Return(entry, nil)
	}
	s.mockNC.EXPECT().KVKeys(gomock.Any(), "registry").Return(keys, nil)
}

// store gives each machine ID its encryption key, and no record for any
// other.
func (s *ClientPublicTestSuite) store(
	keys map[string]string,
) *jobmocks.MockAgentKeyStore {
	st := jobmocks.NewMockAgentKeyStore(s.mockCtrl)
	st.EXPECT().LookupAgentKey(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, id string) (*client.AgentKey, error) {
			key, ok := keys[id]
			if !ok {
				return nil, client.ErrResponseKeyUnknown
			}

			return &client.AgentKey{MachineID: id, EncryptionKey: key}, nil
		}).AnyTimes()

	return st
}

func (s *ClientPublicTestSuite) newClientSeal(
	store client.AgentKeyStore,
) *client.Client {
	c, err := client.New(slog.Default(), s.mockNC, &client.Options{
		Timeout:        30 * time.Second,
		JobsBucket:     "jobs",
		RegistryBucket: "registry",
		StreamName:     "JOBS",
	})
	s.Require().NoError(err)
	if store != nil {
		c.SetAgentKeyStore(store)
	}

	return c
}

func (s *ClientPublicTestSuite) pub(
	kp *ecdh.PrivateKey,
) string {
	return pki.EncodeEncryptionKey(kp.PublicKey())
}

// opened is the env of req's data, opened by own, the key of the agent id.
func (s *ClientPublicTestSuite) opened(
	req *job.Request,
	id string,
	own *ecdh.PrivateKey,
) map[string]secret.Value {
	o, err := seal.Open(req.JobID, req.SealedKeys[id], req.SealedFrom, own)
	s.Require().NoError(err)
	var in createInput
	s.Require().NoError(json.Unmarshal(req.Data, &in))
	s.Require().NoError(secret.Map(&in, o.Value))

	return in.Env
}

func (s *ClientPublicTestSuite) TestRequest() {
	web := job.AgentRegistration{
		MachineID: "m-web",
		Hostname:  "web-01",
		Labels:    map[string]string{"group": "web"},
	}
	db := job.AgentRegistration{MachineID: "m-db", Hostname: "db-01"}
	old := job.AgentRegistration{MachineID: "m-old", Hostname: "old-01"}
	input := createInput{Name: "app", Env: map[string]secret.Value{"TOKEN": "hunter2"}}

	tests := []struct {
		name         string
		store        func() client.AgentKeyStore
		setup        func()
		target       string
		resolved     string
		data         any
		wantErr      string
		validateFunc func(req *job.Request)
	}{
		{
			name:     "when the job client has no agent keys, a job with a secret is refused",
			store:    func() client.AgentKeyStore { return nil },
			target:   "web-01",
			resolved: "m-web",
			data:     input,
			wantErr:  client.ErrNoAgentKeys.Error(),
		},
		{
			name: "when the data has no secret value, nothing is sealed",
			store: func() client.AgentKeyStore {
				return s.store(nil)
			},
			target:   "web-01",
			resolved: "m-web",
			data:     createInput{Name: "app"},
			validateFunc: func(req *job.Request) {
				s.JSONEq(`{"name":"app"}`, string(req.Data))
				s.Empty(req.SealedKeys)
			},
		},
		{
			name: "when the target is _all, the job key is sealed to each agent with a key",
			store: func() client.AgentKeyStore {
				return s.store(map[string]string{"m-web": s.pub(s.web), "m-db": s.pub(s.db)})
			},
			setup:  func() { s.registry(web, db, old) },
			target: "_all",
			data:   input,
			validateFunc: func(req *job.Request) {
				s.NotContains(string(req.Data), "hunter2", "the job never has the value")
				s.Contains(string(req.Data), seal.Prefix)
				s.Len(req.SealedKeys, 2, "an agent with no key gets none")
				s.Equal(map[string]secret.Value{"TOKEN": "hunter2"}, s.opened(req, "m-web", s.web))
				s.Equal(map[string]secret.Value{"TOKEN": "hunter2"}, s.opened(req, "m-db", s.db))
				op := client.ExportOperation(req)
				s.Equal(req.SealedKeys, op["sealed_keys"])
				s.Equal(req.SealedFrom, op["sealed_from"])
			},
		},
		{
			name: "when the target is one host, the job key is sealed only to it",
			store: func() client.AgentKeyStore {
				return s.store(map[string]string{"m-web": s.pub(s.web), "m-db": s.pub(s.db)})
			},
			setup:    func() { s.registry(web, db) },
			target:   "web-01",
			resolved: "m-web",
			data:     input,
			validateFunc: func(req *job.Request) {
				s.Equal([]string{"m-web"}, keysOf(req.SealedKeys))
				s.Equal(map[string]secret.Value{"TOKEN": "hunter2"}, s.opened(req, "m-web", s.web))
			},
		},
		{
			name: "when the agents cannot be listed, it is an error",
			store: func() client.AgentKeyStore {
				return s.store(nil)
			},
			setup: func() {
				s.mockNC.EXPECT().
					KVKeys(gomock.Any(), "registry").
					Return(nil, errors.New("nats: timeout"))
			},
			target:  "_all",
			data:    input,
			wantErr: "list the agents to seal the job to",
		},
		{
			name: "when the key of an agent is not a P-256 key, it is an error",
			store: func() client.AgentKeyStore {
				return s.store(map[string]string{"m-web": "not-a-key"})
			},
			setup:    func() { s.registry(web) },
			target:   "web-01",
			resolved: "m-web",
			data:     input,
			wantErr:  "seal the job key",
		},
		{
			name: "when the job key cannot be made, it is an error",
			store: func() client.AgentKeyStore {
				return s.store(nil)
			},
			setup: func() {
				client.SetNewSealer(func(string) (*seal.Sealer, error) {
					return nil, errors.New("no entropy")
				})
			},
			target:  "web-01",
			data:    input,
			wantErr: "no entropy",
		},
		{
			name:    "when the data does not marshal, it is an error",
			store:   func() client.AgentKeyStore { return nil },
			target:  "web-01",
			data:    map[string]any{"bad": make(chan int)},
			wantErr: "marshal data",
		},
		{
			name: "when sealed data does not marshal, it is an error",
			store: func() client.AgentKeyStore {
				return s.store(nil)
			},
			target: "web-01",
			data: struct {
				Token secret.Value `json:"token"`
				Bad   chan int     `json:"bad"`
			}{Token: "hunter2", Bad: make(chan int)},
			wantErr: "marshal data",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			defer client.ResetNewSealer()
			if tc.setup != nil {
				tc.setup()
			}
			c := s.newClientSeal(tc.store())

			req, err := c.ExportRequest(s.ctx, types.JobTypeModify, tc.target, tc.resolved,
				job.CategoryBehavior, "docker.create", tc.data)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(types.JobTypeModify, req.Type)
			tc.validateFunc(req)
		})
	}
}

func keysOf(
	m map[string]string,
) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	return out
}
