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

package natsserver

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/avfs/avfs/vfs/osfs"

	"github.com/osapi-io/osapi/internal/bus/server"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/cryptoreport"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/telemetry/metrics"
)

// Start sets the namespace of the job subjects, then makes the components.
func Start(
	ctx context.Context,
	d Deps,
) (command.Running, error) {
	n := d.Config.NATS
	d.Logger.Debug(
		"nats server configuration",
		slog.String("config_file", d.ConfigFile),
		slog.Bool("debug", d.Config.Debug),
		slog.String("nats.server.host", n.Server.Host),
		slog.Int("nats.server.port", n.Server.Port),
		slog.String("nats.server.store_dir", n.Server.StoreDir),
		slog.String("nats.server.namespace", n.Server.Namespace),
		slog.String("nats.server.auth.type", n.Server.Auth.Type),
		slog.String("nats.stream.name", n.Stream.Name),
		slog.String("nats.stream.subjects", n.Stream.Subjects),
		slog.String("nats.kv.bucket", n.KV.Bucket),
	)

	job.Init(n.Server.Namespace)

	return Components(ctx, d)
}

// Components makes the NATS server, and its metrics server when the config
// turns it on: the parts that osapi start runs too. ctx is the command's
// context: the heartbeat runs until it ends.
func Components(
	ctx context.Context,
	d Deps,
) (command.Running, error) {
	n := d.Config.NATS
	log := d.Logger.With("component", "nats")
	// The metrics server is made before the NATS connection, so that
	// nats-client records the calls of the bus on it.
	ms, err := metrics.FromConfig(
		n.Server.Metrics, n.Server.KeyDir, "nats.server.key_dir", "nats-metrics", log,
	)
	if err != nil {
		return command.Running{}, err
	}
	s, err := d.NewServer(ctx, log, n, ms.MeterProvider())
	if err != nil {
		return command.Running{}, fmt.Errorf("start the NATS server: %w", err)
	}

	certs := pki.ReadCerts(osfs.NewWithNoIdm(), config.CertSourcesOf(d.Config, "nats"), time.Now())
	report := cryptoreport.NATS(n.Server, certs)
	s.StartHeartbeat(ctx, d.Version, d.Config.Agent.ProcessConditions, &report)

	components := []command.Component{s}
	if ms != nil {
		ms.SetReadinessFunc(ready)
		ms.RegisterSubsystems(metrics.Subsystems(server.SubComponents(n.Server)))
		components = append(components, ms)
	}

	return command.Running{Components: components}, nil
}

// ready is the readiness of the NATS server: New returns it running, so it
// is ready while the process runs.
func ready() error {
	return nil
}
