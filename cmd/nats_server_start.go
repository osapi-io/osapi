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

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/bus/server"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/metrics"
)

// natsServerStartCmd represents the natsServerStart command.
var natsServerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the embedded NATS server",
	Long: "The server makes the streams, consumers and KV buckets that the job system\n" +
		"needs. `osapi start` runs the same server, together with the controller and\n" +
		"an agent, in one process. Use this command to run the bus as its own\n" +
		"process. The `nats.server` section of the config sets the host, the port and\n" +
		"the authentication. Refer to [Configuration](../../../configuration.md).",
	Example: `  osapi nats server start`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		job.Init(appConfig.NATS.Server.Namespace)

		log := logger.With("component", "nats")
		// The metrics server is made before the NATS connection, so that
		// nats-client records the calls of the bus on it.
		metricsServer := newMetricsServer(appConfig.NATS.Server.Metrics, log)
		s, err := server.New(ctx, log, appConfig.NATS, meterProviderOf(metricsServer))
		if err != nil {
			cli.LogFatal(log, "failed to start NATS server", err)
		}

		s.StartHeartbeat(ctx, appConfig.Agent.ProcessConditions)

		if metricsServer != nil {
			metricsServer.SetReadinessFunc(func() error {
				return nil
			})
			metricsServer.RegisterSubsystems(
				metrics.Subsystems(server.SubComponents(appConfig.NATS.Server)),
			)
			metricsServer.Start()
		}

		serve(ctx, []component{s}, func(shutdownCtx context.Context) {
			if metricsServer != nil {
				metricsServer.Stop(shutdownCtx)
			}
		})
	},
}

func init() {
	natsServerCmd.AddCommand(natsServerStartCmd)
}
