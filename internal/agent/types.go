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

package agent

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel/metric"

	"github.com/retr0h/osapi/internal/agent/pki"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/client"
	"github.com/retr0h/osapi/internal/provider/network/netinfo"
	"github.com/retr0h/osapi/internal/provider/node/disk"
	"github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
	"github.com/retr0h/osapi/internal/telemetry/process"
)

// NATSPublisher publishes messages to NATS subjects.
// Satisfied by the nats-client's Client type.
type NATSPublisher interface {
	Publish(ctx context.Context, subject string, data []byte) error
}

// Agent implements job processing with clean lifecycle management.
type Agent struct {
	logger      *slog.Logger
	appConfig   config.Config
	appFs       avfs.VFS
	jobClient   client.JobClient
	streamName  string
	execManager exec.Manager

	// registry dispatches job requests to the appropriate processor.
	registry *ProviderRegistry

	// Providers used directly by the heartbeat and facts collection.
	// These must remain as direct fields so the heartbeat can read
	// system info independently of job processing.
	hostProvider    host.Provider
	diskProvider    disk.Provider
	memProvider     mem.Provider
	loadProvider    load.Provider
	netinfoProvider netinfo.Provider

	// Process provider for self-metrics in heartbeat.
	processProvider process.Provider

	// Registry KV for heartbeat registration.
	registryKV jetstream.KeyValue

	// Facts KV for system facts collection.
	factsKV jetstream.KeyValue

	// startedAt records when the agent process started.
	startedAt time.Time

	// prevConditions tracks condition state between heartbeats.
	prevConditions []job.Condition

	// prevProcessConditions tracks process condition state between heartbeats.
	prevProcessConditions []job.Condition

	// cpuCount cached from facts for HighLoad evaluation.
	cpuCount int

	// cachedFacts holds the latest collected facts for @fact.X resolution.
	cachedFacts *job.FactsRegistration

	// state is the agent's scheduling state (Ready, Draining, Cordoned).
	state string

	// pkiManager handles keypair and enrollment lifecycle (nil when PKI disabled).
	pkiManager *pki.Manager

	// natsClient is used for publishing enrollment requests via NATS.
	// Uses the same NATSClient interface as the rest of the project.
	natsClient NATSPublisher

	// machineID is the permanent host identifier resolved at startup.
	machineID string

	// hostname cached from Start for drain/undrain resubscribe.
	hostname string

	// Lifecycle management.
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Consumer lifecycle for drain/undrain.
	consumerCtx    context.Context
	consumerCancel context.CancelFunc
	consumerWg     sync.WaitGroup

	// subComponents reports the status of internal services in heartbeat.
	subComponents map[string]job.SubComponentInfo

	// Subsystem-specific loggers for background processes.
	heartbeatLogger *slog.Logger
	factsLogger     *slog.Logger

	// OTEL instruments for job metrics.
	jobsProcessed metric.Int64Counter
	jobsActive    metric.Int64UpDownCounter
	jobDuration   metric.Float64Histogram

	// lastHeartbeatTime stores time.Time of the last successful heartbeat write.
	lastHeartbeatTime atomic.Value
}

// JobContext contains the context and data for a single job execution.
type JobContext struct {
	// JobID from the original job request
	JobID string
	// AgentHostname identifies which agent is processing this job
	AgentHostname string
	// JobData contains the raw job request data
	JobData []byte
	// ResponseKV is the key-value bucket for storing responses
	ResponseKV string
}
