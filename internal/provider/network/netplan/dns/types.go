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

// Provider implements the methods to interact with various DNS components.
type Provider interface {
	// GetResolvConfByInterface retrieves the DNS configuration.
	GetResolvConfByInterface(
		interfaceName string,
	) (*GetResult, error)
	// UpdateResolvConfByInterface updates the DNS configuration.
	// Returns an UpdateResult indicating whether the configuration was changed.
	// When overrideDHCP is true, DHCP-provided DNS servers are disabled so
	// only the configured servers are used.
	UpdateResolvConfByInterface(
		servers []string,
		searchDomains []string,
		interfaceName string,
		overrideDHCP bool,
	) (*UpdateResult, error)
	// DeleteNetplanConfig removes the managed DNS Netplan config file.
	DeleteNetplanConfig(
		interfaceName string,
	) (bool, error)
}

// GetResult represents the DNS configuration with servers and search domains.
type GetResult struct {
	// List of DNS server IP addresses (IPv4 or IPv6)
	DNSServers []string
	// List of search domains for DNS resolution
	SearchDomains []string
	// Changed indicates whether system state was modified.
	Changed bool `json:"changed"`
}

// UpdateResult represents the outcome of a DNS update operation.
type UpdateResult struct {
	// Changed indicates whether the DNS configuration was actually modified.
	Changed bool `json:"changed"`
}
