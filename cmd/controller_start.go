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

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/metrics"
	"github.com/osapi-io/osapi/internal/telemetry/tracing"
)

// controllerStartCmd represents the controller start command.
var controllerStartCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start the controller",
	Example: `  osapi controller start`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		shutdownTracer, err := tracing.InitTracer(
			ctx,
			"osapi-controller",
			appConfig.Telemetry.Tracing,
			versionAttr(),
		)
		if err != nil {
			cli.LogFatal(logger, "failed to initialize tracer", err)
		}

		job.Init(appConfig.Controller.NATS.Namespace)

		log := logger.With("component", "controller")

		// Create the metrics server before the API server so the
		// MeterProvider can be injected into otelecho middleware.
		var metricsServer *metrics.Server
		controllerOpts := []controller.Option{controller.WithVersion(appVersion())}

		if appConfig.Controller.Metrics.Enabled {
			metricsServer = metrics.New(
				appConfig.Controller.Metrics.Host,
				appConfig.Controller.Metrics.Port,
				log,
			)

			if metricsServer != nil {
				controllerOpts = append(
					controllerOpts,
					controller.WithMeterProvider(metricsServer.MeterProvider()),
				)
			}
		}

		c, err := controller.New(ctx, log, appConfig, controllerOpts...)
		if err != nil {
			cli.LogFatal(log, "failed to start the controller", err)
		}

		if metricsServer != nil {
			metricsServer.SetReadinessFunc(func() error {
				return c.CheckHealth(context.Background())
			})
			metricsServer.RegisterSubsystems(metrics.Subsystems(c.SubComponents()))
			metricsServer.Start()
		}

		serve(ctx, []component{c}, func(shutdownCtx context.Context) {
			if metricsServer != nil {
				metricsServer.Stop(shutdownCtx)
			}

			_ = shutdownTracer(shutdownCtx)
		})
	},
}

func init() {
	controllerCmd.AddCommand(controllerStartCmd)
}
