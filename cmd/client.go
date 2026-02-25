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
	"log/slog"

	"github.com/osapi-io/osapi-sdk/pkg/osapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/telemetry"
)

var (
	sdkClient      *osapi.Client
	tracerShutdown func(context.Context) error
)

// clientCmd represents the client command.
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "The client subcommand",
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		cli.ValidateDistribution(logger)

		var err error
		tracerShutdown, err = telemetry.InitTracer(
			cmd.Context(),
			"osapi-cli",
			appConfig.Telemetry.Tracing,
		)
		if err != nil {
			cli.LogFatal(logger, "failed to initialize tracer", err)
		}

		logger.Debug(
			"client configuration",
			slog.String("config_file", viper.ConfigFileUsed()),
			slog.Bool("debug", appConfig.Debug),
			slog.String("api.client.url", appConfig.API.URL),
		)

		sdkClient, err = osapi.New(
			appConfig.API.URL,
			appConfig.API.Client.Security.BearerToken,
			osapi.WithLogger(logger),
		)
		if err != nil {
			cli.LogFatal(logger, "failed to create sdk client", err)
		}
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		if tracerShutdown != nil {
			_ = tracerShutdown(context.Background())
		}
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)

	clientCmd.PersistentFlags().
		StringP("url", "", "http://0.0.0.0:8080", "URL the client will connect to")
	clientCmd.PersistentFlags().
		StringP("target", "T", "_any", "Target: _any, _all, hostname, or label (group:web.dev)")

	_ = viper.BindPFlag("api.client.url", clientCmd.PersistentFlags().Lookup("url"))
}
