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

package worker

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/provider/system/disk"
	systemHost "github.com/retr0h/osapi/internal/provider/system/host"
	"github.com/retr0h/osapi/internal/provider/system/load"
	"github.com/retr0h/osapi/internal/provider/system/mem"
)

// processJobOperation handles the actual job processing based on category and operation.
func (w *Worker) processJobOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	switch jobRequest.Category {
	case "system":
		return w.processSystemOperation(jobRequest)
	case "network":
		return w.processNetworkOperation(jobRequest)
	default:
		return nil, fmt.Errorf("unsupported job category: %s", jobRequest.Category)
	}
}

// processSystemOperation handles system-related operations.
func (w *Worker) processSystemOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	// Extract base operation from dotted operation (e.g., "hostname.get" -> "hostname")
	baseOperation := strings.Split(jobRequest.Operation, ".")[0]

	switch baseOperation {
	case "hostname":
		return w.getSystemHostname()
	case "status":
		return w.getSystemStatus()
	case "uptime":
		return w.getSystemUptime()
	case "os", "osinfo":
		return w.getSystemOSInfo()
	case "disk":
		return w.getSystemDisk()
	case "memory", "mem":
		return w.getSystemMemory()
	case "load":
		return w.getSystemLoad()
	default:
		return nil, fmt.Errorf("unsupported system operation: %s", jobRequest.Operation)
	}
}

// processNetworkOperation handles network-related operations.
func (w *Worker) processNetworkOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	// Extract base operation from dotted operation (e.g., "dns.get" -> "dns")
	baseOperation := strings.Split(jobRequest.Operation, ".")[0]

	switch baseOperation {
	case "dns":
		return w.processNetworkDNS(jobRequest)
	case "ping":
		return w.processNetworkPing(jobRequest)
	default:
		return nil, fmt.Errorf("unsupported network operation: %s", jobRequest.Operation)
	}
}

// getSystemHostname retrieves the system hostname and worker labels.
func (w *Worker) getSystemHostname() (json.RawMessage, error) {
	hostProvider := w.getHostProvider()
	hostname, err := hostProvider.GetHostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	result := map[string]interface{}{
		"hostname": hostname,
	}

	if len(w.appConfig.Job.Worker.Labels) > 0 {
		result["labels"] = w.appConfig.Job.Worker.Labels
	}

	return json.Marshal(result)
}

// getSystemStatus retrieves comprehensive system status.
func (w *Worker) getSystemStatus() (json.RawMessage, error) {
	hostProvider := w.getHostProvider()
	diskProvider := w.getDiskProvider()
	memProvider := w.getMemProvider()
	loadProvider := w.getLoadProvider()

	// Get all system information
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

// getSystemUptime retrieves the system uptime.
func (w *Worker) getSystemUptime() (json.RawMessage, error) {
	hostProvider := w.getHostProvider()
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

// getSystemOSInfo retrieves the operating system information.
func (w *Worker) getSystemOSInfo() (json.RawMessage, error) {
	hostProvider := w.getHostProvider()
	osInfo, err := hostProvider.GetOSInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get OS info: %w", err)
	}

	return json.Marshal(osInfo)
}

// getSystemDisk retrieves disk usage statistics.
func (w *Worker) getSystemDisk() (json.RawMessage, error) {
	diskProvider := w.getDiskProvider()
	diskUsage, err := diskProvider.GetLocalUsageStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	result := map[string]interface{}{
		"disks": diskUsage,
	}

	return json.Marshal(result)
}

// getSystemMemory retrieves memory statistics.
func (w *Worker) getSystemMemory() (json.RawMessage, error) {
	memProvider := w.getMemProvider()
	memInfo, err := memProvider.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	return json.Marshal(memInfo)
}

// getSystemLoad retrieves load average statistics.
func (w *Worker) getSystemLoad() (json.RawMessage, error) {
	loadProvider := w.getLoadProvider()
	loadAvg, err := loadProvider.GetAverageStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get load averages: %w", err)
	}

	return json.Marshal(loadAvg)
}

// processNetworkDNS handles DNS configuration operations.
func (w *Worker) processNetworkDNS(
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

		dnsProvider := w.getDNSProvider()
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

	dnsProvider := w.getDNSProvider()
	err := dnsProvider.UpdateResolvConfByInterface(serverStrings, searchStrings, interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to set DNS config: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"message": "DNS configuration updated successfully",
	}

	return json.Marshal(result)
}

// processNetworkPing handles ping operations.
func (w *Worker) processNetworkPing(
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

	pingProvider := w.getPingProvider()
	result, err := pingProvider.Do(address)
	if err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return json.Marshal(result)
}

// Provider accessor methods that return the injected providers
func (w *Worker) getHostProvider() systemHost.Provider {
	return w.hostProvider
}

func (w *Worker) getDiskProvider() disk.Provider {
	return w.diskProvider
}

func (w *Worker) getMemProvider() mem.Provider {
	return w.memProvider
}

func (w *Worker) getLoadProvider() load.Provider {
	return w.loadProvider
}

func (w *Worker) getDNSProvider() dns.Provider {
	return w.dnsProvider
}

func (w *Worker) getPingProvider() ping.Provider {
	return w.pingProvider
}
