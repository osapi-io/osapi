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
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/api"
	"github.com/retr0h/osapi/internal/api/health"
	"github.com/retr0h/osapi/internal/audit"
	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/job"
	jobclient "github.com/retr0h/osapi/internal/job/client"
	"github.com/retr0h/osapi/internal/messaging"
	"github.com/retr0h/osapi/internal/telemetry"
	"github.com/retr0h/osapi/internal/validation"
)

// ServerManager responsible for Server operations.
type ServerManager interface {
	cli.Lifecycle
	// GetAgentHandler returns agent handler for registration.
	GetAgentHandler(jobClient jobclient.JobClient) []func(e *echo.Echo)
	// GetNodeHandler returns node handler for registration.
	GetNodeHandler(jobClient jobclient.JobClient) []func(e *echo.Echo)
	// GetNetworkHandler returns network handler for registration.
	GetNetworkHandler(jobClient jobclient.JobClient) []func(e *echo.Echo)
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
	// GetCommandHandler returns command handler for registration.
	GetCommandHandler(jobClient jobclient.JobClient) []func(e *echo.Echo)
	// GetAuditHandler returns audit handler for registration.
	GetAuditHandler(store audit.Store) []func(e *echo.Echo)
	// RegisterHandlers registers a list of handlers with the Echo instance.
	RegisterHandlers(handlers []func(e *echo.Echo))
}

