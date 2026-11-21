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
	"context"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/hostname"
	"github.com/osapi-io/gohai/pkg/gohai/source"
)

// hostnameFn is the host's name as `hostname` prints it, from gohai's
// hostname collector (injectable for testing).
var hostnameFn = func() (string, error) {
	return machineName(source.Sources{})
}

// machineName is the name `hostname` prints on the host of src.
func machineName(
	src source.Sources,
) (string, error) {
	got, err := hostname.NewWith(src).Collect(context.Background(), nil)
	if err != nil {
		return "", err
	}

	return got.(*hostname.Info).MachineName, nil
}

// HostnameProvider defines the interface for getting hostname
type HostnameProvider interface {
	Hostname() (string, error)
}

// gohaiHostnameProvider implements HostnameProvider with gohai.
type gohaiHostnameProvider struct{}

// Hostname is the host's name, as `hostname` prints it.
func (p gohaiHostnameProvider) Hostname() (string, error) {
	return hostnameFn()
}

// defaultHostnameProvider is the default provider, gohai
var defaultHostnameProvider HostnameProvider = gohaiHostnameProvider{}

// GetAgentHostname returns the hostname that should be used by agents.
// It first checks the configured hostname, then falls back to the host's
// name from gohai.
func GetAgentHostname(
	configuredHostname string,
) (string, error) {
	// If hostname is explicitly configured, use it
	if configuredHostname != "" {
		return configuredHostname, nil
	}

	// Use gohai to get the host's name
	hostname, err := defaultHostnameProvider.Hostname()
	if err != nil || hostname == "" {
		return "unknown", nil
	}

	return hostname, nil
}

// GetAgentHostnameWithProvider returns the hostname using the provided provider.
// This allows for testing with mock providers.
func GetAgentHostnameWithProvider(
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

// GetLocalHostname returns the host's name, from gohai.
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
