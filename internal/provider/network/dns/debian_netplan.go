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

package dns

import (
	"fmt"
	"strings"
)

const (
	netplanDir    = "/etc/netplan"
	dnsFilePrefix = "osapi-dns"
)

func dnsNetplanPath() string {
	return netplanDir + "/" + dnsFilePrefix + ".yaml"
}

func generateDNSNetplanYAML(
	interfaceName string,
	servers []string,
	searchDomains []string,
) []byte {
	var b strings.Builder

	b.WriteString("network:\n")
	b.WriteString("  version: 2\n")
	b.WriteString("  ethernets:\n")
	b.WriteString(fmt.Sprintf("    %s:\n", interfaceName))
	b.WriteString("      nameservers:\n")

	if len(servers) > 0 {
		b.WriteString("        addresses:\n")
		for _, s := range servers {
			b.WriteString(fmt.Sprintf("          - %s\n", s))
		}
	}

	if len(searchDomains) > 0 {
		b.WriteString("        search:\n")
		for _, d := range searchDomains {
			b.WriteString(fmt.Sprintf("          - %s\n", d))
		}
	}

	return []byte(b.String())
}

// resolvePrimaryInterface returns the network interface to use for
// Netplan configuration. It prefers the explicitly provided interface
// name, falls back to the primary_interface from agent facts, and
// defaults to "eth0" as a last resort.
func (u *Debian) resolvePrimaryInterface(
	interfaceName string,
) string {
	if interfaceName != "" {
		return interfaceName
	}

	facts := u.Facts()
	if facts != nil {
		if iface, ok := facts["primary_interface"].(string); ok && iface != "" {
			return iface
		}
	}

	return "eth0"
}
