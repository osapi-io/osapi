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
	ntpProvider ntp.Provider,
	timezoneProvider timezone.Provider,
	powerProvider power.Provider,
	processProvider processProv.Provider,
	userProvider user.Provider,
	packageProvider apt.Provider,
	logProvider logProv.Provider,
	serviceProvider serviceProv.Provider,
	logger *slog.Logger,
) ProcessorFunc {
	logger = logger.With(slog.String("subsystem", "agent.processor.node"))

	return func(ctx context.Context, req job.Request) (json.RawMessage, error) {
		// Extract base operation from dotted operation (e.g., "ntp.get" -> "ntp")
		baseOperation := strings.Split(req.Operation, ".")[0]

		switch baseOperation {
		case "status":
			return getNodeStatus(hostProvider, diskProvider, memoryProvider, loadProvider, logger)
		case "ntp":
			return processNtpOperation(ctx, ntpProvider, logger, req)
		case "timezone":
			return processTimezoneOperation(ctx, timezoneProvider, logger, req)
		case "power":
			return processPowerOperation(ctx, powerProvider, logger, req)
		case "process":
			return processProcessOperation(ctx, processProvider, logger, req)
		case "user":
			return processUserOperation(ctx, userProvider, logger, req)
		case "group":
			return processGroupOperation(ctx, userProvider, logger, req)
		case "sshKey":
			return processSSHKeyOperation(ctx, userProvider, logger, req)
		case "package":
			return processPackageOperation(ctx, packageProvider, logger, req)
		case "log":
			return processLogOperation(ctx, logProvider, logger, req)
		case "service":
			return processServiceOperation(ctx, serviceProvider, logger, req)
		default:
			return nil, fmt.Errorf("unsupported node operation: %s", req.Operation)
		}
	}
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
