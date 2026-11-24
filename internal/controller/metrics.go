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
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/health"
	"github.com/osapi-io/osapi/internal/job"
	jobclient "github.com/osapi-io/osapi/internal/job/client"
)

// natsStatus is the part of a NATS client that the health metrics read.
type natsStatus interface {
	ConnectedURL() string
	ConnectedServerVersion() string
	StreamInfo(ctx context.Context, name string) (*jetstream.StreamInfo, error)
	KVStatus(ctx context.Context, bucket string) (jetstream.KeyValueStatus, error)
	KVKeys(ctx context.Context, bucket string, filters ...string) ([]string, error)
	KVGet(ctx context.Context, bucket string, key string) (jetstream.KeyValueEntry, error)
	ObjectStoreStatus(ctx context.Context, bucket string) (jetstream.ObjectStoreStatus, error)
}

// newHealthChecker checks that NATS is connected and that the jobs bucket
// answers.
func newHealthChecker(
	nc interface {
		ConnectedURL() string
		KVStatus(ctx context.Context, bucket string) (jetstream.KeyValueStatus, error)
	},
	jobsBucket string,
) *health.NATSChecker {
	return &health.NATSChecker{
		NATSCheck: func() error {
			if nc.ConnectedURL() == "" {
				return fmt.Errorf("nats not connected")
			}

			return nil
		},
		KVCheck: func() error {
			_, err := nc.KVStatus(context.Background(), jobsBucket)
			if err != nil {
				return fmt.Errorf("kv bucket not accessible: %w", err)
			}

			return nil
		},
	}
}

// configuredKVBuckets returns the namespaced names of all KV buckets
// declared in osapi.yaml. Only non-empty bucket configs are included.
// configuredKVBuckets returns the name of each KV bucket that cfg names.
func configuredKVBuckets(
	namespace string,
	cfg config.NATS,
) []string {
	var buckets []string
	for _, info := range cfg.AllKVBuckets() {
		if info.Bucket != "" {
			buckets = append(buckets, job.ApplyNamespaceToInfraName(namespace, info.Bucket))
		}
	}

	return buckets
}

// configuredObjectBuckets returns the namespaced names of all Object Store
// buckets declared in osapi.yaml.
// configuredObjectBuckets returns the name of each Object Store bucket that
// cfg names.
func configuredObjectBuckets(
	namespace string,
	cfg config.NATS,
) []string {
	var buckets []string
	for _, info := range cfg.AllObjectStoreBuckets() {
		if info.Bucket != "" {
			buckets = append(
				buckets,
				job.ApplyNamespaceToInfraName(namespace, info.Bucket),
			)
		}
	}

	return buckets
}

// newMetricsProvider reads the state of NATS, the jobs and the components for
// the health status endpoint.
func newMetricsProvider(
	nc natsStatus,
	streamName string,
	kvBuckets []string,
	objBuckets []string,
	jc jobclient.JobClient,
	registryBucket string,
) *health.ClosureMetricsProvider {
	return &health.ClosureMetricsProvider{
		NATSInfoFn: func(_ context.Context) (*health.NATSMetrics, error) {
			url := nc.ConnectedURL()
			if url == "" {
				return nil, fmt.Errorf("NATS client unavailable")
			}

			return &health.NATSMetrics{
				URL:     url,
				Version: nc.ConnectedServerVersion(),
			}, nil
		},
		StreamInfoFn: func(fnCtx context.Context) ([]health.StreamMetrics, error) {
			info, err := nc.StreamInfo(fnCtx, streamName)
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
			results := make([]health.KVMetrics, 0, len(kvBuckets))
			for _, name := range kvBuckets {
				status, err := nc.KVStatus(fnCtx, name)
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
			results := make([]health.ObjectStoreMetrics, 0, len(objBuckets))
			for _, name := range objBuckets {
				status, err := nc.ObjectStoreStatus(fnCtx, name)
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
		JobStatsFn: func(fnCtx context.Context) (*health.JobMetrics, error) {
			stats, err := jc.GetQueueSummary(fnCtx)
			if err != nil {
				return nil, fmt.Errorf("job stats: %w", err)
			}

			return &health.JobMetrics{
				Total:      stats.TotalJobs,
				Submitted:  stats.StatusCounts[types.JobStatusSubmitted],
				Processing: stats.StatusCounts[types.JobStatusProcessing],
				Completed:  stats.StatusCounts[types.JobStatusCompleted],
				Failed:     stats.StatusCounts[types.JobStatusFailed],
				DLQ:        stats.DLQCount,
			}, nil
		},
		AgentStatsFn: func(fnCtx context.Context) (*health.AgentMetrics, error) {
			agents, err := jc.ListAgents(fnCtx)
			if err != nil {
				return nil, fmt.Errorf("agent stats: %w", err)
			}

			details := make([]health.AgentDetail, 0, len(agents))
			ready := 0
			for _, a := range agents {
				// A drained agent is registered but takes no jobs, so it is
				// not ready.
				if a.State == "" || a.State == types.AgentReady {
					ready++
				}
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
				Ready:  ready,
				Agents: details,
			}, nil
		},
		ComponentRegistryFn: func(fnCtx context.Context) ([]health.ComponentEntry, error) {
			if registryBucket == "" {
				return nil, fmt.Errorf("registry KV not configured")
			}

			keys, err := nc.KVKeys(fnCtx, registryBucket)
			if err != nil {
				return nil, fmt.Errorf("list registry keys: %w", err)
			}

			entries := make([]health.ComponentEntry, 0, len(keys))
			for _, key := range keys {
				parts := strings.SplitN(key, ".", 2)
				if len(parts) != 2 {
					continue
				}
				prefix := parts[0]

				kvEntry, err := nc.KVGet(fnCtx, registryBucket, key)
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
					// The agent's scheduling state, by the same rule as agent
					// list: a drain the agent has not applied yet is Draining.
					status := job.EffectiveState(reg.State, jc.CheckDrainFlag(fnCtx, reg.MachineID))
					entry = health.ComponentEntry{
						Type:     "agent",
						Hostname: reg.Hostname,
						Status:   string(status),
						Age:      cli.FormatAge(time.Since(reg.StartedAt)),
						Version:  reg.AgentVersion,
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
					entry.SubComponents = convertSubComponents(reg.SubComponents)
				case "controller", "nats":
					var reg job.ComponentRegistration
					if err := json.Unmarshal(kvEntry.Value(), &reg); err != nil {
						continue
					}
					entry = health.ComponentEntry{
						Type:     reg.Type,
						Hostname: reg.Hostname,
						Status:   string(types.AgentReady),
						Age:      cli.FormatAge(time.Since(reg.StartedAt)),
						Version:  reg.Version,
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
					entry.SubComponents = convertSubComponents(reg.SubComponents)
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

// convertSubComponents converts job.SubComponentInfo map to health.SubComponentInfo map.
func convertSubComponents(
	src map[string]job.SubComponentInfo,
) map[string]health.SubComponentInfo {
	if len(src) == 0 {
		return nil
	}

	result := make(map[string]health.SubComponentInfo, len(src))
	for k, v := range src {
		result[k] = health.SubComponentInfo{
			Status:  v.Status,
			Address: v.Address,
			URL:     v.URL,
		}
	}

	return result
}