// apiServerStartCmd represents the apiServerStart command.
var apiServerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long: `Start the API server.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		shutdownTracer, err := telemetry.InitTracer(ctx, "osapi-api", appConfig.Telemetry.Tracing)
		if err != nil {
			cli.LogFatal(logger, "failed to initialize tracer", err)
		}

		metricsHandler, metricsPath, shutdownMeter, err := telemetry.InitMeter(
			appConfig.Telemetry.Metrics,
		)
		if err != nil {
			cli.LogFatal(logger, "failed to initialize meter", err)
		}

		namespace := appConfig.API.NATS.Namespace
		job.Init(namespace)
		streamName := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.Stream.Name)
		kvBucket := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.KV.Bucket)

		nc, jobsKV := connectNATSAndKV(ctx, kvBucket)

		// Create registry KV bucket for agent discovery
		registryKVConfig := cli.BuildRegistryKVConfig(namespace, appConfig.NATS.Registry)
		registryKV, err := nc.CreateOrUpdateKVBucketWithConfig(ctx, registryKVConfig)
		if err != nil {
			cli.LogFatal(logger, "failed to create registry KV bucket", err)
		}

		jc, err := jobclient.New(logger, nc, &jobclient.Options{
			Timeout:    30 * time.Second,
			KVBucket:   jobsKV,
			RegistryKV: registryKV,
			StreamName: streamName,
		})
		if err != nil {
			cli.LogFatal(logger, "failed to create job client", err)
		}

		validation.RegisterTargetValidator(
			func(ctx context.Context) ([]validation.AgentTarget, error) {
				agents, err := jc.ListAgents(ctx)
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

		checker := newHealthChecker(nc, jobsKV)
		auditStore, auditKV, serverOpts := createAuditStore(ctx, nc, namespace)
		metricsProvider := newMetricsProvider(nc, jobsKV, registryKV, auditKV, streamName, jc)

		var sm ServerManager = api.New(appConfig, logger, serverOpts...)

		registerAPIHandlers(
			sm,
			jc,
			checker,
			metricsProvider,
			metricsHandler,
			metricsPath,
			auditStore,
		)

		sm.Start()
		cli.RunServer(ctx, sm, func() {
			_ = shutdownMeter(context.Background())
			_ = shutdownTracer(context.Background())
			cli.CloseNATSClient(nc)
		})
	},
}

func connectNATSAndKV(
	ctx context.Context,
	kvBucket string,
) (messaging.NATSClient, jetstream.KeyValue) {
	var nc messaging.NATSClient = natsclient.New(logger, &natsclient.Options{
		Host: appConfig.API.NATS.Host,
		Port: appConfig.API.NATS.Port,
		Auth: cli.BuildNATSAuthOptions(appConfig.API.NATS.Auth),
		Name: appConfig.API.NATS.ClientName,
	})

	if err := nc.Connect(); err != nil {
		cli.LogFatal(logger, "failed to connect to NATS for job client", err)
	}

	jobsKV, err := nc.CreateOrUpdateKVBucket(ctx, kvBucket)
	if err != nil {
		cli.LogFatal(logger, "failed to create KV bucket", err)
	}

	return nc, jobsKV
}

func newHealthChecker(
	nc messaging.NATSClient,
	jobsKV jetstream.KeyValue,
) *health.NATSChecker {
	return &health.NATSChecker{
		NATSCheck: func() error {
			natsConn, ok := nc.(*natsclient.Client)
			if !ok || natsConn.NC == nil {
				return fmt.Errorf("NATS client unavailable")
			}

			if natsConn.NC.ConnectedUrl() == "" {
				return fmt.Errorf("NATS not connected")
			}

			return nil
		},
		KVCheck: func() error {
			_, err := jobsKV.Status(context.Background())
			if err != nil {
				return fmt.Errorf("KV bucket not accessible: %w", err)
			}

			return nil
		},
	}
}

func newMetricsProvider(
	nc messaging.NATSClient,
	jobsKV jetstream.KeyValue,
	registryKV jetstream.KeyValue,
	auditKV jetstream.KeyValue,
	streamName string,
	jc jobclient.JobClient,
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
			buckets := []jetstream.KeyValue{jobsKV, registryKV, auditKV}
			results := make([]health.KVMetrics, 0, len(buckets))

			for _, kv := range buckets {
				if kv == nil {
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
		ConsumerStatsFn: func(fnCtx context.Context) (*health.ConsumerMetrics, error) {
			natsConn, ok := nc.(*natsclient.Client)
			if !ok || natsConn.ExtJS == nil {
				return nil, fmt.Errorf("JetStream client unavailable")
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
	}
}

func createAuditStore(
	ctx context.Context,
	nc messaging.NATSClient,
	namespace string,
) (audit.Store, jetstream.KeyValue, []api.Option) {
	if appConfig.NATS.Audit.Bucket == "" {
		return nil, nil, nil
	}

	auditKVConfig := cli.BuildAuditKVConfig(namespace, appConfig.NATS.Audit)
	auditKV, err := nc.CreateOrUpdateKVBucketWithConfig(ctx, auditKVConfig)
	if err != nil {
		cli.LogFatal(logger, "failed to create audit KV bucket", err)
	}

	store := audit.NewKVStore(logger, auditKV)

	return store, auditKV, []api.Option{api.WithAuditStore(store)}
}

func registerAPIHandlers(
	sm ServerManager,
	jc jobclient.JobClient,
	checker health.Checker,
	metricsProvider health.MetricsProvider,
	metricsHandler http.Handler,
	metricsPath string,
	auditStore audit.Store,
) {
	startTime := time.Now()

	handlers := make([]func(e *echo.Echo), 0, 8)
	handlers = append(handlers, sm.GetAgentHandler(jc)...)
	handlers = append(handlers, sm.GetNodeHandler(jc)...)
	handlers = append(handlers, sm.GetNetworkHandler(jc)...)
	handlers = append(handlers, sm.GetJobHandler(jc)...)
	handlers = append(handlers, sm.GetCommandHandler(jc)...)
	handlers = append(
		handlers,
		sm.GetHealthHandler(checker, startTime, "0.1.0", metricsProvider)...)
	handlers = append(handlers, sm.GetMetricsHandler(metricsHandler, metricsPath)...)
	if auditStore != nil {
		handlers = append(handlers, sm.GetAuditHandler(auditStore)...)
	}

	sm.RegisterHandlers(handlers)
}

func init() {
	apiServerCmd.AddCommand(apiServerStartCmd)
}
