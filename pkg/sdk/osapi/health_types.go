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

package osapi

import "github.com/retr0h/osapi/pkg/sdk/osapi/gen"

// HealthStatus represents a liveness check response.
type HealthStatus struct {
	Status string
}

// ReadyStatus represents a readiness check response.
type ReadyStatus struct {
	Status             string
	Error              string
	ServiceUnavailable bool
}

// SystemStatus represents detailed system status.
type SystemStatus struct {
	Status             string
	Version            string
	Uptime             string
	ServiceUnavailable bool
	Components         map[string]ComponentHealth
	NATS               *NATSInfo
	Agents             *AgentStats
	Jobs               *JobStats
	Consumers          *ConsumerStats
	Streams            []StreamInfo
	KVBuckets          []KVBucketInfo
	ObjectStores       []ObjectStoreInfo
}

// ComponentHealth represents a component's health.
type ComponentHealth struct {
	Status string
	Error  string
}

// NATSInfo represents NATS connection info.
type NATSInfo struct {
	URL     string
	Version string
}

// AgentStats represents agent statistics from the health endpoint.
type AgentStats struct {
	Total  int
	Ready  int
	Agents []AgentSummary
}

// AgentSummary represents a summary of an agent from the health endpoint.
type AgentSummary struct {
	Hostname   string
	Labels     string
	Registered string
}

// JobStats represents job queue statistics from the health endpoint.
type JobStats struct {
	Total       int
	Completed   int
	Failed      int
	Processing  int
	Unprocessed int
	Dlq         int
}

// ConsumerStats represents JetStream consumer statistics.
type ConsumerStats struct {
	Total     int
	Consumers []ConsumerDetail
}

// ConsumerDetail represents a single consumer's details.
type ConsumerDetail struct {
	Name        string
	Pending     int
	AckPending  int
	Redelivered int
}

// StreamInfo represents a JetStream stream's info.
type StreamInfo struct {
	Name      string
	Messages  int
	Bytes     int
	Consumers int
}

// KVBucketInfo represents a KV bucket's info.
type KVBucketInfo struct {
	Name  string
	Keys  int
	Bytes int
}

// ObjectStoreInfo represents an Object Store bucket's info.
type ObjectStoreInfo struct {
	Name string
	Size int
}

// healthStatusFromGen converts a gen.HealthResponse to a HealthStatus.
func healthStatusFromGen(
	g *gen.HealthResponse,
) HealthStatus {
	return HealthStatus{
		Status: g.Status,
	}
}

// readyStatusFromGen converts a gen.ReadyResponse to a ReadyStatus.
func readyStatusFromGen(
	g *gen.ReadyResponse,
	serviceUnavailable bool,
) ReadyStatus {
	r := ReadyStatus{
		Status:             g.Status,
		ServiceUnavailable: serviceUnavailable,
	}

	if g.Error != nil {
		r.Error = *g.Error
	}

	return r
}

// systemStatusFromGen converts a gen.StatusResponse to a SystemStatus.
func systemStatusFromGen(
	g *gen.StatusResponse,
	serviceUnavailable bool,
) SystemStatus {
	s := SystemStatus{
		Status:             g.Status,
		Version:            g.Version,
		Uptime:             g.Uptime,
		ServiceUnavailable: serviceUnavailable,
	}

	if g.Components != nil {
		comps := make(map[string]ComponentHealth, len(g.Components))
		for k, v := range g.Components {
			ch := ComponentHealth{
				Status: v.Status,
			}

			if v.Error != nil {
				ch.Error = *v.Error
			}

			comps[k] = ch
		}

		s.Components = comps
	}

	if g.Nats != nil {
		s.NATS = &NATSInfo{
			URL:     g.Nats.Url,
			Version: g.Nats.Version,
		}
	}

	if g.Agents != nil {
		as := &AgentStats{
			Total: g.Agents.Total,
			Ready: g.Agents.Ready,
		}

		if g.Agents.Agents != nil {
			agents := make([]AgentSummary, 0, len(*g.Agents.Agents))
			for _, a := range *g.Agents.Agents {
				summary := AgentSummary{
					Hostname:   a.Hostname,
					Registered: a.Registered,
				}

				if a.Labels != nil {
					summary.Labels = *a.Labels
				}

				agents = append(agents, summary)
			}

			as.Agents = agents
		}

		s.Agents = as
	}

	if g.Jobs != nil {
		s.Jobs = &JobStats{
			Total:       g.Jobs.Total,
			Completed:   g.Jobs.Completed,
			Failed:      g.Jobs.Failed,
			Processing:  g.Jobs.Processing,
			Unprocessed: g.Jobs.Unprocessed,
			Dlq:         g.Jobs.Dlq,
		}
	}

	if g.Consumers != nil {
		cs := &ConsumerStats{
			Total: g.Consumers.Total,
		}

		if g.Consumers.Consumers != nil {
			consumers := make([]ConsumerDetail, 0, len(*g.Consumers.Consumers))
			for _, c := range *g.Consumers.Consumers {
				consumers = append(consumers, ConsumerDetail{
					Name:        c.Name,
					Pending:     c.Pending,
					AckPending:  c.AckPending,
					Redelivered: c.Redelivered,
				})
			}

			cs.Consumers = consumers
		}

		s.Consumers = cs
	}

	if g.Streams != nil {
		streams := make([]StreamInfo, 0, len(*g.Streams))
		for _, st := range *g.Streams {
			streams = append(streams, StreamInfo{
				Name:      st.Name,
				Messages:  st.Messages,
				Bytes:     st.Bytes,
				Consumers: st.Consumers,
			})
		}

		s.Streams = streams
	}

	if g.KvBuckets != nil {
		buckets := make([]KVBucketInfo, 0, len(*g.KvBuckets))
		for _, b := range *g.KvBuckets {
			buckets = append(buckets, KVBucketInfo{
				Name:  b.Name,
				Keys:  b.Keys,
				Bytes: b.Bytes,
			})
		}

		s.KVBuckets = buckets
	}

	if g.ObjectStores != nil {
		stores := make([]ObjectStoreInfo, 0, len(*g.ObjectStores))
		for _, o := range *g.ObjectStores {
			stores = append(stores, ObjectStoreInfo{
				Name: o.Name,
				Size: o.Size,
			})
		}

		s.ObjectStores = stores
	}

	return s
}
