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

package client

import (
	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// DNSConfig is the DNS resolvers of an interface on one host.
type DNSConfig struct {
	Hostname      string   `json:"hostname"`
	Status        string   `json:"status"`
	Error         string   `json:"error,omitempty"`
	Changed       bool     `json:"changed"`
	Servers       []string `json:"servers,omitempty"`
	SearchDomains []string `json:"search_domains,omitempty"`
}

// DNSUpdateResult is an interface's resolvers after an update on one host.
type DNSUpdateResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Changed  bool   `json:"changed"`
}

// DNSDeleteResult is an interface's resolvers after a removal on one host.
type DNSDeleteResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Changed  bool   `json:"changed"`
}

func dnsConfigsFromGen(
	g *gen.NetworkDnsGetResponse,
) Collection[DNSConfig] {
	results := make([]DNSConfig, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, DNSConfig{
			Hostname:      r.Hostname,
			Status:        string(r.Status),
			Error:         derefString(r.Error),
			Servers:       derefStrings(r.Servers),
			SearchDomains: derefStrings(r.SearchDomains),
		})
	}

	return Collection[DNSConfig]{Results: results, JobID: jobIDFromGen(g.JobId)}
}

// dnsResults is the per-host rows of an update or a delete.
func dnsResults(
	rows []gen.NetworkDnsResult,
) []DNSUpdateResult {
	results := make([]DNSUpdateResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, DNSUpdateResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
		})
	}

	return results
}

// derefStrings is the list a pointer holds, or nil.
func derefStrings(
	p *[]string,
) []string {
	if p == nil {
		return nil
	}

	return *p
}
