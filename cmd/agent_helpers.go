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
	hostProvider, diskProvider, memProvider, loadProvider, dnsProvider, pingProvider, commandProvider := providerFactory.CreateProviders()

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
		commandProvider,
		b.registryKV,
	)

	return a, b
}
