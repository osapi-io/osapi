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
	"time"

	"github.com/labstack/echo/v4"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/api"
	"github.com/retr0h/osapi/internal/api/health"
	jobclient "github.com/retr0h/osapi/internal/job/client"
	"github.com/retr0h/osapi/internal/messaging"
)

// ServerManager responsible for Server operations.
type ServerManager interface {
	Lifecycle
	// CreateHandlers initializes handlers and returns a slice of functions to register them.
	CreateHandlers(
		jobClient jobclient.JobClient,
	) []func(e *echo.Echo)
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

		// Create NATS client for job system
		var nc messaging.NATSClient = natsclient.New(logger, &natsclient.Options{
			Host: appConfig.Job.Client.Host,
			Port: appConfig.Job.Client.Port,
			Auth: natsclient.AuthOptions{
				AuthType: natsclient.NoAuth,
			},
			Name: appConfig.Job.Client.ClientName,
		})

		if err := nc.Connect(); err != nil {
			logFatal("failed to connect to NATS for job client", err)
		}

		jobsKV, err := nc.CreateKVBucket(appConfig.Job.KVBucket)
		if err != nil {
			logFatal("failed to create KV bucket", err)
		}

		jc, err := jobclient.New(logger, nc, &jobclient.Options{
			Timeout:  30 * time.Second,
			KVBucket: jobsKV,
		})
		if err != nil {
			logFatal("failed to create job client", err)
		}

		startTime := time.Now()

		checker := &health.NATSChecker{
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
				_, err := jobsKV.Keys()
				if err != nil {
					return fmt.Errorf("KV bucket not accessible: %w", err)
				}

				return nil
			},
		}

		metricsProvider := &health.ClosureMetricsProvider{
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
				info, err := nc.GetStreamInfo(fnCtx, appConfig.Job.StreamName)
				if err != nil {
					return nil, fmt.Errorf("stream info: %w", err)
				}

				return []health.StreamMetrics{
					{
						Name:      appConfig.Job.StreamName,
						Messages:  info.State.Msgs,
						Bytes:     info.State.Bytes,
						Consumers: info.State.Consumers,
					},
				}, nil
			},
			KVInfoFn: func(_ context.Context) ([]health.KVMetrics, error) {
				status, err := jobsKV.Status()
				if err != nil {
					return nil, fmt.Errorf("KV status: %w", err)
				}

				keys, _ := jobsKV.Keys()
				keyCount := len(keys)

				return []health.KVMetrics{
					{
						Name:  status.Bucket(),
						Keys:  keyCount,
						Bytes: status.Bytes(),
					},
				}, nil
			},
			JobStatsFn: func(fnCtx context.Context) (*health.JobMetrics, error) {
				stats, err := jc.GetQueueStats(fnCtx)
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
		}

		healthHandler := health.New(checker, startTime, "0.1.0", metricsProvider)

		var sm ServerManager = api.New(appConfig, logger, api.WithHealthHandler(healthHandler))
		handlers := sm.CreateHandlers(jc)
		sm.RegisterHandlers(handlers)

		sm.Start()
		runServer(ctx, sm, func() {
			if natsConn, ok := nc.(*natsclient.Client); ok && natsConn.NC != nil {
				natsConn.NC.Close()
			}
		})
	},
}

func init() {
	apiServerCmd.AddCommand(apiServerStartCmd)
}
