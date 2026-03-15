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
	"os"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/retr0h/osapi/internal/api"
	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/messaging"
	"github.com/retr0h/osapi/internal/provider/process"
)

// startNATSHeartbeat creates a persistent NATS connection, resolves the
// registry KV bucket, and starts a ComponentHeartbeat for the embedded NATS
// server in a background goroutine. The heartbeat deregisters and the
// connection is closed when ctx is cancelled.
func startNATSHeartbeat(
	ctx context.Context,
	log *slog.Logger,
	host string,
	port int,
	namespace string,
	serverAuth config.NATSServerAuth,
) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Warn(
			"failed to resolve hostname for NATS heartbeat, using 'unknown'",
			slog.String("error", err.Error()),
		)
		hostname = "unknown"
	}

	var nc messaging.NATSClient = natsclient.New(log, &natsclient.Options{
		Host: host,
		Port: port,
		Auth: buildSetupAuth(serverAuth),
		Name: "osapi-nats-heartbeat",
	})

	if err := nc.Connect(); err != nil {
		log.Warn(
			"failed to connect to NATS for heartbeat, skipping",
			slog.String("error", err.Error()),
		)
		return
	}

	registryKVConfig := cli.BuildRegistryKVConfig(namespace, appConfig.NATS.Registry)
	registryKV, err := nc.CreateOrUpdateKVBucketWithConfig(ctx, registryKVConfig)
	if err != nil {
		log.Warn(
			"failed to get registry KV for NATS heartbeat, skipping",
			slog.String("error", err.Error()),
		)
		cli.CloseNATSClient(nc)
		return
	}

	hb := api.NewComponentHeartbeat(
		log,
		registryKV,
		hostname,
		"0.1.0",
		"nats",
		process.New(),
		10*time.Second,
		process.ProcessConditionThresholds{
			MemoryPressureBytes: appConfig.Agent.ProcessConditions.MemoryPressureBytes,
			HighCPUPercent:      appConfig.Agent.ProcessConditions.HighCPUPercent,
		},
	)

	go func() {
		defer cli.CloseNATSClient(nc)
		hb.Start(ctx)
	}()
}

// startNATSHeartbeatFromKV starts a ComponentHeartbeat for the embedded NATS
// server using an already-open registry KV handle. This is used by the
// combined start command which has an existing NATS connection.
func startNATSHeartbeatFromKV(
	ctx context.Context,
	log *slog.Logger,
	registryKV jetstream.KeyValue,
) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Warn(
			"failed to resolve hostname for NATS heartbeat, using 'unknown'",
			slog.String("error", err.Error()),
		)
		hostname = "unknown"
	}

	hb := api.NewComponentHeartbeat(
		log,
		registryKV,
		hostname,
		"0.1.0",
		"nats",
		process.New(),
		10*time.Second,
		process.ProcessConditionThresholds{
			MemoryPressureBytes: appConfig.Agent.ProcessConditions.MemoryPressureBytes,
			HighCPUPercent:      appConfig.Agent.ProcessConditions.HighCPUPercent,
		},
	)

	go hb.Start(ctx)
}
