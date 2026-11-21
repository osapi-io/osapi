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
	logProv "github.com/osapi-io/osapi/internal/provider/node/log"
	"github.com/osapi-io/osapi/internal/provider/node/ntp"
	"github.com/osapi-io/osapi/internal/provider/node/power"
	processProv "github.com/osapi-io/osapi/internal/provider/node/process"
	serviceProv "github.com/osapi-io/osapi/internal/provider/node/service"
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
	ntpProvider ntp.Provider,
	powerProvider power.Provider,
	processProvider processProv.Provider,
	userProvider user.Provider,
	logProvider logProv.Provider,
	serviceProvider serviceProv.Provider,
	logger *slog.Logger,
) ProcessorFunc {
	logger = logger.With(slog.String("subsystem", "agent.processor.node"))

	return func(ctx context.Context, req job.Request) (json.RawMessage, error) {
		// Extract base operation from dotted operation (e.g., "ntp.get" -> "ntp")
		baseOperation := strings.Split(req.Operation, ".")[0]

		switch baseOperation {
		case "ntp":
			return processNtpOperation(ctx, ntpProvider, logger, req)
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
		case "log":
			return processLogOperation(ctx, logProvider, logger, req)
		case "service":
			return processServiceOperation(ctx, serviceProvider, logger, req)
		default:
			return nil, fmt.Errorf("unsupported node operation: %s", req.Operation)
		}
	}
}
