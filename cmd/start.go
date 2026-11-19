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

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/bus/server"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/reference"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/metrics"
	"github.com/osapi-io/osapi/internal/telemetry/tracing"
)

// startCmd represents the top-level start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start all components (NATS, controller, agent)",
	Long: "This runs the embedded NATS server, the API server and the agent together.\n" +
		"Use it for a deployment on one host. For a deployment on more than one\n" +
		"host, start the NATS server, the controller and each agent alone.",
	Annotations: map[string]string{
		reference.Notes: "The NATS server starts first, and waits until it is ready. Then osapi\n" +
			"makes the streams and buckets, starts the API server and starts the agent.\n" +
			"On `SIGINT` or `SIGTERM`, the three components stop at the same time, in\n" +
			"10 seconds or less. Then osapi closes the NATS connections and the\n" +
			"telemetry exporters.",
	},
	Example: `  osapi start`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		if err := cli.ValidateDistribution(logger); err != nil {
			cli.LogFatal(logger, "unsupported host", err)
		}

		shutdownTracer, err := tracing.InitTracer(
			ctx,
			"osapi",
			appConfig.Telemetry.Tracing,
			versionAttr(),
		)
		if err != nil {
			cli.LogFatal(logger, "failed to initialize tracer", err)
		}

		job.Init(appConfig.Controller.NATS.Namespace)

		// Each metrics server is made before the NATS connection of its
		// component, so that nats-client records the calls of that
		// connection on it. Each server starts when its component is ready.
		natsLog := logger.With("component", "nats")
		controllerLog := logger.With("component", "controller")
		agentLog := logger.With("component", "agent")
		natsMetrics := newMetricsServer(appConfig.NATS.Server.Metrics, natsLog)
		controllerMetrics := newMetricsServer(appConfig.Controller.Metrics, controllerLog)
		agentMetrics := newMetricsServer(appConfig.Agent.Metrics, agentLog)
		var metricsServers []*metrics.Server

		natsServer, err := server.New(ctx, natsLog, appConfig.NATS, meterProviderOf(natsMetrics))
		if err != nil {
			cli.LogFatal(natsLog, "failed to start NATS server", err)
		}

		// The controller metrics server also serves the HTTP metrics of the
		// otelecho middleware.
		controllerOpts := []controller.Option{controller.WithVersion(appVersion())}
		if controllerMetrics != nil {
			controllerOpts = append(
				controllerOpts,
				controller.WithMeterProvider(controllerMetrics.MeterProvider()),
			)
		}

		c, err := controller.New(ctx, controllerLog, appConfig, controllerOpts...)
		if err != nil {
			cli.LogFatal(logger, "failed to start the controller", err)
		}

		if controllerMetrics != nil {
			controllerMetrics.SetReadinessFunc(func() error {
				return c.CheckHealth(context.Background())
			})
			controllerMetrics.RegisterSubsystems(metrics.Subsystems(c.SubComponents()))
			controllerMetrics.Start()
			metricsServers = append(metricsServers, controllerMetrics)
		}

		natsServer.StartHeartbeat(ctx, appVersion(), appConfig.Agent.ProcessConditions)

		agentServer, agentBundle, err := agent.Setup(
			ctx,
			agentLog,
			appFs,
			appConfig,
			meterProviderOf(agentMetrics),
		)
		if err != nil {
			cli.LogFatal(logger, "failed to start the agent", err)
		}

		if agentMetrics != nil {
			agentMetrics.SetReadinessFunc(func() error {
				return agentServer.IsReady()
			})
			agentServer.SetMeterProvider(agentMetrics.MeterProvider())
			agentMetrics.RegisterSubsystems(metrics.Subsystems(agentServer.SubComponents()))
			agentMetrics.RegisterHeartbeatAge(agentServer.LastHeartbeatTime)
			agentMetrics.Start()
			metricsServers = append(metricsServers, agentMetrics)
		}

		if natsMetrics != nil {
			natsMetrics.SetReadinessFunc(func() error {
				return nil
			})
			natsMetrics.RegisterSubsystems(
				metrics.Subsystems(server.SubComponents(appConfig.NATS.Server)),
			)
			natsMetrics.Start()
			metricsServers = append(metricsServers, natsMetrics)
		}

		components := []component{c, agentServer, natsServer}
		serve(ctx, components, func(shutdownCtx context.Context) {
			for _, s := range metricsServers {
				s.Stop(shutdownCtx)
			}

			_ = shutdownTracer(shutdownCtx)
			bus.Close(agentBundle.NATS)
		})
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
