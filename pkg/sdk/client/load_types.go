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

// LoadResult represents load average query result from a single agent.
type LoadResult struct {
	Hostname    string       `json:"hostname"`
	Status      string       `json:"status"`
	Error       string       `json:"error,omitempty"`
	Changed     bool         `json:"changed"`
	LoadAverage *LoadAverage `json:"load_average,omitempty"`
}

// loadAverageFromGen converts a gen.LoadAverageResponse to a LoadAverage.
func loadAverageFromGen(
	g *gen.LoadAverageResponse,
) *LoadAverage {
	if g == nil {
		return nil
	}

	return &LoadAverage{
		OneMin:     g.N1min,
		FiveMin:    g.N5min,
		FifteenMin: g.N15min,
	}
}

// loadCollectionFromGen converts a gen.LoadCollectionResponse to a Collection[LoadResult].
func loadCollectionFromGen(
	g *gen.LoadCollectionResponse,
) Collection[LoadResult] {
	results := make([]LoadResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, LoadResult{
			Hostname:    r.Hostname,
			Status:      string(r.Status),
			Error:       derefString(r.Error),
			Changed:     derefBool(r.Changed),
			LoadAverage: loadAverageFromGen(r.LoadAverage),
		})
	}

	return Collection[LoadResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
