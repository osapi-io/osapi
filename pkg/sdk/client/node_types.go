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

import (
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// Collection is a generic wrapper for collection responses from node queries.
type Collection[T any] struct {
	Results []T    `json:"results"`
	JobID   string `json:"job_id"`
}

// Disk represents disk usage information.
type Disk struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
	Used  int    `json:"used"`
	Free  int    `json:"free"`
}

// HostnameResult represents a hostname query result from a single agent.
type HostnameResult struct {
	Hostname string            `json:"hostname"`
	Error    string            `json:"error,omitempty"`
	Changed  bool              `json:"changed"`
	Labels   map[string]string `json:"labels,omitempty"`
}

// NodeStatus represents full node status from a single agent.
type NodeStatus struct {
	Hostname    string       `json:"hostname"`
	Uptime      string       `json:"uptime,omitempty"`
	Error       string       `json:"error,omitempty"`
	Changed     bool         `json:"changed"`
	Disks       []Disk       `json:"disks,omitempty"`
	LoadAverage *LoadAverage `json:"load_average,omitempty"`
	Memory      *Memory      `json:"memory,omitempty"`
	OSInfo      *OSInfo      `json:"os_info,omitempty"`
}

// DiskResult represents disk query result from a single agent.
type DiskResult struct {
	Hostname string `json:"hostname"`
	Error    string `json:"error,omitempty"`
	Changed  bool   `json:"changed"`
	Disks    []Disk `json:"disks,omitempty"`
}

// MemoryResult represents memory query result from a single agent.
type MemoryResult struct {
	Hostname string  `json:"hostname"`
	Error    string  `json:"error,omitempty"`
	Changed  bool    `json:"changed"`
	Memory   *Memory `json:"memory,omitempty"`
}

// LoadResult represents load average query result from a single agent.
type LoadResult struct {
	Hostname    string       `json:"hostname"`
	Error       string       `json:"error,omitempty"`
	Changed     bool         `json:"changed"`
	LoadAverage *LoadAverage `json:"load_average,omitempty"`
}

// OSInfoResult represents OS info query result from a single agent.
type OSInfoResult struct {
	Hostname string  `json:"hostname"`
	Error    string  `json:"error,omitempty"`
	Changed  bool    `json:"changed"`
	OSInfo   *OSInfo `json:"os_info,omitempty"`
}

// UptimeResult represents uptime query result from a single agent.
type UptimeResult struct {
	Hostname string `json:"hostname"`
	Uptime   string `json:"uptime,omitempty"`
	Error    string `json:"error,omitempty"`
	Changed  bool   `json:"changed"`
}

// DNSConfig represents DNS configuration from a single agent.
type DNSConfig struct {
	Hostname      string   `json:"hostname"`
	Error         string   `json:"error,omitempty"`
	Changed       bool     `json:"changed"`
	Servers       []string `json:"servers,omitempty"`
	SearchDomains []string `json:"search_domains,omitempty"`
}

// DNSUpdateResult represents DNS update result from a single agent.
type DNSUpdateResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Changed  bool   `json:"changed"`
}

// PingResult represents ping result from a single agent.
type PingResult struct {
	Hostname        string  `json:"hostname"`
	Error           string  `json:"error,omitempty"`
	Changed         bool    `json:"changed"`
	PacketsSent     int     `json:"packets_sent"`
	PacketsReceived int     `json:"packets_received"`
	PacketLoss      float64 `json:"packet_loss"`
	MinRtt          string  `json:"min_rtt,omitempty"`
	AvgRtt          string  `json:"avg_rtt,omitempty"`
	MaxRtt          string  `json:"max_rtt,omitempty"`
}

// CommandResult represents command execution result from a single agent.
type CommandResult struct {
	Hostname   string `json:"hostname"`
	Stdout     string `json:"stdout,omitempty"`
	Stderr     string `json:"stderr,omitempty"`
	Error      string `json:"error,omitempty"`
	ExitCode   int    `json:"exit_code"`
	Changed    bool   `json:"changed"`
	DurationMs int64  `json:"duration_ms"`
}

