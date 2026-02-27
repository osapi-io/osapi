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
	Short: "Start all components (NATS, API server, agent)",
	Long: `Start the embedded NATS server, API server, and agent in a single process.

This is the recommended way to run OSAPI on a single host. All three components
start in order (NATS → API → agent) and shut down gracefully on SIGINT/SIGTERM.
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

		metricsHandler, metricsPath, shutdownMeter, err := telemetry.InitMeter(
			appConfig.Telemetry.Metrics,
		)
		if err != nil {
			cli.LogFatal(logger, "failed to initialize meter", err)
		}

		job.Init(appConfig.NATS.Server.Namespace)

		natsServer := setupNATSServer(ctx, logger.With("component", "nats"))
		sm, apiBundle := setupAPIServer(
			ctx, logger.With("component", "api"),
			appConfig.API.NATS, metricsHandler, metricsPath,
		)
		agentServer, agentBundle := setupAgent(
			ctx, logger.With("component", "agent"), appConfig.Agent.NATS,
		)

		composite := &compositeLifecycle{
			components: []cli.Lifecycle{
				sm,
				agentServer,
				&natsLifecycle{server: natsServer},
			},
		}

		composite.Start()
		cli.RunServer(ctx, composite, func() {
			_ = shutdownMeter(context.Background())
			_ = shutdownTracer(context.Background())
			cli.CloseNATSClient(agentBundle.nc)
			cli.CloseNATSClient(apiBundle.nc)
		})
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
