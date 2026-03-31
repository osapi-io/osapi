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

// PowerResult represents the result of a power operation for one host.
type PowerResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Action   string `json:"action,omitempty"`
	Delay    int    `json:"delay,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// PowerOpts contains options for power operations (reboot, shutdown).
type PowerOpts struct {
	// Delay is the number of seconds to wait before executing the operation.
	Delay int
	// Message is an optional message to broadcast before the operation.
	Message string
}

// powerCollectionFromReboot converts a gen.PowerRebootResponse
// to a Collection[PowerResult].
func powerCollectionFromReboot(
	g *gen.PowerRebootResponse,
) Collection[PowerResult] {
	results := make([]PowerResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, powerResultFromGen(r))
	}

	return Collection[PowerResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// powerCollectionFromShutdown converts a gen.PowerShutdownResponse
// to a Collection[PowerResult].
func powerCollectionFromShutdown(
	g *gen.PowerShutdownResponse,
) Collection[PowerResult] {
	results := make([]PowerResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, powerResultFromGen(r))
	}

	return Collection[PowerResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// powerResultFromGen converts a single gen.PowerResult to a PowerResult.
func powerResultFromGen(
	r gen.PowerResult,
) PowerResult {
	return PowerResult{
		Hostname: r.Hostname,
		Status:   string(r.Status),
		Action:   derefString(r.Action),
		Delay:    derefInt(r.Delay),
		Changed:  derefBool(r.Changed),
		Error:    derefString(r.Error),
	}
}
