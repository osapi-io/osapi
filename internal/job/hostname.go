// Copyright (c) 2025 John Dewey

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

package job

import (
	"github.com/shirou/gopsutil/v4/host"
)

// hostInfoFn is the function used to get host info (injectable for testing).
var hostInfoFn = host.Info

// HostnameProvider defines the interface for getting hostname
type HostnameProvider interface {
	Hostname() (string, error)
}

// gopsutilHostnameProvider implements HostnameProvider using gopsutil/host
type gopsutilHostnameProvider struct{}

func (p gopsutilHostnameProvider) Hostname() (string, error) {
	info, err := hostInfoFn()
	if err != nil {
		return "", err
	}
	return info.Hostname, nil
}

// defaultHostnameProvider is the default provider using gopsutil
var defaultHostnameProvider HostnameProvider = gopsutilHostnameProvider{}

// GetWorkerHostname returns the hostname that should be used by workers.
// It first checks the configured hostname, then falls back to system hostname using gopsutil.
// This function respects configuration while using gopsutil for system detection.
func GetWorkerHostname(
	configuredHostname string,
) (string, error) {
	// If hostname is explicitly configured, use it
	if configuredHostname != "" {
		return configuredHostname, nil
	}

	// Use gopsutil to get system hostname
	hostname, err := defaultHostnameProvider.Hostname()
	if err != nil || hostname == "" {
		return "unknown", nil
	}

	return hostname, nil
}

// GetWorkerHostnameWithProvider returns the hostname using the provided provider.
// This allows for testing with mock providers.
func GetWorkerHostnameWithProvider(
	configuredHostname string,
	provider HostnameProvider,
) (string, error) {
	// If hostname is explicitly configured, use it
	if configuredHostname != "" {
		return configuredHostname, nil
	}

	// Use the provided provider
	hostname, err := provider.Hostname()
	if err != nil || hostname == "" {
		return "unknown", nil
	}

	return hostname, nil
}

// GetLocalHostname returns the current system hostname using gopsutil.
// This is for backward compatibility with existing code.
func GetLocalHostname() (string, error) {
	return GetLocalHostnameWithProvider(defaultHostnameProvider)
}

// GetLocalHostnameWithProvider returns the hostname using the provided provider.
// This allows for testing with mock providers.
func GetLocalHostnameWithProvider(
	provider HostnameProvider,
) (string, error) {
	hostname, err := provider.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}
