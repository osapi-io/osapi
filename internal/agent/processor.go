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

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/node/disk"
	nodeHost "github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
	"github.com/retr0h/osapi/internal/provider/node/ntp"
	"github.com/retr0h/osapi/internal/provider/node/power"
	processProv "github.com/retr0h/osapi/internal/provider/node/process"
	"github.com/retr0h/osapi/internal/provider/node/sysctl"
	"github.com/retr0h/osapi/internal/provider/node/timezone"
	"github.com/retr0h/osapi/internal/provider/node/user"
)

// processJobOperation handles the actual job processing based on category and operation.
func (a *Agent) processJobOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	a.logger.Debug("dispatching to provider",
		slog.String("category", jobRequest.Category),
		slog.String("operation", jobRequest.Operation),
	)

	return a.registry.Dispatch(jobRequest)
}

// NewNodeProcessor returns a ProcessorFunc that handles node-related operations.
func NewNodeProcessor(
	hostProvider nodeHost.Provider,
	diskProvider disk.Provider,
	memProvider mem.Provider,
	loadProvider load.Provider,
	sysctlProvider sysctl.Provider,
	ntpProvider ntp.Provider,
	timezoneProvider timezone.Provider,
	powerProvider power.Provider,
	processProvider processProv.Provider,
	userProvider user.Provider,
	appConfig config.Config,
	logger *slog.Logger,
) ProcessorFunc {
	return func(req job.Request) (json.RawMessage, error) {
		// Extract base operation from dotted operation (e.g., "hostname.get" -> "hostname")
		baseOperation := strings.Split(req.Operation, ".")[0]

		switch baseOperation {
		case "hostname":
			if req.Type == job.TypeModify {
				return setNodeHostname(hostProvider, req, logger)
			}
			return getNodeHostname(hostProvider, appConfig, logger)
		case "status":
			return getNodeStatus(hostProvider, diskProvider, memProvider, loadProvider, logger)
		case "uptime":
			return getNodeUptime(hostProvider, logger)
		case "os", "osinfo":
			return getNodeOSInfo(hostProvider, logger)
		case "disk":
			return getNodeDisk(diskProvider, logger)
		case "memory", "mem":
			return getNodeMemory(memProvider, logger)
		case "load":
			return getNodeLoad(loadProvider, logger)
		case "sysctl":
			return processSysctlOperation(sysctlProvider, logger, req)
		case "ntp":
			return processNtpOperation(ntpProvider, logger, req)
		case "timezone":
			return processTimezoneOperation(timezoneProvider, logger, req)
		case "power":
			return processPowerOperation(powerProvider, logger, req)
		case "process":
			return processProcessOperation(processProvider, logger, req)
		case "user":
			return processUserOperation(userProvider, logger, req)
		case "group":
			return processGroupOperation(userProvider, logger, req)
		default:
			return nil, fmt.Errorf("unsupported node operation: %s", req.Operation)
		}
	}
}

// getNodeHostname retrieves the node hostname and agent labels.
func getNodeHostname(
	hostProvider nodeHost.Provider,
	appConfig config.Config,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing host.GetHostname")

	hostname, err := hostProvider.GetHostname()
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"hostname": hostname,
		"changed":  false,
	}

	if len(appConfig.Agent.Labels) > 0 {
		result["labels"] = appConfig.Agent.Labels
	}

	return json.Marshal(result)
}

// setNodeHostname sets the node hostname via the host provider.
func setNodeHostname(
	hostProvider nodeHost.Provider,
	req job.Request,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing host.UpdateHostname")

	var data struct {
		Hostname string `json:"hostname"`
	}
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return nil, fmt.Errorf("invalid hostname update data: %w", err)
	}

	result, err := hostProvider.UpdateHostname(data.Hostname)
	if err != nil {
		return nil, err
	}

	resp := map[string]interface{}{
		"hostname": data.Hostname,
		"changed":  result.Changed,
	}

	return json.Marshal(resp)
}

// getNodeStatus retrieves comprehensive node status.
func getNodeStatus(
	hostProvider nodeHost.Provider,
	diskProvider disk.Provider,
	memProvider mem.Provider,
	loadProvider load.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing node.GetStatus")

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
		"changed":       false,
	}

	return json.Marshal(result)
}

// getNodeUptime retrieves the system uptime.
func getNodeUptime(
	hostProvider nodeHost.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing host.GetUptime")

	uptime, err := hostProvider.GetUptime()
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"uptime_seconds": uptime.Seconds(),
		"uptime":         uptime.String(),
		"changed":        false,
	}

	return json.Marshal(result)
}

// getNodeOSInfo retrieves the operating system information.
func getNodeOSInfo(
	hostProvider nodeHost.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing host.GetOSInfo")

	osInfo, err := hostProvider.GetOSInfo()
	if err != nil {
		return nil, err
	}

	return json.Marshal(osInfo)
}

// getNodeDisk retrieves disk usage statistics.
func getNodeDisk(
	diskProvider disk.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing disk.GetLocalUsageStats")

	diskUsage, err := diskProvider.GetLocalUsageStats()
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"disks":   diskUsage,
		"changed": false,
	}

	return json.Marshal(result)
}

// getNodeMemory retrieves memory statistics.
func getNodeMemory(
	memProvider mem.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing mem.GetStats")

	memInfo, err := memProvider.GetStats()
	if err != nil {
		return nil, err
	}

	return json.Marshal(memInfo)
}

// getNodeLoad retrieves load average statistics.
func getNodeLoad(
	loadProvider load.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing load.GetAverageStats")

	loadAvg, err := loadProvider.GetAverageStats()
	if err != nil {
		return nil, err
	}

	return json.Marshal(loadAvg)
}
