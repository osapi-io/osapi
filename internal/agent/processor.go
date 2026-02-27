// Copyright (c) 2025 John Dewey

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
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/command"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/provider/node/disk"
	nodeHost "github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
)

// processJobOperation handles the actual job processing based on category and operation.
func (a *Agent) processJobOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	a.logger.Debug("dispatching to provider",
		slog.String("category", jobRequest.Category),
		slog.String("operation", jobRequest.Operation),
	)

	switch jobRequest.Category {
	case "node":
		return a.processNodeOperation(jobRequest)
	case "network":
		return a.processNetworkOperation(jobRequest)
	case "command":
		return a.processCommandOperation(jobRequest)
	default:
		return nil, fmt.Errorf("unsupported job category: %s", jobRequest.Category)
	}
}

// processNodeOperation handles system-related operations.
func (a *Agent) processNodeOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	// Extract base operation from dotted operation (e.g., "hostname.get" -> "hostname")
	baseOperation := strings.Split(jobRequest.Operation, ".")[0]

	switch baseOperation {
	case "hostname":
		return a.getNodeHostname()
	case "status":
		return a.getNodeStatus()
	case "uptime":
		return a.getNodeUptime()
	case "os", "osinfo":
		return a.getNodeOSInfo()
	case "disk":
		return a.getNodeDisk()
	case "memory", "mem":
		return a.getNodeMemory()
	case "load":
		return a.getNodeLoad()
	default:
		return nil, fmt.Errorf("unsupported node operation: %s", jobRequest.Operation)
	}
}

// processNetworkOperation handles network-related operations.
func (a *Agent) processNetworkOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	// Extract base operation from dotted operation (e.g., "dns.get" -> "dns")
	baseOperation := strings.Split(jobRequest.Operation, ".")[0]

	switch baseOperation {
	case "dns":
		return a.processNetworkDNS(jobRequest)
	case "ping":
		return a.processNetworkPing(jobRequest)
	default:
		return nil, fmt.Errorf("unsupported network operation: %s", jobRequest.Operation)
	}
}

// getNodeHostname retrieves the node hostname and agent labels.
func (a *Agent) getNodeHostname() (json.RawMessage, error) {
	a.logger.Debug("executing host.GetHostname")
	hostProvider := a.getHostProvider()
	hostname, err := hostProvider.GetHostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	result := map[string]interface{}{
		"hostname": hostname,
	}

	if len(a.appConfig.Agent.Labels) > 0 {
		result["labels"] = a.appConfig.Agent.Labels
	}

	return json.Marshal(result)
}

// getNodeStatus retrieves comprehensive node status.
func (a *Agent) getNodeStatus() (json.RawMessage, error) {
	a.logger.Debug("executing node.GetStatus")
	hostProvider := a.getHostProvider()
	diskProvider := a.getDiskProvider()
	memProvider := a.getMemProvider()
	loadProvider := a.getLoadProvider()

	// Get all node information
	hostname, _ := hostProvider.GetHostname()
	osInfo, _ := hostProvider.GetOSInfo()
	uptime, _ := hostProvider.GetUptime()
	diskUsage, _ := diskProvider.GetLocalUsageStats()
	memInfo, _ := memProvider.GetStats()
	loadAvg, _ := loadProvider.GetAverageStats()

	result := map[string]interface{}{
		"hostname":      hostname,
		"os_info":       osInfo,
		"uptime":        uptime,
		"disk_usage":    diskUsage,
		"memory_stats":  memInfo,
		"load_averages": loadAvg,
	}

	return json.Marshal(result)
}

// getNodeUptime retrieves the system uptime.
func (a *Agent) getNodeUptime() (json.RawMessage, error) {
	a.logger.Debug("executing host.GetUptime")
	hostProvider := a.getHostProvider()
	uptime, err := hostProvider.GetUptime()
	if err != nil {
		return nil, fmt.Errorf("failed to get uptime: %w", err)
	}

	result := map[string]interface{}{
		"uptime_seconds": uptime.Seconds(),
		"uptime":         uptime.String(),
	}

	return json.Marshal(result)
}

// getNodeOSInfo retrieves the operating system information.
func (a *Agent) getNodeOSInfo() (json.RawMessage, error) {
	a.logger.Debug("executing host.GetOSInfo")
	hostProvider := a.getHostProvider()
	osInfo, err := hostProvider.GetOSInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get OS info: %w", err)
	}

	return json.Marshal(osInfo)
}

