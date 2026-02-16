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

package worker

import (
	"log/slog"
	"strings"

	"github.com/shirou/gopsutil/v4/host"

	"github.com/retr0h/osapi/internal/cmdexec"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/provider/system/disk"
	systemHost "github.com/retr0h/osapi/internal/provider/system/host"
	"github.com/retr0h/osapi/internal/provider/system/load"
	"github.com/retr0h/osapi/internal/provider/system/mem"
)

// ProviderFactory creates platform-specific providers for the worker.
type ProviderFactory struct {
	logger *slog.Logger
}

// NewProviderFactory creates a new provider factory.
func NewProviderFactory(
	logger *slog.Logger,
) *ProviderFactory {
	return &ProviderFactory{
		logger: logger,
	}
}

// CreateProviders creates all providers needed for the worker.
func (f *ProviderFactory) CreateProviders() (
	systemHost.Provider,
	disk.Provider,
	mem.Provider,
	load.Provider,
	dns.Provider,
	ping.Provider,
) {
	info, _ := host.Info()
	platform := strings.ToLower(info.Platform)

	// Create system providers
	var hostProvider systemHost.Provider
	switch platform {
	case "ubuntu":
		hostProvider = systemHost.NewUbuntuProvider()
	default:
		hostProvider = systemHost.NewLinuxProvider()
	}

	var diskProvider disk.Provider
	switch platform {
	case "ubuntu":
		diskProvider = disk.NewUbuntuProvider(f.logger)
	default:
		diskProvider = disk.NewLinuxProvider()
	}

	var memProvider mem.Provider
	switch platform {
	case "ubuntu":
		memProvider = mem.NewUbuntuProvider()
	default:
		memProvider = mem.NewLinuxProvider()
	}

	var loadProvider load.Provider
	switch platform {
	case "ubuntu":
		loadProvider = load.NewUbuntuProvider()
	default:
		loadProvider = load.NewLinuxProvider()
	}

	// Create network providers
	var dnsProvider dns.Provider
	execManager := cmdexec.New(f.logger)
	switch platform {
	case "ubuntu":
		dnsProvider = dns.NewUbuntuProvider(f.logger, execManager)
	default:
		dnsProvider = dns.NewLinuxProvider()
	}

	var pingProvider ping.Provider
	switch platform {
	case "ubuntu":
		pingProvider = ping.NewUbuntuProvider()
	default:
		pingProvider = ping.NewLinuxProvider()
	}

	return hostProvider, diskProvider, memProvider, loadProvider, dnsProvider, pingProvider
}
