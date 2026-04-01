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

// Package log provides log viewing operations.
package log

import (
	"context"
)

// Provider implements log viewing operations.
type Provider interface {
	// Query returns journal entries with optional filtering.
	Query(ctx context.Context, opts QueryOpts) ([]Entry, error)
	// QueryUnit returns journal entries for a specific systemd unit.
	QueryUnit(ctx context.Context, unit string, opts QueryOpts) ([]Entry, error)
}

// QueryOpts contains optional filters for log queries.
type QueryOpts struct {
	Lines    int    `json:"lines,omitempty"`
	Since    string `json:"since,omitempty"`
	Priority string `json:"priority,omitempty"`
}

// Entry represents a single journal entry.
type Entry struct {
	Timestamp string `json:"timestamp"`
	Unit      string `json:"unit,omitempty"`
	Priority  string `json:"priority"`
	Message   string `json:"message"`
	PID       int    `json:"pid,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
}
