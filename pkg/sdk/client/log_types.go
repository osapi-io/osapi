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

// LogEntryResult represents the result of a log query for one host.
type LogEntryResult struct {
	Hostname string     `json:"hostname"`
	Status   string     `json:"status"`
	Entries  []LogEntry `json:"entries,omitempty"`
	Error    string     `json:"error,omitempty"`
}

// LogEntry represents a single journal entry.
type LogEntry struct {
	Timestamp string `json:"timestamp,omitempty"`
	Unit      string `json:"unit,omitempty"`
	Priority  string `json:"priority,omitempty"`
	Message   string `json:"message,omitempty"`
	PID       int    `json:"pid,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
}

// LogQueryOpts contains options for log query operations.
type LogQueryOpts struct {
	// Lines is the maximum number of log lines to return.
	Lines *int
	// Since filters entries since this time (e.g., "1h", "2026-01-01 00:00:00").
	Since *string
	// Priority filters by log priority level (e.g., "err", "warning", "info").
	Priority *string
}

// logCollectionFromGen converts a gen.LogCollectionResponse to a
// Collection[LogEntryResult].
func logCollectionFromGen(
	g *gen.LogCollectionResponse,
) Collection[LogEntryResult] {
	results := make([]LogEntryResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, logEntryResultFromGen(r))
	}

	return Collection[LogEntryResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// logEntryResultFromGen converts a gen.LogResultEntry to a LogEntryResult.
func logEntryResultFromGen(
	r gen.LogResultEntry,
) LogEntryResult {
	result := LogEntryResult{
		Hostname: r.Hostname,
		Status:   string(r.Status),
		Error:    derefString(r.Error),
	}

	if r.Entries != nil {
		entries := make([]LogEntry, 0, len(*r.Entries))
		for _, e := range *r.Entries {
			entries = append(entries, logEntryInfoFromGen(e))
		}
		result.Entries = entries
	}

	return result
}

// logEntryInfoFromGen converts a gen.LogEntryInfo to a LogEntry.
func logEntryInfoFromGen(
	e gen.LogEntryInfo,
) LogEntry {
	return LogEntry{
		Timestamp: derefString(e.Timestamp),
		Unit:      derefString(e.Unit),
		Priority:  derefString(e.Priority),
		Message:   derefString(e.Message),
		PID:       derefInt(e.Pid),
		Hostname:  derefString(e.Hostname),
	}
}
