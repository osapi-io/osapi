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

package agent

import (
	"context"
	"log/slog"

	"github.com/osapi-io/osapi/internal/provider"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/exec"
	dockerProv "github.com/osapi-io/osapi/internal/provider/container/docker"
	fileProv "github.com/osapi-io/osapi/internal/provider/file"
	"github.com/osapi-io/osapi/internal/provider/network/netinfo"
	"github.com/osapi-io/osapi/internal/provider/network/netplan/dns"
	ifaceProv "github.com/osapi-io/osapi/internal/provider/network/netplan/iface"
	routeProv "github.com/osapi-io/osapi/internal/provider/network/netplan/route"
	"github.com/osapi-io/osapi/internal/provider/network/ping"
	aptProv "github.com/osapi-io/osapi/internal/provider/node/apt"
	certProv "github.com/osapi-io/osapi/internal/provider/node/certificate"
	"github.com/osapi-io/osapi/internal/provider/node/disk"
	nodeHost "github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	logProv "github.com/osapi-io/osapi/internal/provider/node/log"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
	ntpProv "github.com/osapi-io/osapi/internal/provider/node/ntp"
	powerProv "github.com/osapi-io/osapi/internal/provider/node/power"
	processProv "github.com/osapi-io/osapi/internal/provider/node/process"
	serviceProv "github.com/osapi-io/osapi/internal/provider/node/service"
	timezoneProv "github.com/osapi-io/osapi/internal/provider/node/timezone"
	userProv "github.com/osapi-io/osapi/internal/provider/node/user"
	cronProv "github.com/osapi-io/osapi/internal/provider/schedule/cron"
)

// dockerNewFn makes the Docker client (injectable for testing). docker checks
// the error before it uses the client.
var dockerNewFn = func() (dockerProv.Provider, error) {
	return dockerProv.New()
}

// hostEnv is what the providers of a host are made from. Each method returns
// the provider of one area for platform: "debian", "darwin", or any other
// value for the Linux fallback, which does nothing.
type hostEnv struct {
	log       *slog.Logger
	fs        avfs.VFS
	exec      exec.Manager
	platform  string
	container bool
	hostname  string
	// fileState and fileStateBucket record the files that the providers
	// deploy. fileState is nil when the agent has no file-state bucket.
	fileState       provider.FileState
	fileStateBucket string
	files           fileProv.Provider
}

func (e hostEnv) host() nodeHost.Provider {
	if e.platform == "darwin" {
		return nodeHost.NewDarwinProvider()
	}

	return nodeHost.NewLinuxProvider()
}

func (e hostEnv) disk() disk.Provider {
	switch e.platform {
	case "debian":
		return disk.NewDebianProvider(e.log)
	case "darwin":
		return disk.NewDarwinProvider(e.log)
	default:
		return disk.NewLinuxProvider()
	}
}

func (e hostEnv) memory() memory.Provider {
	switch e.platform {
	case "debian":
		return memory.NewDebianProvider()
	case "darwin":
		return memory.NewDarwinProvider()
	default:
		return memory.NewLinuxProvider()
	}
}

func (e hostEnv) load() load.Provider {
	switch e.platform {
	case "debian":
		return load.NewDebianProvider()
	case "darwin":
		return load.NewDarwinProvider()
	default:
		return load.NewLinuxProvider()
	}
}

func (e hostEnv) dns() dns.Provider {
	switch e.platform {
	case "debian":
		if e.container {
			return dns.NewDebianDockerProvider(e.log, e.fs)
		}
		return dns.NewDebianProvider(
			e.log,
			e.fs,
			e.fileState,
			e.fileStateBucket,
			e.exec,
			e.hostname,
		)
	case "darwin":
		return dns.NewDarwinProvider(e.log, e.exec)
	default:
		return dns.NewLinuxProvider()
	}
}

func (e hostEnv) ping() ping.Provider {
	switch e.platform {
	case "debian":
		return ping.NewDebianProvider()
	case "darwin":
		return ping.NewDarwinProvider()
	default:
		return ping.NewLinuxProvider()
	}
}

func (e hostEnv) netinfo() netinfo.Provider {
	if e.platform == "darwin" {
		return netinfo.NewDarwinProvider(e.exec)
	}

	return netinfo.NewLinuxProvider()
}

// docker returns the Docker provider, or nil if Docker does not answer. The
// agent runs without the docker operations then.
func (e hostEnv) docker(
	ctx context.Context,
) dockerProv.Provider {
	c, err := dockerNewFn()
	if err != nil {
		e.log.Info("Docker client creation failed, docker operations disabled",
			slog.String("error", err.Error()))
		return nil
	}
	if err := c.Ping(ctx); err != nil {
		e.log.Info("Docker not available, docker operations disabled",
			slog.String("error", err.Error()))
		return nil
	}

	return c
}

