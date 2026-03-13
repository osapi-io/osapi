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
	"log/slog"

	"github.com/retr0h/osapi/internal/agent"
	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	fileProv "github.com/retr0h/osapi/internal/provider/file"
)

// setupAgent connects to NATS, creates providers, and builds the agent
// Lifecycle. It is used by the standalone agent start and combined start
// commands.
func setupAgent(
	ctx context.Context,
	log *slog.Logger,
	connCfg config.NATSConnection,
) (cli.Lifecycle, *natsBundle) {
	namespace := connCfg.Namespace
	streamName := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.Stream.Name)
	kvBucket := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.KV.Bucket)

	b := connectNATSBundle(ctx, log, connCfg, kvBucket, namespace, streamName)

	providerFactory := agent.NewProviderFactory(log)
	hostProvider, diskProvider, memProvider, loadProvider, dnsProvider, pingProvider, netinfoProvider, commandProvider, dockerProvider := providerFactory.CreateProviders()

	// Create file provider if Object Store and file-state KV are configured
	hostname, _ := job.GetAgentHostname(appConfig.Agent.Hostname)
	fileProvider := createFileProvider(ctx, log, b, namespace, hostname)

	a := agent.New(
		appFs,
		appConfig,
		log,
		b.jobClient,
		streamName,
		hostProvider,
		diskProvider,
		memProvider,
		loadProvider,
		dnsProvider,
		pingProvider,
		netinfoProvider,
		commandProvider,
		fileProvider,
		dockerProvider,
		b.registryKV,
		b.factsKV,
	)

	return a, b
}

// createFileProvider creates a file provider if Object Store and file-state KV
// are configured. Returns nil if either is unavailable.
func createFileProvider(
	ctx context.Context,
	log *slog.Logger,
	b *natsBundle,
	namespace string,
	hostname string,
) fileProv.Provider {
	if appConfig.NATS.Objects.Bucket == "" || appConfig.NATS.FileState.Bucket == "" {
		return nil
	}

	objStoreName := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.Objects.Bucket)
	objStore, err := b.nc.ObjectStore(ctx, objStoreName)
	if err != nil {
		log.Warn("Object Store not available, file operations disabled",
			slog.String("bucket", objStoreName),
			slog.String("error", err.Error()),
		)
		return nil
	}

	fileStateKVConfig := cli.BuildFileStateKVConfig(namespace, appConfig.NATS.FileState)
	fileStateKV, err := b.nc.CreateOrUpdateKVBucketWithConfig(ctx, fileStateKVConfig)
	if err != nil {
		log.Warn("file state KV not available, file operations disabled",
			slog.String("bucket", fileStateKVConfig.Bucket),
			slog.String("error", err.Error()),
		)
		return nil
	}

	return fileProv.New(log, appFs, objStore, fileStateKV, hostname)
}
