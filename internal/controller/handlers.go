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
	"net/http"
	"os"
	"time"

	"github.com/avfs/avfs/vfs/osfs"
	"github.com/labstack/echo/v5"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior/dispatch"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/admission"
	uihandler "github.com/osapi-io/osapi/internal/controller/api/ui"
	"github.com/osapi-io/osapi/internal/controller/certrotation"
	endpointserver "github.com/osapi-io/osapi/internal/controller/endpoint/server"
	admissionEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/admission"
	"github.com/osapi-io/osapi/internal/controller/endpoints/agent"
	auditEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/audit"
	healthEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/health"
	jobEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/job"
	pkiEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/pki"
	secretEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/secret"
	"github.com/osapi-io/osapi/internal/controller/endpoints/telemetry"
	"github.com/osapi-io/osapi/internal/controller/health"
	"github.com/osapi-io/osapi/internal/controller/keyrotation"
	jobclient "github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/secret/vault"
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
	keyRotation     *keyrotation.Rotation
	certRotation    *certrotation.Rotation
	vaults          *vault.Set
	// version is the build version of the binary.
	version string
	// health answers the health endpoints. New makes it and starts its
	// metrics refresh, which Stop ends; with none, handlers makes one with
	// no refresh.
	health *health.Health
}

// handlers returns each REST API handler. The UI handler is only there
// when it is enabled.
func handlers(
	d handlerDeps,
) []func(e *echo.Echo) {
	startTime := time.Now()
	signingKeys := d.cfg.Controller.API.Security.Keys()

	// Build custom roles map from config.
	var customRoles authtoken.CustomRoles
	if cfgRoles := d.cfg.Controller.API.Security.Roles; len(cfgRoles) > 0 {
		customRoles = make(authtoken.CustomRoles, len(cfgRoles))
		for name, role := range cfgRoles {
			customRoles[name] = role.Permissions
		}
	}

	reporter := d.health
	if reporter == nil {
		reporter = health.New(d.log, d.checker, startTime, d.version, d.metricsProvider)
	}

	handlers := make([]func(e *echo.Echo), 0, 16)
	handlers = append(
		handlers,
		endpointserver.Handler(
			d.log,
			signingKeys,
			customRoles,
			agentDeps(d),
			admissionDeps(d),
			telemetry.Deps{Traces: traceReader(d.log, d.cfg.Telemetry.Tracing)},
			auditEndpoint.Deps{Store: d.auditStore},
			jobEndpoint.Deps{Jobs: d.jobs},
			healthEndpoint.Deps{Health: reporter},
			fileDeps(d.nats, d.objBucket, d.fileStateBucket),
			secretsDeps(d),
			secretDeps(d),
			pkiDeps(d),
			// Each behavior's endpoint group, which the gen packages that
			// internal/behaviors/all/gen imports register.
			dispatch.Deps{JobClient: d.jobs},
		)...,
	)

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
// person views the traces, which for Jaeger is also its query API, with the
// TLS of cfg.TLS. It is nil when no URL is set, or when its TLS does not
// load, which it logs; the trace route then answers 503.
func traceReader(
	log *slog.Logger,
	cfg config.TracingConfig,
) telemetry.TraceReader {
	if cfg.URL == "" {
		return nil
	}
	if cfg.TLS == (config.ClientTLS{}) {
		return traces.NewJaeger(cfg.URL, nil)
	}
	tlsConfig, err := cfg.TLS.Config()
	if err != nil {
		log.Error("read traces: telemetry.tracing.tls", slog.String("error", err.Error()))

		return nil
	}
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = tlsConfig

	return traces.NewJaeger(cfg.URL, &http.Client{Timeout: 10 * time.Second, Transport: t})
}

// agentDeps is what the agent group uses: the job client, and the
// admission watcher as the joiner. A controller with no watcher gives no
// joiner, not an interface that holds a nil pointer.
func agentDeps(
	d handlerDeps,
) agent.Deps {
	deps := agent.Deps{Jobs: d.jobs}
	if d.admission != nil {
		deps.Joiner = d.admission
		deps.Admissions = d.admission
	}

	return deps
}

// pkiDeps is what the pki group uses: the job client, whose agents
// report their certificates, and the certificates of the controller and
// the NATS server on this host, read from its files at each request.
func pkiDeps(
	d handlerDeps,
) pkiEndpoint.Deps {
	host, _ := os.Hostname()
	fsys := osfs.NewWithNoIdm()
	srcs := config.CertSourcesOf(d.cfg, "controller", "nats")

	return pkiEndpoint.Deps{
		Agents:       d.jobs,
		Hostname:     host,
		Local:        func() []pki.Cert { return pki.ReadCerts(fsys, srcs, time.Now()) },
		Rotation:     d.keyRotation,
		CertRotation: d.certRotation,
	}
}

// admissionDeps is what the admission group's handlers use: the watcher of
// the enrollment, which keeps the records, and the job client.
func admissionDeps(
	d handlerDeps,
) admissionEndpoint.Deps {
	return admissionEndpoint.Deps{Store: d.admission, Jobs: d.jobs}
}

// secretsDeps is what each route resolves a reference with. A controller
// with no vault set gives no vaults, not an interface that holds a nil
// pointer, so a route answers that the vault does not exist.
func secretsDeps(
	d handlerDeps,
) endpointserver.Secrets {
	var deps endpointserver.Secrets
	if d.vaults != nil {
		deps.Vaults = d.vaults
	}

	return deps
}

// secretDeps is what the secret group's handlers use. A controller with no
// vault set gives a nil interface, not an interface that holds a nil
// pointer, so the group answers that the vault is not there.
func secretDeps(
	d handlerDeps,
) secretEndpoint.Deps {
	var deps secretEndpoint.Deps
	if d.vaults != nil {
		deps.Vaults = d.vaults
	}

	return deps
}
