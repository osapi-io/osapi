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

package cmd

import (
	"context"
	"io"
	"log/slog"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider/command"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/provider/node/disk"
	nodeHost "github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
	"github.com/retr0h/osapi/internal/provider/registry"
	"github.com/retr0h/osapi/pkg/sdk/platform"
)

// PingParams holds the input for the ping.do operation.
type PingParams struct {
	Address string `json:"address"`
}

// DNSGetParams holds the input for dns.get-resolv-conf.
type DNSGetParams struct {
	InterfaceName string `json:"interface_name"`
}

// DNSUpdateParams holds the input for dns.update-resolv-conf.
type DNSUpdateParams struct {
	Servers       []string `json:"servers"`
	SearchDomains []string `json:"search_domains"`
	InterfaceName string   `json:"interface_name"`
}

// buildProviderRegistry creates a registry with all known provider operations,
// using platform detection to select the correct provider variant (Ubuntu,
// Darwin, or generic Linux).
func buildProviderRegistry() *registry.Registry {
	reg := registry.New()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	execManager := exec.New(logger)
	plat := platform.Detect()

	registerHostProvider(reg, plat)
	registerDiskProvider(reg, plat, logger)
	registerMemProvider(reg, plat)
	registerLoadProvider(reg, plat)
	registerPingProvider(reg, plat)
	registerDNSProvider(reg, plat, logger, execManager)
	registerCommandProvider(reg, logger, execManager)

	return reg
}

func registerHostProvider(
	reg *registry.Registry,
	platform string,
) {
	var p nodeHost.Provider
	switch platform {
	case "ubuntu":
		p = nodeHost.NewUbuntuProvider()
	case "darwin":
		p = nodeHost.NewDarwinProvider()
	default:
		p = nodeHost.NewLinuxProvider()
	}

	reg.Register(registry.Registration{
		Name: "host",
		Operations: map[string]registry.OperationSpec{
			"get-hostname": {
				Run: func(_ context.Context, _ any) (any, error) {
					return p.GetHostname()
				},
			},
			"get-uptime": {
				Run: func(_ context.Context, _ any) (any, error) {
					return p.GetUptime()
				},
			},
			"get-os-info": {
				Run: func(_ context.Context, _ any) (any, error) {
					return p.GetOSInfo()
				},
			},
			"get-architecture": {
				Run: func(_ context.Context, _ any) (any, error) {
					return p.GetArchitecture()
				},
			},
			"get-kernel-version": {
				Run: func(_ context.Context, _ any) (any, error) {
					return p.GetKernelVersion()
				},
			},
			"get-fqdn": {
				Run: func(_ context.Context, _ any) (any, error) {
					return p.GetFQDN()
				},
			},
			"get-cpu-count": {
				Run: func(_ context.Context, _ any) (any, error) {
					return p.GetCPUCount()
				},
			},
			"get-service-manager": {
				Run: func(_ context.Context, _ any) (any, error) {
					return p.GetServiceManager()
				},
			},
			"get-package-manager": {
				Run: func(_ context.Context, _ any) (any, error) {
					return p.GetPackageManager()
				},
			},
		},
	})
}

func registerDiskProvider(
	reg *registry.Registry,
	platform string,
	logger *slog.Logger,
) {
	var p disk.Provider
	switch platform {
	case "ubuntu":
		p = disk.NewUbuntuProvider(logger)
	case "darwin":
		p = disk.NewDarwinProvider(logger)
	default:
		p = disk.NewLinuxProvider()
	}

	reg.Register(registry.Registration{
		Name: "disk",
		Operations: map[string]registry.OperationSpec{
			"get-local-usage-stats": {
				Run: func(_ context.Context, _ any) (any, error) {
					return p.GetLocalUsageStats()
				},
			},
		},
	})
}

func registerMemProvider(
	reg *registry.Registry,
	platform string,
) {
	var p mem.Provider
	switch platform {
	case "ubuntu":
		p = mem.NewUbuntuProvider()
	case "darwin":
		p = mem.NewDarwinProvider()
	default:
		p = mem.NewLinuxProvider()
	}

	reg.Register(registry.Registration{
		Name: "mem",
		Operations: map[string]registry.OperationSpec{
			"get-stats": {
				Run: func(_ context.Context, _ any) (any, error) {
					return p.GetStats()
				},
			},
		},
	})
}

func registerLoadProvider(
	reg *registry.Registry,
	platform string,
) {
	var p load.Provider
	switch platform {
	case "ubuntu":
		p = load.NewUbuntuProvider()
	case "darwin":
		p = load.NewDarwinProvider()
	default:
		p = load.NewLinuxProvider()
	}

	reg.Register(registry.Registration{
		Name: "load",
		Operations: map[string]registry.OperationSpec{
			"get-average-stats": {
				Run: func(_ context.Context, _ any) (any, error) {
					return p.GetAverageStats()
				},
			},
		},
	})
}

func registerPingProvider(
	reg *registry.Registry,
	platform string,
) {
	var p ping.Provider
	switch platform {
	case "ubuntu":
		p = ping.NewUbuntuProvider()
	case "darwin":
		p = ping.NewDarwinProvider()
	default:
		p = ping.NewLinuxProvider()
	}

	reg.Register(registry.Registration{
		Name: "ping",
		Operations: map[string]registry.OperationSpec{
			"do": {
				NewParams: func() any { return &PingParams{} },
				Run: func(_ context.Context, params any) (any, error) {
					pp := params.(*PingParams)
					return p.Do(pp.Address)
				},
			},
		},
	})
}

func registerDNSProvider(
	reg *registry.Registry,
	platform string,
	logger *slog.Logger,
	em exec.Manager,
) {
	var p dns.Provider
	switch platform {
	case "ubuntu":
		p = dns.NewUbuntuProvider(logger, em)
	case "darwin":
		p = dns.NewDarwinProvider(logger, em)
	default:
		p = dns.NewLinuxProvider()
	}

	reg.Register(registry.Registration{
		Name: "dns",
		Operations: map[string]registry.OperationSpec{
			"get-resolv-conf": {
				NewParams: func() any { return &DNSGetParams{} },
				Run: func(_ context.Context, params any) (any, error) {
					pp := params.(*DNSGetParams)
					return p.GetResolvConfByInterface(pp.InterfaceName)
				},
			},
			"update-resolv-conf": {
				NewParams: func() any { return &DNSUpdateParams{} },
				Run: func(_ context.Context, params any) (any, error) {
					pp := params.(*DNSUpdateParams)
					return p.UpdateResolvConfByInterface(
						pp.Servers,
						pp.SearchDomains,
						pp.InterfaceName,
					)
				},
			},
		},
	})
}

func registerCommandProvider(
	reg *registry.Registry,
	logger *slog.Logger,
	em exec.Manager,
) {
	p := command.New(logger, em)

	reg.Register(registry.Registration{
		Name: "command",
		Operations: map[string]registry.OperationSpec{
			"exec": {
				NewParams: func() any { return &command.ExecParams{} },
				Run: func(_ context.Context, params any) (any, error) {
					pp := params.(*command.ExecParams)
					return p.Exec(*pp)
				},
			},
			"shell": {
				NewParams: func() any { return &command.ShellParams{} },
				Run: func(_ context.Context, params any) (any, error) {
					pp := params.(*command.ShellParams)
					return p.Shell(*pp)
				},
			},
		},
	})
}