// loadAverageFromGen converts a gen.LoadAverageResponse to a LoadAverage.
func loadAverageFromGen(
	g *gen.LoadAverageResponse,
) *LoadAverage {
	if g == nil {
		return nil
	}

	return &LoadAverage{
		OneMin:     g.N1min,
		FiveMin:    g.N5min,
		FifteenMin: g.N15min,
	}
}

// memoryFromGen converts a gen.MemoryResponse to a Memory.
func memoryFromGen(
	g *gen.MemoryResponse,
) *Memory {
	if g == nil {
		return nil
	}

	return &Memory{
		Total: g.Total,
		Used:  g.Used,
		Free:  g.Free,
	}
}

// osInfoFromGen converts a gen.OSInfoResponse to an OSInfo.
func osInfoFromGen(
	g *gen.OSInfoResponse,
) *OSInfo {
	if g == nil {
		return nil
	}

	return &OSInfo{
		Distribution: g.Distribution,
		Version:      g.Version,
	}
}

// disksFromGen converts a gen.DisksResponse to a slice of Disk.
func disksFromGen(
	g *gen.DisksResponse,
) []Disk {
	if g == nil {
		return nil
	}

	disks := make([]Disk, 0, len(*g))
	for _, d := range *g {
		disks = append(disks, Disk{
			Name:  d.Name,
			Total: d.Total,
			Used:  d.Used,
			Free:  d.Free,
		})
	}

	return disks
}

// derefString safely dereferences a string pointer, returning empty string for nil.
func derefString(
	s *string,
) string {
	if s == nil {
		return ""
	}

	return *s
}

// derefInt safely dereferences an int pointer, returning zero for nil.
func derefInt(
	i *int,
) int {
	if i == nil {
		return 0
	}

	return *i
}

// derefInt64 safely dereferences an int64 pointer, returning zero for nil.
func derefInt64(
	i *int64,
) int64 {
	if i == nil {
		return 0
	}

	return *i
}

// derefFloat64 safely dereferences a float64 pointer, returning zero for nil.
func derefFloat64(
	f *float64,
) float64 {
	if f == nil {
		return 0
	}

	return *f
}

// derefBool safely dereferences a bool pointer, returning false for nil.
func derefBool(
	b *bool,
) bool {
	if b == nil {
		return false
	}

	return *b
}

// jobIDFromGen extracts a job ID string from an optional UUID pointer.
func jobIDFromGen(
	id *openapi_types.UUID,
) string {
	if id == nil {
		return ""
	}

	return id.String()
}

