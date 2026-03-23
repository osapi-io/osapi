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

package host

import (
	"time"
)

// Provider implements the methods to interact with various Host components.
type Provider interface {
	// GetUptime retrieves the system uptime.
	GetUptime() (time.Duration, error)
	// GetHostname retrieves the hostname of the system.
	GetHostname() (string, error)
	// GetOSInfo retrieves information about the operating system, including the
	// distribution name and version.
	GetOSInfo() (*Result, error)
	// GetArchitecture retrieves the system CPU architecture (e.g., x86_64, arm64).
	GetArchitecture() (string, error)
	// GetKernelVersion retrieves the running kernel version string.
	GetKernelVersion() (string, error)
	// GetFQDN retrieves the fully qualified domain name of the system.
	GetFQDN() (string, error)
	// GetCPUCount retrieves the number of logical CPUs available.
	GetCPUCount() (int, error)
	// GetServiceManager detects the system's service manager (e.g., systemd).
	GetServiceManager() (string, error)
	// GetPackageManager detects the system's package manager (e.g., apt, dnf, yum).
	GetPackageManager() (string, error)
}

// Result represents the operating system information.
type Result struct {
	// The name of the Linux distribution (e.g., Debian, CentOS).
	Distribution string
	// The version of the Linux distribution (e.g., 20.04, 8.3).
	Version string
	// Changed indicates whether system state was modified.
	Changed bool `json:"changed"`
}
