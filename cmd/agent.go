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

// agentCmd represents the agent command.
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage the agent process",
	Long: `Manage the node agent process. The agent runs on each managed host,
processes jobs, and reports status back to the control plane.`,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		cli.ValidateDistribution(logger)

		logger.Debug(
			"agent configuration",
			slog.String("config_file", viper.ConfigFileUsed()),
			slog.Bool("debug", appConfig.Debug),
			slog.String("agent.nats.host", appConfig.Agent.NATS.Host),
			slog.Int("agent.nats.port", appConfig.Agent.NATS.Port),
			slog.String("agent.nats.client_name", appConfig.Agent.NATS.ClientName),
			slog.String("agent.nats.namespace", appConfig.Agent.NATS.Namespace),
			slog.String("agent.nats.auth.type", appConfig.Agent.NATS.Auth.Type),
			slog.String("agent.queue_group", appConfig.Agent.QueueGroup),
			slog.String("agent.hostname", appConfig.Agent.Hostname),
			slog.Int("agent.max_jobs", appConfig.Agent.MaxJobs),
			slog.String("agent.consumer.name", appConfig.Agent.Consumer.Name),
		)
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	// Agent configuration flags
	agentCmd.PersistentFlags().
		StringP("agent-host", "", "localhost", "NATS server hostname for agent")
	agentCmd.PersistentFlags().
		IntP("agent-port", "", 4222, "NATS server port for agent")
	agentCmd.PersistentFlags().
		StringP("agent-client-name", "", "osapi-agent", "NATS client name for agent")
	agentCmd.PersistentFlags().
		StringP("agent-queue-group", "", "job-agents", "NATS queue group for load balancing")
	agentCmd.PersistentFlags().
		StringP("agent-hostname", "", "", "Agent hostname (defaults to system hostname)")
	agentCmd.PersistentFlags().
		IntP("agent-max-jobs", "", 10, "Maximum concurrent jobs per agent")

	// Consumer configuration flags
	agentCmd.PersistentFlags().
		IntP("consumer-max-deliver", "", 5, "Maximum delivery attempts before DLQ")
	agentCmd.PersistentFlags().
		StringP("consumer-ack-wait", "", "2m", "Time to wait for acknowledgment before retry")
	agentCmd.PersistentFlags().
		IntP("consumer-max-ack-pending", "", 1000, "Maximum unacknowledged messages")
	agentCmd.PersistentFlags().
		StringP("consumer-replay-policy", "", "instant", "Replay policy: instant or original")
	agentCmd.PersistentFlags().
		StringSliceP("consumer-back-off", "", []string{"30s", "2m", "5m", "15m", "30m"}, "Retry backoff intervals")

	// Bind flags to viper config
	_ = viper.BindPFlag(
		"agent.nats.host",
		agentCmd.PersistentFlags().Lookup("agent-host"),
	)
	_ = viper.BindPFlag(
		"agent.nats.port",
		agentCmd.PersistentFlags().Lookup("agent-port"),
	)
	_ = viper.BindPFlag(
		"agent.nats.client_name",
		agentCmd.PersistentFlags().Lookup("agent-client-name"),
	)
	_ = viper.BindPFlag(
		"agent.queue_group",
		agentCmd.PersistentFlags().Lookup("agent-queue-group"),
	)
	_ = viper.BindPFlag(
		"agent.hostname",
		agentCmd.PersistentFlags().Lookup("agent-hostname"),
	)
	_ = viper.BindPFlag(
		"agent.max_jobs",
		agentCmd.PersistentFlags().Lookup("agent-max-jobs"),
	)

	// Bind consumer configuration flags
	_ = viper.BindPFlag(
		"agent.consumer.max_deliver",
		agentCmd.PersistentFlags().Lookup("consumer-max-deliver"),
	)
	_ = viper.BindPFlag(
		"agent.consumer.ack_wait",
		agentCmd.PersistentFlags().Lookup("consumer-ack-wait"),
	)
	_ = viper.BindPFlag(
		"agent.consumer.max_ack_pending",
		agentCmd.PersistentFlags().Lookup("consumer-max-ack-pending"),
	)
	_ = viper.BindPFlag(
		"agent.consumer.replay_policy",
		agentCmd.PersistentFlags().Lookup("consumer-replay-policy"),
	)
	_ = viper.BindPFlag(
		"agent.consumer.back_off",
		agentCmd.PersistentFlags().Lookup("consumer-back-off"),
	)
}
