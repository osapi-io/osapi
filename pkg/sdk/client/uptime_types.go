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

// UptimeResult represents uptime query result from a single agent.
type UptimeResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Uptime   string `json:"uptime,omitempty"`
	Error    string `json:"error,omitempty"`
	Changed  bool   `json:"changed"`
}

// uptimeCollectionFromGen converts a gen.UptimeCollectionResponse to a Collection[UptimeResult].
func uptimeCollectionFromGen(
	g *gen.UptimeCollectionResponse,
) Collection[UptimeResult] {
	results := make([]UptimeResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, UptimeResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Uptime:   derefString(r.Uptime),
			Error:    derefString(r.Error),
			Changed:  derefBool(r.Changed),
		})
	}

	return Collection[UptimeResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
