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
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/agent"
	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	fileProv "github.com/retr0h/osapi/internal/provider/file"
	"github.com/retr0h/osapi/internal/telemetry/process"
	cronProv "github.com/retr0h/osapi/internal/provider/scheduled/cron"
	"github.com/retr0h/osapi/pkg/sdk/platform"
)

// Ensure cli import is used (for BuildFileStateKVConfig etc.)
var _ = cli.BuildFileStateKVConfig

// setupAgent connects to NATS, creates providers, and builds the agent.
// It is used by the standalone agent start and combined start commands.
func setupAgent(
	ctx context.Context,
	log *slog.Logger,
	connCfg config.NATSConnection,
) (*agent.Agent, *natsBundle, map[string]job.SubComponentInfo) {
	namespace := connCfg.Namespace
	streamName := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.Stream.Name)
	kvBucket := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.KV.Bucket)

	b := connectNATSBundle(ctx, log, connCfg, kvBucket, namespace, streamName)

	providerFactory := agent.NewProviderFactory(log, appFs)
	hostProvider, diskProvider, memProvider, loadProvider, dnsProvider, pingProvider, netinfoProvider, commandProvider, dockerProvider := providerFactory.CreateProviders()

	// Create file provider if Object Store and file-state KV are configured.
	hostname, _ := job.GetAgentHostname(appConfig.Agent.Hostname)
	fileProvider, fileStateKV := createFileProvider(ctx, log, b, namespace, hostname)

	// Create cron provider — depends on file provider for SHA-tracked deploys.
	cronProvider := createCronProvider(log, fileProvider, fileStateKV, hostname)

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
		cronProvider,
		process.New(),
		b.registryKV,
		b.factsKV,
	)

	enabledOrDisabled := func(enabled bool) string {
		if enabled {
			return "ok"
		}

		return "disabled"
	}

	subComponents := map[string]job.SubComponentInfo{
		"agent.facts":     {Status: "ok"},
		"agent.heartbeat": {Status: "ok"},
		"agent.metrics": {
			Status: enabledOrDisabled(appConfig.Agent.Metrics.Enabled),
			Address: fmt.Sprintf(
				"http://%s:%d",
				appConfig.Agent.Metrics.Host,
				appConfig.Agent.Metrics.Port,
			),
		},
	}
	a.SetSubComponents(subComponents)

	return a, b, subComponents
}

// createFileProvider creates a file provider if Object Store and file-state KV
// are configured. Returns nil provider and nil KV if either is unavailable.
// The stateKV is also returned so the cron provider can check managed files.
func createFileProvider(
	ctx context.Context,
	log *slog.Logger,
	b *natsBundle,
	namespace string,
	hostname string,
) (fileProv.Provider, jetstream.KeyValue) {
	if appConfig.NATS.Objects.Bucket == "" || appConfig.NATS.FileState.Bucket == "" {
		return nil, nil
	}

	objStoreName := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.Objects.Bucket)
	objStore, err := b.nc.ObjectStore(ctx, objStoreName)
	if err != nil {
		log.Warn("Object Store not available, file operations disabled",
			slog.String("bucket", objStoreName),
			slog.String("error", err.Error()),
		)
		return nil, nil
	}

	fileStateKVConfig := cli.BuildFileStateKVConfig(namespace, appConfig.NATS.FileState)
	fileStateKV, err := b.nc.CreateOrUpdateKVBucketWithConfig(ctx, fileStateKVConfig)
	if err != nil {
		log.Warn("file state KV not available, file operations disabled",
			slog.String("bucket", fileStateKVConfig.Bucket),
			slog.String("error", err.Error()),
		)
		return nil, nil
	}

	// Seed system templates into the object store (idempotent).
	if err := agent.SeedSystemTemplates(ctx, log, objStore); err != nil {
		log.Warn("failed to seed system templates",
			slog.String("error", err.Error()),
		)
	}

	return fileProv.New(log, appFs, objStore, fileStateKV, hostname), fileStateKV
}

// createCronProvider creates a platform-specific cron provider. On Debian, the
// cron provider delegates file writes to the file provider. On other platforms,
// all operations return ErrUnsupported.
func createCronProvider(
	log *slog.Logger,
	fileProvider fileProv.Provider,
	fileStateKV jetstream.KeyValue,
	hostname string,
) cronProv.Provider {
	plat := platform.Detect()

	switch plat {
	case "debian":
		if fileProvider == nil {
			log.Warn("file provider not available, cron operations disabled")
			return cronProv.NewLinuxProvider()
		}
		return cronProv.NewDebianProvider(log, appFs, fileProvider, fileStateKV, hostname)
	case "darwin":
		return cronProv.NewDarwinProvider()
	default:
		return cronProv.NewLinuxProvider()
	}
}
