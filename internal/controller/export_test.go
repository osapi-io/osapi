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

	"github.com/labstack/echo/v5"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api/file"
	"github.com/osapi-io/osapi/internal/controller/api/health"
	"github.com/osapi-io/osapi/internal/job"
	jobclient "github.com/osapi-io/osapi/internal/job/client"
)

// ExportNewMetricsProvider calls newMetricsProvider.
func ExportNewMetricsProvider(
	nc interface {
		ConnectedURL() string
		ConnectedServerVersion() string
		GetStreamInfo(ctx context.Context, streamName string) (*jetstream.StreamInfo, error)
		KeyValue(ctx context.Context, bucket string) (jetstream.KeyValue, error)
		ObjectStore(ctx context.Context, name string) (jetstream.ObjectStore, error)
	},
	streamName string,
	kvBuckets []string,
	objBuckets []string,
	jc jobclient.JobClient,
	registryKV jetstream.KeyValue,
) *health.ClosureMetricsProvider {
	return newMetricsProvider(nc, streamName, kvBuckets, objBuckets, jc, registryKV)
}

// ExportNewHealthChecker calls newHealthChecker.
func ExportNewHealthChecker(
	nc interface{ ConnectedURL() string },
	jobsKV jetstream.KeyValue,
) *health.NATSChecker {
	return newHealthChecker(nc, jobsKV)
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
) (file.ObjectStoreManager, jetstream.KeyValue) {
	return openFileBuckets(ctx, log, nc, namespace, cfg)
}

// ExportHandlers calls handlers with the parts that a test sets.
func ExportHandlers(
	log *slog.Logger,
	cfg config.Config,
	auditStore audit.Store,
	objStore file.ObjectStoreManager,
) []func(e *echo.Echo) {
	return handlers(handlerDeps{
		log:        log,
		cfg:        cfg,
		auditStore: auditStore,
		objStore:   objStore,
	})
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
	registryKV jetstream.KeyValue,
	cfg config.NotificationsConfig,
) {
	startConditionWatcher(ctx, log, registryKV, cfg)
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
