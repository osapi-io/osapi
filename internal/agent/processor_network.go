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
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/network/netplan/dns"
	"github.com/retr0h/osapi/internal/provider/network/netplan/iface"
	"github.com/retr0h/osapi/internal/provider/network/netplan/route"
	"github.com/retr0h/osapi/internal/provider/network/ping"
)

// NewNetworkProcessor returns a ProcessorFunc that handles network-related operations.
func NewNetworkProcessor(
	dnsProvider dns.Provider,
	pingProvider ping.Provider,
	interfaceProvider iface.Provider,
	routeProvider route.Provider,
	logger *slog.Logger,
) ProcessorFunc {
	return func(req job.Request) (json.RawMessage, error) {
		// Extract base operation from dotted operation (e.g., "dns.get" -> "dns")
		baseOperation := strings.Split(req.Operation, ".")[0]

		switch baseOperation {
		case "dns":
			return processNetworkDNS(dnsProvider, logger, req)
		case "ping":
			return processNetworkPing(pingProvider, logger, req)
		case "interface":
			return processInterfaceOperation(interfaceProvider, logger, req)
		case "route":
			return processRouteOperation(routeProvider, logger, req)
		default:
			return nil, fmt.Errorf("unsupported network operation: %s", req.Operation)
		}
	}
}

// processNetworkDNS handles DNS configuration operations.
func processNetworkDNS(
	dnsProvider dns.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var dnsData map[string]interface{}
	if err := json.Unmarshal(jobRequest.Data, &dnsData); err != nil {
		return nil, fmt.Errorf("failed to parse DNS data: %w", err)
	}

	// Extract sub-operation: "dns.delete" -> "delete"
	parts := strings.Split(jobRequest.Operation, ".")
	subOp := ""
	if len(parts) >= 2 {
		subOp = parts[1]
	}

	if jobRequest.Type == job.TypeQuery {
		// Get DNS configuration
		interfaceName, _ := dnsData["interface"].(string)
		if interfaceName == "" {
			interfaceName = "eth0" // Default interface
		}

		logger.Debug("executing dns.GetResolvConfByInterface",
			slog.String("interface", interfaceName),
		)

		config, err := dnsProvider.GetResolvConfByInterface(interfaceName)
		if err != nil {
			return nil, err
		}

		return json.Marshal(config)
	}

	// Delete DNS configuration
	if subOp == "delete" {
		interfaceName, _ := dnsData["interface"].(string)

		logger.Debug("executing dns.DeleteResolvConf",
			slog.String("interface", interfaceName),
		)

		changed, err := dnsProvider.DeleteNetplanConfig(interfaceName)
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"success": true,
			"changed": changed,
			"message": "DNS configuration deleted",
		}

		return json.Marshal(result)
	}

	// Set DNS configuration
	servers, _ := dnsData["servers"].([]interface{})
	searchDomains, _ := dnsData["search_domains"].([]interface{})
	interfaceName, _ := dnsData["interface"].(string)
	overrideDHCP, _ := dnsData["override_dhcp"].(bool)

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

	logger.Debug("executing dns.UpdateResolvConfByInterface",
		slog.String("interface", interfaceName),
		slog.Int("servers", len(serverStrings)),
	)

	dnsResult, err := dnsProvider.UpdateResolvConfByInterface(
		serverStrings,
		searchStrings,
		interfaceName,
		overrideDHCP,
	)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"success": true,
		"changed": dnsResult.Changed,
		"message": "DNS configuration updated successfully",
	}

	return json.Marshal(result)
}

// processNetworkPing handles ping operations.
func processNetworkPing(
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

	logger.Debug("executing ping.Do",
		slog.String("address", address),
	)

	result, err := pingProvider.Do(address)
	if err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return json.Marshal(result)
}
