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

	"github.com/retr0h/osapi/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// nodeAgentCmd represents the nodeAgent command.
var nodeAgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "The agent subcommand",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		cli.ValidateDistribution(logger)

		logger.Debug(
			"node agent configuration",
			slog.String("config_file", viper.ConfigFileUsed()),
			slog.Bool("debug", appConfig.Debug),
			slog.String("job.worker.nats.host", appConfig.Job.Worker.NATS.Host),
			slog.Int("job.worker.nats.port", appConfig.Job.Worker.NATS.Port),
			slog.String("job.worker.nats.client_name", appConfig.Job.Worker.NATS.ClientName),
			slog.String("job.worker.nats.namespace", appConfig.Job.Worker.NATS.Namespace),
			slog.String("job.worker.nats.auth.type", appConfig.Job.Worker.NATS.Auth.Type),
			slog.String("job.worker.queue_group", appConfig.Job.Worker.QueueGroup),
			slog.String("job.worker.hostname", appConfig.Job.Worker.Hostname),
			slog.Int("job.worker.max_jobs", appConfig.Job.Worker.MaxJobs),
			slog.String("job.worker.consumer.name", appConfig.Job.Worker.Consumer.Name),
		)
	},
}

func init() {
	nodeCmd.AddCommand(nodeAgentCmd)

	// Worker configuration flags
	nodeAgentCmd.PersistentFlags().
		StringP("worker-host", "", "localhost", "NATS server hostname for agent")
	nodeAgentCmd.PersistentFlags().
		IntP("worker-port", "", 4222, "NATS server port for agent")
	nodeAgentCmd.PersistentFlags().
		StringP("worker-client-name", "", "osapi-job-worker", "NATS client name for agent")
	nodeAgentCmd.PersistentFlags().
		StringP("worker-queue-group", "", "job-workers", "NATS queue group for load balancing")
	nodeAgentCmd.PersistentFlags().
		StringP("worker-hostname", "", "", "Agent hostname (defaults to system hostname)")
	nodeAgentCmd.PersistentFlags().
		IntP("worker-max-jobs", "", 10, "Maximum concurrent jobs per agent")

	// Consumer configuration flags
	nodeAgentCmd.PersistentFlags().
		IntP("consumer-max-deliver", "", 5, "Maximum delivery attempts before DLQ")
	nodeAgentCmd.PersistentFlags().
		StringP("consumer-ack-wait", "", "2m", "Time to wait for acknowledgment before retry")
	nodeAgentCmd.PersistentFlags().
		IntP("consumer-max-ack-pending", "", 1000, "Maximum unacknowledged messages")
	nodeAgentCmd.PersistentFlags().
		StringP("consumer-replay-policy", "", "instant", "Replay policy: instant or original")
	nodeAgentCmd.PersistentFlags().
		StringSliceP("consumer-back-off", "", []string{"30s", "2m", "5m", "15m", "30m"}, "Retry backoff intervals")

	// Bind flags to viper config
	_ = viper.BindPFlag(
		"job.worker.nats.host",
		nodeAgentCmd.PersistentFlags().Lookup("worker-host"),
	)
	_ = viper.BindPFlag(
		"job.worker.nats.port",
		nodeAgentCmd.PersistentFlags().Lookup("worker-port"),
	)
	_ = viper.BindPFlag(
		"job.worker.nats.client_name",
		nodeAgentCmd.PersistentFlags().Lookup("worker-client-name"),
	)
	_ = viper.BindPFlag(
		"job.worker.queue_group",
		nodeAgentCmd.PersistentFlags().Lookup("worker-queue-group"),
	)
	_ = viper.BindPFlag(
		"job.worker.hostname",
		nodeAgentCmd.PersistentFlags().Lookup("worker-hostname"),
	)
	_ = viper.BindPFlag(
		"job.worker.max_jobs",
		nodeAgentCmd.PersistentFlags().Lookup("worker-max-jobs"),
	)

	// Bind consumer configuration flags
	_ = viper.BindPFlag(
		"job.worker.consumer.max_deliver",
		nodeAgentCmd.PersistentFlags().Lookup("consumer-max-deliver"),
	)
	_ = viper.BindPFlag(
		"job.worker.consumer.ack_wait",
		nodeAgentCmd.PersistentFlags().Lookup("consumer-ack-wait"),
	)
	_ = viper.BindPFlag(
		"job.worker.consumer.max_ack_pending",
		nodeAgentCmd.PersistentFlags().Lookup("consumer-max-ack-pending"),
	)
	_ = viper.BindPFlag(
		"job.worker.consumer.replay_policy",
		nodeAgentCmd.PersistentFlags().Lookup("consumer-replay-policy"),
	)
	_ = viper.BindPFlag(
		"job.worker.consumer.back_off",
		nodeAgentCmd.PersistentFlags().Lookup("consumer-back-off"),
	)
}
