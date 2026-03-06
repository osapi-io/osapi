// Copyright (c) 2026 John Dewey
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package dns

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

// UpdateResolvConfByInterface updates the DNS configuration for a macOS
// network interface using `networksetup`. It resolves the interface name
// (e.g., "en0") to a network service name (e.g., "Wi-Fi") via
// `networksetup -listallhardwareports`, then applies DNS servers and
// search domains.
//
// This command requires root privileges on macOS.
func (d *Darwin) UpdateResolvConfByInterface(
	servers []string,
	searchDomains []string,
	interfaceName string,
) (*UpdateResult, error) {
	d.logger.Info(
		"setting DNS configuration via networksetup",
		slog.String("servers", strings.Join(servers, ", ")),
		slog.String("search_domains", strings.Join(searchDomains, ", ")),
		slog.String("interface", interfaceName),
	)

	if len(servers) == 0 && len(searchDomains) == 0 {
		return nil, fmt.Errorf("no DNS servers or search domains provided; nothing to update")
	}

	existingConfig, err := d.GetResolvConfByInterface(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get current DNS configuration: %w", err)
	}

	// Use existing values if new values are not provided
	if len(servers) == 0 {
		servers = existingConfig.DNSServers
	}
	if len(searchDomains) == 0 {
		searchDomains = existingConfig.SearchDomains
	}

	// Compare desired config against existing to detect no-op
	if slices.Equal(servers, existingConfig.DNSServers) &&
		slices.Equal(searchDomains, existingConfig.SearchDomains) {
		d.logger.Info("dns configuration unchanged, skipping update")
		return &UpdateResult{Changed: false}, nil
	}

	// Resolve interface name to network service name
	serviceName, err := d.resolveServiceName(interfaceName)
	if err != nil {
		return nil, err
	}

	// Set DNS servers
	if len(servers) > 0 {
		args := append([]string{"-setdnsservers", serviceName}, servers...)
		output, err := d.execManager.RunCmd("networksetup", args)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to set DNS servers with networksetup: %w - %s",
				err,
				output,
			)
		}
	}

	// Set search domains
	if len(searchDomains) > 0 {
		args := append([]string{"-setsearchdomains", serviceName}, searchDomains...)
		output, err := d.execManager.RunCmd("networksetup", args)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to set search domains with networksetup: %w - %s",
				err,
				output,
			)
		}
	}

	return &UpdateResult{Changed: true}, nil
}

// resolveServiceName maps a BSD interface name (e.g., "en0") to its
// macOS network service name (e.g., "Wi-Fi") by parsing the output of
// `networksetup -listallhardwareports`.
//
// Example output:
//
//	Hardware Port: Wi-Fi
//	Device: en0
//	Ethernet Address: a4:83:e7:1a:2b:3c
//
//	Hardware Port: Thunderbolt Ethernet
//	Device: en1
//	Ethernet Address: 00:11:22:33:44:55
func (d *Darwin) resolveServiceName(
	interfaceName string,
) (string, error) {
	output, err := d.execManager.RunCmd("networksetup", []string{"-listallhardwareports"})
	if err != nil {
		return "", fmt.Errorf("failed to list hardware ports: %w - %s", err, output)
	}

	var currentService string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Hardware Port:") {
			currentService = strings.TrimPrefix(line, "Hardware Port: ")
		}
		if strings.HasPrefix(line, "Device:") {
			device := strings.TrimSpace(strings.TrimPrefix(line, "Device:"))
			if device == interfaceName {
				return currentService, nil
			}
		}
	}

	return "", fmt.Errorf("no network service found for interface %q", interfaceName)
}
