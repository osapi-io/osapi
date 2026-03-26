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

import gopsutil "github.com/shirou/gopsutil/v4/process"

// ExportNewProviderWithPID exposes the private provider constructor for testing.
func ExportNewProviderWithPID(
	pid int32,
) Provider {
	return &provider{pid: pid}
}

// SetNewProcessFn overrides the newProcessFn injectable for testing.
func SetNewProcessFn(
	fn func(int32) (*gopsutil.Process, error),
) {
	newProcessFn = fn
}

// ResetNewProcessFn restores the default newProcessFn.
func ResetNewProcessFn() {
	newProcessFn = gopsutil.NewProcess
}

// SetCPUPercentFn overrides the cpuPercentFn injectable for testing.
func SetCPUPercentFn(
	fn func(*gopsutil.Process) (float64, error),
) {
	cpuPercentFn = fn
}

// ResetCPUPercentFn restores the default cpuPercentFn.
func ResetCPUPercentFn() {
	cpuPercentFn = func(proc *gopsutil.Process) (float64, error) {
		return proc.CPUPercent()
	}
}

// SetMemoryInfoFn overrides the memoryInfoFn injectable for testing.
func SetMemoryInfoFn(
	fn func(*gopsutil.Process) (uint64, error),
) {
	memoryInfoFn = fn
}

// ResetMemoryInfoFn restores the default memoryInfoFn.
func ResetMemoryInfoFn() {
	memoryInfoFn = func(proc *gopsutil.Process) (uint64, error) {
		info, err := proc.MemoryInfo()
		if err != nil {
			return 0, err
		}

		return info.RSS, nil
	}
}
