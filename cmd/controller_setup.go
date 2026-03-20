// Copyright (c) 2024 John Dewey

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

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/retr0h/osapi/internal/controller"
	"github.com/retr0h/osapi/internal/controller/api"
	"github.com/retr0h/osapi/internal/controller/api/file"
	"github.com/retr0h/osapi/internal/controller/api/health"
	"github.com/retr0h/osapi/internal/audit"
	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	jobclient "github.com/retr0h/osapi/internal/job/client"
	"github.com/retr0h/osapi/internal/messaging"
	"github.com/retr0h/osapi/internal/controller/notify"
	"github.com/retr0h/osapi/internal/provider/process"
	"github.com/retr0h/osapi/internal/validation"
)

// ServerManager responsible for Server operations.
type ServerManager interface {
	cli.Lifecycle
	// GetAgentHandler returns agent handler for registration.
	GetAgentHandler(jobClient jobclient.JobClient) []func(e *echo.Echo)
	// GetNodeHandler returns node handler for registration.
	GetNodeHandler(jobClient jobclient.JobClient) []func(e *echo.Echo)
	// GetJobHandler returns job handler for registration.
	GetJobHandler(jobClient jobclient.JobClient) []func(e *echo.Echo)
	// GetHealthHandler returns health handler for registration.
	GetHealthHandler(
		checker health.Checker,
		startTime time.Time,
		version string,
		metrics health.MetricsProvider,
	) []func(e *echo.Echo)
	// GetMetricsHandler returns Prometheus metrics handler for registration.
	GetMetricsHandler(metricsHandler http.Handler, path string) []func(e *echo.Echo)
	// GetAuditHandler returns audit handler for registration.
	GetAuditHandler(store audit.Store) []func(e *echo.Echo)
	// GetDockerHandler returns Docker handler for registration.
	GetDockerHandler(jobClient jobclient.JobClient) []func(e *echo.Echo)
	// GetFileHandler returns file handler for registration.
	GetFileHandler(objStore file.ObjectStoreManager) []func(e *echo.Echo)
	// RegisterHandlers registers a list of handlers with the Echo instance.
	RegisterHandlers(handlers []func(e *echo.Echo))
}

// natsBundle holds the NATS connection, job client, and KV handles created
// by connectNATSBundle.
type natsBundle struct {
	nc         messaging.NATSClient
	jobClient  jobclient.JobClient
	jobsKV     jetstream.KeyValue
	registryKV jetstream.KeyValue
	factsKV    jetstream.KeyValue
	objStore   file.ObjectStoreManager
}

// setupController connects to NATS, creates the API server with all handlers,
// and returns the server manager and NATS bundle. It is used by the standalone
// API server start and combined start commands.
func setupController(
	ctx context.Context,
	log *slog.Logger,
	connCfg config.NATSConnection,
	metricsHandler http.Handler,
	metricsPath string,
) (ServerManager, *natsBundle) {
	namespace := connCfg.Namespace
	streamName := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.Stream.Name)
	kvBucket := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.KV.Bucket)

	b := connectNATSBundle(ctx, log, connCfg, kvBucket, namespace, streamName)

	validation.RegisterTargetValidator(
		func(ctx context.Context) ([]validation.AgentTarget, error) {
			agents, err := b.jobClient.ListAgents(ctx)
			if err != nil {
				return nil, err
			}
			targets := make([]validation.AgentTarget, 0, len(agents))
			for _, a := range agents {
				targets = append(targets, validation.AgentTarget{
					Hostname: a.Hostname,
					Labels:   a.Labels,
				})
			}
			return targets, nil
		},
	)

	checker := newHealthChecker(b.nc, b.jobsKV)
	auditStore, serverOpts := createAuditStore(ctx, log, b.nc, namespace)
	kvBuckets := configuredKVBuckets(namespace)
	objBuckets := configuredObjectBuckets(namespace)
	metricsProvider := newMetricsProvider(
		b.nc,
		streamName,
		kvBuckets,
		objBuckets,
		b.jobClient,
		b.registryKV,
	)

	sm := api.New(appConfig, log, serverOpts...)
	registerControllerHandlers(
		sm, b.jobClient, checker, metricsProvider,
		metricsHandler, metricsPath, auditStore, b.objStore,
	)

	startControllerHeartbeat(ctx, log, b.registryKV)
	startConditionWatcher(ctx, log, b.registryKV)

	return sm, b
}

