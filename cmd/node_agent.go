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
			slog.String("node.agent.nats.host", appConfig.Node.Agent.NATS.Host),
			slog.Int("node.agent.nats.port", appConfig.Node.Agent.NATS.Port),
			slog.String("node.agent.nats.client_name", appConfig.Node.Agent.NATS.ClientName),
			slog.String("node.agent.nats.namespace", appConfig.Node.Agent.NATS.Namespace),
			slog.String("node.agent.nats.auth.type", appConfig.Node.Agent.NATS.Auth.Type),
			slog.String("node.agent.queue_group", appConfig.Node.Agent.QueueGroup),
			slog.String("node.agent.hostname", appConfig.Node.Agent.Hostname),
			slog.Int("node.agent.max_jobs", appConfig.Node.Agent.MaxJobs),
			slog.String("node.agent.consumer.name", appConfig.Node.Agent.Consumer.Name),
		)
	},
}

func init() {
	nodeCmd.AddCommand(nodeAgentCmd)

	// Agent configuration flags
	nodeAgentCmd.PersistentFlags().
		StringP("agent-host", "", "localhost", "NATS server hostname for agent")
	nodeAgentCmd.PersistentFlags().
		IntP("agent-port", "", 4222, "NATS server port for agent")
	nodeAgentCmd.PersistentFlags().
		StringP("agent-client-name", "", "osapi-node-agent", "NATS client name for agent")
	nodeAgentCmd.PersistentFlags().
		StringP("agent-queue-group", "", "job-workers", "NATS queue group for load balancing")
	nodeAgentCmd.PersistentFlags().
		StringP("agent-hostname", "", "", "Agent hostname (defaults to system hostname)")
	nodeAgentCmd.PersistentFlags().
		IntP("agent-max-jobs", "", 10, "Maximum concurrent jobs per agent")

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
		"node.agent.nats.host",
		nodeAgentCmd.PersistentFlags().Lookup("agent-host"),
	)
	_ = viper.BindPFlag(
		"node.agent.nats.port",
		nodeAgentCmd.PersistentFlags().Lookup("agent-port"),
	)
	_ = viper.BindPFlag(
		"node.agent.nats.client_name",
		nodeAgentCmd.PersistentFlags().Lookup("agent-client-name"),
	)
	_ = viper.BindPFlag(
		"node.agent.queue_group",
		nodeAgentCmd.PersistentFlags().Lookup("agent-queue-group"),
	)
	_ = viper.BindPFlag(
		"node.agent.hostname",
		nodeAgentCmd.PersistentFlags().Lookup("agent-hostname"),
	)
	_ = viper.BindPFlag(
		"node.agent.max_jobs",
		nodeAgentCmd.PersistentFlags().Lookup("agent-max-jobs"),
	)

	// Bind consumer configuration flags
	_ = viper.BindPFlag(
		"node.agent.consumer.max_deliver",
		nodeAgentCmd.PersistentFlags().Lookup("consumer-max-deliver"),
	)
	_ = viper.BindPFlag(
		"node.agent.consumer.ack_wait",
		nodeAgentCmd.PersistentFlags().Lookup("consumer-ack-wait"),
	)
	_ = viper.BindPFlag(
		"node.agent.consumer.max_ack_pending",
		nodeAgentCmd.PersistentFlags().Lookup("consumer-max-ack-pending"),
	)
	_ = viper.BindPFlag(
		"node.agent.consumer.replay_policy",
		nodeAgentCmd.PersistentFlags().Lookup("consumer-replay-policy"),
	)
	_ = viper.BindPFlag(
		"node.agent.consumer.back_off",
		nodeAgentCmd.PersistentFlags().Lookup("consumer-back-off"),
	)
}
