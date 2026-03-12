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

package agent

import (
	"context"
	"log/slog"
	"strings"

	"github.com/shirou/gopsutil/v4/host"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider/command"
	containerProv "github.com/retr0h/osapi/internal/provider/container"
	"github.com/retr0h/osapi/internal/provider/container/runtime/docker"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/provider/network/netinfo"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/provider/node/disk"
	nodeHost "github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
)

// factoryHostInfoFn is the function used to get host info (injectable for testing).
var factoryHostInfoFn = host.Info

// factoryDockerNewFn is the function used to create a Docker driver (injectable for testing).
var factoryDockerNewFn = docker.New

// ProviderFactory creates platform-specific providers for the agent.
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

// CreateProviders creates all providers needed for the agent.
func (f *ProviderFactory) CreateProviders() (
	nodeHost.Provider,
	disk.Provider,
	mem.Provider,
	load.Provider,
	dns.Provider,
	ping.Provider,
	netinfo.Provider,
	command.Provider,
	containerProv.Provider,
) {
	info, _ := factoryHostInfoFn()
	platform := strings.ToLower(info.Platform)
	if platform == "" && strings.ToLower(info.OS) == "darwin" {
		platform = "darwin"
	}

	if platform == "darwin" {
		f.logger.Info("running on darwin")
	}

	// Create system providers
	var hostProvider nodeHost.Provider
	switch platform {
	case "ubuntu":
		hostProvider = nodeHost.NewUbuntuProvider()
	case "darwin":
		hostProvider = nodeHost.NewDarwinProvider()
	default:
		hostProvider = nodeHost.NewLinuxProvider()
	}

	var diskProvider disk.Provider
	switch platform {
	case "ubuntu":
		diskProvider = disk.NewUbuntuProvider(f.logger)
	case "darwin":
		diskProvider = disk.NewDarwinProvider(f.logger)
	default:
		diskProvider = disk.NewLinuxProvider()
	}

	var memProvider mem.Provider
	switch platform {
	case "ubuntu":
		memProvider = mem.NewUbuntuProvider()
	case "darwin":
		memProvider = mem.NewDarwinProvider()
	default:
		memProvider = mem.NewLinuxProvider()
	}

	var loadProvider load.Provider
	switch platform {
	case "ubuntu":
		loadProvider = load.NewUbuntuProvider()
	case "darwin":
		loadProvider = load.NewDarwinProvider()
	default:
		loadProvider = load.NewLinuxProvider()
	}

	// Create network providers
	var dnsProvider dns.Provider
	execManager := exec.New(f.logger)
	switch platform {
	case "ubuntu":
		dnsProvider = dns.NewUbuntuProvider(f.logger, execManager)
	case "darwin":
		dnsProvider = dns.NewDarwinProvider(f.logger, execManager)
	default:
		dnsProvider = dns.NewLinuxProvider()
	}

	var pingProvider ping.Provider
	switch platform {
	case "ubuntu":
		pingProvider = ping.NewUbuntuProvider()
	case "darwin":
		pingProvider = ping.NewDarwinProvider()
	default:
		pingProvider = ping.NewLinuxProvider()
	}

	// Create network info provider
	var netinfoProvider netinfo.Provider
	switch platform {
	case "darwin":
		netinfoProvider = netinfo.NewDarwinProvider(execManager)
	default:
		netinfoProvider = netinfo.NewLinuxProvider()
	}

	// Create command provider (cross-platform, uses exec.Manager)
	commandProvider := command.New(f.logger, execManager)

	// Create container provider (conditional on Docker availability)
	var containerProvider containerProv.Provider
	dockerDriver, err := factoryDockerNewFn()
	if err == nil {
		if pingErr := dockerDriver.Ping(context.Background()); pingErr == nil {
			containerProvider = containerProv.New(dockerDriver)
		} else {
			f.logger.Info("Docker not available, container operations disabled",
				slog.String("error", pingErr.Error()))
		}
	} else {
		f.logger.Info("Docker client creation failed, container operations disabled",
			slog.String("error", err.Error()))
	}

	return hostProvider, diskProvider, memProvider, loadProvider, dnsProvider, pingProvider, netinfoProvider, commandProvider, containerProvider
}
