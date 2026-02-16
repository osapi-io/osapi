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
	"time"

	"github.com/nats-io/nats.go"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/retr0h/osapi/internal/job/client"
	"github.com/retr0h/osapi/internal/messaging"
)

var (
	natsClient messaging.NATSClient
	jobsKV     nats.KeyValue
	jobClient  client.JobClient
)

// clientJobCmd represents the clientJob command.
var clientJobCmd = &cobra.Command{
	Use:   "job",
	Short: "The job subcommand for direct NATS interaction",
	Long: `The job subcommand allows direct interaction with the NATS job queue
for testing and debugging purposes. This bypasses the API and talks directly
to NATS using the nats-client library.`,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		validateDistribution()

		ctx := cmd.Context()

		logger.Debug(
			"job client configuration",
			slog.Bool("debug", appConfig.Debug),
			slog.String("client.host", appConfig.Job.Client.Host),
			slog.Int("client.port", appConfig.Job.Client.Port),
			slog.String("client.client_name", appConfig.Job.Client.ClientName),
			slog.String("stream_name", appConfig.StreamName),
			slog.String("stream_subjects", appConfig.StreamSubjects),
			slog.String("kv_bucket", appConfig.KVBucket),
		)

		// Create NATS client
		var nc messaging.NATSClient = natsclient.New(logger, &natsclient.Options{
			Host: appConfig.Job.Client.Host,
			Port: appConfig.Job.Client.Port,
			Auth: natsclient.AuthOptions{
				AuthType: natsclient.NoAuth,
			},
			Name: appConfig.Job.Client.ClientName,
		})
		natsClient = nc

		if err := natsClient.Connect(); err != nil {
			logFatal("failed to connect to NATS", err)
		}

		// Setup JOBS stream for subject routing
		streamConfig := &nats.StreamConfig{
			Name:     appConfig.StreamName,
			Subjects: []string{appConfig.StreamSubjects},
		}

		if err := natsClient.CreateOrUpdateStreamWithConfig(ctx, streamConfig); err != nil {
			logFatal("failed to setup JetStream", err)
		}

		// Setup DLQ stream for failed jobs using advisory messages
		dlqMaxAge, _ := time.ParseDuration(appConfig.DLQ.MaxAge)
		var dlqStorage nats.StorageType
		if appConfig.DLQ.Storage == "memory" {
			dlqStorage = nats.MemoryStorage
		} else {
			dlqStorage = nats.FileStorage
		}

		dlqStreamConfig := &nats.StreamConfig{
			Name: "JOBS-DLQ",
			Subjects: []string{
				"$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES." + appConfig.StreamName + ".*",
			},
			Storage:  dlqStorage,
			MaxAge:   dlqMaxAge,
			MaxMsgs:  appConfig.DLQ.MaxMsgs,
			Replicas: appConfig.DLQ.Replicas,
			Metadata: map[string]string{
				"dead_letter_queue": "true",
			},
		}
		if err := natsClient.CreateOrUpdateStreamWithConfig(ctx, dlqStreamConfig); err != nil {
			logFatal("failed to setup DLQ stream", err)
		}

		// Create/get the jobs KV bucket
		var err error
		jobsKV, err = natsClient.CreateKVBucket(appConfig.KVBucket)
		if err != nil {
			logFatal("failed to create KV bucket", err)
		}

		// Create job client
		var jc client.JobClient
		jc, err = client.New(logger, natsClient, &client.Options{
			Timeout:  30 * time.Second, // Default timeout
			KVBucket: jobsKV,
		})
		if err != nil {
			logFatal("failed to create job client", err)
		}
		jobClient = jc
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		if natsClient != nil {
			if nc, ok := natsClient.(*natsclient.Client); ok && nc.NC != nil {
				nc.NC.Close()
			}
		}
	},
}

