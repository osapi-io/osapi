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
	"io/fs"
	"log/slog"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/behavior/server"
	"github.com/osapi-io/osapi/internal/config"
	agentAPI "github.com/osapi-io/osapi/internal/controller/api/agent"
	auditAPI "github.com/osapi-io/osapi/internal/controller/api/audit"
	factsAPI "github.com/osapi-io/osapi/internal/controller/api/facts"
	"github.com/osapi-io/osapi/internal/controller/api/file"
	"github.com/osapi-io/osapi/internal/controller/api/health"
	jobAPI "github.com/osapi-io/osapi/internal/controller/api/job"
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
	"github.com/osapi-io/osapi/internal/controller/enrollment"
	jobclient "github.com/osapi-io/osapi/internal/job/client"
	uiassets "github.com/osapi-io/osapi/ui"
)

// handlerDeps is what the REST API handlers use.
type handlerDeps struct {
	log             *slog.Logger
	cfg             config.Config
	jobs            jobclient.JobClient
	checker         health.Checker
	metricsProvider health.MetricsProvider
	auditStore      audit.Store
	objStore        file.ObjectStoreManager
	fileStateKV     jetstream.KeyValue
	enrollment      *enrollment.Watcher
}

// handlers returns each REST API handler. The audit, file and UI handlers are
// only there when their parts are.
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

	healthSubs := map[string]health.SubComponentInfo{
		"controller.api": {
			Status:  "ok",
			Address: httpAddr("0.0.0.0", d.cfg.Controller.API.Port),
		},
		"controller.heartbeat": {Status: "ok"},
		"controller.metrics": {
			Status:  enabledOrDisabled(d.cfg.Controller.Metrics.Enabled),
			Address: httpAddr(d.cfg.Controller.Metrics.Host, d.cfg.Controller.Metrics.Port),
		},
		"controller.notifier": {
			Status: enabledOrDisabled(d.cfg.Controller.Notifications.Enabled),
		},
		"controller.tracing": {Status: enabledOrDisabled(d.cfg.Telemetry.Tracing.Enabled)},
	}

	handlers := make([]func(e *echo.Echo), 0, 16)
	handlers = append(
		handlers,
		agentAPI.Handler(d.log, d.jobs, signingKey, customRoles, d.enrollment)...,
	)
	handlers = append(handlers, nodeAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, jobAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(
		handlers,
		health.Handler(
			d.log,
			d.checker,
			startTime,
			"0.1.0",
			d.metricsProvider,
			healthSubs,
			signingKey,
			customRoles,
		)...,
	)
	handlers = append(handlers, hostnameAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, dockerAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, scheduleAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	handlers = append(handlers, sysctlAPI.Handler(d.log, d.jobs, signingKey, customRoles)...)
	// Behaviors: each one's generated routes, registered by the gen packages
	// internal/behaviors/all/gen imports, served under /api/next beside the
	// hand-written routes.
	handlers = append(handlers, server.Handler(d.log, d.jobs, signingKey, customRoles)...)
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
	if d.auditStore != nil {
		handlers = append(
			handlers,
			auditAPI.Handler(d.log, d.auditStore, signingKey, customRoles)...)
	}
	if d.objStore != nil {
		handlers = append(
			handlers,
			file.Handler(d.log, d.objStore, d.fileStateKV, signingKey, customRoles)...,
		)
	}
	handlers = append(handlers, factsAPI.Handler(d.log, signingKey, customRoles)...)

	if d.cfg.Controller.UI.UIEnabled() {
		// The embedded assets are under dist/ in the ui package. fs.Sub fails
		// only for a path that is not valid, and "dist" is valid.
		distFS, _ := fs.Sub(uiassets.Assets, "dist")
		handlers = append(handlers, uihandler.Handler(distFS)...)
		d.log.Info("embedded UI enabled", slog.String("path", "/"))
	}

	return handlers
}
