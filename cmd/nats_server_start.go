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

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/telemetry/metrics"
)

// natsServerStartCmd represents the natsServerStart command.
var natsServerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the embedded NATS server",
	Long: `Start the embedded NATS server with JetStream enabled.
Configures streams, consumers, and KV buckets needed by the job system.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		job.Init(appConfig.NATS.Server.Namespace)

		log := logger.With("component", "nats")
		s := setupNATSServer(ctx, log)

		startNATSHeartbeat(
			ctx,
			log,
			appConfig.NATS.Server.Host,
			appConfig.NATS.Server.Port,
			appConfig.NATS.Server.Namespace,
			appConfig.NATS.Server.Auth,
		)

		var metricsServer *metrics.Server
		if appConfig.NATS.Server.Metrics.Enabled {
			metricsServer = metrics.New(
				appConfig.NATS.Server.Metrics.Host,
				appConfig.NATS.Server.Metrics.Port,
				log,
			)

			if metricsServer != nil {
				metricsServer.SetReadinessFunc(func() error {
					return nil
				})
				metricsServer.RegisterSubsystems(
					subsystemStatuses(buildNATSSubComponents()),
				)
			}

			metricsServer.Start()
		}

		var ns cli.Lifecycle = &natsLifecycle{server: s}
		cli.RunServer(ctx, ns, func() {
			if metricsServer != nil {
				metricsServer.Stop(context.Background())
			}
		})
	},
}

func init() {
	natsServerCmd.AddCommand(natsServerStartCmd)
}