func connectNATSBundle(
	ctx context.Context,
	log *slog.Logger,
	connCfg config.NATSConnection,
	kvBucket string,
	namespace string,
	streamName string,
) *natsBundle {
	var nc messaging.NATSClient = natsclient.New(log, &natsclient.Options{
		Host: connCfg.Host,
		Port: connCfg.Port,
		Auth: cli.BuildNATSAuthOptions(connCfg.Auth),
		Name: connCfg.ClientName,
	})

	if err := nc.Connect(); err != nil {
		cli.LogFatal(log, "failed to connect to NATS", err)
	}

	jobsKV, err := nc.CreateOrUpdateKVBucket(ctx, kvBucket)
	if err != nil {
		cli.LogFatal(log, "failed to create KV bucket", err)
	}

	registryKVConfig := cli.BuildRegistryKVConfig(namespace, appConfig.NATS.Registry)
	registryKV, err := nc.CreateOrUpdateKVBucketWithConfig(ctx, registryKVConfig)
	if err != nil {
		cli.LogFatal(log, "failed to create registry KV bucket", err)
	}

	var factsKV jetstream.KeyValue
	if appConfig.NATS.Facts.Bucket != "" {
		factsKVConfig := cli.BuildFactsKVConfig(namespace, appConfig.NATS.Facts)
		factsKV, err = nc.CreateOrUpdateKVBucketWithConfig(ctx, factsKVConfig)
		if err != nil {
			cli.LogFatal(log, "failed to create facts KV bucket", err)
		}
	}

	var stateKV jetstream.KeyValue
	if appConfig.NATS.State.Bucket != "" {
		stateKVConfig := cli.BuildStateKVConfig(namespace, appConfig.NATS.State)
		stateKV, err = nc.CreateOrUpdateKVBucketWithConfig(ctx, stateKVConfig)
		if err != nil {
			cli.LogFatal(log, "failed to create state KV bucket", err)
		}
	}

	// Create or update Object Store bucket for file management API
	var objStore file.ObjectStoreManager
	if appConfig.NATS.Objects.Bucket != "" {
		objStoreConfig := cli.BuildObjectStoreConfig(namespace, appConfig.NATS.Objects)
		var objErr error
		objStore, objErr = nc.CreateOrUpdateObjectStore(ctx, objStoreConfig)
		if objErr != nil {
			log.Warn("Object Store not available, file endpoints disabled",
				slog.String("bucket", objStoreConfig.Bucket),
				slog.String("error", objErr.Error()))
		}
	}

	jc, err := jobclient.New(log, nc, &jobclient.Options{
		Timeout:    30 * time.Second,
		KVBucket:   jobsKV,
		RegistryKV: registryKV,
		FactsKV:    factsKV,
		StateKV:    stateKV,
		StreamName: streamName,
	})
	if err != nil {
		cli.LogFatal(log, "failed to create job client", err)
	}

	return &natsBundle{
		nc:         nc,
		jobClient:  jc,
		jobsKV:     jobsKV,
		registryKV: registryKV,
		factsKV:    factsKV,
		objStore:   objStore,
	}
}

func newHealthChecker(
	nc messaging.NATSClient,
	jobsKV jetstream.KeyValue,
) *health.NATSChecker {
	return &health.NATSChecker{
		NATSCheck: func() error {
			natsConn, ok := nc.(*natsclient.Client)
			if !ok || natsConn.NC == nil {
				return fmt.Errorf("nats client unavailable")
			}

			if natsConn.NC.ConnectedUrl() == "" {
				return fmt.Errorf("nats not connected")
			}

			return nil
		},
		KVCheck: func() error {
			_, err := jobsKV.Status(context.Background())
			if err != nil {
				return fmt.Errorf("kv bucket not accessible: %w", err)
			}

			return nil
		},
	}
}

