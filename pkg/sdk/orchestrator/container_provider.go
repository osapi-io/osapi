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

package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ContainerProvider provides typed access to provider operations inside
// a container via the RuntimeTarget's ExecProvider method. This runs
// `osapi provider run <provider> <operation>` inside the container,
// so the osapi binary must be present at /osapi in the container.
//
// The SDK owns the input/output type contracts. The CLI's `provider run`
// command and this typed layer use the same serialization format.
type ContainerProvider struct {
	target RuntimeTarget
}

// NewContainerProvider creates a provider bound to a RuntimeTarget.
func NewContainerProvider(
	target RuntimeTarget,
) *ContainerProvider {
	return &ContainerProvider{target: target}
}

// ── Result types ─────────────────────────────────────────────────────
//
// These mirror the internal provider result types and define the JSON
// contract between the `osapi provider run` CLI and SDK consumers.

// HostOSInfo contains operating system information.
type HostOSInfo struct {
	Distribution string `json:"Distribution"`
	Version      string `json:"Version"`
	Changed      bool   `json:"changed"`
}

// MemStats contains memory statistics in bytes.
type MemStats struct {
	Total     uint64 `json:"Total"`
	Available uint64 `json:"Available"`
	Free      uint64 `json:"Free"`
	Cached    uint64 `json:"Cached"`
	Changed   bool   `json:"changed"`
}

// LoadStats contains load average statistics.
type LoadStats struct {
	Load1   float32 `json:"Load1"`
	Load5   float32 `json:"Load5"`
	Load15  float32 `json:"Load15"`
	Changed bool    `json:"changed"`
}

// DiskUsage contains disk usage statistics for a single mount point.
type DiskUsage struct {
	Name    string `json:"Name"`
	Total   uint64 `json:"Total"`
	Used    uint64 `json:"Used"`
	Free    uint64 `json:"Free"`
	Changed bool   `json:"changed"`
}

// CommandResult contains the output of a command execution.
type CommandResult struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
	Changed    bool   `json:"changed"`
}

// ── Input types ──────────────────────────────────────────────────────
//
// These mirror the internal provider param types.

// ExecParams contains parameters for direct command execution.
type ExecParams struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	Cwd     string   `json:"cwd,omitempty"`
	Timeout int      `json:"timeout,omitempty"`
}

