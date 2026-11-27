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

package agentcmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/metrics"
)

// Start checks the host, starts the tracer, sets the namespace of the job
// subjects, then makes the components. The tracer stops after them.
func Start(
	ctx context.Context,
	d Deps,
) (command.Running, error) {
	if err := d.CheckHost(); err != nil {
		return command.Running{}, fmt.Errorf("unsupported host: %w", err)
	}
	a := d.Config.Agent
	d.Logger.Debug(
		"agent configuration",
		slog.String("config_file", d.ConfigFile),
		slog.Bool("debug", d.Config.Debug),
		slog.String("agent.nats.host", a.NATS.Host),
		slog.Int("agent.nats.port", a.NATS.Port),
		slog.String("agent.nats.client_name", a.NATS.ClientName),
		slog.String("agent.nats.namespace", a.NATS.Namespace),
		slog.String("agent.nats.auth.type", a.NATS.Auth.Type),
		slog.String("agent.hostname", a.Hostname),
		slog.Int("agent.max_jobs", a.MaxJobs),
		slog.String("agent.consumer.name", a.Consumer.Name),
	)

	stop, err := d.Tracer(ctx, "osapi-agent")
	if err != nil {
		return command.Running{}, fmt.Errorf("start the tracer: %w", err)
	}
	job.Init(a.NATS.Namespace)

	r, err := Components(ctx, d)
	if err != nil {
		return command.Running{}, err
	}

	return command.Then(r, func(ctx context.Context) { _ = stop(ctx) }), nil
}

// Components makes the agent, and its metrics server when the config turns
// it on: the parts that osapi start runs too. The cleanup closes the
// agent's NATS connection.
func Components(
	ctx context.Context,
	d Deps,
) (command.Running, error) {
	log := d.Logger.With("component", "agent")
	// The metrics server is made before the NATS connection, so that
	// nats-client records the calls of the agent on it.
	ms := metrics.FromConfig(
		d.Config.Agent.Metrics,
		d.Config.Agent.PKI.KeyDir,
		"agent-metrics",
		log,
	)
	a, closeConn, err := d.NewAgent(ctx, log, d.FS, d.Config, ms.MeterProvider())
	if err != nil {
		return command.Running{}, fmt.Errorf("start the agent: %w", err)
	}
	a.SetVersion(d.Version)
	if d.LogLevel != nil {
		a.SetLogLevel(d.LogLevel)
	}

	components := []command.Component{a}
	if ms != nil {
		ms.SetReadinessFunc(a.IsReady)
		a.SetMeterProvider(ms.MeterProvider())
		ms.RegisterSubsystems(metrics.Subsystems(a.SubComponents()))
		ms.RegisterHeartbeatAge(a.LastHeartbeatTime)
		components = append(components, ms)
	}

	return command.Running{
		Components: components,
		Cleanup:    func(context.Context) { closeConn() },
	}, nil
}
