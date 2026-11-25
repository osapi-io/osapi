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
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/osapi-io/osapi/internal/behavior"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/gohai/pkg/gohai"
	"github.com/osapi-io/gohai/pkg/gohai/family"
	"github.com/osapi-io/gohai/pkg/gohai/source"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/exec"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/telemetry/process"
	"github.com/osapi-io/osapi/internal/validation"
	"github.com/osapi-io/osapi/pkg/sdk/types"
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
	e.fileStateBucket, e.objects = openFiles(ctx, e, conn.NATS, namespace, cfg.NATS)
	if e.fileStateBucket != "" {
		e.fileState = conn.NATS
	}

	hostProvider := e.host()
	diskProvider := e.disk()
	memoryProvider := e.memory()
	loadProvider := e.load()

	// A behavior's provider reads the agent's facts through h.Facts. The
	// registry is made before the agent, thus facts reads the agent when it
	// is set below.
	facts := &agentFacts{}
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
		buildRegistry(e, cfg, facts.get),
		conn.RegistryBucket,
		conn.FactsBucket,
		execManager,
		conn.NATS,
	)
	facts.agent = a
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
	ObjectReader
	ObjectStoreStatus(ctx context.Context, bucket string) (jetstream.ObjectStoreStatus, error)
	KVStatus(ctx context.Context, bucket string) (jetstream.KeyValueStatus, error)
	KVGet(ctx context.Context, bucket string, key string) (jetstream.KeyValueEntry, error)
	KVPut(ctx context.Context, bucket string, key string, value []byte) (uint64, error)
}

// openFiles returns the file-state bucket and the objects that behaviors
// read. Each is empty if cfg names no Object Store or file-state bucket, or
// if either does not open. The agent runs without the file operations then.
func openFiles(
	ctx context.Context,
	e hostEnv,
	nc fileStores,
	namespace string,
	cfg config.NATS,
) (string, host.Objects) {
	if cfg.Objects.Bucket == "" || cfg.FileState.Bucket == "" {
		return "", nil
	}

	objBucket := job.ApplyNamespaceToInfraName(namespace, cfg.Objects.Bucket)
	if _, err := nc.ObjectStoreStatus(ctx, objBucket); err != nil {
		e.log.Warn(
			"Object Store not available, file operations disabled",
			slog.String("bucket", objBucket),
			slog.String("error", err.Error()),
		)
		return "", nil
	}

	// The agent opens the buckets and does not make them: its NATS user
	// cannot. osapi nats setup makes them.
	fileState := job.ApplyNamespaceToInfraName(namespace, cfg.FileState.Bucket)
	if _, err := nc.KVStatus(ctx, fileState); err != nil {
		e.log.Warn(
			"file state KV not available, file operations disabled",
			slog.String("bucket", fileState),
			slog.String("error", err.Error()),
		)
		return "", nil
	}

	return fileState, storeObjects{store: nc, bucket: objBucket}
}

// ObjectReader reads the objects of the Object Store bucket. A
// *natsclient.Client is one.
type ObjectReader interface {
	ObjectGetBytes(ctx context.Context, bucket string, name string) ([]byte, error)
	ObjectInfo(ctx context.Context, bucket string, name string) (*jetstream.ObjectInfo, error)
}

// storeObjects is the object store of the fleet as behaviors read it.
type storeObjects struct {
	store  ObjectReader
	bucket string
}

// Object is the content of the object name and its content type, from the
// Osapi-Content-Type header that the upload gave it: raw when it has none.
func (o storeObjects) Object(
	ctx context.Context,
	name string,
) ([]byte, types.ContentType, error) {
	content, err := o.store.ObjectGetBytes(ctx, o.bucket, name)
	if errors.Is(err, jetstream.ErrObjectNotFound) {
		return nil, "", fmt.Errorf("object %q: %w", name, behavior.ErrNotFound)
	}
	if err != nil {
		return nil, "", fmt.Errorf("object %q: %w", name, err)
	}
	contentType := types.ContentRaw
	if info, err := o.store.ObjectInfo(ctx, o.bucket, name); err == nil && info.Headers != nil {
		if ct := info.Headers.Get("Osapi-Content-Type"); ct != "" {
			contentType = types.ContentType(ct)
		}
	}

	return content, contentType, nil
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
		"agent.pki":       {Status: "ok"},
		"agent.metrics": {
			Status:  enabledOrDisabled(cfg.Agent.Metrics.Enabled),
			Address: cfg.Agent.Metrics.Address(),
			URL:     cfg.Agent.Metrics.URL,
		},
	}
}

// buildRegistry makes the processor of each job category, with the providers
// of e, and registers each one.
func buildRegistry(
	e hostEnv,
	cfg config.Config,
	facts func() map[string]any,
) *ProviderRegistry {
	r := NewProviderRegistry()

	r.Register("node", NewNodeProcessor(e.log))

	// Behaviors run beside the categories above.
	// Each builds its own provider for this platform from what it declares.
	// A behavior's provider reads the agent's facts as h.Facts.
	behaviorProviders := BuildBehaviorProviders(host.Host{
		Platform:  behavior.PlatformName(e.platform),
		Container: e.container,
		Hostname:  e.hostname,
		Logger:    e.log,
		FS:        e.fs,
		Exec:      e.exec,
		// A method that runs a caller's command, such as docker exec, reads
		// this, as the command processor does.
		CommandEnabled:  cfg.Agent.Command.Enabled,
		FileState:       e.fileState,
		FileStateBucket: e.fileStateBucket,
		Objects:         e.objects,
		Facts:           facts,
		Now:             time.Now,
	}, e.log)
	r.Register(job.CategoryBehavior, NewBehaviorProcessor(behaviorProviders, e.log))

	return r
}
