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

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/metric"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/exec"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/provider/node/disk"
	"github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
	"github.com/osapi-io/osapi/internal/telemetry/process"
)

// NATSClient is the access to NATS that the agent uses itself: the core
// messages of enrollment, and the writes of its registration and its facts.
// A *natsclient.Client is one.
type NATSClient interface {
	PublishCore(ctx context.Context, subject string, data []byte) error
	Subscribe(
		ctx context.Context,
		subject string,
		handler natsclient.MsgHandler,
	) (*nats.Subscription, error)
	KVPut(ctx context.Context, bucket string, key string, value []byte) (uint64, error)
	KVDelete(ctx context.Context, bucket string, key string) error
	KVGet(ctx context.Context, bucket string, key string) (jetstream.KeyValueEntry, error)
}

// Agent implements job processing with clean lifecycle management.
type Agent struct {
	// renamed tells the heartbeat that a hostname update completed, so it
	// reads the new hostname and registers with it at once, not at its next
	// tick: until then the controller, which took the new name from the
	// update's answer, finds no verified registration of this agent.
	renamed     chan chan struct{}
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
	hostProvider   host.Provider
	diskProvider   disk.Provider
	memoryProvider memory.Provider
	loadProvider   load.Provider

	// Process provider for self-metrics in heartbeat.
	processProvider process.Provider

	// registryBucket is the KV bucket of the heartbeat registration, or "".
	registryBucket string

	// seen is what the agent last reported, which the timeline compares
	// each heartbeat and each facts write with (osapi-io/osapi#1454).
	seen seenState

	// factsBucket is the KV bucket of the system facts, or "".
	factsBucket string

	// startedAt records when the agent process started.
	startedAt time.Time

	// version is the build version of the agent binary, from SetVersion. It
	// is in the started event of each job and in the registration.
	version string
	// fileLabels are the labels of the agent's config file, which the
	// labels of the controller's config are laid over.
	fileLabels map[string]string
	// fileMaxJobs is the max_jobs of the agent's config file, which the
	// controller's config replaces when it sets one.
	fileMaxJobs int
	// logLevel is the level of the process's logger, from SetLogLevel, and
	// fileLogLevel the level it had then: the one of --debug or debug in the
	// config file, which the controller's config replaces when it sets one.
	// A nil logLevel is a logger whose level the agent does not change.
	logLevel     *slog.LevelVar
	fileLogLevel slog.Level
	// fileFactsInterval is the facts interval of the agent's config file,
	// which the controller's config replaces when it sets one, and
	// factsInterval the one that the controller's config set, or 0 for the
	// file's. factsReset gives the facts loop a new one.
	fileFactsInterval string
	factsInterval     time.Duration
	factsReset        chan time.Duration
	// configVersion is the version of the controller's config that the
	// agent applied, 0 for none.
	configVersion int

	// prevConditions tracks condition state between heartbeats.
	prevConditions []job.Condition

	// prevProcessConditions tracks process condition state between heartbeats.
	prevProcessConditions []job.Condition
	// seenWrittenAt is when the heartbeat last wrote the agent's last-seen
	// time; it writes it at most once each seenInterval.
	seenWrittenAt time.Time

	// cpuCount cached from facts for HighLoad evaluation.
	cpuCount int

	// cachedFacts holds the latest collected facts for @fact.X resolution.
	// The refresh goroutine replaces it while job goroutines read it, so it is
	// swapped as a whole rather than mutated in place.
	cachedFacts atomic.Pointer[job.FactsRegistration]

	// lifecycleMu guards state and the consumer context below. A transition is
	// a state change and a consumer start or stop together, so the two must be
	// taken as one: draining an agent at the moment its enrollment is accepted
	// otherwise interleaves stopping and starting the consumers, and the agent ends
	// up cordoned with live consumers or ready with none.
	lifecycleMu sync.RWMutex

	// state is the agent's scheduling state (Ready, Draining, Cordoned).
	// Guarded by lifecycleMu.
	state types.AgentState

	// pkiManager handles keypair and enrollment lifecycle. Nil before Start
	// loads the keys; the agent takes no job until then.
	pkiManager *pki.Manager

	// natsClient is used for publishing enrollment requests via NATS.
	// Uses the same NATSClient interface as the rest of the project.
	natsClient NATSClient

	// machineID is the permanent host identifier resolved at startup.
	machineID string

	// hostname cached from Start for drain/undrain resubscribe.
	hostname string

	// Lifecycle management.
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Consumer lifecycle for drain/undrain. Guarded by lifecycleMu.
	consumerCtx    context.Context
	consumerCancel context.CancelFunc
	consumerWg     sync.WaitGroup
	// consumerSetWg counts the consumers of the running set only, so a drain
	// waits for the set it cancelled and not for one started after it.
	// Guarded by lifecycleMu.
	consumerSetWg *sync.WaitGroup
	// drainStuck is set when a drain's consumers have not stopped within
	// drainStuckAfter; the heartbeat reports it as the DrainStuck condition.
	// Guarded by lifecycleMu.
	drainStuck bool
	// consumerErrs maps each job consumer of the running set that cannot
	// be created or bound to its last error; the heartbeat reports it as
	// the ConsumersUnavailable condition. Guarded by consumerErrMu.
	consumerErrs  map[string]string
	consumerErrMu sync.Mutex

	// missingPrograms are the behaviors' programs that this host lacks,
	// found at start and again on each heartbeat, which reports them. Only
	// the heartbeat goroutine writes it after the start.
	missingPrograms []string

	// subComponents reports the status of internal services in heartbeat.
	subComponents map[string]job.SubComponentInfo

	// Subsystem-specific loggers for background processes.
	heartbeatLogger *slog.Logger
	factsLogger     *slog.Logger
	pkiLogger       *slog.Logger

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
}
