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

import "github.com/osapi-io/osapi/pkg/sdk/client/gen"

// LogEntryOpts selects journal entries. Each field is optional.
type LogEntryOpts struct {
	// Unit is a systemd unit, such as nginx.service.
	Unit string
	// Lines is how many of the newest entries; 100 when it is zero.
	Lines int
	// Since is a time, in any form that journalctl --since takes.
	Since string
	// Priority is a priority name, such as err: that one and above.
	Priority string
}

// LogEntryResult is one journal entry of one host. A host that failed or
// skipped has a row with its status and no entry.
type LogEntryResult struct {
	Hostname  string `json:"hostname"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp,omitempty"`
	Unit      string `json:"unit,omitempty"`
	Priority  string `json:"priority,omitempty"`
	Message   string `json:"message,omitempty"`
	PID       int    `json:"pid,omitempty"`
	Error     string `json:"error,omitempty"`
}

// LogSourceResult is one syslog identifier of one host's journal.
type LogSourceResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name,omitempty"`
	Error    string `json:"error,omitempty"`
}

// logEntryCollectionFromGen converts the rows of a log entry list.
func logEntryCollectionFromGen(
	g *gen.LogEntryCollectionResponse,
) Collection[LogEntryResult] {
	results := make([]LogEntryResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, LogEntryResult{
			Hostname:  r.Hostname,
			Status:    string(r.Status),
			Timestamp: derefString(r.Timestamp),
			Unit:      derefString(r.Unit),
			Priority:  derefString(r.Priority),
			Message:   derefString(r.Message),
			PID:       derefInt(r.Pid),
			Error:     derefString(r.Error),
		})
	}

	return Collection[LogEntryResult]{Results: results, JobID: jobIDFromGen(g.JobId)}
}

// logSourceCollectionFromGen converts the rows of a log source list.
func logSourceCollectionFromGen(
	g *gen.LogSourceCollectionResponse,
) Collection[LogSourceResult] {
	results := make([]LogSourceResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, LogSourceResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			Error:    derefString(r.Error),
		})
	}

	return Collection[LogSourceResult]{Results: results, JobID: jobIDFromGen(g.JobId)}
}
