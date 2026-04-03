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
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// DNSConfig represents DNS configuration from a single agent.
type DNSConfig struct {
	Hostname      string   `json:"hostname"`
	Status        string   `json:"status"`
	Error         string   `json:"error,omitempty"`
	Changed       bool     `json:"changed"`
	Servers       []string `json:"servers,omitempty"`
	SearchDomains []string `json:"search_domains,omitempty"`
}

// DNSUpdateResult represents DNS update result from a single agent.
type DNSUpdateResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Changed  bool   `json:"changed"`
}

// DNSDeleteResult represents DNS delete result from a single agent.
type DNSDeleteResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Changed  bool   `json:"changed"`
}

// dnsConfigCollectionFromGen converts a gen.DNSConfigCollectionResponse to a Collection[DNSConfig].
func dnsConfigCollectionFromGen(
	g *gen.DNSConfigCollectionResponse,
) Collection[DNSConfig] {
	results := make([]DNSConfig, 0, len(g.Results))
	for _, r := range g.Results {
		dc := DNSConfig{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
		}

		if r.Servers != nil {
			dc.Servers = *r.Servers
		}

		if r.SearchDomains != nil {
			dc.SearchDomains = *r.SearchDomains
		}

		results = append(results, dc)
	}

	return Collection[DNSConfig]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// dnsUpdateCollectionFromGen converts a gen.DNSUpdateCollectionResponse to a Collection[DNSUpdateResult].
func dnsUpdateCollectionFromGen(
	g *gen.DNSUpdateCollectionResponse,
) Collection[DNSUpdateResult] {
	results := make([]DNSUpdateResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, DNSUpdateResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
		})
	}

	return Collection[DNSUpdateResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// dnsDeleteCollectionFromGen converts a gen.DNSDeleteCollectionResponse
// to a Collection[DNSDeleteResult].
func dnsDeleteCollectionFromGen(
	g *gen.DNSDeleteCollectionResponse,
) Collection[DNSDeleteResult] {
	results := make([]DNSDeleteResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, DNSDeleteResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
		})
	}

	return Collection[DNSDeleteResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
