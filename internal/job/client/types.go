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
	"github.com/retr0h/osapi/internal/provider/command"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/provider/network/ping"
)

// JobClient defines the interface for interacting with the jobs system.
type JobClient interface {
	// Job queue management operations
	CreateJob(
		ctx context.Context,
		operationData map[string]interface{},
		targetHostname string,
	) (*CreateJobResult, error)
	GetQueueStats(
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

	// Query operations — all return (jobID, result..., error)
	QueryNodeStatus(
		ctx context.Context,
		hostname string,
	) (string, *job.NodeStatusResponse, error)
	QueryNodeStatusAny(
		ctx context.Context,
	) (string, *job.NodeStatusResponse, error)
	QueryNodeStatusAll(
		ctx context.Context,
	) (string, []*job.NodeStatusResponse, map[string]string, error)
	QueryNodeStatusBroadcast(
		ctx context.Context,
		target string,
	) (string, []*job.NodeStatusResponse, map[string]string, error)
	QueryNodeHostname(
		ctx context.Context,
		hostname string,
	) (string, string, *job.AgentInfo, error)
	QueryNodeHostnameAll(
		ctx context.Context,
	) (string, map[string]*job.AgentInfo, map[string]string, error)
	QueryNodeHostnameBroadcast(
		ctx context.Context,
		target string,
	) (string, map[string]*job.AgentInfo, map[string]string, error)
	QueryNetworkDNS(
		ctx context.Context,
		hostname string,
		iface string,
	) (string, *dns.Config, string, error)
	QueryNetworkDNSAll(
		ctx context.Context,
		iface string,
	) (string, map[string]*dns.Config, map[string]string, error)
	QueryNetworkDNSBroadcast(
		ctx context.Context,
		target string,
		iface string,
	) (string, map[string]*dns.Config, map[string]string, error)

	// Modify operations — all return (jobID, result..., error)
	ModifyNetworkDNS(
		ctx context.Context,
		hostname string,
		servers []string,
		searchDomains []string,
		iface string,
	) (string, string, bool, error)
	ModifyNetworkDNSAny(
		ctx context.Context,
		servers []string,
		searchDomains []string,
		iface string,
	) (string, string, bool, error)
	ModifyNetworkDNSAll(
		ctx context.Context,
		servers []string,
		searchDomains []string,
		iface string,
	) (string, map[string]error, map[string]bool, error)
	ModifyNetworkDNSBroadcast(
		ctx context.Context,
		target string,
		servers []string,
		searchDomains []string,
		iface string,
	) (string, map[string]error, map[string]bool, error)
	QueryNetworkPing(
		ctx context.Context,
		hostname string,
		address string,
	) (string, *ping.Result, string, error)
	QueryNetworkPingAny(
		ctx context.Context,
		address string,
	) (string, *ping.Result, string, error)
	QueryNetworkPingAll(
		ctx context.Context,
		address string,
	) (string, map[string]*ping.Result, map[string]string, error)
	QueryNetworkPingBroadcast(
		ctx context.Context,
		target string,
		address string,
	) (string, map[string]*ping.Result, map[string]string, error)

	// Command operations — all return (jobID, result..., error)
	ModifyCommandExec(
		ctx context.Context,
		hostname string,
		cmdName string,
		args []string,
		cwd string,
		timeout int,
	) (string, *command.Result, string, error)
	ModifyCommandExecBroadcast(
		ctx context.Context,
		target string,
		cmdName string,
		args []string,
		cwd string,
		timeout int,
	) (string, map[string]*command.Result, map[string]string, error)
	ModifyCommandShell(
		ctx context.Context,
		hostname string,
		cmdStr string,
		cwd string,
		timeout int,
	) (string, *command.Result, string, error)
	ModifyCommandShellBroadcast(
		ctx context.Context,
		target string,
		cmdStr string,
		cwd string,
		timeout int,
	) (string, map[string]*command.Result, map[string]string, error)

	// Worker discovery
	ListAgents(
		ctx context.Context,
	) ([]job.AgentInfo, error)

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

	// Worker operations - used by job workers for processing
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
	Jobs       []*job.QueuedJob
	TotalCount int
}

// computedJobStatus represents the computed status from events
type computedJobStatus struct {
	Status       string
	Error        string
	Hostname     string
	UpdatedAt    string
	AgentStates map[string]job.AgentState
	Timeline     []job.TimelineEvent
}

// Ensure Client implements JobClient interface
var _ JobClient = (*Client)(nil)