// configuredKVBuckets returns the namespaced names of all KV buckets
// declared in osapi.yaml. Only non-empty bucket configs are included.
func configuredKVBuckets(namespace string) []string {
	var buckets []string
	add := func(name string) {
		if name != "" {
			buckets = append(buckets, job.ApplyNamespaceToInfraName(namespace, name))
		}
	}

	add(appConfig.NATS.KV.Bucket)
	add(appConfig.NATS.KV.ResponseBucket)
	add(appConfig.NATS.Audit.Bucket)
	add(appConfig.NATS.Registry.Bucket)
	add(appConfig.NATS.Facts.Bucket)
	add(appConfig.NATS.State.Bucket)
	add(appConfig.NATS.FileState.Bucket)

	return buckets
}

// configuredObjectBuckets returns the namespaced names of all Object Store
// buckets declared in osapi.yaml.
func configuredObjectBuckets(namespace string) []string {
	var buckets []string
	if appConfig.NATS.Objects.Bucket != "" {
		buckets = append(
			buckets,
			job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.Objects.Bucket),
		)
	}

	return buckets
}

func newMetricsProvider(
	nc messaging.NATSClient,
	streamName string,
	kvBuckets []string,
	objBuckets []string,
	jc jobclient.JobClient,
	registryKV jetstream.KeyValue,
) *health.ClosureMetricsProvider {
	return &health.ClosureMetricsProvider{
		NATSInfoFn: func(_ context.Context) (*health.NATSMetrics, error) {
			natsConn, ok := nc.(*natsclient.Client)
			if !ok || natsConn.NC == nil {
				return nil, fmt.Errorf("NATS client unavailable")
			}

			metrics := &health.NATSMetrics{
				URL: natsConn.NC.ConnectedUrl(),
			}

			if wrapper, ok := natsConn.NC.(*natsclient.NATSConnWrapper); ok &&
				wrapper.Conn != nil {
				metrics.Version = wrapper.Conn.ConnectedServerVersion()
			}

			return metrics, nil
		},
		StreamInfoFn: func(fnCtx context.Context) ([]health.StreamMetrics, error) {
			info, err := nc.GetStreamInfo(fnCtx, streamName)
			if err != nil {
				return nil, fmt.Errorf("stream info: %w", err)
			}

			return []health.StreamMetrics{
				{
					Name:      streamName,
					Messages:  info.State.Msgs,
					Bytes:     info.State.Bytes,
					Consumers: info.State.Consumers,
				},
			}, nil
		},
		KVInfoFn: func(fnCtx context.Context) ([]health.KVMetrics, error) {
			natsConn, ok := nc.(*natsclient.Client)
			if !ok || natsConn.ExtJS == nil {
				return nil, fmt.Errorf("jetstream client unavailable")
			}

			results := make([]health.KVMetrics, 0, len(kvBuckets))
			for _, name := range kvBuckets {
				kv, err := natsConn.ExtJS.KeyValue(fnCtx, name)
				if err != nil {
					continue
				}
				status, err := kv.Status(fnCtx)
				if err != nil {
					continue
				}
				results = append(results, health.KVMetrics{
					Name:  status.Bucket(),
					Keys:  int(status.Values()),
					Bytes: status.Bytes(),
				})
			}

			return results, nil
		},
		ObjectStoreInfoFn: func(fnCtx context.Context) ([]health.ObjectStoreMetrics, error) {
			natsConn, ok := nc.(*natsclient.Client)
			if !ok || natsConn.ExtJS == nil {
				return nil, fmt.Errorf("jetstream client unavailable")
			}

			results := make([]health.ObjectStoreMetrics, 0, len(objBuckets))
			for _, name := range objBuckets {
				obj, err := natsConn.ExtJS.ObjectStore(fnCtx, name)
				if err != nil {
					continue
				}
				status, err := obj.Status(fnCtx)
				if err != nil {
					continue
				}
				results = append(results, health.ObjectStoreMetrics{
					Name: status.Bucket(),
					Size: status.Size(),
				})
			}

			return results, nil
		},
		ConsumerStatsFn: func(fnCtx context.Context) (*health.ConsumerMetrics, error) {
			natsConn, ok := nc.(*natsclient.Client)
			if !ok || natsConn.ExtJS == nil {
				return nil, fmt.Errorf("jetstream client unavailable")
			}

			stream, err := natsConn.ExtJS.Stream(fnCtx, streamName)
			if err != nil {
				return nil, fmt.Errorf("stream: %w", err)
			}

			var details []health.ConsumerDetail
			lister := stream.ListConsumers(fnCtx)
			for ci := range lister.Info() {
				details = append(details, health.ConsumerDetail{
					Name:        ci.Name,
					Pending:     ci.NumPending,
					AckPending:  ci.NumAckPending,
					Redelivered: ci.NumRedelivered,
				})
			}
			if lister.Err() != nil {
				return nil, fmt.Errorf("list consumers: %w", lister.Err())
			}

			return &health.ConsumerMetrics{
				Total:     len(details),
				Consumers: details,
			}, nil
		},
		JobStatsFn: func(fnCtx context.Context) (*health.JobMetrics, error) {
			stats, err := jc.GetQueueSummary(fnCtx)
			if err != nil {
				return nil, fmt.Errorf("job stats: %w", err)
			}

			return &health.JobMetrics{
				Total:       stats.TotalJobs,
				Unprocessed: stats.StatusCounts["submitted"],
				Processing:  stats.StatusCounts["processing"],
				Completed:   stats.StatusCounts["completed"],
				Failed:      stats.StatusCounts["failed"],
				DLQ:         stats.DLQCount,
			}, nil
		},
		AgentStatsFn: func(fnCtx context.Context) (*health.AgentMetrics, error) {
			agents, err := jc.ListAgents(fnCtx)
			if err != nil {
				return nil, fmt.Errorf("agent stats: %w", err)
			}

			details := make([]health.AgentDetail, 0, len(agents))
			for _, a := range agents {
				labels := ""
				for k, v := range a.Labels {
					if labels != "" {
						labels += ", "
					}
					labels += k + "=" + v
				}
				registered := cli.FormatAge(time.Since(a.RegisteredAt)) + " ago"
				details = append(details, health.AgentDetail{
					Hostname:   a.Hostname,
					Labels:     labels,
					Registered: registered,
				})
			}

			return &health.AgentMetrics{
				Total:  len(agents),
				Ready:  len(agents), // presence in KV = Ready
				Agents: details,
			}, nil
		},
		ComponentRegistryFn: func(fnCtx context.Context) ([]health.ComponentEntry, error) {
			if registryKV == nil {
				return nil, fmt.Errorf("registry KV not configured")
			}

			keys, err := registryKV.Keys(fnCtx)
			if err != nil {
				if strings.Contains(err.Error(), "no keys found") {
					return []health.ComponentEntry{}, nil
				}
				return nil, fmt.Errorf("list registry keys: %w", err)
			}

			entries := make([]health.ComponentEntry, 0, len(keys))
			for _, key := range keys {
				parts := strings.SplitN(key, ".", 2)
				if len(parts) != 2 {
					continue
				}
				prefix := parts[0]

				kvEntry, err := registryKV.Get(fnCtx, key)
				if err != nil {
					continue
				}

				var entry health.ComponentEntry

				switch prefix {
				case "agents":
					var reg job.AgentRegistration
					if err := json.Unmarshal(kvEntry.Value(), &reg); err != nil {
						continue
					}
					entry = health.ComponentEntry{
						Type:     "agent",
						Hostname: reg.Hostname,
						Status:   "Ready",
						Age:      cli.FormatAge(time.Since(reg.StartedAt)),
					}
					for _, c := range reg.Conditions {
						if c.Status {
							entry.Conditions = append(entry.Conditions, c.Type)
						}
					}
					if reg.Process != nil {
						entry.CPUPercent = reg.Process.CPUPercent
						entry.MemBytes = reg.Process.RSSBytes
					}
				case "controller", "nats":
					var reg job.ComponentRegistration
					if err := json.Unmarshal(kvEntry.Value(), &reg); err != nil {
						continue
					}
					entry = health.ComponentEntry{
						Type:     reg.Type,
						Hostname: reg.Hostname,
						Status:   "Ready",
						Age:      cli.FormatAge(time.Since(reg.StartedAt)),
					}
					for _, c := range reg.Conditions {
						if c.Status {
							entry.Conditions = append(entry.Conditions, c.Type)
						}
					}
					if reg.Process != nil {
						entry.CPUPercent = reg.Process.CPUPercent
						entry.MemBytes = reg.Process.RSSBytes
					}
				default:
					continue
				}

				entries = append(entries, entry)
			}

			sort.Slice(entries, func(i, j int) bool {
				if entries[i].Type != entries[j].Type {
					return entries[i].Type < entries[j].Type
				}
				return entries[i].Hostname < entries[j].Hostname
			})

			return entries, nil
		},
	}
}

