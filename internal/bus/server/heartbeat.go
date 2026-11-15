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

package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/heartbeat"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/process"
)

// hostnameFn returns the name of the host (injectable for testing).
var hostnameFn = os.Hostname

// StartHeartbeat writes the registration of the server to the registry KV
// bucket until ctx ends. If the config names no registry bucket, it logs a
// warning and returns, because the server runs without a heartbeat.
func (s *Server) StartHeartbeat(
	ctx context.Context,
	conditions config.ProcessConditions,
) {
	if s.registryKV == nil {
		s.log.Warn("no registry KV bucket, NATS heartbeat skipped")
		return
	}

	hostname, err := hostnameFn()
	if err != nil {
		s.log.Warn(
			"failed to resolve hostname for NATS heartbeat, using 'unknown'",
			slog.String("error", err.Error()),
		)
		hostname = "unknown"
	}

	hb := heartbeat.New(s.log, heartbeat.Options{
		RegistryKV: s.registryKV,
		Hostname:   hostname,
		Version:    "0.1.0",
		Component:  "nats",
		Process:    process.New(),
		Interval:   10 * time.Second,
		Thresholds: process.ConditionThresholds{
			MemoryPressureBytes: conditions.MemoryPressureBytes,
			HighCPUPercent:      conditions.HighCPUPercent,
		},
		SubComponents: SubComponents(s.cfg),
	})
	go hb.Start(ctx)
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
			URL:     sc.URL,
		},
		"nats.metrics": {Status: "disabled"},
	}

	if sc.Metrics.Enabled {
		scs["nats.metrics"] = job.SubComponentInfo{
			Status:  "ok",
			Address: fmt.Sprintf("http://%s:%d", sc.Metrics.Host, sc.Metrics.Port),
			URL:     sc.Metrics.URL,
		}
	}

	return scs
}
