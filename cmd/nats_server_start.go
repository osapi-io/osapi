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
	"fmt"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	natsembedded "github.com/osapi-io/nats-server/pkg/server"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/messaging"
)

// natsLifecycle adapts the embedded NATS server to the Lifecycle interface.
type natsLifecycle struct {
	server *natsembedded.Server
}

func (n *natsLifecycle) Start() {}

func (n *natsLifecycle) Stop(_ context.Context) { n.server.Stop() }

// natsServerStartCmd represents the natsServerStart command.
var natsServerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the embedded NATS server",
	Long: `Start the embedded NATS server with JetStream enabled.
Configures streams, consumers, and KV buckets needed by the job system.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		host := appConfig.NATS.Server.Host
		port := appConfig.NATS.Server.Port
		storeDir := appConfig.NATS.Server.StoreDir

		opts := &natsembedded.Options{
			Options: &natsserver.Options{
				Host:      host,
				Port:      port,
				JetStream: true,
				StoreDir:  storeDir,
				NoSigs:    true,
				NoLog:     false,
			},
			ReadyTimeout: 5 * time.Second,
		}

		s := natsembedded.New(logger, opts)
		if err := s.Start(); err != nil {
			logFatal("failed to start embedded NATS server", err)
		}

		if err := setupJetStream(ctx, host, port); err != nil {
			s.Stop()
			logFatal("failed to setup JetStream infrastructure", err)
		}

		logger.Info(
			"embedded NATS server running",
			"host", host,
			"port", port,
			"store_dir", storeDir,
		)

		var ns Lifecycle = &natsLifecycle{server: s}
		runServer(ctx, ns)
	},
}

func setupJetStream(
	ctx context.Context,
	host string,
	port int,
) error {
	var nc messaging.NATSClient = natsclient.New(logger, &natsclient.Options{
		Host: host,
		Port: port,
		Auth: natsclient.AuthOptions{
			AuthType: natsclient.NoAuth,
		},
		Name: "osapi-nats-setup",
	})

	if err := nc.Connect(); err != nil {
		return fmt.Errorf("connect to NATS: %w", err)
	}
	defer func() {
		if natsConn, ok := nc.(*natsclient.Client); ok && natsConn.NC != nil {
			natsConn.NC.Close()
		}
	}()

	// Create JOBS stream
	streamMaxAge, _ := time.ParseDuration(appConfig.Job.Stream.MaxAge)
	var streamStorage nats.StorageType
	if appConfig.Job.Stream.Storage == "memory" {
		streamStorage = nats.MemoryStorage
	} else {
		streamStorage = nats.FileStorage
	}

	var streamDiscard nats.DiscardPolicy
	if appConfig.Job.Stream.Discard == "new" {
		streamDiscard = nats.DiscardNew
	} else {
		streamDiscard = nats.DiscardOld
	}

	streamConfig := &nats.StreamConfig{
		Name:     appConfig.Job.StreamName,
		Subjects: []string{appConfig.Job.StreamSubjects},
		MaxAge:   streamMaxAge,
		MaxMsgs:  appConfig.Job.Stream.MaxMsgs,
		Storage:  streamStorage,
		Replicas: appConfig.Job.Stream.Replicas,
		Discard:  streamDiscard,
	}

	if err := nc.CreateOrUpdateStreamWithConfig(ctx, streamConfig); err != nil {
		return fmt.Errorf("create JOBS stream: %w", err)
	}

	// Create consumer
	ackWait, _ := time.ParseDuration(appConfig.Job.Consumer.AckWait)

	backOff := make([]time.Duration, 0, len(appConfig.Job.Consumer.BackOff))
	for _, b := range appConfig.Job.Consumer.BackOff {
		d, _ := time.ParseDuration(b)
		backOff = append(backOff, d)
	}

	var replayPolicy jetstream.ReplayPolicy
	if appConfig.Job.Consumer.ReplayPolicy == "original" {
		replayPolicy = jetstream.ReplayOriginalPolicy
	} else {
		replayPolicy = jetstream.ReplayInstantPolicy
	}

	consumerConfig := jetstream.ConsumerConfig{
		Durable:       appConfig.Job.ConsumerName,
		AckWait:       ackWait,
		MaxDeliver:    appConfig.Job.Consumer.MaxDeliver,
		MaxAckPending: appConfig.Job.Consumer.MaxAckPending,
		ReplayPolicy:  replayPolicy,
		BackOff:       backOff,
	}

	if err := nc.CreateOrUpdateConsumerWithConfig(ctx, appConfig.Job.StreamName, consumerConfig); err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}

	// Create KV buckets
	if _, err := nc.CreateKVBucket(appConfig.Job.KVBucket); err != nil {
		return fmt.Errorf("create KV bucket %s: %w", appConfig.Job.KVBucket, err)
	}

	if _, err := nc.CreateKVBucket(appConfig.Job.KVResponseBucket); err != nil {
		return fmt.Errorf("create KV bucket %s: %w", appConfig.Job.KVResponseBucket, err)
	}

	// Create DLQ stream
	dlqMaxAge, _ := time.ParseDuration(appConfig.Job.DLQ.MaxAge)
	var dlqStorage nats.StorageType
	if appConfig.Job.DLQ.Storage == "memory" {
		dlqStorage = nats.MemoryStorage
	} else {
		dlqStorage = nats.FileStorage
	}

	dlqStreamConfig := &nats.StreamConfig{
		Name: appConfig.Job.StreamName + "-DLQ",
		Subjects: []string{
			"$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES." + appConfig.Job.StreamName + ".*",
		},
		Storage:  dlqStorage,
		MaxAge:   dlqMaxAge,
		MaxMsgs:  appConfig.Job.DLQ.MaxMsgs,
		Replicas: appConfig.Job.DLQ.Replicas,
		Metadata: map[string]string{
			"dead_letter_queue": "true",
		},
	}

	if err := nc.CreateOrUpdateStreamWithConfig(ctx, dlqStreamConfig); err != nil {
		return fmt.Errorf("create DLQ stream: %w", err)
	}

	logger.Info("JetStream infrastructure configured successfully")

	return nil
}

func init() {
	natsServerCmd.AddCommand(natsServerStartCmd)
}
