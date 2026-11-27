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

package controllercmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/metrics"
)

// Start logs the config, with no secret in it, starts the tracer, sets
// the namespace of the job subjects, then makes the components. The tracer
// stops after them.
func Start(
	ctx context.Context,
	d Deps,
) (command.Running, error) {
	c := d.Config.Controller
	d.Logger.Debug(
		"controller configuration",
		slog.String("config_file", d.ConfigFile),
		slog.Bool("debug", d.Config.Debug),
		slog.Int("controller.api.port", c.API.Port),
		slog.String("controller.nats.host", c.NATS.Host),
		slog.Int("controller.nats.port", c.NATS.Port),
		slog.String("controller.nats.client_name", c.NATS.ClientName),
		slog.String("controller.nats.namespace", c.NATS.Namespace),
		slog.String("controller.nats.auth.type", c.NATS.Auth.Type),
		slog.String(
			"controller.api.security.cors.allow_origins",
			strings.Join(c.API.Security.CORS.AllowOrigins, ","),
		),
	)

	stop, err := d.Tracer(ctx, "osapi-controller")
	if err != nil {
		return command.Running{}, fmt.Errorf("start the tracer: %w", err)
	}
	job.Init(d.Config.Controller.NATS.Namespace)

	r, err := Components(ctx, d)
	if err != nil {
		return command.Running{}, err
	}

	return command.Then(r, func(ctx context.Context) { _ = stop(ctx) }), nil
}

// Components makes the controller, and its metrics server when the config
// turns it on: the parts that osapi start runs too. The metrics server also
// serves the HTTP metrics of the API, so it is made first.
func Components(
	ctx context.Context,
	d Deps,
) (command.Running, error) {
	log := d.Logger.With("component", "controller")
	ms := metrics.FromConfig(
		d.Config.Controller.Metrics,
		d.Config.Controller.PKI.KeyDir,
		"controller.pki.key_dir",
		"controller-metrics",
		log,
	)
	opts := []controller.Option{controller.WithVersion(d.Version)}
	if ms != nil {
		opts = append(opts, controller.WithMeterProvider(ms.MeterProvider()))
	}

	c, err := d.NewController(ctx, log, d.Config, opts...)
	if err != nil {
		return command.Running{}, fmt.Errorf("start the controller: %w", err)
	}

	components := []command.Component{c}
	if ms != nil {
		ms.SetReadinessFunc(readiness(c))
		ms.RegisterSubsystems(metrics.Subsystems(c.SubComponents()))
		components = append(components, ms)
	}

	return command.Running{Components: components}, nil
}

// readiness is the readiness of the metrics server: the health of c.
func readiness(
	c Controller,
) func() error {
	return func() error { return c.CheckHealth(context.Background()) }
}
