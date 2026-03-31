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

// TimezoneResult represents timezone information from a query operation.
type TimezoneResult struct {
	Hostname  string `json:"hostname"`
	Status    string `json:"status"`
	Timezone  string `json:"timezone,omitempty"`
	UTCOffset string `json:"utc_offset,omitempty"`
	Error     string `json:"error,omitempty"`
}

// TimezoneMutationResult represents the result of a timezone update.
type TimezoneMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Timezone string `json:"timezone,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// TimezoneUpdateOpts contains options for updating the system timezone.
type TimezoneUpdateOpts struct {
	// Timezone is the IANA timezone name (e.g., "America/New_York"). Required.
	Timezone string
}

// timezoneCollectionFromGen converts a gen.TimezoneCollectionResponse
// to a Collection[TimezoneResult].
func timezoneCollectionFromGen(
	g *gen.TimezoneCollectionResponse,
) Collection[TimezoneResult] {
	results := make([]TimezoneResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, TimezoneResult{
			Hostname:  r.Hostname,
			Status:    string(r.Status),
			Timezone:  derefString(r.Timezone),
			UTCOffset: derefString(r.UtcOffset),
			Error:     derefString(r.Error),
		})
	}

	return Collection[TimezoneResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// timezoneMutationCollectionFromUpdate converts a gen.TimezoneUpdateResponse
// to a Collection[TimezoneMutationResult].
func timezoneMutationCollectionFromUpdate(
	g *gen.TimezoneUpdateResponse,
) Collection[TimezoneMutationResult] {
	results := make([]TimezoneMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, TimezoneMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Timezone: derefString(r.Timezone),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[TimezoneMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
