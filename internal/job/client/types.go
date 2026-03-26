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

package client

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/retr0h/osapi/internal/job"
)

const (
	// DefaultPageSize is the default number of jobs per page.
	DefaultPageSize = 10
	// MaxPageSize is the maximum allowed page size.
	MaxPageSize = 100
)

// JobClient defines the interface for interacting with the jobs system.
type JobClient interface {
	// Generic dispatch operations — used by API handlers to submit jobs
	// without importing typed wrapper methods.
	Query(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, *job.Response, error)
	QueryBroadcast(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, map[string]*job.Response, map[string]string, error)
	Modify(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, *job.Response, error)
	ModifyBroadcast(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, map[string]*job.Response, map[string]string, error)

	// Job queue management operations
	GetQueueSummary(
		ctx context.Context,
	) (*job.QueueStats, error)
	GetJobStatus(
		ctx context.Context,
		jobID string,
	) (*job.QueuedJob, error)
	ListJobs(
		ctx context.Context,
		statusFilter string,
		limit int,
		offset int,
	) (*ListJobsResult, error)

	// Agent discovery
	ListAgents(
		ctx context.Context,
	) ([]job.AgentInfo, error)
	GetAgent(
		ctx context.Context,
		hostname string,
	) (*job.AgentInfo, error)

	// Agent timeline
	WriteAgentTimelineEvent(
		ctx context.Context,
		hostname, event, message string,
	) error
	GetAgentTimeline(
		ctx context.Context,
		hostname string,
	) ([]job.TimelineEvent, error)

	// Agent drain flag
	CheckDrainFlag(
		ctx context.Context,
		hostname string,
	) bool
	SetDrainFlag(
		ctx context.Context,
		hostname string,
	) error
	DeleteDrainFlag(
		ctx context.Context,
		hostname string,
	) error

	// Job deletion
	DeleteJob(
		ctx context.Context,
		jobID string,
	) error

	// Job retry
	RetryJob(
		ctx context.Context,
		jobID string,
		targetHostname string,
	) (*CreateJobResult, error)

	// Agent operations - used by agents for processing
	WriteStatusEvent(
		ctx context.Context,
		jobID string,
		event string,
		hostname string,
		data map[string]interface{},
	) error
	WriteJobResponse(
		ctx context.Context,
		jobID string,
		hostname string,
		responseData []byte,
		status string,
		errorMsg string,
		changed *bool,
	) error
	ConsumeJobs(
		ctx context.Context,
		streamName string,
		consumerName string,
		handler func(jetstream.Msg) error,
		opts *natsclient.ConsumeOptions,
	) error
	GetJobData(
		ctx context.Context,
		jobKey string,
	) ([]byte, error)
	CreateOrUpdateConsumer(
		ctx context.Context,
		streamName string,
		consumerConfig jetstream.ConsumerConfig,
	) error
}

// CreateJobResult represents the result of creating a job.
type CreateJobResult struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Revision  uint64 `json:"revision"`
	Timestamp string `json:"timestamp"`
}

// ListJobsResult represents the result of listing jobs with pagination.
type ListJobsResult struct {
	Jobs         []*job.QueuedJob
	TotalCount   int
	StatusCounts map[string]int
}

// computedJobStatus represents the computed status from events
type computedJobStatus struct {
	Status      string
	Error       string
	Hostname    string
	UpdatedAt   string
	AgentStates map[string]job.AgentState
	Timeline    []job.TimelineEvent
}

// lightJobInfo holds status derived from KV key names only (no reads).
type lightJobInfo struct {
	Status string
}

// Ensure Client implements JobClient interface
var _ JobClient = (*Client)(nil)
