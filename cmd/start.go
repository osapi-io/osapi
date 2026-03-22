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

package cmd

import (
	"context"
	"sync"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/ops"
	"github.com/retr0h/osapi/internal/telemetry"
)

// compositeLifecycle manages multiple Lifecycle components, starting them
// sequentially and stopping them concurrently.
type compositeLifecycle struct {
	components []cli.Lifecycle
}

func (c *compositeLifecycle) Start() {
	for _, comp := range c.components {
		comp.Start()
	}
}

func (c *compositeLifecycle) Stop(ctx context.Context) {
	var wg sync.WaitGroup
	for _, comp := range c.components {
		wg.Add(1)
		go func(lc cli.Lifecycle) {
			defer wg.Done()
			lc.Stop(ctx)
		}(comp)
	}
	wg.Wait()
}

// startCmd represents the top-level start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start all components (NATS, controller, agent)",
	Long: `Start the embedded NATS server, controller, and agent in a single process.

This is the recommended way to run OSAPI on a single host. All three components
start in order (NATS → controller → agent) and shut down gracefully on SIGINT/SIGTERM.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		cli.ValidateDistribution(logger)

		shutdownTracer, err := telemetry.InitTracer(
			ctx,
			"osapi",
			appConfig.Telemetry.Tracing,
		)
		if err != nil {
			cli.LogFatal(logger, "failed to initialize tracer", err)
		}

		job.Init(appConfig.NATS.Server.Namespace)

		natsLog := logger.With("component", "nats")
		natsServer := setupNATSServer(ctx, natsLog)
		sm, controllerBundle := setupController(
			ctx, logger.With("component", "controller"),
			appConfig.Controller.NATS,
		)

		startNATSHeartbeatFromKV(ctx, natsLog, controllerBundle.registryKV)

		agentServer, agentBundle := setupAgent(
			ctx, logger.With("component", "agent"), appConfig.Agent.NATS,
		)

		// Start per-component ops servers for /metrics.
		var opsServers []*ops.Server

		if appConfig.Controller.Metrics.IsEnabled() {
			s := ops.New(
				appConfig.Controller.Metrics.Port,
				logger.With("component", "controller-ops"),
			)
			s.Start()

			opsServers = append(opsServers, s)
		}

		if appConfig.Agent.Metrics.IsEnabled() {
			s := ops.New(
				appConfig.Agent.Metrics.Port,
				logger.With("component", "agent-ops"),
			)
			s.Start()

			opsServers = append(opsServers, s)
		}

		if appConfig.NATS.Server.Metrics.IsEnabled() {
			s := ops.New(
				appConfig.NATS.Server.Metrics.Port,
				logger.With("component", "nats-ops"),
			)
			s.Start()

			opsServers = append(opsServers, s)
		}

		composite := &compositeLifecycle{
			components: []cli.Lifecycle{
				sm,
				agentServer,
				&natsLifecycle{server: natsServer},
			},
		}

		composite.Start()
		cli.RunServer(ctx, composite, func() {
			for _, o := range opsServers {
				o.Stop(context.Background())
			}

			_ = shutdownTracer(context.Background())
			cli.CloseNATSClient(agentBundle.nc)
			cli.CloseNATSClient(controllerBundle.nc)
		})
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
