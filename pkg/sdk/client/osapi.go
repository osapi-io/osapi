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

// Package client provides a Go SDK for the OSAPI REST API.
//
// Create a client with New() and use the domain-specific services
// to interact with the API:
//
//	client := client.New("http://localhost:8080", "your-jwt-token")
//
//	// Get hostname
//	resp, err := client.Hostname.Get(ctx, "_any")
//
//	// Execute a command
//	resp, err := client.Command.Exec(ctx, client.ExecRequest{
//	    Command: "uptime",
//	    Target:  "_all",
//	})
package client

import (
	"log/slog"
	"net/http"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// Client is the top-level OSAPI SDK client. Use New() to create one.
type Client struct {
	// Agent provides agent discovery and details operations.
	Agent *AgentService

	// Status provides full node status queries (OS, disk, memory, load).
	Status *StatusService

	// Hostname provides hostname query and update operations.
	Hostname *HostnameService

	// Disk provides disk usage query operations.
	Disk *DiskService

	// Memory provides memory usage query operations.
	Memory *MemoryService

	// Load provides load average query operations.
	Load *LoadService

	// Uptime provides uptime query operations.
	Uptime *UptimeService

	// OS provides operating system info query operations.
	OS *OSService

	// DNS provides DNS configuration query and update operations.
	DNS *DNSService

	// Ping provides network ping operations.
	Ping *PingService

	// Command provides command execution operations (exec, shell).
	Command *CommandService

	// FileDeploy provides file deployment operations on target hosts.
	FileDeploy *FileDeployService

	// Job provides job queue operations (create, get, list, delete, retry).
	Job *JobService

	// Health provides health check operations (liveness, readiness, status).
	Health *HealthService

	// Audit provides audit log operations (list, get, export).
	Audit *AuditService

	// File provides file management operations (upload, list, get, delete).
	File *FileService

	// Docker provides Docker container management operations (create, list,
	// inspect, start, stop, remove, exec, pull).
	Docker *DockerService

	// Cron provides cron schedule management operations (list, get,
	// create, update, delete).
	Cron *CronService

	// Sysctl provides sysctl parameter management operations (list, get,
	// set, delete).
	Sysctl *SysctlService

	// NTP provides NTP management operations (get, create, update, delete).
	NTP *NTPService

	// Timezone provides system timezone management operations (get, update).
	Timezone *TimezoneService

	// Process provides process management operations (list, get, signal).
	Process *ProcessService

	// Power provides power management operations (reboot, shutdown).
	Power *PowerService

	httpClient    *gen.ClientWithResponses
	baseURL       string
	logger        *slog.Logger
	baseTransport http.RoundTripper
}

// Option configures the Client.
type Option func(*Client)

// WithLogger sets a custom logger. Defaults to slog.Default().
func WithLogger(
	logger *slog.Logger,
) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithHTTPTransport sets a custom base HTTP transport.
func WithHTTPTransport(
	transport http.RoundTripper,
) Option {
	return func(c *Client) {
		c.baseTransport = transport
	}
}

// New creates an OSAPI SDK client.
func New(
	baseURL string,
	bearerToken string,
	opts ...Option,
) *Client {
	c := &Client{
		baseURL:       baseURL,
		logger:        slog.Default(),
		baseTransport: http.DefaultTransport,
	}

	for _, opt := range opts {
		opt(c)
	}

	transport := &authTransport{
		base:       c.baseTransport,
		authHeader: "Bearer " + bearerToken,
		logger:     c.logger,
	}

	hc := &http.Client{
		Transport: transport,
	}

	// Error is unreachable: the only ClientOption passed (WithHTTPClient) cannot
	// fail, and NewClientWithResponses only errors when a ClientOption does.
	// Invalid URLs are caught later at HTTP call time with a clear parse error.
	httpClient, _ := gen.NewClientWithResponses(baseURL, gen.WithHTTPClient(hc))

	c.httpClient = httpClient
	c.Agent = &AgentService{client: httpClient}
	c.Status = &StatusService{client: httpClient}
	c.Hostname = &HostnameService{client: httpClient}
	c.Disk = &DiskService{client: httpClient}
	c.Memory = &MemoryService{client: httpClient}
	c.Load = &LoadService{client: httpClient}
	c.Uptime = &UptimeService{client: httpClient}
	c.OS = &OSService{client: httpClient}
	c.DNS = &DNSService{client: httpClient}
	c.Ping = &PingService{client: httpClient}
	c.Command = &CommandService{client: httpClient}
	c.FileDeploy = &FileDeployService{client: httpClient}
	c.Job = &JobService{client: httpClient}
	c.Health = &HealthService{client: httpClient}
	c.Audit = &AuditService{client: httpClient}
	c.File = &FileService{client: httpClient}
	c.Docker = &DockerService{client: httpClient}
	c.Cron = &CronService{client: httpClient}
	c.Sysctl = &SysctlService{client: httpClient}
	c.NTP = &NTPService{client: httpClient}
	c.Timezone = &TimezoneService{client: httpClient}
	c.Process = &ProcessService{client: httpClient}
	c.Power = &PowerService{client: httpClient}

	return c
}
