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
	"io/fs"
	"log/slog"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior/dispatch"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/admission"
	commandAPI "github.com/osapi-io/osapi/internal/controller/api/node/command"
	uihandler "github.com/osapi-io/osapi/internal/controller/api/ui"
	endpointserver "github.com/osapi-io/osapi/internal/controller/endpoint/server"
	admissionEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/admission"
	"github.com/osapi-io/osapi/internal/controller/endpoints/agent"
	auditEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/audit"
	controllerEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/controller"
	factsEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/facts"
	healthEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/health"
	jobEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/job"
	natsEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/nats"
	"github.com/osapi-io/osapi/internal/controller/endpoints/telemetry"
	"github.com/osapi-io/osapi/internal/controller/health"
	jobclient "github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/telemetry/traces"
	uiassets "github.com/osapi-io/osapi/ui"

	// Registers the routes of every endpoint group that is served.
	_ "github.com/osapi-io/osapi/internal/controller/endpoints/all/gen"
)

// handlerDeps is what the REST API handlers use.
type handlerDeps struct {
	log             *slog.Logger
	cfg             config.Config
	jobs            jobclient.JobClient
	checker         health.Checker
	metricsProvider health.MetricsProvider
	auditStore      audit.Store
	nats            fileBuckets
	objBucket       string
	fileStateBucket string
	admission       *admission.Watcher
	// version is the build version of the binary.
	version string
}

// handlers returns each REST API handler. The UI handler is only there
// when it is enabled.
func handlers(
	d handlerDeps,
) []func(e *echo.Echo) {
	startTime := time.Now()
	signingKey := d.cfg.Controller.API.Security.SigningKey

	// Build custom roles map from config.
	var customRoles authtoken.CustomRoles
	if cfgRoles := d.cfg.Controller.API.Security.Roles; len(cfgRoles) > 0 {
		customRoles = make(authtoken.CustomRoles, len(cfgRoles))
		for name, role := range cfgRoles {
			customRoles[name] = role.Permissions
		}
	}

	// The controller's sub-components come to the health answer with its
	// registry entry, from its heartbeat.
	reporter := health.New(d.log, d.checker, startTime, d.version, d.metricsProvider)
	reporter.StartMetricsRefresh(context.Background())

	handlers := make([]func(e *echo.Echo), 0, 16)
	handlers = append(
		handlers,
		endpointserver.Handler(
			d.log,
			signingKey,
			customRoles,
			agent.Deps{Jobs: d.jobs, Health: reporter},
			admissionDeps(d),
			telemetry.Deps{Traces: traceReader(d.cfg.Telemetry.Tracing)},
			auditEndpoint.Deps{Store: d.auditStore},
			jobEndpoint.Deps{Jobs: d.jobs},
			healthEndpoint.Deps{Health: reporter},
			controllerEndpoint.Deps{Health: reporter},
			natsEndpoint.Deps{Health: reporter},
			fileDeps(d.nats, d.objBucket, d.fileStateBucket),
			factsEndpoint.Deps{},
			// Each behavior's endpoint group, which the gen packages that
			// internal/behaviors/all/gen imports register.
			dispatch.Deps{JobClient: d.jobs},
		)...,
	)
	handlers = append(handlers, commandAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)

	if d.cfg.Controller.UI.UIEnabled() {
		// The embedded assets are under dist/ in the ui package. fs.Sub fails
		// only for a path that is not valid, and "dist" is valid.
		distFS, _ := fs.Sub(uiassets.Assets, "dist")
		handlers = append(handlers, uihandler.Handler(distFS)...)
		d.log.Info("embedded UI enabled", slog.String("path", "/"))
	}

	return handlers
}

// traceReader reads traces from the backend at cfg.URL, the address where a
// person views the traces, which for Jaeger is also its query API. It is nil
// when no URL is set, and the trace route then answers 503.
func traceReader(
	cfg config.TracingConfig,
) telemetry.TraceReader {
	if cfg.URL == "" {
		return nil
	}

	return traces.NewJaeger(cfg.URL, nil)
}

// admissionDeps is what the admission group's handlers use. A controller
// with PKI off has no watcher, and the group must see a nil interface, not
// an interface that holds a nil pointer, to answer that admission is off.
func admissionDeps(
	d handlerDeps,
) admissionEndpoint.Deps {
	deps := admissionEndpoint.Deps{Jobs: d.jobs}
	if d.admission != nil {
		deps.Store = d.admission
	}

	return deps
}
