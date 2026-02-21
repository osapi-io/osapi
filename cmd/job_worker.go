// Copyright (c) 2025 John Dewey

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

// jobWorkerCmd represents the jobWorker command.
var jobWorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "The worker subcommand",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		validateDistribution()

		logger.Debug(
			"job worker configuration",
			slog.String("config_file", viper.ConfigFileUsed()),
			slog.Bool("debug", appConfig.Debug),
			slog.String("worker.nats.host", appConfig.Job.Worker.NATS.Host),
			slog.Int("worker.nats.port", appConfig.Job.Worker.NATS.Port),
			slog.String("worker.nats.client_name", appConfig.Job.Worker.NATS.ClientName),
			slog.String("worker.queue_group", appConfig.Job.Worker.QueueGroup),
			slog.String("worker.hostname", appConfig.Job.Worker.Hostname),
			slog.Int("worker.max_jobs", appConfig.Job.Worker.MaxJobs),
		)
	},
}

func init() {
	jobCmd.AddCommand(jobWorkerCmd)

	// Worker configuration flags
	jobWorkerCmd.PersistentFlags().
		StringP("worker-host", "", "localhost", "NATS server hostname for worker")
	jobWorkerCmd.PersistentFlags().
		IntP("worker-port", "", 4222, "NATS server port for worker")
	jobWorkerCmd.PersistentFlags().
		StringP("worker-client-name", "", "osapi-job-worker", "NATS client name for worker")
	jobWorkerCmd.PersistentFlags().
		StringP("worker-queue-group", "", "job-workers", "NATS queue group for load balancing")
	jobWorkerCmd.PersistentFlags().
		StringP("worker-hostname", "", "", "Worker hostname (defaults to system hostname)")
	jobWorkerCmd.PersistentFlags().
		IntP("worker-max-jobs", "", 10, "Maximum concurrent jobs per worker")

	// Consumer configuration flags
	jobWorkerCmd.PersistentFlags().
		IntP("consumer-max-deliver", "", 5, "Maximum delivery attempts before DLQ")
	jobWorkerCmd.PersistentFlags().
		StringP("consumer-ack-wait", "", "2m", "Time to wait for acknowledgment before retry")
	jobWorkerCmd.PersistentFlags().
		IntP("consumer-max-ack-pending", "", 1000, "Maximum unacknowledged messages")
	jobWorkerCmd.PersistentFlags().
		StringP("consumer-replay-policy", "", "instant", "Replay policy: instant or original")
	jobWorkerCmd.PersistentFlags().
		StringSliceP("consumer-back-off", "", []string{"30s", "2m", "5m", "15m", "30m"}, "Retry backoff intervals")

	// Bind flags to viper config
	_ = viper.BindPFlag(
		"job.worker.nats.host",
		jobWorkerCmd.PersistentFlags().Lookup("worker-host"),
	)
	_ = viper.BindPFlag(
		"job.worker.nats.port",
		jobWorkerCmd.PersistentFlags().Lookup("worker-port"),
	)
	_ = viper.BindPFlag(
		"job.worker.nats.client_name",
		jobWorkerCmd.PersistentFlags().Lookup("worker-client-name"),
	)
	_ = viper.BindPFlag(
		"job.worker.queue_group",
		jobWorkerCmd.PersistentFlags().Lookup("worker-queue-group"),
	)
	_ = viper.BindPFlag(
		"job.worker.hostname",
		jobWorkerCmd.PersistentFlags().Lookup("worker-hostname"),
	)
	_ = viper.BindPFlag(
		"job.worker.max_jobs",
		jobWorkerCmd.PersistentFlags().Lookup("worker-max-jobs"),
	)

	// Bind consumer configuration flags
	_ = viper.BindPFlag(
		"job.consumer.max_deliver",
		jobWorkerCmd.PersistentFlags().Lookup("consumer-max-deliver"),
	)
	_ = viper.BindPFlag(
		"job.consumer.ack_wait",
		jobWorkerCmd.PersistentFlags().Lookup("consumer-ack-wait"),
	)
	_ = viper.BindPFlag(
		"job.consumer.max_ack_pending",
		jobWorkerCmd.PersistentFlags().Lookup("consumer-max-ack-pending"),
	)
	_ = viper.BindPFlag(
		"job.consumer.replay_policy",
		jobWorkerCmd.PersistentFlags().Lookup("consumer-replay-policy"),
	)
	_ = viper.BindPFlag(
		"job.consumer.back_off",
		jobWorkerCmd.PersistentFlags().Lookup("consumer-back-off"),
	)
}
