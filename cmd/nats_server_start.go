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
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	natsembedded "github.com/osapi-io/nats-server/pkg/server"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
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
		namespace := appConfig.NATS.Server.Namespace

		// Initialize subject namespace
		job.Init(namespace)

		serverAuth := appConfig.NATS.Server.Auth
		serverOpts := buildNATSServerOpts(host, port, storeDir, serverAuth)

		opts := &natsembedded.Options{
			Options:      serverOpts,
			ReadyTimeout: 5 * time.Second,
		}

		s := natsembedded.New(logger, opts)
		if err := s.Start(); err != nil {
			cli.LogFatal(logger, "failed to start embedded NATS server", err)
		}

		if err := setupJetStream(ctx, host, port, namespace, serverAuth); err != nil {
			s.Stop()
			cli.LogFatal(logger, "failed to setup JetStream infrastructure", err)
		}

		logger.Info(
			"embedded NATS server running",
			"host", host,
			"port", port,
			"store_dir", storeDir,
			"namespace", namespace,
			"auth.type", serverAuth.Type,
		)

		var ns cli.Lifecycle = &natsLifecycle{server: s}
		cli.RunServer(ctx, ns)
	},
}

func setupJetStream(
	ctx context.Context,
	host string,
	port int,
	namespace string,
	serverAuth config.NATSServerAuth,
) error {
	var nc messaging.NATSClient = natsclient.New(logger, &natsclient.Options{
		Host: host,
		Port: port,
		Auth: buildSetupAuth(serverAuth),
		Name: "osapi-nats-setup",
	})

	if err := nc.Connect(); err != nil {
		return fmt.Errorf("connect to NATS: %w", err)
	}
	defer cli.CloseNATSClient(nc)

	// Apply namespace to infrastructure names
	streamName := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.Stream.Name)
	streamSubjects := job.ApplyNamespaceToSubjects(namespace, appConfig.NATS.Stream.Subjects)
	kvBucket := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.KV.Bucket)
	kvResponseBucket := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.KV.ResponseBucket)

	// Create JOBS stream
	streamMaxAge, _ := time.ParseDuration(appConfig.NATS.Stream.MaxAge)
	streamStorage := cli.ParseJetstreamStorageType(appConfig.NATS.Stream.Storage)

	var streamDiscard jetstream.DiscardPolicy
	if appConfig.NATS.Stream.Discard == "new" {
		streamDiscard = jetstream.DiscardNew
	} else {
		streamDiscard = jetstream.DiscardOld
	}

	streamConfig := jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{streamSubjects},
		MaxAge:   streamMaxAge,
		MaxMsgs:  appConfig.NATS.Stream.MaxMsgs,
		Storage:  streamStorage,
		Replicas: appConfig.NATS.Stream.Replicas,
		Discard:  streamDiscard,
	}

	if err := nc.CreateOrUpdateStreamWithConfig(ctx, streamConfig); err != nil {
		return fmt.Errorf("create JOBS stream: %w", err)
	}

	// Create KV buckets
	if _, err := nc.CreateOrUpdateKVBucket(ctx, kvBucket); err != nil {
		return fmt.Errorf("create KV bucket %s: %w", kvBucket, err)
	}

	if _, err := nc.CreateOrUpdateKVBucket(ctx, kvResponseBucket); err != nil {
		return fmt.Errorf("create KV bucket %s: %w", kvResponseBucket, err)
	}

	// Create audit KV bucket with configured settings
	if appConfig.NATS.Audit.Bucket != "" {
		auditKVConfig := cli.BuildAuditKVConfig(namespace, appConfig.NATS.Audit)
		if _, err := nc.CreateOrUpdateKVBucketWithConfig(ctx, auditKVConfig); err != nil {
			return fmt.Errorf("create audit KV bucket %s: %w", auditKVConfig.Bucket, err)
		}
	}

	// Create DLQ stream
	dlqMaxAge, _ := time.ParseDuration(appConfig.NATS.DLQ.MaxAge)
	dlqStorage := cli.ParseJetstreamStorageType(appConfig.NATS.DLQ.Storage)

	dlqStreamConfig := jetstream.StreamConfig{
		Name: streamName + "-DLQ",
		Subjects: []string{
			"$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES." + streamName + ".*",
		},
		Storage:  dlqStorage,
		MaxAge:   dlqMaxAge,
		MaxMsgs:  appConfig.NATS.DLQ.MaxMsgs,
		Replicas: appConfig.NATS.DLQ.Replicas,
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

func buildNATSServerOpts(
	host string,
	port int,
	storeDir string,
	serverAuth config.NATSServerAuth,
) *natsserver.Options {
	opts := &natsserver.Options{
		Host:      host,
		Port:      port,
		JetStream: true,
		StoreDir:  storeDir,
		NoSigs:    true,
		NoLog:     false,
	}

	switch serverAuth.Type {
	case "user_pass":
		users := make([]*natsserver.User, 0, len(serverAuth.Users))
		for _, u := range serverAuth.Users {
			users = append(users, &natsserver.User{
				Username: u.Username,
				Password: u.Password,
			})
		}
		opts.Users = users
	case "nkey":
		nkeys := make([]*natsserver.NkeyUser, 0, len(serverAuth.NKeys))
		for _, nk := range serverAuth.NKeys {
			nkeys = append(nkeys, &natsserver.NkeyUser{
				Nkey: nk,
			})
		}
		opts.Nkeys = nkeys
	}

	return opts
}

func buildSetupAuth(
	serverAuth config.NATSServerAuth,
) natsclient.AuthOptions {
	if serverAuth.Type == "user_pass" && len(serverAuth.Users) > 0 {
		return natsclient.AuthOptions{
			AuthType: natsclient.UserPassAuth,
			Username: serverAuth.Users[0].Username,
			Password: serverAuth.Users[0].Password,
		}
	}

	return natsclient.AuthOptions{
		AuthType: natsclient.NoAuth,
	}
}

func init() {
	natsServerCmd.AddCommand(natsServerStartCmd)
}
