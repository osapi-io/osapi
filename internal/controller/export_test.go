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
	"os"

	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/osapi-io/osapi/internal/controller/admission"

	"github.com/osapi-io/osapi/internal/templates"

	"github.com/osapi-io/osapi/internal/controller/api"

	"github.com/avfs/avfs"

	pkiEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/pki"
	secretEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/secret"
	"github.com/osapi-io/osapi/internal/secret/vault"

	fileEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/file"

	"github.com/osapi-io/osapi/internal/controller/endpoints/telemetry"

	"github.com/labstack/echo/v5"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/config"
	endpointserver "github.com/osapi-io/osapi/internal/controller/endpoint/server"
	"github.com/osapi-io/osapi/internal/controller/health"
	"github.com/osapi-io/osapi/internal/controller/notify"
	"github.com/osapi-io/osapi/internal/job"
	jobclient "github.com/osapi-io/osapi/internal/job/client"
)

// ExportNewMetricsProvider calls newMetricsProvider.
func ExportNewMetricsProvider(
	nc natsStatus,
	streamName string,
	kvBuckets []string,
	objBuckets []string,
	jc jobclient.JobClient,
	registryBucket string,
) *health.ClosureMetricsProvider {
	return newMetricsProvider(nc, streamName, kvBuckets, objBuckets, jc, registryBucket)
}

// ExportNewHealthChecker calls newHealthChecker.
func ExportNewHealthChecker(
	nc natsStatus,
	jobsBucket string,
) *health.NATSChecker {
	return newHealthChecker(nc, jobsBucket)
}

// ExportConfiguredKVBuckets calls configuredKVBuckets.
func ExportConfiguredKVBuckets(
	namespace string,
	cfg config.NATS,
) []string {
	return configuredKVBuckets(namespace, cfg)
}

// ExportConfiguredObjectBuckets calls configuredObjectBuckets.
func ExportConfiguredObjectBuckets(
	namespace string,
	cfg config.NATS,
) []string {
	return configuredObjectBuckets(namespace, cfg)
}

// ExportNewAuditStore calls newAuditStore.
func ExportNewAuditStore(
	ctx context.Context,
	log *slog.Logger,
	nc auditStreams,
	namespace string,
	cfg config.NATSAudit,
) (*audit.StreamStore, error) {
	return newAuditStore(ctx, log, nc, namespace, cfg)
}

// ExportOpenFileBuckets calls openFileBuckets.
func ExportOpenFileBuckets(
	ctx context.Context,
	log *slog.Logger,
	nc fileBuckets,
	namespace string,
	cfg config.NATS,
) (string, string) {
	return openFileBuckets(ctx, log, nc, namespace, cfg)
}

// ExportFileDeps calls fileDeps.
func ExportFileDeps(
	nc fileBuckets,
	objBucket string,
	stateBucket string,
) fileEndpoint.Deps {
	return fileDeps(nc, objBucket, stateBucket)
}

// ExportSecretDeps calls secretDeps with the vaults.
func ExportSecretDeps(
	vaults *vault.Set,
) secretEndpoint.Deps {
	return secretDeps(handlerDeps{vaults: vaults})
}

// ExportPKIDeps calls pkiDeps with cfg.
func ExportPKIDeps(
	cfg config.Config,
) pkiEndpoint.Deps {
	return pkiDeps(handlerDeps{cfg: cfg})
}

// ExportMakeVaultKeys calls makeVaultKeys.
func ExportMakeVaultKeys(
	ctx context.Context,
	log *slog.Logger,
	vfs avfs.VFS,
	specs []vault.Spec,
) error {
	return makeVaultKeys(ctx, log, vfs, specs)
}

// ExportHandlers calls handlers with the parts that a test sets.
func ExportHandlers(
	log *slog.Logger,
	cfg config.Config,
	auditStore audit.Store,
	nats fileBuckets,
	objBucket string,
) []func(e *echo.Echo) {
	return handlers(handlerDeps{
		log:        log,
		cfg:        cfg,
		auditStore: auditStore,
		nats:       nats,
		objBucket:  objBucket,
	})
}