// getNodeDisk retrieves disk usage statistics.
func (a *Agent) getNodeDisk() (json.RawMessage, error) {
	a.logger.Debug("executing disk.GetLocalUsageStats")
	diskProvider := a.getDiskProvider()
	diskUsage, err := diskProvider.GetLocalUsageStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	result := map[string]interface{}{
		"disks": diskUsage,
	}

	return json.Marshal(result)
}

// getNodeMemory retrieves memory statistics.
func (a *Agent) getNodeMemory() (json.RawMessage, error) {
	a.logger.Debug("executing mem.GetStats")
	memProvider := a.getMemProvider()
	memInfo, err := memProvider.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	return json.Marshal(memInfo)
}

// getNodeLoad retrieves load average statistics.
func (a *Agent) getNodeLoad() (json.RawMessage, error) {
	a.logger.Debug("executing load.GetAverageStats")
	loadProvider := a.getLoadProvider()
	loadAvg, err := loadProvider.GetAverageStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get load averages: %w", err)
	}

	return json.Marshal(loadAvg)
}

// processNetworkDNS handles DNS configuration operations.
func (a *Agent) processNetworkDNS(
	jobRequest job.Request,
) (json.RawMessage, error) {
	var dnsData map[string]interface{}
	if err := json.Unmarshal(jobRequest.Data, &dnsData); err != nil {
		return nil, fmt.Errorf("failed to parse DNS data: %w", err)
	}

	if jobRequest.Type == job.TypeQuery {
		// Get DNS configuration
		interfaceName, _ := dnsData["interface"].(string)
		if interfaceName == "" {
			interfaceName = "eth0" // Default interface
		}

		a.logger.Debug("executing dns.GetResolvConfByInterface",
			slog.String("interface", interfaceName),
		)
		dnsProvider := a.getDNSProvider()
		config, err := dnsProvider.GetResolvConfByInterface(interfaceName)
		if err != nil {
			return nil, fmt.Errorf("failed to get DNS config: %w", err)
		}

		return json.Marshal(config)
	}

	// Set DNS configuration
	servers, _ := dnsData["servers"].([]interface{})
	searchDomains, _ := dnsData["search_domains"].([]interface{})
	interfaceName, _ := dnsData["interface"].(string)

	var serverStrings []string
	for _, s := range servers {
		if str, ok := s.(string); ok {
			serverStrings = append(serverStrings, str)
		}
	}

	var searchStrings []string
	for _, s := range searchDomains {
		if str, ok := s.(string); ok {
			searchStrings = append(searchStrings, str)
		}
	}

	a.logger.Debug("executing dns.UpdateResolvConfByInterface",
		slog.String("interface", interfaceName),
		slog.Int("servers", len(serverStrings)),
	)
	dnsProvider := a.getDNSProvider()
	dnsResult, err := dnsProvider.UpdateResolvConfByInterface(
		serverStrings,
		searchStrings,
		interfaceName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set DNS config: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"changed": dnsResult.Changed,
		"message": "DNS configuration updated successfully",
	}

	return json.Marshal(result)
}

// processNetworkPing handles ping operations.
func (a *Agent) processNetworkPing(
	jobRequest job.Request,
) (json.RawMessage, error) {
	var pingData map[string]interface{}
	if err := json.Unmarshal(jobRequest.Data, &pingData); err != nil {
		return nil, fmt.Errorf("failed to parse ping data: %w", err)
	}

	address, ok := pingData["address"].(string)
	if !ok {
		return nil, fmt.Errorf("missing ping address")
	}

	a.logger.Debug("executing ping.Do",
		slog.String("address", address),
	)
	pingProvider := a.getPingProvider()
	result, err := pingProvider.Do(address)
	if err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return json.Marshal(result)
}

// Provider accessor methods that return the injected providers
func (a *Agent) getHostProvider() nodeHost.Provider {
	return a.hostProvider
}

func (a *Agent) getDiskProvider() disk.Provider {
	return a.diskProvider
}

func (a *Agent) getMemProvider() mem.Provider {
	return a.memProvider
}

func (a *Agent) getLoadProvider() load.Provider {
	return a.loadProvider
}

func (a *Agent) getDNSProvider() dns.Provider {
	return a.dnsProvider
}

func (a *Agent) getPingProvider() ping.Provider {
	return a.pingProvider
}

func (a *Agent) getCommandProvider() command.Provider {
	return a.commandProvider
}
