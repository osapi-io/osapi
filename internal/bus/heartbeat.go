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

package bus

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/heartbeat"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/process"
)

// hostnameFn returns the name of the host (injectable for testing).
var hostnameFn = os.Hostname

// StartHeartbeat connects to the embedded NATS server and writes its
// registration to the registry KV bucket until ctx ends. Then it closes the
// connection. If it cannot connect or open the bucket, it logs a warning and
// returns, because the server runs without a heartbeat.
func StartHeartbeat(
	ctx context.Context,
	log *slog.Logger,
	cfg config.NATS,
	conditions config.ProcessConditions,
) {
	sc := cfg.Server
	nc := natsclient.New(log, &natsclient.Options{
		Host: sc.Host,
		Port: sc.Port,
		Auth: ServerClientAuth(sc.Auth),
		Name: "osapi-nats-heartbeat",
	})
	if err := nc.Connect(); err != nil {
		log.Warn(
			"failed to connect to NATS for heartbeat, skipping",
			slog.String("error", err.Error()),
		)
		return
	}

	registryKV, err := nc.CreateOrUpdateKVBucketWithConfig(
		ctx,
		RegistryKV(sc.Namespace, cfg.Registry),
	)
	if err != nil {
		log.Warn(
			"failed to get registry KV for NATS heartbeat, skipping",
			slog.String("error", err.Error()),
		)
		Close(nc)
		return
	}

	hb := newHeartbeat(log, registryKV, sc, conditions)
	go func() {
		defer Close(nc)
		hb.Start(ctx)
	}()
}

// StartHeartbeatOn writes the registration of the embedded NATS server to
// registryKV until ctx ends. osapi start uses it, because it has a connection
// open already.
func StartHeartbeatOn(
	ctx context.Context,
	log *slog.Logger,
	registryKV jetstream.KeyValue,
	sc config.NATSServer,
	conditions config.ProcessConditions,
) {
	go newHeartbeat(log, registryKV, sc, conditions).Start(ctx)
}

func newHeartbeat(
	log *slog.Logger,
	registryKV jetstream.KeyValue,
	sc config.NATSServer,
	conditions config.ProcessConditions,
) *heartbeat.Heartbeat {
	hostname, err := hostnameFn()
	if err != nil {
		log.Warn(
			"failed to resolve hostname for NATS heartbeat, using 'unknown'",
			slog.String("error", err.Error()),
		)
		hostname = "unknown"
	}

	return heartbeat.New(
		log,
		registryKV,
		hostname,
		"0.1.0",
		"nats",
		process.New(),
		10*time.Second,
		process.ConditionThresholds{
			MemoryPressureBytes: conditions.MemoryPressureBytes,
			HighCPUPercent:      conditions.HighCPUPercent,
		},
		SubComponents(sc),
	)
}

// SubComponents returns the parts of the embedded NATS server that its
// heartbeat and its metrics server report.
func SubComponents(
	sc config.NATSServer,
) map[string]job.SubComponentInfo {
	scs := map[string]job.SubComponentInfo{
		"nats.heartbeat": {Status: "ok"},
		"nats.server": {
			Status:  "ok",
			Address: fmt.Sprintf("nats://%s:%d", sc.Host, sc.Port),
		},
		"nats.metrics": {Status: "disabled"},
	}

	if sc.Metrics.Enabled {
		scs["nats.metrics"] = job.SubComponentInfo{
			Status:  "ok",
			Address: fmt.Sprintf("http://%s:%d", sc.Metrics.Host, sc.Metrics.Port),
		}
	}

	return scs
}
