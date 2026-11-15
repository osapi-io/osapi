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
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// natsServerCmd represents the natsServer command.
var natsServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the embedded NATS server",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		logger.Debug(
			"nats server configuration",
			slog.String("config_file", viper.ConfigFileUsed()),
			slog.Bool("debug", appConfig.Debug),
			slog.String("nats.server.host", appConfig.NATS.Server.Host),
			slog.Int("nats.server.port", appConfig.NATS.Server.Port),
			slog.String("nats.server.store_dir", appConfig.NATS.Server.StoreDir),
			slog.String("nats.server.namespace", appConfig.NATS.Server.Namespace),
			slog.String("nats.server.auth.type", appConfig.NATS.Server.Auth.Type),
			slog.String("nats.stream.name", appConfig.NATS.Stream.Name),
			slog.String("nats.stream.subjects", appConfig.NATS.Stream.Subjects),
			slog.String("nats.kv.bucket", appConfig.NATS.KV.Bucket),
			slog.String("nats.kv.response_bucket", appConfig.NATS.KV.ResponseBucket),
		)
	},
}

func init() {
	natsCmd.AddCommand(natsServerCmd)
}
