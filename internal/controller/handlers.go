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
	"github.com/osapi-io/osapi/internal/behavior/dispatch"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/admission"
	nodeAPI "github.com/osapi-io/osapi/internal/controller/api/node"
	certificateAPI "github.com/osapi-io/osapi/internal/controller/api/node/certificate"
	commandAPI "github.com/osapi-io/osapi/internal/controller/api/node/command"
	dockerAPI "github.com/osapi-io/osapi/internal/controller/api/node/container"
	nodeFileAPI "github.com/osapi-io/osapi/internal/controller/api/node/file"
	hostnameAPI "github.com/osapi-io/osapi/internal/controller/api/node/hostname"
	logAPI "github.com/osapi-io/osapi/internal/controller/api/node/log"
	networkAPI "github.com/osapi-io/osapi/internal/controller/api/node/network"
	ntpAPI "github.com/osapi-io/osapi/internal/controller/api/node/ntp"
	packageAPI "github.com/osapi-io/osapi/internal/controller/api/node/package"
	powerAPI "github.com/osapi-io/osapi/internal/controller/api/node/power"
	processAPI "github.com/osapi-io/osapi/internal/controller/api/node/process"
	scheduleAPI "github.com/osapi-io/osapi/internal/controller/api/node/schedule"
	serviceAPI "github.com/osapi-io/osapi/internal/controller/api/node/service"
	sysctlAPI "github.com/osapi-io/osapi/internal/controller/api/node/sysctl"
	timezoneAPI "github.com/osapi-io/osapi/internal/controller/api/node/timezone"
	userAPI "github.com/osapi-io/osapi/internal/controller/api/node/user"
	uihandler "github.com/osapi-io/osapi/internal/controller/api/ui"
	endpointserver "github.com/osapi-io/osapi/internal/controller/endpoint/server"
	admissionEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/admission"
	"github.com/osapi-io/osapi/internal/controller/endpoints/agent"
	auditEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/audit"
	factsEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/facts"
	healthEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/health"
	jobEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/job"
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
	var customRoles map[string][]string
	if cfgRoles := d.cfg.Controller.API.Security.Roles; len(cfgRoles) > 0 {
		customRoles = make(map[string][]string, len(cfgRoles))
		for name, role := range cfgRoles {
			customRoles[name] = role.Permissions
		}
	}

	// The same sub-components the heartbeat reports, so the two copies in the
	// health response cannot disagree.
	healthSubs := convertSubComponents(subComponents(d.cfg))
	reporter := health.New(d.log, d.checker, startTime, d.version, d.metricsProvider, healthSubs)
	reporter.StartMetricsRefresh(context.Background())

	handlers := make([]func(e *echo.Echo), 0, 16)
	handlers = append(
		handlers,
		endpointserver.Handler(
			d.log,
			signingKey,
			customRoles,
			agent.Deps{Jobs: d.jobs},
			admissionDeps(d),
			telemetry.Deps{Traces: traceReader(d.cfg.Telemetry.Tracing)},
			auditEndpoint.Deps{Store: d.auditStore},
			jobEndpoint.Deps{Jobs: d.jobs},
			healthEndpoint.Deps{Health: reporter},
			fileDeps(d.nats, d.objBucket, d.fileStateBucket),
			factsEndpoint.Deps{},
			// Each behavior's endpoint group, which the gen packages that
			// internal/behaviors/all/gen imports register, served under
			// /api/next beside the hand-written routes.
			dispatch.Deps{JobClient: d.jobs},
		)...,
	)
	handlers = append(handlers, nodeAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, hostnameAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, dockerAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, scheduleAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, sysctlAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, ntpAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, powerAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, timezoneAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, networkAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, commandAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, processAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, userAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, packageAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, certificateAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, serviceAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, logAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, nodeFileAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)

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
