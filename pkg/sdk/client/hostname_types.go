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

// HostnameResult represents a hostname query result from a single agent.
type HostnameResult struct {
	Hostname string            `json:"hostname"`
	Status   string            `json:"status"`
	Error    string            `json:"error,omitempty"`
	Changed  bool              `json:"changed"`
	Labels   map[string]string `json:"labels,omitempty"`
}

// HostnameUpdateResult represents a hostname update result from a single agent.
type HostnameUpdateResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Changed  bool   `json:"changed"`
}

// hostnameCollectionFromGen converts a gen.HostnameCollectionResponse to a Collection[HostnameResult].
func hostnameCollectionFromGen(
	g *gen.HostnameCollectionResponse,
) Collection[HostnameResult] {
	results := make([]HostnameResult, 0, len(g.Results))
	for _, r := range g.Results {
		hr := HostnameResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
		}

		if r.Labels != nil {
			hr.Labels = *r.Labels
		}

		results = append(results, hr)
	}

	return Collection[HostnameResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// hostnameUpdateCollectionFromGen converts a gen.HostnameUpdateCollectionResponse to a Collection[HostnameUpdateResult].
func hostnameUpdateCollectionFromGen(
	r *gen.HostnameUpdateCollectionResponse,
) Collection[HostnameUpdateResult] {
	results := make([]HostnameUpdateResult, 0, len(r.Results))
	for _, item := range r.Results {
		result := HostnameUpdateResult{
			Hostname: item.Hostname,
			Status:   string(item.Status),
			Changed:  derefBool(item.Changed),
			Error:    derefString(item.Error),
		}
		results = append(results, result)
	}

	return Collection[HostnameUpdateResult]{
		Results: results,
		JobID:   jobIDFromGen(r.JobId),
	}
}