// ShellParams contains parameters for shell command execution.
type ShellParams struct {
	Command string `json:"command"`
	Cwd     string `json:"cwd,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// PingParams contains parameters for the ping operation.
type PingParams struct {
	Address string `json:"address"`
}

// PingResult contains the result of a ping operation.
type PingResult struct {
	PacketsSent     int           `json:"PacketsSent"`
	PacketsReceived int           `json:"PacketsReceived"`
	PacketLoss      float64       `json:"PacketLoss"`
	MinRTT          time.Duration `json:"MinRTT"`
	AvgRTT          time.Duration `json:"AvgRTT"`
	MaxRTT          time.Duration `json:"MaxRTT"`
	Changed         bool          `json:"changed"`
}

// DNSGetParams contains parameters for DNS configuration retrieval.
type DNSGetParams struct {
	InterfaceName string `json:"interface_name"`
}

// DNSGetResult contains DNS configuration.
type DNSGetResult struct {
	DNSServers    []string `json:"DNSServers"`
	SearchDomains []string `json:"SearchDomains"`
}

// DNSUpdateParams contains parameters for DNS configuration update.
type DNSUpdateParams struct {
	Servers       []string `json:"servers"`
	SearchDomains []string `json:"search_domains"`
	InterfaceName string   `json:"interface_name"`
}

// DNSUpdateResult contains the result of a DNS update operation.
type DNSUpdateResult struct {
	Changed bool `json:"changed"`
}

// ── Helper ───────────────────────────────────────────────────────────

// run executes a provider operation and unmarshals the JSON result.
func run[T any](
	ctx context.Context,
	cp *ContainerProvider,
	provider string,
	operation string,
	params any,
) (*T, error) {
	var data []byte
	if params != nil {
		var err error
		data, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal %s/%s params: %w", provider, operation, err)
		}
	}

	out, err := cp.target.ExecProvider(ctx, provider, operation, data)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("unmarshal %s/%s result: %w", provider, operation, err)
	}

	return &result, nil
}

// runScalar executes a provider operation that returns a JSON scalar
// (string, int, float).
func runScalar[T any](
	ctx context.Context,
	cp *ContainerProvider,
	provider string,
	operation string,
) (T, error) {
	var zero T

	out, err := cp.target.ExecProvider(ctx, provider, operation, nil)
	if err != nil {
		return zero, err
	}

	var result T
	if err := json.Unmarshal(out, &result); err != nil {
		return zero, fmt.Errorf("unmarshal %s/%s result: %w", provider, operation, err)
	}

	return result, nil
}

// ── Host operations ──────────────────────────────────────────────────

// GetHostname returns the container's hostname.
func (cp *ContainerProvider) GetHostname(
	ctx context.Context,
) (string, error) {
	return runScalar[string](ctx, cp, "host", "get-hostname")
}

// GetOSInfo returns the container's OS distribution and version.
func (cp *ContainerProvider) GetOSInfo(
	ctx context.Context,
) (*HostOSInfo, error) {
	return run[HostOSInfo](ctx, cp, "host", "get-os-info", nil)
}

// GetArchitecture returns the CPU architecture (e.g., x86_64, aarch64).
func (cp *ContainerProvider) GetArchitecture(
	ctx context.Context,
) (string, error) {
	return runScalar[string](ctx, cp, "host", "get-architecture")
}

// GetKernelVersion returns the kernel version string.
func (cp *ContainerProvider) GetKernelVersion(
	ctx context.Context,
) (string, error) {
	return runScalar[string](ctx, cp, "host", "get-kernel-version")
}

// GetUptime returns the system uptime as a duration.
func (cp *ContainerProvider) GetUptime(
	ctx context.Context,
) (time.Duration, error) {
	return runScalar[time.Duration](ctx, cp, "host", "get-uptime")
}

// GetFQDN returns the fully qualified domain name.
func (cp *ContainerProvider) GetFQDN(
	ctx context.Context,
) (string, error) {
	return runScalar[string](ctx, cp, "host", "get-fqdn")
}

// GetCPUCount returns the number of logical CPUs.
func (cp *ContainerProvider) GetCPUCount(
	ctx context.Context,
) (int, error) {
	return runScalar[int](ctx, cp, "host", "get-cpu-count")
}

// GetServiceManager returns the system service manager (e.g., systemd).
func (cp *ContainerProvider) GetServiceManager(
	ctx context.Context,
) (string, error) {
	return runScalar[string](ctx, cp, "host", "get-service-manager")
}

// GetPackageManager returns the system package manager (e.g., apt, dnf).
func (cp *ContainerProvider) GetPackageManager(
	ctx context.Context,
) (string, error) {
	return runScalar[string](ctx, cp, "host", "get-package-manager")
}

// ── Memory operations ────────────────────────────────────────────────

// GetMemStats returns memory statistics.
func (cp *ContainerProvider) GetMemStats(
	ctx context.Context,
) (*MemStats, error) {
	return run[MemStats](ctx, cp, "mem", "get-stats", nil)
}

// ── Load operations ──────────────────────────────────────────────────

// GetLoadStats returns load average statistics.
func (cp *ContainerProvider) GetLoadStats(
	ctx context.Context,
) (*LoadStats, error) {
	return run[LoadStats](ctx, cp, "load", "get-average-stats", nil)
}

// ── Disk operations ──────────────────────────────────────────────────

// GetDiskUsage returns disk usage statistics for all local mounts.
func (cp *ContainerProvider) GetDiskUsage(
	ctx context.Context,
) ([]DiskUsage, error) {
	out, err := cp.target.ExecProvider(ctx, "disk", "get-local-usage-stats", nil)
	if err != nil {
		return nil, err
	}

	var result []DiskUsage
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("unmarshal disk/get-local-usage-stats result: %w", err)
	}

	return result, nil
}

// ── Command operations ───────────────────────────────────────────────

// Exec runs a command inside the container via the command provider.
func (cp *ContainerProvider) Exec(
	ctx context.Context,
	params ExecParams,
) (*CommandResult, error) {
	return run[CommandResult](ctx, cp, "command", "exec", &params)
}

// Shell runs a shell command inside the container via the command provider.
func (cp *ContainerProvider) Shell(
	ctx context.Context,
	params ShellParams,
) (*CommandResult, error) {
	return run[CommandResult](ctx, cp, "command", "shell", &params)
}

// ── Ping operations ──────────────────────────────────────────────────

// Ping pings an address from inside the container.
func (cp *ContainerProvider) Ping(
	ctx context.Context,
	address string,
) (*PingResult, error) {
	return run[PingResult](ctx, cp, "ping", "do", &PingParams{Address: address})
}

// ── DNS operations ───────────────────────────────────────────────────

// GetDNS returns the DNS configuration for the given interface.
func (cp *ContainerProvider) GetDNS(
	ctx context.Context,
	interfaceName string,
) (*DNSGetResult, error) {
	return run[DNSGetResult](ctx, cp, "dns", "get-resolv-conf", &DNSGetParams{
		InterfaceName: interfaceName,
	})
}

// UpdateDNS updates the DNS configuration for the given interface.
func (cp *ContainerProvider) UpdateDNS(
	ctx context.Context,
	params DNSUpdateParams,
) (*DNSUpdateResult, error) {
	return run[DNSUpdateResult](ctx, cp, "dns", "update-resolv-conf", &params)
}
