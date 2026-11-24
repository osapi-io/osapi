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
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/osapi-io/osapi/pkg/sdk/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/cli"
	clicommand "github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/config"
	endpointcommand "github.com/osapi-io/osapi/internal/controller/endpoint/command"
	"github.com/osapi-io/osapi/internal/telemetry/tracing"

	// Every behavior with a CLI registers its generated commands.
	_ "github.com/osapi-io/osapi/internal/behaviors/all/gen"
)

var (
	sdkClient      *client.Client
	tracerShutdown func(context.Context) error
	// commandSpan is the root span of one client command. Each request of
	// the command sends its context, so the trace starts here and not at the
	// controller.
	commandSpan trace.Span
)

// clientCmd represents the client command.
var clientCmd = &cobra.Command{
	Use: "client",
	// Its help is its purpose in command.RootGroups, which DescribeRoot gives it.
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		var err error
		tracerShutdown, err = tracing.InitTracer(
			cmd.Context(),
			"osapi-cli",
			appConfig.Telemetry.Tracing,
			versionAttr(),
		)
		if err != nil {
			cli.LogFatal(logger, "failed to initialize tracer", err)
		}

		// The command opens its own span only when it sends spans somewhere. A
		// span that is not exported would still go to the controller as the
		// parent of the request, and the trace would show a parent that
		// never arrives. Without an exporter, the controller starts the
		// trace, and the command prints its ID from the response.
		if exports(appConfig.Telemetry.Tracing) {
			var ctx context.Context
			ctx, commandSpan = otel.Tracer("osapi-cli").Start(cmd.Context(), cmd.CommandPath(),
				trace.WithSpanKind(trace.SpanKindClient))
			cmd.SetContext(ctx)
		}

		logger.Debug(
			"client configuration",
			slog.String("config_file", viper.ConfigFileUsed()),
			slog.Bool("debug", appConfig.Debug),
			slog.String("controller.client.url", appConfig.Controller.Client.URL),
		)

		// The bearer token can be a reference, which the command resolves
		// from the vaults of its config.
		if err := resolveReferences(cmd.Context(), "controller.client"); err != nil {
			cli.LogFatal(logger, "failed to resolve the bearer token", err)
		}

		sdkClient = client.New(
			appConfig.Controller.Client.URL,
			appConfig.Controller.Client.Security.BearerToken,
			client.WithLogger(logger),
			client.WithHTTPTransport(cli.TraceTransport{Base: http.DefaultTransport}),
		)
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		// The trace of a result that had no summary line to carry it.
		cli.FlushTraceID()
		// A command that exits through LogFatal does not reach here, so its
		// span is not sent. The controller still has the request.
		if commandSpan != nil {
			commandSpan.End()
		}
		if tracerShutdown != nil {
			_ = tracerShutdown(context.Background())
		}
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)

	clientCmd.PersistentFlags().
		StringP("url", "", "http://0.0.0.0:8080", "URL the client will connect to")

	_ = viper.BindPFlag("controller.client.url", clientCmd.PersistentFlags().Lookup("url"))

	// node exists before the groups are attached, so a behavior's group
	// joins it: sysctl is under node.
	clientCmd.AddCommand(clientNodeCmd)

	// The generated commands of the endpoint groups go under client. A group
	// that also has hand-written commands, such as agent with status,
	// is named, so its generated commands join that command; so is
	// a command that a generated one goes under, such as file stale.
	endpointcommand.Use(func() endpointcommand.Caller { return sdkClient })
	if err := clicommand.Attach(clientCmd, clicommand.Deps{
		JSON:      func() bool { return jsonOutput },
		Error:     func(err error) { cli.HandleError(err, logger) },
		PrintJSON: cli.PrintJSON,
		Print:     cli.Printer{},
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// The help of node lists each behavior under the label of its area.
	clicommand.GroupByArea(clientNodeCmd)
}

// exports reports whether the tracing settings send spans to an exporter.
func exports(
	cfg config.TracingConfig,
) bool {
	return cfg.Enabled && cfg.Exporter != "" && cfg.Exporter != "none"
}
