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

package agent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/gohai/pkg/gohai"
	"github.com/osapi-io/gohai/pkg/gohai/family"
	"github.com/osapi-io/gohai/pkg/gohai/source"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/exec"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/provider/command"
	fileProv "github.com/osapi-io/osapi/internal/provider/file"
	"github.com/osapi-io/osapi/internal/provider/node/disk"
	nodeHost "github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
	"github.com/osapi-io/osapi/internal/telemetry/process"
	"github.com/osapi-io/osapi/internal/validation"
)

// detectFn and containerFn tell the platform of the host (injectable for
// testing).
var (
	detectFn    = family.Detect
	containerFn = func() bool { return inContainer(source.Sources{}) }
)

// inContainer reports whether the agent runs in a container, from gohai's
// virtualization collector: docker, podman, lxc, nspawn and the others of
// Ohai's cascade. A zero src is the real host.
func inContainer(
	src source.Sources,
) bool {
	facts, err := gohaiFacts(gohai.WithCollectors("virtualization"), gohai.WithSources(src))

	return err == nil && facts.Virtualization.InContainer()
}

// gohaiFacts collects the facts that opts select.
func gohaiFacts(
	opts ...gohai.Option,
) (*gohai.Facts, error) {
	g, err := gohai.New(opts...)
	if err != nil {
		return nil, err
	}

	return g.Collect(context.Background())
}

// Setup connects to NATS as cfg.Agent.NATS tells, and makes the provider of
// each operation for this host. The metrics of each call to NATS go to mp,
// which can be nil. It returns the agent and its connection. The
// caller closes the connection after the agent stops.
func Setup(
	ctx context.Context,
	log *slog.Logger,
	fs avfs.VFS,
	cfg config.Config,
	mp *sdkmetric.MeterProvider,
) (*Agent, *client.Connection, error) {
	namespace := cfg.Agent.NATS.Namespace
	timeout, _ := time.ParseDuration(cfg.Controller.API.JobTimeout)

	conn, err := client.Connect(
		ctx,
		log,
		cfg.Agent.NATS,
		cfg.NATS,
		timeout,
		validation.ResolveTarget,
		mp,
	)
	if err != nil {
		return nil, nil, err
	}

	// The agent makes the Object Store bucket, so that it can open it below
	// when the controller has not started yet.
	if cfg.NATS.Objects.Bucket != "" {
		objCfg := bus.ObjectStore(namespace, cfg.NATS.Objects)
		if err := conn.NATS.ObjectStoreCreateOrUpdate(ctx, objCfg); err != nil {
			log.Warn("Object Store not available, file endpoints disabled",
				slog.String("bucket", objCfg.Bucket),
				slog.String("error", err.Error()))
		}
	}

	plat := detectFn()
	if plat == "darwin" {
		log.Info("running on darwin")
	}

	execManager := exec.New(log, cfg.Agent.PrivilegeEscalation.Enabled)
	hostname, _ := job.GetAgentHostname(cfg.Agent.Hostname)
	e := hostEnv{
		log:       log,
		fs:        fs,
		exec:      execManager,
		platform:  plat,
		container: containerFn(),
		hostname:  hostname,
	}
	e.files, e.fileStateBucket = openFiles(ctx, e, conn.NATS, namespace, cfg.NATS)
	if e.fileStateBucket != "" {
		e.fileState = conn.NATS
	}

	hostProvider := e.host()
	diskProvider := e.disk()
	memoryProvider := e.memory()
	loadProvider := e.load()

	a := New(
		fs,
		cfg,
		log,
		conn.Jobs,
		job.ApplyNamespaceToInfraName(namespace, cfg.NATS.Stream.Name),
		hostProvider,
		diskProvider,
		memoryProvider,
		loadProvider,
		process.New(),
		buildRegistry(
			ctx,
			e,
			cfg,
			hostProvider,
			diskProvider,
			memoryProvider,
			loadProvider,
		),
		conn.RegistryBucket,
		conn.FactsBucket,
		execManager,
		conn.NATS,
	)
	a.SetSubComponents(subComponents(cfg))

	return a, conn, nil
}

// SubComponents returns the parts of the agent that its heartbeat and its
// metrics server report.
func (a *Agent) SubComponents() map[string]job.SubComponentInfo {
	return a.subComponents
}

// fileStores is the access to NATS that openFiles uses. A *natsclient.Client
// is one.
type fileStores interface {
	fileProv.Objects
	TemplateStore
	ObjectStoreStatus(ctx context.Context, bucket string) (jetstream.ObjectStoreStatus, error)
	KVCreateOrUpdate(ctx context.Context, config jetstream.KeyValueConfig) error
	KVGet(ctx context.Context, bucket string, key string) (jetstream.KeyValueEntry, error)
	KVPut(ctx context.Context, bucket string, key string, value []byte) (uint64, error)
}

