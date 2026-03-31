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

// Package process provides process management operations.
package process

import (
	"context"
	"syscall"

	gopsutil "github.com/shirou/gopsutil/v4/process"
)

// Provider implements process management operations.
type Provider interface {
	// List returns all running processes.
	List(ctx context.Context) ([]Info, error)
	// Get returns details for a specific process by PID.
	Get(ctx context.Context, pid int) (*Info, error)
	// Signal sends a signal to a process by PID.
	Signal(ctx context.Context, pid int, signal string) (*SignalResult, error)
}

// Querier provides methods to query a single process.
// *gopsutil.Process satisfies this interface.
type Querier interface {
	Name() (string, error)
	Username() (string, error)
	Status() ([]string, error)
	CPUPercent() (float64, error)
	MemoryPercent() (float32, error)
	MemoryInfo() (*gopsutil.MemoryInfoStat, error)
	Cmdline() (string, error)
	CreateTime() (int64, error)
}

// Item pairs a PID with a Querier. PID is exposed
// separately because gopsutil.Process.Pid is a struct field, which
// interfaces cannot express.
type Item struct {
	PID     int32
	Querier Querier
}

// Lister provides methods to list and look up processes.
type Lister interface {
	Processes() ([]Item, error)
	NewProcess(pid int32) (Querier, error)
}

// Signaler sends signals to processes.
type Signaler interface {
	Kill(pid int, sig syscall.Signal) error
}

// Info represents a running process.
type Info struct {
	PID        int     `json:"pid"`
	Name       string  `json:"name"`
	User       string  `json:"user"`
	State      string  `json:"state"`
	CPUPercent float64 `json:"cpu_percent"`
	MemPercent float32 `json:"mem_percent"`
	MemRSS     int64   `json:"mem_rss"`
	Command    string  `json:"command"`
	StartTime  string  `json:"start_time"`
}

// SignalResult represents the outcome of sending a signal to a process.
type SignalResult struct {
	PID     int    `json:"pid"`
	Signal  string `json:"signal"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