func createAuditStore(
	ctx context.Context,
	log *slog.Logger,
	nc messaging.NATSClient,
	namespace string,
) (audit.Store, []api.Option) {
	if appConfig.NATS.Audit.Bucket == "" {
		return nil, nil
	}

	auditKVConfig := cli.BuildAuditKVConfig(namespace, appConfig.NATS.Audit)
	auditKV, err := nc.CreateOrUpdateKVBucketWithConfig(ctx, auditKVConfig)
	if err != nil {
		cli.LogFatal(log, "failed to create audit KV bucket", err)
	}

	store := audit.NewKVStore(log, auditKV)

	return store, []api.Option{api.WithAuditStore(store)}
}

func registerControllerHandlers(
	sm ServerManager,
	jc jobclient.JobClient,
	checker health.Checker,
	metricsProvider health.MetricsProvider,
	metricsHandler http.Handler,
	metricsPath string,
	auditStore audit.Store,
	objStore file.ObjectStoreManager,
) {
	startTime := time.Now()

	handlers := make([]func(e *echo.Echo), 0, 8)
	handlers = append(handlers, sm.GetAgentHandler(jc)...)
	handlers = append(handlers, sm.GetNodeHandler(jc)...)
	handlers = append(handlers, sm.GetJobHandler(jc)...)
	handlers = append(
		handlers,
		sm.GetHealthHandler(checker, startTime, "0.1.0", metricsProvider)...)
	handlers = append(handlers, sm.GetDockerHandler(jc)...)
	handlers = append(handlers, sm.GetMetricsHandler(metricsHandler, metricsPath)...)
	if auditStore != nil {
		handlers = append(handlers, sm.GetAuditHandler(auditStore)...)
	}
	if objStore != nil {
		handlers = append(handlers, sm.GetFileHandler(objStore)...)
	}

	sm.RegisterHandlers(handlers)
}