// openFiles makes the file provider, and returns it with the file-state
// bucket. Both are nil if cfg names no Object Store or file-state bucket, or
// if either does not open. The agent runs without the file operations then.
func openFiles(
	ctx context.Context,
	e hostEnv,
	nc fileStores,
	namespace string,
	cfg config.NATS,
) (fileProv.Provider, string) {
	if cfg.Objects.Bucket == "" || cfg.FileState.Bucket == "" {
		return nil, ""
	}

	objBucket := job.ApplyNamespaceToInfraName(namespace, cfg.Objects.Bucket)
	if _, err := nc.ObjectStoreStatus(ctx, objBucket); err != nil {
		e.log.Warn(
			"Object Store not available, file operations disabled",
			slog.String("bucket", objBucket),
			slog.String("error", err.Error()),
		)
		return nil, ""
	}

	fileStateCfg := bus.FileStateKV(namespace, cfg.FileState)
	if err := nc.KVCreateOrUpdate(ctx, fileStateCfg); err != nil {
		e.log.Warn(
			"file state KV not available, file operations disabled",
			slog.String("bucket", fileStateCfg.Bucket),
			slog.String("error", err.Error()),
		)
		return nil, ""
	}

	if err := SeedSystemTemplates(ctx, e.log, nc, objBucket); err != nil {
		e.log.Warn(
			"failed to seed system templates",
			slog.String("error", err.Error()),
		)
	}

	return fileProv.New(
		e.log, e.fs, nc, objBucket, nc, fileStateCfg.Bucket, e.hostname, e.exec,
	), fileStateCfg.Bucket
}

// subComponents returns the parts of the agent that its heartbeat reports.
func subComponents(
	cfg config.Config,
) map[string]job.SubComponentInfo {
	enabledOrDisabled := func(enabled bool) string {
		if enabled {
			return "ok"
		}

		return "disabled"
	}

	return map[string]job.SubComponentInfo{
		"agent.facts":     {Status: "ok"},
		"agent.heartbeat": {Status: "ok"},
		"agent.pki":       {Status: enabledOrDisabled(cfg.Agent.PKI.Enabled)},
		"agent.metrics": {
			Status: enabledOrDisabled(cfg.Agent.Metrics.Enabled),
			Address: fmt.Sprintf(
				"http://%s:%d",
				cfg.Agent.Metrics.Host,
				cfg.Agent.Metrics.Port,
			),
			URL: cfg.Agent.Metrics.URL,
		},
	}
}

// buildRegistry makes the processor of each job category, with the providers
// of e, and registers each one.
func buildRegistry(
	ctx context.Context,
	e hostEnv,
	cfg config.Config,
	hostProvider nodeHost.Provider,
	diskProvider disk.Provider,
	memoryProvider memory.Provider,
	loadProvider load.Provider,
) *ProviderRegistry {
	ntpProvider := e.ntp()
	powerProvider := e.power()
	processProvider := e.process()
	userProvider := e.user()
	packageProvider := e.packages()
	logProvider := e.logs()
	serviceProvider := e.service()
	dnsProvider := e.dns()
	pingProvider := e.ping()
	interfaceProvider, routeProvider := e.netplan()
	commandProvider := command.New(e.log, e.exec)
	dockerProvider := e.docker(ctx)
	cronProvider := e.cron()
	certificateProvider := e.certificate()

	r := NewProviderRegistry()

	r.Register(
		"node",
		NewNodeProcessor(
			ntpProvider,
			powerProvider,
			processProvider,
			userProvider,
			packageProvider,
			logProvider,
			serviceProvider,
			e.log,
		),
		hostProvider,
		diskProvider,
		memoryProvider,
		loadProvider,
		ntpProvider,
		powerProvider,
		processProvider,
		userProvider,
		packageProvider,
		logProvider,
		serviceProvider,
	)

	r.Register(
		"network",
		NewNetworkProcessor(dnsProvider, pingProvider, interfaceProvider, routeProvider, e.log),
		dnsProvider, pingProvider, interfaceProvider, routeProvider,
	)

	r.Register(
		"command",
		NewCommandProcessor(commandProvider, cfg.Agent.Command.Enabled, e.log),
		commandProvider,
	)

	r.Register("file", NewFileProcessor(e.files, e.log), e.files)

	r.Register(
		"docker",
		NewDockerProcessor(dockerProvider, cfg.Agent.Command.Enabled, e.log),
		dockerProvider,
	)

	r.Register("schedule", NewScheduleProcessor(cronProvider, e.log), cronProvider)

	r.Register(
		"certificate",
		NewCertificateProcessor(certificateProvider, e.log),
		certificateProvider,
	)

	// Behaviors run beside the categories above.
	// Each builds its own provider for this platform from what it declares.
	behaviorProviders := BuildBehaviorProviders(host.Host{
		Platform:        e.platform,
		Container:       e.container,
		Hostname:        e.hostname,
		Logger:          e.log,
		FS:              e.fs,
		Exec:            e.exec,
		FileState:       e.fileState,
		FileStateBucket: e.fileStateBucket,
		Now:             time.Now,
	}, e.log)
	r.Register(
		job.CategoryBehavior,
		NewBehaviorProcessor(behaviorProviders, e.log),
		behaviorProviders.All()...,
	)

	return r
}
