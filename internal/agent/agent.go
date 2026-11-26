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
	"fmt"
	"log/slog"
	"maps"
	"time"

	"github.com/avfs/avfs"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/exec"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/provider/node/disk"
	nodeHost "github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
	"github.com/osapi-io/osapi/internal/telemetry/process"

	"github.com/osapi-io/osapi/internal/telemetry/instruments"
)

// New creates a new agent instance.
func New(
	appFs avfs.VFS,
	appConfig config.Config,
	logger *slog.Logger,
	jobClient client.JobClient,
	streamName string,
	hostProvider nodeHost.Provider,
	diskProvider disk.Provider,
	memoryProvider memory.Provider,
	loadProvider load.Provider,
	processProvider process.Provider,
	registry *ProviderRegistry,
	registryBucket string,
	factsBucket string,
	execManager exec.Manager,
	natsClient NATSClient,
) *Agent {
	a := &Agent{
		renamed:         make(chan chan struct{}, 1),
		logger:          logger.With(slog.String("subsystem", "agent")),
		appConfig:       appConfig,
		fileLabels:      maps.Clone(appConfig.Agent.Labels),
		fileMaxJobs:     appConfig.Agent.MaxJobs,
		appFs:           appFs,
		jobClient:       jobClient,
		streamName:      streamName,
		hostProvider:    hostProvider,
		diskProvider:    diskProvider,
		memoryProvider:  memoryProvider,
		loadProvider:    loadProvider,
		processProvider: processProvider,
		registry:        registry,
		registryBucket:  registryBucket,
		factsBucket:     factsBucket,
		execManager:     execManager,
		natsClient:      natsClient,
		heartbeatLogger: logger.With(slog.String("subsystem", "agent.heartbeat")),
		factsLogger:     logger.With(slog.String("subsystem", "agent.facts")),
		pkiLogger:       logger.With(slog.String("subsystem", "agent.pki")),
	}

	return a
}

// IsReady returns nil when the agent is ready to process jobs,
// or an error describing why it is not.
func (a *Agent) IsReady() error {
	if a.ctx == nil || a.ctx.Err() != nil {
		return fmt.Errorf("agent not started")
	}

	return nil
}

// SetMeterProvider creates OTEL instruments for job metrics.
func (a *Agent) SetMeterProvider(
	mp *sdkmetric.MeterProvider,
) {
	meter := mp.Meter("osapi-agent")
	a.jobsProcessed, _ = meter.Int64Counter(
		instruments.JobsProcessed.Name,
		metric.WithDescription(instruments.JobsProcessed.Doc),
	)
	a.jobsActive, _ = meter.Int64UpDownCounter(
		instruments.JobsActive.Name,
		metric.WithDescription(instruments.JobsActive.Doc),
	)
	a.jobDuration, _ = meter.Float64Histogram(
		instruments.JobDuration.Name,
		metric.WithDescription(instruments.JobDuration.Doc),
		metric.WithUnit(instruments.JobDuration.Unit),
	)

	// Behaviors record their calls through the same provider.
	behavior.SetMeterProvider(mp)
}

// LastHeartbeatTime returns the timestamp of the last successful
// heartbeat write. Returns zero time if no heartbeat has been written.
func (a *Agent) LastHeartbeatTime() time.Time {
	if t, ok := a.lastHeartbeatTime.Load().(time.Time); ok {
		return t
	}
	return time.Time{}
}

// SetSubComponents sets the sub-component info published in heartbeats.
func (a *Agent) SetSubComponents(
	scs map[string]job.SubComponentInfo,
) {
	a.subComponents = scs
}

// SetLogLevel gives the agent the level of the process's logger, so that
// the log level of the controller's config of the agent changes it. Its
// level now is the level of the agent's config file.
func (a *Agent) SetLogLevel(
	level *slog.LevelVar,
) {
	a.logLevel = level
	a.fileLogLevel = level.Level()
}

// SetVersion sets the build version of the agent binary, such as v1.4.0.
// osapi agent start calls it with the version the binary was built with;
// the agent writes it in the started event of each job and in its
// registration.
func (a *Agent) SetVersion(
	version string,
) {
	a.version = version
}