// startConditionWatcher creates a LogNotifier and a Watcher, then starts the
// watcher in a background goroutine when notifications are enabled. The watcher
// monitors the registry KV bucket for condition transitions.
func startConditionWatcher(
	ctx context.Context,
	log *slog.Logger,
	registryKV jetstream.KeyValue,
) {
	if !appConfig.Notifications.Enabled {
		return
	}

	if registryKV == nil {
		return
	}

	var renotifyInterval time.Duration
	if appConfig.Notifications.RenotifyInterval != "" {
		d, err := time.ParseDuration(appConfig.Notifications.RenotifyInterval)
		if err != nil {
			log.Warn(
				"invalid renotify_interval, using default (0 = disabled)",
				slog.String("value", appConfig.Notifications.RenotifyInterval),
				slog.String("error", err.Error()),
			)
		} else {
			renotifyInterval = d
		}
	}

	notifier := notify.NewLogNotifier(log)
	watcher := notify.NewWatcher(registryKV, notifier, log, renotifyInterval)

	go func() {
		if err := watcher.Start(ctx); err != nil {
			log.Warn(
				"condition watcher stopped with error",
				slog.String("error", err.Error()),
			)
		}
	}()
}

// startControllerHeartbeat resolves the local hostname, creates a ComponentHeartbeat
// for the API server, and starts it in a background goroutine. The heartbeat
// deregisters when ctx is cancelled.
func startControllerHeartbeat(
	ctx context.Context,
	log *slog.Logger,
	registryKV jetstream.KeyValue,
) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Warn(
			"failed to resolve hostname for controller heartbeat, using 'unknown'",
			slog.String("error", err.Error()),
		)
		hostname = "unknown"
	}

	hb := controller.NewComponentHeartbeat(
		log,
		registryKV,
		hostname,
		"0.1.0",
			"controller",
		process.New(),
		10*time.Second,
		process.ConditionThresholds{
			MemoryPressureBytes: appConfig.Agent.ProcessConditions.MemoryPressureBytes,
			HighCPUPercent:      appConfig.Agent.ProcessConditions.HighCPUPercent,
		},
	)

	go hb.Start(ctx)
}
