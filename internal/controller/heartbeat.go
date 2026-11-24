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

package controller

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/heartbeat"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/process"
)

// hostnameFn returns the name of the host (injectable for testing).
var hostnameFn = os.Hostname

// startHeartbeat writes the registration of the controller to registryKV
// until ctx ends. Then the heartbeat removes it.
func startHeartbeat(
	ctx context.Context,
	log *slog.Logger,
	registry heartbeat.RegistryStore,
	registryBucket string,
	version string,
	subs map[string]job.SubComponentInfo,
	conditions config.ProcessConditions,
) {
	hostname, err := hostnameFn()
	if err != nil {
		log.Warn(
			"failed to resolve hostname for controller heartbeat, using 'unknown'",
			slog.String("error", err.Error()),
		)
		hostname = "unknown"
	}

	hb := heartbeat.New(log, heartbeat.Options{
		KV:             registry,
		RegistryBucket: registryBucket,
		Hostname:       hostname,
		Version:        version,
		Component:      "controller",
		Process:        process.New(),
		Interval:       10 * time.Second,
		Thresholds: process.ConditionThresholds{
			MemoryPressureBytes: conditions.MemoryPressureBytes,
			HighCPUPercent:      conditions.HighCPUPercent,
		},
		SubComponents: subs,
	})
	go hb.Start(ctx)
}

// subComponents returns the parts of the controller that its heartbeat and
// its metrics server report.
func subComponents(
	cfg config.Config,
) map[string]job.SubComponentInfo {
	// An empty API host binds every interface, which is what 0.0.0.0 shows.
	apiHost := cfg.Controller.API.Host
	if apiHost == "" {
		apiHost = "0.0.0.0"
	}
	api := httpAddr(apiHost, cfg.Controller.API.Port)

	return map[string]job.SubComponentInfo{
		"controller.api": {
			Status:  "ok",
			Address: api,
			URL:     cfg.Controller.API.URL,
		},
		"controller.heartbeat": {Status: "ok"},
		"controller.metrics": {
			Status:  enabledOrDisabled(cfg.Controller.Metrics.Enabled),
			Address: httpAddr(cfg.Controller.Metrics.Host, cfg.Controller.Metrics.Port),
			URL:     cfg.Controller.Metrics.URL,
		},
		"controller.pki":      {Status: "ok"},
		"controller.notifier": {Status: enabledOrDisabled(cfg.Controller.Notifications.Enabled)},
		"controller.tracing": {
			Status:  enabledOrDisabled(cfg.Telemetry.Tracing.Enabled),
			Address: tracingAddr(cfg.Telemetry.Tracing),
			URL:     cfg.Telemetry.Tracing.URL,
		},
		"controller.ui": {
			Status:  enabledOrDisabled(cfg.Controller.UI.UIEnabled()),
			Address: api,
			URL:     cfg.Controller.UI.URL,
		},
	}
}

// tracingAddr is where the controller sends its spans, the OTLP gRPC
// endpoint. Tracing listens on nothing, so this is the one sub-component whose
// address is a destination, not a bind address. stdout exports nowhere.
func tracingAddr(
	tc config.TracingConfig,
) string {
	if !tc.Enabled || tc.Exporter != "otlp" || tc.OTLPEndpoint == "" {
		return ""
	}

	return "grpc://" + tc.OTLPEndpoint
}

func httpAddr(
	host string,
	port int,
) string {
	return "http://" + net.JoinHostPort(host, strconv.Itoa(port))
}

func enabledOrDisabled(
	enabled bool,
) string {
	if enabled {
		return "ok"
	}

	return "disabled"
}
