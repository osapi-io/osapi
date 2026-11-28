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

package process

import "github.com/osapi-io/osapi/pkg/sdk/types"

// ListInput is list's input: it takes none.
type ListInput struct{}

// GetInput names the process to read.
type GetInput struct {
	PID int `json:"pid" validate:"required,min=1,max=4194304" doc:"The process ID, such as 1234"`
}

// SignalInput is the signal to send, and the process to send it to.
type SignalInput struct {
	PID    int          `json:"pid"    validate:"required,min=1,max=4194304" doc:"The process ID, such as 1234"                   telemetry:"attr"`
	Signal types.Signal `json:"signal" validate:"required,enum"              doc:"The signal: TERM, KILL, HUP, INT, USR1 or USR2" telemetry:"attr"`
}

// Entry is one process of the host.
type Entry struct {
	PID        int     `json:"pid"                   doc:"The process ID"`
	Name       string  `json:"name,omitempty"        doc:"The name of the process"`
	User       string  `json:"user,omitempty"        doc:"The user that runs it"`
	State      string  `json:"state,omitempty"       doc:"Its state, such as running or sleep"`
	CPUPercent float64 `json:"cpu_percent,omitempty" doc:"Its CPU time as a percent of the time since it started"`
	MemPercent float32 `json:"mem_percent,omitempty" doc:"Its resident memory as a percent of the host's memory"`
	MemRSS     int64   `json:"mem_rss,omitempty"     doc:"Its resident memory in bytes"                           render:"bytes"`
	Command    string  `json:"command,omitempty"     doc:"Its full command line"`
	StartTime  string  `json:"start_time,omitempty"  doc:"When it started, in RFC 3339"`
}

// Result is what a signal did.
type Result struct {
	PID     int          `json:"pid"     doc:"The process ID"`
	Signal  types.Signal `json:"signal"  doc:"The signal that was sent"`
	Changed bool         `json:"changed" doc:"Whether the signal was sent"`
}