// hostnameCollectionFromGen converts a gen.HostnameCollectionResponse to a Collection[HostnameResult].
func hostnameCollectionFromGen(
	g *gen.HostnameCollectionResponse,
) Collection[HostnameResult] {
	results := make([]HostnameResult, 0, len(g.Results))
	for _, r := range g.Results {
		hr := HostnameResult{
			Hostname: r.Hostname,
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
		}

		if r.Labels != nil {
			hr.Labels = *r.Labels
		}

		results = append(results, hr)
	}

	return Collection[HostnameResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// nodeStatusCollectionFromGen converts a gen.NodeStatusCollectionResponse to a Collection[NodeStatus].
func nodeStatusCollectionFromGen(
	g *gen.NodeStatusCollectionResponse,
) Collection[NodeStatus] {
	results := make([]NodeStatus, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, NodeStatus{
			Hostname:    r.Hostname,
			Uptime:      derefString(r.Uptime),
			Error:       derefString(r.Error),
			Changed:     derefBool(r.Changed),
			Disks:       disksFromGen(r.Disks),
			LoadAverage: loadAverageFromGen(r.LoadAverage),
			Memory:      memoryFromGen(r.Memory),
			OSInfo:      osInfoFromGen(r.OsInfo),
		})
	}

	return Collection[NodeStatus]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// diskCollectionFromGen converts a gen.DiskCollectionResponse to a Collection[DiskResult].
func diskCollectionFromGen(
	g *gen.DiskCollectionResponse,
) Collection[DiskResult] {
	results := make([]DiskResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, DiskResult{
			Hostname: r.Hostname,
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
			Disks:    disksFromGen(r.Disks),
		})
	}

	return Collection[DiskResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// memoryCollectionFromGen converts a gen.MemoryCollectionResponse to a Collection[MemoryResult].
func memoryCollectionFromGen(
	g *gen.MemoryCollectionResponse,
) Collection[MemoryResult] {
	results := make([]MemoryResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, MemoryResult{
			Hostname: r.Hostname,
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
			Memory:   memoryFromGen(r.Memory),
		})
	}

	return Collection[MemoryResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// loadCollectionFromGen converts a gen.LoadCollectionResponse to a Collection[LoadResult].
func loadCollectionFromGen(
	g *gen.LoadCollectionResponse,
) Collection[LoadResult] {
	results := make([]LoadResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, LoadResult{
			Hostname:    r.Hostname,
			Error:       derefString(r.Error),
			Changed:     derefBool(r.Changed),
			LoadAverage: loadAverageFromGen(r.LoadAverage),
		})
	}

	return Collection[LoadResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// osInfoCollectionFromGen converts a gen.OSInfoCollectionResponse to a Collection[OSInfoResult].
func osInfoCollectionFromGen(
	g *gen.OSInfoCollectionResponse,
) Collection[OSInfoResult] {
	results := make([]OSInfoResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, OSInfoResult{
			Hostname: r.Hostname,
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
			OSInfo:   osInfoFromGen(r.OsInfo),
		})
	}

	return Collection[OSInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// uptimeCollectionFromGen converts a gen.UptimeCollectionResponse to a Collection[UptimeResult].
func uptimeCollectionFromGen(
	g *gen.UptimeCollectionResponse,
) Collection[UptimeResult] {
	results := make([]UptimeResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, UptimeResult{
			Hostname: r.Hostname,
			Uptime:   derefString(r.Uptime),
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
		})
	}

	return Collection[UptimeResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// dnsConfigCollectionFromGen converts a gen.DNSConfigCollectionResponse to a Collection[DNSConfig].
func dnsConfigCollectionFromGen(
	g *gen.DNSConfigCollectionResponse,
) Collection[DNSConfig] {
	results := make([]DNSConfig, 0, len(g.Results))
	for _, r := range g.Results {
		dc := DNSConfig{
			Hostname: r.Hostname,
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
		}

		if r.Servers != nil {
			dc.Servers = *r.Servers
		}

		if r.SearchDomains != nil {
			dc.SearchDomains = *r.SearchDomains
		}

		results = append(results, dc)
	}

	return Collection[DNSConfig]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// dnsUpdateCollectionFromGen converts a gen.DNSUpdateCollectionResponse to a Collection[DNSUpdateResult].
func dnsUpdateCollectionFromGen(
	g *gen.DNSUpdateCollectionResponse,
) Collection[DNSUpdateResult] {
	results := make([]DNSUpdateResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, DNSUpdateResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
		})
	}

	return Collection[DNSUpdateResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// pingCollectionFromGen converts a gen.PingCollectionResponse to a Collection[PingResult].
func pingCollectionFromGen(
	g *gen.PingCollectionResponse,
) Collection[PingResult] {
	results := make([]PingResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, PingResult{
			Hostname:        r.Hostname,
			Error:           derefString(r.Error),
			Changed:         derefBool(r.Changed),
			PacketsSent:     derefInt(r.PacketsSent),
			PacketsReceived: derefInt(r.PacketsReceived),
			PacketLoss:      derefFloat64(r.PacketLoss),
			MinRtt:          derefString(r.MinRtt),
			AvgRtt:          derefString(r.AvgRtt),
			MaxRtt:          derefString(r.MaxRtt),
		})
	}

	return Collection[PingResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// commandCollectionFromGen converts a gen.CommandResultCollectionResponse to a Collection[CommandResult].
func commandCollectionFromGen(
	g *gen.CommandResultCollectionResponse,
) Collection[CommandResult] {
	results := make([]CommandResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, CommandResult{
			Hostname:   r.Hostname,
			Stdout:     derefString(r.Stdout),
			Stderr:     derefString(r.Stderr),
			Error:      derefString(r.Error),
			ExitCode:   derefInt(r.ExitCode),
			Changed:    derefBool(r.Changed),
			DurationMs: derefInt64(r.DurationMs),
		})
	}

	return Collection[CommandResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
