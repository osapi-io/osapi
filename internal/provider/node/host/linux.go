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

// Package host provides system host information.
package host

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/shirou/gopsutil/v4/host"

	"github.com/osapi-io/osapi/internal/provider"
)

// Compile-time check: Linux must satisfy Provider and FactsSetter.
var (
	_ Provider             = (*Linux)(nil)
	_ provider.FactsSetter = (*Linux)(nil)
)

// Linux implements the Host Provider interface for any Linux host. Each read
// is the same on every distribution: gopsutil for the hostname, OS, kernel,
// architecture, CPU count and uptime, a look for apt, dnf or yum, and a stat
// of the systemd directory. Every Linux host, Debian included, uses it. The
// hostname behavior sets the name.
type Linux struct {
	provider.FactsAware

	InfoFn     func() (*host.InfoStat, error)
	HostnameFn func() (string, error)
	NumCPUFn   func() int
	StatFn     func(name string) (os.FileInfo, error)
	LookPathFn func(file string) (string, error)
}

// NewLinuxProvider factory to create a new Linux instance.
func NewLinuxProvider() *Linux {
	return &Linux{
		InfoFn:     host.Info,
		HostnameFn: os.Hostname,
		NumCPUFn:   runtime.NumCPU,
		StatFn:     os.Stat,
		LookPathFn: exec.LookPath,
	}
}
