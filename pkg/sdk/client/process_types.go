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
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// ProcessInfoResult is one process of one host.
type ProcessInfoResult struct {
	Hostname   string  `json:"hostname"`
	Status     string  `json:"status"`
	PID        int     `json:"pid,omitempty"`
	Name       string  `json:"name,omitempty"`
	User       string  `json:"user,omitempty"`
	State      string  `json:"state,omitempty"`
	CPUPercent float64 `json:"cpu_percent,omitempty"`
	MemPercent float32 `json:"mem_percent,omitempty"`
	MemRSS     int64   `json:"mem_rss,omitempty"`
	Command    string  `json:"command,omitempty"`
	StartTime  string  `json:"start_time,omitempty"`
	Error      string  `json:"error,omitempty"`
}

// ProcessSignalResult is what a signal did on one host.
type ProcessSignalResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	PID      int    `json:"pid,omitempty"`
	Signal   string `json:"signal,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// ProcessSignalOpts is the signal to send.
type ProcessSignalOpts struct {
	// Signal is TERM, KILL, HUP, INT, USR1 or USR2. Required.
	Signal string
}

func processInfoCollectionFromGen(
	rows []gen.ProcessEntry,
	jobID *openapi_types.UUID,
) Collection[ProcessInfoResult] {
	results := make([]ProcessInfoResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, ProcessInfoResult{
			Hostname:   r.Hostname,
			Status:     string(r.Status),
			PID:        derefInt(r.Pid),
			Name:       derefString(r.Name),
			User:       derefString(r.User),
			State:      derefString(r.State),
			CPUPercent: derefFloat64(r.CpuPercent),
			MemPercent: derefFloat32(r.MemPercent),
			MemRSS:     derefInt64(r.MemRss),
			Command:    derefString(r.Command),
			StartTime:  derefString(r.StartTime),
			Error:      derefString(r.Error),
		})
	}

	return Collection[ProcessInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(jobID),
	}
}

func processSignalCollectionFromGen(
	rows []gen.ProcessResult,
	jobID *openapi_types.UUID,
) Collection[ProcessSignalResult] {
	results := make([]ProcessSignalResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, ProcessSignalResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			PID:      derefInt(r.Pid),
			Signal:   derefString(r.Signal),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[ProcessSignalResult]{
		Results: results,
		JobID:   jobIDFromGen(jobID),
	}
}
