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
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/afero"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/client"
	"github.com/retr0h/osapi/internal/provider"
	"github.com/retr0h/osapi/internal/provider/command"
	dockerProv "github.com/retr0h/osapi/internal/provider/docker"
	fileProv "github.com/retr0h/osapi/internal/provider/file"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/provider/network/netinfo"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/provider/node/disk"
	"github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
	"github.com/retr0h/osapi/internal/provider/process"
	cronProv "github.com/retr0h/osapi/internal/provider/scheduled/cron"
)

// New creates a new agent instance.
func New(
	appFs afero.Fs,
	appConfig config.Config,
	logger *slog.Logger,
	jobClient client.JobClient,
	streamName string,
	hostProvider host.Provider,
	diskProvider disk.Provider,
	memProvider mem.Provider,
	loadProvider load.Provider,
	dnsProvider dns.Provider,
	pingProvider ping.Provider,
	netinfoProvider netinfo.Provider,
	commandProvider command.Provider,
	fileProvider fileProv.Provider,
	dockerProvider dockerProv.Provider,
	cronProvider cronProv.Provider,
	processProvider process.Provider,
	registryKV jetstream.KeyValue,
	factsKV jetstream.KeyValue,
) *Agent {
	logger = logger.With(slog.String("subsystem", "agent"))

	a := &Agent{
		logger:          logger,
		appConfig:       appConfig,
		appFs:           appFs,
		jobClient:       jobClient,
		streamName:      streamName,
		hostProvider:    hostProvider,
		diskProvider:    diskProvider,
		memProvider:     memProvider,
		loadProvider:    loadProvider,
		dnsProvider:     dnsProvider,
		pingProvider:    pingProvider,
		netinfoProvider: netinfoProvider,
		commandProvider: commandProvider,
		fileProvider:    fileProvider,
		dockerProvider:  dockerProvider,
		cronProvider:    cronProvider,
		processProvider: processProvider,
		registryKV:      registryKV,
		factsKV:         factsKV,
		heartbeatLogger: logger.With(slog.String("subsystem", "heartbeat")),
		factsLogger:     logger.With(slog.String("subsystem", "facts")),
	}

	// Wire agent facts into all providers so they can access the latest
	// facts at execution time (e.g., for template rendering).
	provider.WireProviderFacts(
		a.GetFacts,
		hostProvider,
		diskProvider,
		memProvider,
		loadProvider,
		dnsProvider,
		pingProvider,
		netinfoProvider,
		commandProvider,
		fileProvider,
		dockerProvider,
	)

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
		"jobs_processed_total",
		metric.WithDescription("Total jobs processed"),
	)
	a.jobsActive, _ = meter.Int64UpDownCounter(
		"jobs_active",
		metric.WithDescription("Currently executing jobs"),
	)
	a.jobDuration, _ = meter.Float64Histogram(
		"job_duration_seconds",
		metric.WithDescription("Job execution duration in seconds"),
	)
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
