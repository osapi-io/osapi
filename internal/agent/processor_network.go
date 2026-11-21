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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/provider/network/netplan/iface"
	"github.com/osapi-io/osapi/internal/provider/network/netplan/route"
	"github.com/osapi-io/osapi/internal/provider/network/ping"
)

// NewNetworkProcessor returns a ProcessorFunc that handles network-related operations.
func NewNetworkProcessor(
	pingProvider ping.Provider,
	interfaceProvider iface.Provider,
	routeProvider route.Provider,
	logger *slog.Logger,
) ProcessorFunc {
	logger = logger.With(slog.String("subsystem", "agent.processor.network"))

	return func(ctx context.Context, req job.Request) (json.RawMessage, error) {
		// Extract base operation from dotted operation (e.g., "ping.do" -> "ping")
		baseOperation := strings.Split(req.Operation, ".")[0]

		switch baseOperation {
		case "ping":
			return processNetworkPing(ctx, pingProvider, logger, req)
		case "interface":
			return processInterfaceOperation(ctx, interfaceProvider, logger, req)
		case "route":
			return processRouteOperation(ctx, routeProvider, logger, req)
		default:
			return nil, fmt.Errorf("unsupported network operation: %s", req.Operation)
		}
	}
}

// processNetworkPing handles ping operations.
func processNetworkPing(
	ctx context.Context,
	pingProvider ping.Provider,
	logger *slog.Logger,
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

	logger.Debug(
		"executing ping.Do",
		slog.String("address", address),
	)

	result, err := pingProvider.Do(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return json.Marshal(result)
}
