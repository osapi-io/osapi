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

import (
	"syscall"

	gopsutil "github.com/shirou/gopsutil/v4/process"
)

// SetListProcesses overrides the listProcesses function for testing.
func SetListProcesses(
	fn func() ([]*gopsutil.Process, error),
) {
	listProcesses = fn
}

// ResetListProcesses restores the default listProcesses function.
func ResetListProcesses() {
	listProcesses = defaultListProcesses
}

// SetGetProcess overrides the getProcess function for testing.
func SetGetProcess(
	fn func(int32) (*gopsutil.Process, error),
) {
	getProcess = fn
}

// ResetGetProcess restores the default getProcess function.
func ResetGetProcess() {
	getProcess = defaultGetProcess
}

// SetKillProcess overrides the killProcess function for testing.
func SetKillProcess(
	fn func(int, syscall.Signal) error,
) {
	killProcess = fn
}

// ResetKillProcess restores the default killProcess function.
func ResetKillProcess() {
	killProcess = defaultKillProcess
}

// SetGatherInfoFromP overrides the gatherInfoFromP function for testing.
func SetGatherInfoFromP(
	fn func(*gopsutil.Process) (*Info, error),
) {
	gatherInfoFromP = fn
}

// ResetGatherInfoFromP restores the default gatherInfoFromP function.
func ResetGatherInfoFromP() {
	gatherInfoFromP = defaultGatherInfoFromP
}
