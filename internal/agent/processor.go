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
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/provider/node/apt"
	"github.com/osapi-io/osapi/internal/provider/node/disk"
	nodeHost "github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	logProv "github.com/osapi-io/osapi/internal/provider/node/log"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
	"github.com/osapi-io/osapi/internal/provider/node/ntp"
	"github.com/osapi-io/osapi/internal/provider/node/power"
	processProv "github.com/osapi-io/osapi/internal/provider/node/process"
	serviceProv "github.com/osapi-io/osapi/internal/provider/node/service"
	"github.com/osapi-io/osapi/internal/provider/node/sysctl"
	"github.com/osapi-io/osapi/internal/provider/node/timezone"
	"github.com/osapi-io/osapi/internal/provider/node/user"
)

// processJobOperation handles the actual job processing based on category and operation.
func (a *Agent) processJobOperation(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	a.logger.Debug(
		"dispatching to provider",
		slog.String("category", jobRequest.Category),
		slog.String("operation", jobRequest.Operation),
	)

	return a.registry.Dispatch(ctx, jobRequest)
}

// NewNodeProcessor returns a ProcessorFunc that handles node-related operations.
func NewNodeProcessor(
	hostProvider nodeHost.Provider,
	diskProvider disk.Provider,
	memoryProvider memory.Provider,
	loadProvider load.Provider,
	sysctlProvider sysctl.Provider,
	ntpProvider ntp.Provider,
	timezoneProvider timezone.Provider,
	powerProvider power.Provider,
	processProvider processProv.Provider,
	userProvider user.Provider,
	packageProvider apt.Provider,
	logProvider logProv.Provider,
	serviceProvider serviceProv.Provider,
	appConfig config.Config,
	logger *slog.Logger,
) ProcessorFunc {
	return func(ctx context.Context, req job.Request) (json.RawMessage, error) {
		// Extract base operation from dotted operation (e.g., "hostname.get" -> "hostname")
		baseOperation := strings.Split(req.Operation, ".")[0]

		switch baseOperation {
		case "hostname":
			if req.Type == job.TypeModify {
				return setNodeHostname(ctx, hostProvider, req, logger)
			}
			return getNodeHostname(hostProvider, appConfig, logger)
		case "status":
			return getNodeStatus(hostProvider, diskProvider, memoryProvider, loadProvider, logger)
		case "uptime":
			return getNodeUptime(hostProvider, logger)
		case "os", "osinfo":
			return getNodeOSInfo(hostProvider, logger)
		case "disk":
			return getNodeDisk(diskProvider, logger)
		case "memory", "mem":
			return getNodeMemory(memoryProvider, logger)
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
		case "sshKey":
			return processSSHKeyOperation(userProvider, logger, req)
		case "package":
			return processPackageOperation(packageProvider, logger, req)
		case "log":
			return processLogOperation(logProvider, logger, req)
		case "service":
			return processServiceOperation(serviceProvider, logger, req)
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
	ctx context.Context,
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

	result, err := hostProvider.UpdateHostname(ctx, data.Hostname)
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
	memoryProvider memory.Provider,
	loadProvider load.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing node.GetStatus")

	// Six independent reads. One failing is not a reason to answer nothing, so
	// each is recorded under the field it would have filled and the rest are
	// returned: a zero-valued field otherwise reads as a host with nothing to
	// report.
	fieldErrors := map[string]string{}

	record := func(field string, err error) {
		if err != nil {
			fieldErrors[field] = err.Error()
			logger.Warn(
				"node status read failed",
				slog.String("field", field),
				slog.String("error", err.Error()),
			)
		}
	}

	hostname, err := hostProvider.GetHostname()
	record("hostname", err)

	osInfo, err := hostProvider.GetOSInfo()
	record("os_info", err)

	uptime, err := hostProvider.GetUptime()
	record("uptime", err)

	diskUsage, err := diskProvider.GetLocalUsageStats()
	record("disk_usage", err)

	memInfo, err := memoryProvider.GetStats()
	record("memory_stats", err)

	loadAvg, err := loadProvider.GetAverageStats()
	record("load_averages", err)

	result := map[string]interface{}{
		"hostname":      hostname,
		"os_info":       osInfo,
		"uptime":        uptime,
		"disk_usage":    diskUsage,
		"memory_stats":  memInfo,
		"load_averages": loadAvg,
		"changed":       false,
	}

	if len(fieldErrors) > 0 {
		result["field_errors"] = fieldErrors
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
	memoryProvider memory.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing memory.GetStats")

	memInfo, err := memoryProvider.GetStats()
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
