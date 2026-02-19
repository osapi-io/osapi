// Copyright (c) 2024 John Dewey

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

	"github.com/retr0h/osapi/internal/client/gen"
)

// CombinedHandler is a superset of all smaller handler interfaces.
type CombinedHandler interface {
	HealthHandler
	NetworkHandler
	SystemHandler
	JobHandler
}

// HealthHandler defines an interface for interacting with Health client operations.
type HealthHandler interface {
	// GetHealth get the health liveness API endpoint.
	GetHealth(
		ctx context.Context,
	) (*gen.GetHealthResponse, error)
	// GetHealthReady get the health readiness API endpoint.
	GetHealthReady(
		ctx context.Context,
	) (*gen.GetHealthReadyResponse, error)
	// GetHealthDetailed get the health detailed API endpoint.
	GetHealthDetailed(
		ctx context.Context,
	) (*gen.GetHealthDetailedResponse, error)
}

// JobHandler defines an interface for interacting with Job client operations.
type JobHandler interface {
	// PostJob creates a new job via the REST API.
	PostJob(
		ctx context.Context,
		operation map[string]interface{},
		targetHostname string,
	) (*gen.PostJobResponse, error)

	// GetJobByID retrieves a specific job by ID via the REST API.
	GetJobByID(
		ctx context.Context,
		id string,
	) (*gen.GetJobByIDResponse, error)

	// DeleteJobByID deletes a specific job by ID via the REST API.
	DeleteJobByID(
		ctx context.Context,
		id string,
	) (*gen.DeleteJobByIDResponse, error)

	// GetJobs retrieves jobs, optionally filtered by status, via the REST API.
	GetJobs(
		ctx context.Context,
		status string,
	) (*gen.GetJobResponse, error)

	// GetJobQueueStats retrieves queue statistics via the REST API.
	GetJobQueueStats(
		ctx context.Context,
	) (*gen.GetJobStatusResponse, error)

	// GetJobWorkers retrieves active workers via the REST API.
	GetJobWorkers(
		ctx context.Context,
	) (*gen.GetJobWorkersResponse, error)
}

// NetworkHandler defines an interface for interacting with Network client operations.
type NetworkHandler interface {
	// GetNetworkDNSByInterface get the network dns get API endpoint.
	GetNetworkDNSByInterface(
		ctx context.Context,
		hostname string,
		interfaceName string,
	) (*gen.GetNetworkDNSByInterfaceResponse, error)

	// PutNetworkDNS put the network dns put API endpoint.
	PutNetworkDNS(
		ctx context.Context,
		hostname string,
		servers []string,
		searchDomains []string,
		interfaceName string,
	) (*gen.PutNetworkDNSResponse, error)
	// PostNetworkPing post the network ping API endpoint.
	PostNetworkPing(
		ctx context.Context,
		hostname string,
		address string,
	) (*gen.PostNetworkPingResponse, error)
}

// SystemHandler defines an interface for interacting with System client operations.
type SystemHandler interface {
	// GetSystemStatus get the system status API endpoint.
	GetSystemStatus(
		ctx context.Context,
		hostname string,
	) (*gen.GetSystemStatusResponse, error)
	// GetSystemHostname get the system hostname API endpoint.
	GetSystemHostname(
		ctx context.Context,
		hostname string,
	) (*gen.GetSystemHostnameResponse, error)
}
