// Copyright (c) 2024 John Dewey

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

package dns

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/retr0h/osapi/internal/provider/network/netplan"
)

// UpdateResolvConfByInterface updates the DNS configuration for a specific
// network interface by generating a Netplan drop-in file and applying it.
// The function preserves existing settings for values that are not specified
// and delegates idempotency to the Netplan state tracker.
//
// The read path still uses resolvectl to query current DNS state. The write
// path generates a Netplan YAML file under /etc/netplan/osapi-dns.yaml and
// applies it via `netplan generate` + `netplan apply`.
func (u *Debian) UpdateResolvConfByInterface(
	servers []string,
	searchDomains []string,
	interfaceName string,
	overrideDHCP bool,
) (*UpdateResult, error) {
	u.logger.Info(
		"setting dns configuration via netplan",
		slog.String("servers", strings.Join(servers, ", ")),
		slog.String("search_domains", strings.Join(searchDomains, ", ")),
	)

	if len(servers) == 0 && len(searchDomains) == 0 {
		return nil, fmt.Errorf("no DNS servers or search domains provided; nothing to update")
	}

	existingConfig, err := u.GetResolvConfByInterface(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get current resolvectl configuration: %w", err)
	}

	// Use existing values if new values are not provided.
	if len(servers) == 0 {
		servers = existingConfig.DNSServers
	}
	if len(searchDomains) == 0 {
		searchDomains = existingConfig.SearchDomains
	}

	// Filter out root domain marker before generating YAML.
	filteredDomains := make([]string, 0, len(searchDomains))
	for _, domain := range searchDomains {
		if domain != "." {
			filteredDomains = append(filteredDomains, domain)
		}
	}

	// Resolve the interface name for the Netplan config.
	resolvedInterface := u.resolvePrimaryInterface(interfaceName)

	// Detect the interface type from netplan status to use the
	// correct YAML section (ethernets, wifis, bridges, etc.).
	ifaceType := "ethernets"

	status, statusErr := netplan.GetStatus(u.execManager)
	if statusErr == nil {
		if iface, ok := status[resolvedInterface]; ok {
			ifaceType = netplanSectionForType(iface.Type)
		}
	}

	// Generate the Netplan YAML content.
	content := generateDNSNetplanYAML(
		resolvedInterface,
		ifaceType,
		servers,
		filteredDomains,
		overrideDHCP,
	)

	// Apply via the shared Netplan helper (handles write, validate,
	// apply, and KV state tracking with SHA-based idempotency).
	changed, applyErr := netplan.ApplyConfig(
		context.TODO(),
		u.logger,
		u.fs,
		u.stateKV,
		u.execManager,
		u.hostname,
		dnsNetplanPath(),
		content,
		map[string]string{
			"domain":    "dns",
			"interface": resolvedInterface,
		},
	)
	if applyErr != nil {
		return nil, fmt.Errorf("dns update via netplan: %w", applyErr)
	}

	return &UpdateResult{Changed: changed}, nil
}
