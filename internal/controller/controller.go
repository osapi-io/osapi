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

// Package controller runs the controller process: the REST API, and the
// watchers and the heartbeat that run beside it.
package controller

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api"
	"github.com/osapi-io/osapi/internal/controller/api/health"
	"github.com/osapi-io/osapi/internal/controller/enrollment"
	"github.com/osapi-io/osapi/internal/job"
	jobclient "github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/validation"
)

// Controller is the controller process.
type Controller struct {
	api           *api.Server
	conn          *jobclient.Connection
	checker       *health.NATSChecker
	subComponents map[string]job.SubComponentInfo
}

// Option changes how New makes the controller.
type Option func(*options)

type options struct {
	meterProvider *sdkmetric.MeterProvider
}

// WithMeterProvider sends the HTTP metrics, the NATS metrics and the job
// metrics of the controller to mp.
func WithMeterProvider(
	mp *sdkmetric.MeterProvider,
) Option {
	return func(o *options) {
		o.meterProvider = mp
	}
}

// New connects to NATS as cfg.Controller.NATS tells, and makes the REST API
// with each handler. It starts the heartbeat, and the enrollment and
// condition watchers that cfg turns on. They run until ctx ends. Start serves
// the API.
func New(
	ctx context.Context,
	log *slog.Logger,
	cfg config.Config,
	opts ...Option,
) (*Controller, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	namespace := cfg.Controller.NATS.Namespace
	timeout, _ := time.ParseDuration(cfg.Controller.API.JobTimeout)

	conn, err := jobclient.Connect(
		ctx,
		log,
		cfg.Controller.NATS,
		cfg.NATS,
		timeout,
		validation.ResolveTarget,
		o.meterProvider,
	)
	if err != nil {
		return nil, err
	}

	c, err := newController(ctx, log, cfg, o, conn, namespace)
	if err != nil {
		bus.Close(conn.NATS)

		return nil, err
	}

	return c, nil
}

func newController(
	ctx context.Context,
	log *slog.Logger,
	cfg config.Config,
	o options,
	conn *jobclient.Connection,
	namespace string,
) (*Controller, error) {
	if o.meterProvider != nil {
		conn.Jobs.SetMeterProvider(o.meterProvider)
	}
	registerTargetValidator(conn.Jobs)

	var apiOpts []api.Option
	if o.meterProvider != nil {
		apiOpts = append(apiOpts, api.WithMeterProvider(o.meterProvider))
	}

	auditStore, err := newAuditStore(ctx, log, conn.NATS, namespace, cfg.NATS.Audit)
	if err != nil {
		return nil, err
	}
	if auditStore != nil {
		apiOpts = append(apiOpts, api.WithAuditStore(auditStore))
	}

	var watcher *enrollment.Watcher
	if cfg.Controller.PKI.Enabled {
		watcher, err = startEnrollment(ctx, log, conn, namespace, cfg)
		if err != nil {
			return nil, err
		}
	}

	objBucket, fileStateBucket := openFileBuckets(ctx, log, conn.NATS, namespace, cfg.NATS)
	checker := newHealthChecker(conn.NATS, conn.JobsBucket)
	metricsProvider := newMetricsProvider(
		conn.NATS,
		job.ApplyNamespaceToInfraName(namespace, cfg.NATS.Stream.Name),
		configuredKVBuckets(namespace, cfg.NATS),
		configuredObjectBuckets(namespace, cfg.NATS),
		conn.Jobs,
		conn.RegistryBucket,
	)

	srv := api.New(cfg, log, apiOpts...)
	srv.RegisterHandlers(handlers(handlerDeps{
		log:             log,
		cfg:             cfg,
		jobs:            conn.Jobs,
		checker:         checker,
		metricsProvider: metricsProvider,
		auditStore:      auditStore,
		nats:            conn.NATS,
		objBucket:       objBucket,
		fileStateBucket: fileStateBucket,
		enrollment:      watcher,
	}))

	subs := subComponents(cfg)
	startHeartbeat(ctx, log, conn.NATS, conn.RegistryBucket, subs, cfg.Agent.ProcessConditions)
	startConditionWatcher(ctx, log, conn.NATS, conn.RegistryBucket, cfg.Controller.Notifications)

	return &Controller{
		api:           srv,
		conn:          conn,
		checker:       checker,
		subComponents: subs,
	}, nil
}

// Start serves the REST API without blocking.
func (c *Controller) Start() {
	c.api.Start()
}

// Stop asks the REST API to stop, then closes the NATS connection.
func (c *Controller) Stop(
	ctx context.Context,
) {
	c.api.Stop(ctx)
	bus.Close(c.conn.NATS)
}

// CheckHealth tells whether NATS and the jobs bucket answer.
func (c *Controller) CheckHealth(
	ctx context.Context,
) error {
	return c.checker.CheckHealth(ctx)
}

// SubComponents returns the parts of the controller that its heartbeat and
// its metrics server report.
func (c *Controller) SubComponents() map[string]job.SubComponentInfo {
	return c.subComponents
}

// registerTargetValidator lets the validation of a target look up the
// agents that the job client knows.
func registerTargetValidator(
	jobs jobclient.JobClient,
) {
	validation.RegisterTargetValidator(
		func(ctx context.Context) ([]validation.AgentTarget, error) {
			agents, err := jobs.ListAgents(ctx)
			if err != nil {
				return nil, err
			}
			targets := make([]validation.AgentTarget, 0, len(agents))
			for _, a := range agents {
				targets = append(targets, validation.AgentTarget{
					MachineID: a.MachineID,
					Hostname:  a.Hostname,
					State:     a.State,
					Labels:    a.Labels,
					Verified:  a.Verified,
				})
			}
			return targets, nil
		},
	)
}

// background runs fn in a goroutine, and logs the error that it returns.
func background(
	ctx context.Context,
	log *slog.Logger,
	name string,
	fn func(context.Context) error,
) {
	go func() {
		if err := fn(ctx); err != nil {
			log.Warn(
				fmt.Sprintf("%s stopped with error", name),
				slog.String("error", err.Error()),
			)
		}
	}()
}
