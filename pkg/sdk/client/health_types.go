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

package client

import "github.com/osapi-io/osapi/pkg/sdk/client/gen"

// HealthStatus represents a liveness check response.
type HealthStatus struct {
	Status string `json:"status"`
}

// ReadyStatus represents a readiness check response.
type ReadyStatus struct {
	Status             string `json:"status"`
	Error              string `json:"error,omitempty"`
	ServiceUnavailable bool   `json:"service_unavailable"`
}

// SystemStatus represents detailed system status.
type SystemStatus struct {
	Status       string            `json:"status"`
	Version      string            `json:"version"`
	Uptime       string            `json:"uptime"`
	Checks       []Check           `json:"checks,omitempty"`
	NATS         *NATSInfo         `json:"nats,omitempty"`
	Agents       *AgentStats       `json:"agents,omitempty"`
	Jobs         *JobStats         `json:"jobs,omitempty"`
	Consumers    int               `json:"consumers"`
	Streams      []StreamInfo      `json:"streams,omitempty"`
	KVBuckets    []KVBucketInfo    `json:"kv_buckets,omitempty"`
	ObjectStores []ObjectStoreInfo `json:"object_stores,omitempty"`
	Registry     []RegistryEntry   `json:"registry,omitempty"`
}

// RegistryEntry represents a unified component registration in the health registry.
type RegistryEntry struct {
	Type       string   `json:"type"`
	Hostname   string   `json:"hostname"`
	Status     string   `json:"status"`
	Conditions []string `json:"conditions,omitempty"`
	Age        string   `json:"age"`
	CPUPercent float64  `json:"cpu_percent"`
	MemBytes   int64    `json:"mem_bytes"`
	// Version is the build version of the component's binary.
	Version string `json:"version,omitempty"`
	// Address is the listen address of a sub-component.
	Address string `json:"address,omitempty"`
	// URL is where a person opens a sub-component; empty when not
	// configured.
	URL string `json:"url,omitempty"`
	// Components are the sub-components of the component, sorted by name.
	Components []RegistryEntry `json:"components,omitempty"`
}

// Check is one check that the controller made for its answer, such as its
// connection to NATS.
type Check struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// NATSInfo represents NATS connection info.
type NATSInfo struct {
	URL     string `json:"url"`
	Version string `json:"version"`
}

// AgentStats represents agent statistics from the health endpoint.
type AgentStats struct {
	Total  int            `json:"total"`
	Ready  int            `json:"ready"`
	Agents []AgentSummary `json:"agents,omitempty"`
}

// AgentSummary represents a summary of an agent from the health endpoint.
type AgentSummary struct {
	Hostname   string `json:"hostname"`
	Labels     string `json:"labels,omitempty"`
	Registered string `json:"registered"`
}

// JobStats represents job queue statistics from the health endpoint.
type JobStats struct {
	Total       int `json:"total"`
	Completed   int `json:"completed"`
	Failed      int `json:"failed"`
	Processing  int `json:"processing"`
	Unprocessed int `json:"unprocessed"`
	Dlq         int `json:"dlq"`
}

// StreamInfo represents a JetStream stream's info.
type StreamInfo struct {
	Name      string `json:"name"`
	Messages  int    `json:"messages"`
	Bytes     int    `json:"bytes"`
	Consumers int    `json:"consumers"`
}

// KVBucketInfo represents a KV bucket's info.
type KVBucketInfo struct {
	Name  string `json:"name"`
	Keys  int    `json:"keys"`
	Bytes int    `json:"bytes"`
}

// ObjectStoreInfo represents an Object Store bucket's info.
type ObjectStoreInfo struct {
	Name string `json:"name"`
	Size int    `json:"size"`
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
) SystemStatus {
	s := SystemStatus{
		Status:  g.Status,
		Version: g.Version,
		Uptime:  g.Uptime,
	}

	for _, c := range g.Checks {
		check := Check{Name: c.Name, Status: c.Status}
		if c.Error != nil {
			check.Error = *c.Error
		}
		s.Checks = append(s.Checks, check)
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

	s.Consumers = g.Consumers

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

	if g.Registry != nil {
		s.Registry = registryFromGen(*g.Registry)
	}

	return s
}

// registryFromGen converts the registry entries of the answer, with the
// sub-components under each.
func registryFromGen(
	in []gen.ComponentEntry,
) []RegistryEntry {
	entries := make([]RegistryEntry, 0, len(in))
	for _, e := range in {
		entry := RegistryEntry{}
		if e.Type != nil {
			entry.Type = *e.Type
		}
		if e.Hostname != nil {
			entry.Hostname = *e.Hostname
		}
		if e.Status != nil {
			entry.Status = *e.Status
		}
		if e.Version != nil {
			entry.Version = *e.Version
		}
		if e.Age != nil {
			entry.Age = *e.Age
		}
		if e.CpuPercent != nil {
			entry.CPUPercent = float64(*e.CpuPercent)
		}
		if e.MemBytes != nil {
			entry.MemBytes = *e.MemBytes
		}
		if e.Conditions != nil {
			entry.Conditions = *e.Conditions
		}
		if e.Address != nil {
			entry.Address = *e.Address
		}
		if e.Url != nil {
			entry.URL = *e.Url
		}
		if e.Components != nil {
			entry.Components = registryFromGen(*e.Components)
		}
		entries = append(entries, entry)
	}

	return entries
}