// ExportSecretsDeps calls secretsDeps with vaults.
func ExportSecretsDeps(
	vaults *vault.Set,
) endpointserver.Secrets {
	return secretsDeps(handlerDeps{vaults: vaults})
}

// ExportConvertSubComponents calls convertSubComponents.
func ExportConvertSubComponents(
	src map[string]job.SubComponentInfo,
) map[string]health.SubComponentInfo {
	return convertSubComponents(src)
}

// ExportSubComponents calls subComponents.
func ExportSubComponents(
	cfg config.Config,
) map[string]job.SubComponentInfo {
	return subComponents(cfg)
}

// ExportStartConditionWatcher calls startConditionWatcher.
func ExportStartConditionWatcher(
	ctx context.Context,
	log *slog.Logger,
	registry notify.Registry,
	registryBucket string,
	cfg config.NotificationsConfig,
) {
	startConditionWatcher(ctx, log, registry, registryBucket, cfg)
}

// ExportBackground calls background.
func ExportBackground(
	ctx context.Context,
	log *slog.Logger,
	name string,
	fn func(context.Context) error,
) {
	background(ctx, log, name, fn)
}

// SetHostnameFn replaces the function that returns the name of the host.
func SetHostnameFn(
	fn func() (string, error),
) {
	hostnameFn = fn
}

// ResetHostnameFn puts back the function that returns the name of the host.
func ResetHostnameFn() {
	hostnameFn = os.Hostname
}

// ExportRegisterTargetValidator calls registerTargetValidator.
func ExportRegisterTargetValidator(
	jobs jobclient.JobClient,
) {
	registerTargetValidator(jobs)
}

// ExportTraceReader calls traceReader.
func ExportTraceReader(
	log *slog.Logger,
	cfg config.TracingConfig,
) telemetry.TraceReader {
	return traceReader(log, cfg)
}

// ExportTLSOptions calls tlsOptions.
func ExportTLSOptions(
	cfg config.Config,
) ([]api.Option, error) {
	return tlsOptions(cfg)
}

// SetSeedTemplates swaps the template seeder, and returns a func that puts it
// back.
func SetSeedTemplates(
	fn func(context.Context, *slog.Logger, templates.TemplateStore, string) error,
) func() {
	old := seedTemplates
	seedTemplates = fn

	return func() { seedTemplates = old }
}

// ExportJoinAgentUser calls joinAgentUser.
func ExportJoinAgentUser(
	cfg config.Config,
	namespace string,
) (admission.AgentUser, admission.Mint, error) {
	return joinAgentUser(cfg, namespace)
}

// ExportCalloutService calls calloutService, with no watcher decisions.
func ExportCalloutService(
	cfg config.Config,
) error {
	_, err := calloutService(cfg, "osapi", nil, nil, slog.Default())

	return err
}

// ExportCalloutConnect calls calloutConnect.
func ExportCalloutConnect(
	cfg config.Config,
) error {
	c, err := calloutConnect(slog.Default(), cfg)
	if err == nil {
		c.NC.Close()
	}

	return err
}

// SetCalloutConnect swaps the connection of the auth callout.
func SetCalloutConnect(
	fn func(*slog.Logger, config.Config) (*natsclient.Client, error),
) func() {
	was := calloutConnect
	calloutConnect = fn

	return func() { calloutConnect = was }
}

// ExportRevocationConnect calls revocationConnect.
func ExportRevocationConnect(
	cfg config.Config,
) error {
	c, err := revocationConnect(slog.Default(), cfg)
	if err == nil {
		c.NC.Close()
	}

	return err
}

// SetRevocationConnect swaps the connection of the NATS user revocation.
func SetRevocationConnect(
	fn func(*slog.Logger, config.Config) (*natsclient.Client, error),
) func() {
	was := revocationConnect
	revocationConnect = fn

	return func() { revocationConnect = was }
}