func init() {
	clientCmd.AddCommand(clientJobCmd)

	clientJobCmd.PersistentFlags().
		StringP("nats-host", "", "localhost", "NATS server hostname")
	clientJobCmd.PersistentFlags().
		IntP("nats-port", "", 4222, "NATS server port")
	clientJobCmd.PersistentFlags().
		StringP("client-name", "", "osapi-jobs-cli", "NATS client name")
	clientJobCmd.PersistentFlags().
		StringP("kv-bucket", "", "job-queue", "KV bucket name for job storage")
	clientJobCmd.PersistentFlags().
		StringP("stream-name", "", "JOBS", "JetStream stream name")
	clientJobCmd.PersistentFlags().
		StringP("stream-subjects", "", "jobs.>", "JetStream stream subjects pattern")
	clientJobCmd.PersistentFlags().
		StringP("kv-response-bucket", "", "job-responses", "KV bucket name for job responses")
	clientJobCmd.PersistentFlags().
		StringP("consumer-name", "", "jobs-worker", "JetStream consumer name")

	// Stream configuration
	clientJobCmd.PersistentFlags().
		StringP("stream-max-age", "", "24h", "JetStream stream max age")
	clientJobCmd.PersistentFlags().
		IntP("stream-max-msgs", "", 10000, "JetStream stream max messages")
	clientJobCmd.PersistentFlags().
		StringP("stream-storage", "", "file", "JetStream stream storage type")
	clientJobCmd.PersistentFlags().
		IntP("stream-replicas", "", 1, "JetStream stream replicas")
	clientJobCmd.PersistentFlags().
		StringP("stream-discard", "", "old", "JetStream stream discard policy")

	// Consumer configuration
	clientJobCmd.PersistentFlags().
		IntP("consumer-max-deliver", "", 5, "JetStream consumer max deliver attempts")
	clientJobCmd.PersistentFlags().
		StringP("consumer-ack-wait", "", "30s", "JetStream consumer ack wait time")
	clientJobCmd.PersistentFlags().
		IntP("consumer-max-ack-pending", "", 100, "JetStream consumer max ack pending")
	clientJobCmd.PersistentFlags().
		StringP("consumer-replay-policy", "", "instant", "JetStream consumer replay policy")

	// KeyValue bucket configuration
	clientJobCmd.PersistentFlags().
		StringP("kv-ttl", "", "1h", "KV bucket TTL")
	clientJobCmd.PersistentFlags().
		IntP("kv-max-bytes", "", 104857600, "KV bucket max bytes (100MB)")
	clientJobCmd.PersistentFlags().
		StringP("kv-storage", "", "file", "KV bucket storage type")
	clientJobCmd.PersistentFlags().
		IntP("kv-replicas", "", 1, "KV bucket replicas")

	// DLQ configuration
	clientJobCmd.PersistentFlags().
		StringP("dlq-max-age", "", "7d", "DLQ stream max age")
	clientJobCmd.PersistentFlags().
		IntP("dlq-max-msgs", "", 1000, "DLQ stream max messages")
	clientJobCmd.PersistentFlags().
		StringP("dlq-storage", "", "file", "DLQ stream storage type")
	clientJobCmd.PersistentFlags().
		IntP("dlq-replicas", "", 1, "DLQ stream replicas")

	// Server configuration
	clientJobCmd.PersistentFlags().
		StringP("server-host", "", "localhost", "Job server hostname")
	clientJobCmd.PersistentFlags().
		IntP("server-port", "", 4222, "Job server port")

	// Bind flags to viper config
	_ = viper.BindPFlag("job.client.host", clientJobCmd.PersistentFlags().Lookup("nats-host"))
	_ = viper.BindPFlag("job.client.port", clientJobCmd.PersistentFlags().Lookup("nats-port"))
	_ = viper.BindPFlag(
		"job.client.client_name",
		clientJobCmd.PersistentFlags().Lookup("client-name"),
	)
	_ = viper.BindPFlag("job.kv_bucket", clientJobCmd.PersistentFlags().Lookup("kv-bucket"))
	_ = viper.BindPFlag("job.stream_name", clientJobCmd.PersistentFlags().Lookup("stream-name"))
	_ = viper.BindPFlag(
		"job.stream_subjects",
		clientJobCmd.PersistentFlags().Lookup("stream-subjects"),
	)
	_ = viper.BindPFlag(
		"job.kv_response_bucket",
		clientJobCmd.PersistentFlags().Lookup("kv-response-bucket"),
	)
	_ = viper.BindPFlag(
		"job.consumer_name",
		clientJobCmd.PersistentFlags().Lookup("consumer-name"),
	)

	// Stream configuration bindings
	_ = viper.BindPFlag(
		"job.stream.max_age",
		clientJobCmd.PersistentFlags().Lookup("stream-max-age"),
	)
	_ = viper.BindPFlag(
		"job.stream.max_msgs",
		clientJobCmd.PersistentFlags().Lookup("stream-max-msgs"),
	)
	_ = viper.BindPFlag(
		"job.stream.storage",
		clientJobCmd.PersistentFlags().Lookup("stream-storage"),
	)
	_ = viper.BindPFlag(
		"job.stream.replicas",
		clientJobCmd.PersistentFlags().Lookup("stream-replicas"),
	)
	_ = viper.BindPFlag(
		"job.stream.discard",
		clientJobCmd.PersistentFlags().Lookup("stream-discard"),
	)

	// Consumer configuration bindings
	_ = viper.BindPFlag(
		"job.consumer.max_deliver",
		clientJobCmd.PersistentFlags().Lookup("consumer-max-deliver"),
	)
	_ = viper.BindPFlag(
		"job.consumer.ack_wait",
		clientJobCmd.PersistentFlags().Lookup("consumer-ack-wait"),
	)
	_ = viper.BindPFlag(
		"job.consumer.max_ack_pending",
		clientJobCmd.PersistentFlags().Lookup("consumer-max-ack-pending"),
	)
	_ = viper.BindPFlag(
		"job.consumer.replay_policy",
		clientJobCmd.PersistentFlags().Lookup("consumer-replay-policy"),
	)

	// KeyValue configuration bindings
	_ = viper.BindPFlag("job.kv.ttl", clientJobCmd.PersistentFlags().Lookup("kv-ttl"))
	_ = viper.BindPFlag("job.kv.max_bytes", clientJobCmd.PersistentFlags().Lookup("kv-max-bytes"))
	_ = viper.BindPFlag("job.kv.storage", clientJobCmd.PersistentFlags().Lookup("kv-storage"))
	_ = viper.BindPFlag("job.kv.replicas", clientJobCmd.PersistentFlags().Lookup("kv-replicas"))

	// DLQ configuration bindings
	_ = viper.BindPFlag("job.dlq.max_age", clientJobCmd.PersistentFlags().Lookup("dlq-max-age"))
	_ = viper.BindPFlag("job.dlq.max_msgs", clientJobCmd.PersistentFlags().Lookup("dlq-max-msgs"))
	_ = viper.BindPFlag("job.dlq.storage", clientJobCmd.PersistentFlags().Lookup("dlq-storage"))
	_ = viper.BindPFlag("job.dlq.replicas", clientJobCmd.PersistentFlags().Lookup("dlq-replicas"))

	// Server configuration bindings
	_ = viper.BindPFlag("job.server.host", clientJobCmd.PersistentFlags().Lookup("server-host"))
	_ = viper.BindPFlag("job.server.port", clientJobCmd.PersistentFlags().Lookup("server-port"))
}
