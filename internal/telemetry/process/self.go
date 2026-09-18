// Copyright (c) 2024 John Dewey

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

// Package process provides resource usage metrics for the current process.
package process

import (
	"fmt"
	"os"
	"runtime"

	gopsutil "github.com/shirou/gopsutil/v4/process"
)

// Injectable functions for testing.
var (
	newProcessFn = gopsutil.NewProcess
	cpuPercentFn = func(proc *gopsutil.Process) (float64, error) { return proc.CPUPercent() }
	// procMemoryInfoFn wraps the raw gopsutil call so the error branch in
	// defaultMemoryInfoFn below can be exercised: gopsutil's MemoryInfo()
	// never fails in practice on any platform this runs on, so the failure
	// path needs its own seam to test.
	procMemoryInfoFn = func(proc *gopsutil.Process) (*gopsutil.MemoryInfoStat, error) { return proc.MemoryInfo() }
	memoryInfoFn     = defaultMemoryInfoFn
)

// defaultMemoryInfoFn is the default value of memoryInfoFn, named so
// ResetMemoryInfoFn in export_test.go can restore this exact function
// rather than a duplicate closure at a different source location.
func defaultMemoryInfoFn(
	proc *gopsutil.Process,
) (uint64, error) {
	info, err := procMemoryInfoFn(proc)
	if err != nil {
		return 0, err
	}

	return info.RSS, nil
}

type provider struct {
	pid int32
}

// New creates a new Provider that reports metrics for the current process.
func New() Provider {
	return &provider{pid: int32(os.Getpid())}
}

// GetMetrics retrieves CPU percent, RSS memory, and goroutine count for
// the current process.
func (p *provider) GetMetrics() (*Metrics, error) {
	proc, err := newProcessFn(p.pid)
	if err != nil {
		return nil, fmt.Errorf("get process: %w", err)
	}

	cpuPercent, err := cpuPercentFn(proc)
	if err != nil {
		return nil, fmt.Errorf("get cpu percent: %w", err)
	}

	rss, err := memoryInfoFn(proc)
	if err != nil {
		return nil, fmt.Errorf("get memory info: %w", err)
	}

	return &Metrics{
		CPUPercent: cpuPercent,
		RSSBytes:   int64(rss),
		Goroutines: runtime.NumGoroutine(),
	}, nil
}
