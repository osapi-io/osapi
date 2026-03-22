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

package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/controller/api"
	"github.com/retr0h/osapi/internal/job"
	jobclient "github.com/retr0h/osapi/internal/job/client"
	"github.com/retr0h/osapi/internal/telemetry/metrics"
	"github.com/retr0h/osapi/internal/telemetry/tracing"
)

// controllerStartCmd represents the controller start command.
var controllerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the controller",
	Long:  `Start the control plane process (API server, heartbeat, notifications).`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		shutdownTracer, err := tracing.InitTracer(
			ctx,
			"osapi-controller",
			appConfig.Telemetry.Tracing,
		)
		if err != nil {
			cli.LogFatal(logger, "failed to initialize tracer", err)
		}

		job.Init(appConfig.Controller.NATS.Namespace)

		log := logger.With("component", "controller")

		// Create the metrics server before the API server so the
		// MeterProvider can be injected into otelecho middleware.
		var metricsServer *metrics.Server
		var controllerOpts []api.Option

		if appConfig.Controller.Metrics.Enabled {
			metricsServer = metrics.New(
				appConfig.Controller.Metrics.Host,
				appConfig.Controller.Metrics.Port,
				log.With("subsystem", "metrics"),
			)

			if metricsServer != nil {
				controllerOpts = append(
					controllerOpts,
					api.WithMeterProvider(metricsServer.MeterProvider()),
				)
			}
		}

		sm, b := setupController(ctx, log, appConfig.Controller.NATS, controllerOpts...)

		if metricsServer != nil {
			if jc, ok := b.jobClient.(*jobclient.Client); ok {
				jc.SetMeterProvider(metricsServer.MeterProvider())
			}
		}

		if metricsServer != nil {
			metricsServer.SetReadinessFunc(func() error {
				return b.checker.CheckHealth(context.Background())
			})
			metricsServer.RegisterSubsystems(subsystemStatuses(b.subComponents))
			metricsServer.Start()
		}

		sm.Start()
		cli.RunServer(ctx, sm, func() {
			if metricsServer != nil {
				metricsServer.Stop(context.Background())
			}

			_ = shutdownTracer(context.Background())
			cli.CloseNATSClient(b.nc)
		})
	},
}

func init() {
	controllerCmd.AddCommand(controllerStartCmd)
}
