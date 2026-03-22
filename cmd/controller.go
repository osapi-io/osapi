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
	"log/slog"
	"strings"

	"github.com/ggwhite/go-masker/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/config"
)

// controllerCmd represents the controller command.
var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Manage the controller process",
	Long: `Manage the control plane process. The controller runs the REST API,
component heartbeat, and condition notification watcher.`,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		masker := masker.NewMaskerMarshaler()
		maskedConfig, err := masker.Struct(&appConfig)
		if err != nil {
			cli.LogFatal(logger, "failed to mask config", err)
		}

		maskedAppConfig, ok := maskedConfig.(*config.Config)
		if !ok {
			cli.LogFatal(logger, "failed to type assert maskedConfig", nil)
		}

		logger.Debug(
			"controller configuration",
			slog.String("config_file", viper.ConfigFileUsed()),
			slog.Bool("debug", appConfig.Debug),
			slog.Int("controller.api.port", appConfig.Controller.API.Port),
			slog.String("controller.nats.host", appConfig.Controller.NATS.Host),
			slog.Int("controller.nats.port", appConfig.Controller.NATS.Port),
			slog.String("controller.nats.client_name", appConfig.Controller.NATS.ClientName),
			slog.String("controller.nats.namespace", appConfig.Controller.NATS.Namespace),
			slog.String("controller.nats.auth.type", appConfig.Controller.NATS.Auth.Type),
			slog.String(
				"controller.api.security.cors.allow_origins",
				strings.Join(appConfig.Controller.API.Security.CORS.AllowOrigins, ","),
			),
			slog.String(
				"controller.api.security.signing_key",
				maskedAppConfig.Controller.API.Security.SigningKey,
			),
		)
	},
}

func init() {
	rootCmd.AddCommand(controllerCmd)

	controllerCmd.PersistentFlags().
		IntP("port", "p", 8080, "Port the server will bind to")

	_ = viper.BindPFlag("controller.api.port", controllerCmd.PersistentFlags().Lookup("port"))
}
