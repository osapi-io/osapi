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

	"github.com/retr0h/osapi/internal/job"
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
	GetQueueStats(ctx context.Context) (*job.QueueStats, error)
	GetJobStatus(ctx context.Context, jobID string) (*job.QueuedJob, error)
	ListJobs(ctx context.Context, statusFilter string) ([]*job.QueuedJob, error)

	// Query operations
	QuerySystemStatus(ctx context.Context, hostname string) (*job.SystemStatusResponse, error)
	QuerySystemStatusAny(ctx context.Context) (*job.SystemStatusResponse, error)
	QuerySystemHostname(ctx context.Context, hostname string) (string, error)
	QueryNetworkDNS(ctx context.Context, hostname, iface string) (*dns.Config, error)

	// Modify operations
	ModifyNetworkDNS(
		ctx context.Context,
		hostname string,
		servers []string,
		searchDomains []string,
		iface string,
	) error
	ModifyNetworkDNSAny(
		ctx context.Context,
		servers []string,
		searchDomains []string,
		iface string,
	) error
	ModifyNetworkPing(
		ctx context.Context,
		hostname string,
		address string,
	) (*ping.Result, error)
	ModifyNetworkPingAny(
		ctx context.Context,
		address string,
	) (*ping.Result, error)
}

// Ensure Client implements JobClient interface
var _ JobClient = (*Client)(nil)
