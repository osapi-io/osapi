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

// ProcessInfoResult represents the result of a process list/get operation
// for one host.
type ProcessInfoResult struct {
	Hostname  string        `json:"hostname"`
	Status    string        `json:"status"`
	Processes []ProcessInfo `json:"processes,omitempty"`
	Error     string        `json:"error,omitempty"`
}

// ProcessInfo represents information about a single running process.
type ProcessInfo struct {
	PID        int     `json:"pid"`
	Name       string  `json:"name,omitempty"`
	User       string  `json:"user,omitempty"`
	State      string  `json:"state,omitempty"`
	CPUPercent float64 `json:"cpu_percent,omitempty"`
	MemPercent float32 `json:"mem_percent,omitempty"`
	MemRSS     int64   `json:"mem_rss,omitempty"`
	Command    string  `json:"command,omitempty"`
	StartTime  string  `json:"start_time,omitempty"`
}

// ProcessSignalResult represents the result of sending a signal to a process
// for one host.
type ProcessSignalResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	PID      int    `json:"pid,omitempty"`
	Signal   string `json:"signal,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// ProcessSignalOpts contains options for the process signal operation.
type ProcessSignalOpts struct {
	// Signal is the signal name to send (e.g., TERM, KILL, HUP).
	Signal string
}

// processInfoCollectionFromList converts a gen.ProcessCollectionResponse
// to a Collection[ProcessInfoResult].
func processInfoCollectionFromList(
	g *gen.ProcessCollectionResponse,
) Collection[ProcessInfoResult] {
	results := make([]ProcessInfoResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, processInfoResultFromListEntry(r))
	}

	return Collection[ProcessInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// processInfoCollectionFromGet converts a gen.ProcessGetResponse
// to a Collection[ProcessInfoResult].
func processInfoCollectionFromGet(
	g *gen.ProcessGetResponse,
) Collection[ProcessInfoResult] {
	results := make([]ProcessInfoResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, processInfoResultFromGetEntry(r))
	}

	return Collection[ProcessInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// processSignalCollectionFromGen converts a gen.ProcessSignalResponse
// to a Collection[ProcessSignalResult].
func processSignalCollectionFromGen(
	g *gen.ProcessSignalResponse,
) Collection[ProcessSignalResult] {
	results := make([]ProcessSignalResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, processSignalResultFromGen(r))
	}

	return Collection[ProcessSignalResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// processInfoResultFromListEntry converts a gen.ProcessEntry to a
// ProcessInfoResult.
func processInfoResultFromListEntry(
	r gen.ProcessEntry,
) ProcessInfoResult {
	result := ProcessInfoResult{
		Hostname: r.Hostname,
		Status:   string(r.Status),
		Error:    derefString(r.Error),
	}

	if r.Processes != nil {
		procs := make([]ProcessInfo, 0, len(*r.Processes))
		for _, p := range *r.Processes {
			procs = append(procs, processInfoFromGen(p))
		}
		result.Processes = procs
	}

	return result
}

// processInfoResultFromGetEntry converts a gen.ProcessGetEntry to a
// ProcessInfoResult.
func processInfoResultFromGetEntry(
	r gen.ProcessGetEntry,
) ProcessInfoResult {
	result := ProcessInfoResult{
		Hostname: r.Hostname,
		Status:   string(r.Status),
		Error:    derefString(r.Error),
	}

	if r.Process != nil {
		result.Processes = []ProcessInfo{processInfoFromGen(*r.Process)}
	}

	return result
}

// processInfoFromGen converts a gen.ProcessInfo to a ProcessInfo.
func processInfoFromGen(
	p gen.ProcessInfo,
) ProcessInfo {
	return ProcessInfo{
		PID:        derefInt(p.Pid),
		Name:       derefString(p.Name),
		User:       derefString(p.User),
		State:      derefString(p.State),
		CPUPercent: derefFloat64(p.CpuPercent),
		MemPercent: derefFloat32(p.MemPercent),
		MemRSS:     derefInt64(p.MemRss),
		Command:    derefString(p.Command),
		StartTime:  derefString(p.StartTime),
	}
}

// processSignalResultFromGen converts a gen.ProcessSignalResult to a
// ProcessSignalResult.
func processSignalResultFromGen(
	r gen.ProcessSignalResult,
) ProcessSignalResult {
	return ProcessSignalResult{
		Hostname: r.Hostname,
		Status:   string(r.Status),
		PID:      derefInt(r.Pid),
		Signal:   derefString(r.Signal),
		Changed:  derefBool(r.Changed),
		Error:    derefString(r.Error),
	}
}