func (e hostEnv) cron() cronProv.Provider {
	switch e.platform {
	case "debian":
		if e.files == nil {
			e.log.Warn("file provider not available, cron operations disabled")
			return cronProv.NewLinuxProvider()
		}
		return cronProv.NewDebianProvider(
			e.log,
			e.fs,
			e.files,
			e.fileState,
			e.fileStateBucket,
			e.hostname,
		)
	case "darwin":
		return cronProv.NewDarwinProvider()
	default:
		return cronProv.NewLinuxProvider()
	}
}

func (e hostEnv) ntp() ntpProv.Provider {
	switch e.platform {
	case "debian":
		if e.container {
			e.log.Info("running in container, NTP operations disabled")
			return ntpProv.NewLinuxProvider()
		}
		return ntpProv.NewDebianProvider(e.log, e.fs, e.exec)
	case "darwin":
		return ntpProv.NewDarwinProvider()
	default:
		return ntpProv.NewLinuxProvider()
	}
}

func (e hostEnv) timezone() timezoneProv.Provider {
	switch e.platform {
	case "debian":
		if e.container {
			e.log.Info("running in container, timezone operations disabled")
			return timezoneProv.NewLinuxProvider()
		}
		return timezoneProv.NewDebianProvider(e.log, e.exec)
	case "darwin":
		return timezoneProv.NewDarwinProvider()
	default:
		return timezoneProv.NewLinuxProvider()
	}
}

func (e hostEnv) power() powerProv.Provider {
	switch e.platform {
	case "debian":
		if e.container {
			e.log.Info("running in container, power operations disabled")
			return powerProv.NewLinuxProvider()
		}
		return powerProv.NewDebianProvider(e.log, e.exec)
	case "darwin":
		return powerProv.NewDarwinProvider()
	default:
		return powerProv.NewLinuxProvider()
	}
}

func (e hostEnv) process() processProv.Provider {
	switch e.platform {
	case "debian":
		return processProv.NewDebianProvider(
			e.log,
			processProv.NewGopsutilLister(),
			processProv.NewSyscallSignaler(),
		)
	case "darwin":
		return processProv.NewDarwinProvider()
	default:
		return processProv.NewLinuxProvider()
	}
}

func (e hostEnv) user() userProv.Provider {
	switch e.platform {
	case "debian":
		return userProv.NewDebianProvider(e.log, e.fs, e.exec)
	case "darwin":
		return userProv.NewDarwinProvider()
	default:
		return userProv.NewLinuxProvider()
	}
}

func (e hostEnv) packages() aptProv.Provider {
	switch e.platform {
	case "debian":
		return aptProv.NewDebianProvider(e.log, e.exec)
	case "darwin":
		return aptProv.NewDarwinProvider()
	default:
		return aptProv.NewLinuxProvider()
	}
}

func (e hostEnv) logs() logProv.Provider {
	switch e.platform {
	case "debian":
		if e.container {
			return logProv.NewLinuxProvider()
		}
		return logProv.NewDebianProvider(e.log, e.exec)
	case "darwin":
		return logProv.NewDarwinProvider()
	default:
		return logProv.NewLinuxProvider()
	}
}

func (e hostEnv) certificate() certProv.Provider {
	switch e.platform {
	case "debian":
		if e.files == nil {
			e.log.Warn("file provider not available, certificate operations disabled")
			return certProv.NewLinuxProvider()
		}
		return certProv.NewDebianProvider(
			e.log,
			e.fs,
			e.files,
			e.fileState,
			e.fileStateBucket,
			e.exec,
			e.hostname,
		)
	case "darwin":
		return certProv.NewDarwinProvider()
	default:
		return certProv.NewLinuxProvider()
	}
}

func (e hostEnv) service() serviceProv.Provider {
	switch e.platform {
	case "debian":
		if e.container {
			e.log.Info("running in container, service operations disabled")
			return serviceProv.NewLinuxProvider()
		}
		if e.files == nil {
			e.log.Warn("file provider not available, service operations disabled")
			return serviceProv.NewLinuxProvider()
		}
		return serviceProv.NewDebianProvider(
			e.log,
			e.fs,
			e.files,
			e.fileState,
			e.fileStateBucket,
			e.exec,
			e.hostname,
		)
	case "darwin":
		return serviceProv.NewDarwinProvider()
	default:
		return serviceProv.NewLinuxProvider()
	}
}

func (e hostEnv) netplan() (ifaceProv.Provider, routeProv.Provider) {
	switch e.platform {
	case "debian":
		if e.fileState == nil {
			e.log.Warn("file state KV not available, netplan operations disabled")
			return ifaceProv.NewLinuxProvider(), routeProv.NewLinuxProvider()
		}
		return ifaceProv.NewDebianProvider(
				e.log,
				e.fs,
				e.fileState,
				e.fileStateBucket,
				e.exec,
				e.hostname,
			),
			routeProv.NewDebianProvider(
				e.log,
				e.fs,
				e.fileState,
				e.fileStateBucket,
				e.exec,
				e.hostname,
			)
	case "darwin":
		return ifaceProv.NewDarwinProvider(), routeProv.NewDarwinProvider()
	default:
		return ifaceProv.NewLinuxProvider(), routeProv.NewLinuxProvider()
	}
}
