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
	"fmt"
	"log/slog"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/agent"
	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/command"
	dockerProv "github.com/retr0h/osapi/internal/provider/container/docker"
	fileProv "github.com/retr0h/osapi/internal/provider/file"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/provider/network/netinfo"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	aptProv "github.com/retr0h/osapi/internal/provider/node/apt"
	"github.com/retr0h/osapi/internal/provider/node/disk"
	nodeHost "github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	logProv "github.com/retr0h/osapi/internal/provider/node/log"
	"github.com/retr0h/osapi/internal/provider/node/mem"
	ntpProv "github.com/retr0h/osapi/internal/provider/node/ntp"
	powerProv "github.com/retr0h/osapi/internal/provider/node/power"
	processProv "github.com/retr0h/osapi/internal/provider/node/process"
	sysctlProv "github.com/retr0h/osapi/internal/provider/node/sysctl"
	timezoneProv "github.com/retr0h/osapi/internal/provider/node/timezone"
	userProv "github.com/retr0h/osapi/internal/provider/node/user"
	cronProv "github.com/retr0h/osapi/internal/provider/scheduled/cron"
	"github.com/retr0h/osapi/internal/telemetry/process"
	"github.com/retr0h/osapi/pkg/sdk/platform"
)

// Ensure cli import is used (for BuildFileStateKVConfig etc.)
var _ = cli.BuildFileStateKVConfig

// dockerNewFn is injectable for testing.
var dockerNewFn = dockerProv.New

// setupAgent connects to NATS, creates providers, and builds the agent.
// It is used by the standalone agent start and combined start commands.
func setupAgent(
	ctx context.Context,
	log *slog.Logger,
	connCfg config.NATSConnection,
) (*agent.Agent, *natsBundle, map[string]job.SubComponentInfo) {
	namespace := connCfg.Namespace
	streamName := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.Stream.Name)
	kvBucket := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.KV.Bucket)

	b := connectNATSBundle(ctx, log, connCfg, kvBucket, namespace, streamName)

	plat := platform.Detect()
	if plat == "darwin" {
		log.Info("running on darwin")
	}

	execManager := exec.New(log)

	// --- Node providers ---
	var hostProvider nodeHost.Provider
	switch plat {
	case "debian":
		if platform.IsContainer() {
			hostProvider = nodeHost.NewDebianDockerProvider()
		} else {
			hostProvider = nodeHost.NewDebianProvider(execManager)
		}
	case "darwin":
		hostProvider = nodeHost.NewDarwinProvider()
	default:
		hostProvider = nodeHost.NewLinuxProvider()
	}

	var diskProvider disk.Provider
	switch plat {
	case "debian":
		diskProvider = disk.NewDebianProvider(log)
	case "darwin":
		diskProvider = disk.NewDarwinProvider(log)
	default:
		diskProvider = disk.NewLinuxProvider()
	}

	var memProvider mem.Provider
	switch plat {
	case "debian":
		memProvider = mem.NewDebianProvider()
	case "darwin":
		memProvider = mem.NewDarwinProvider()
	default:
		memProvider = mem.NewLinuxProvider()
	}

	var loadProvider load.Provider
	switch plat {
	case "debian":
		loadProvider = load.NewDebianProvider()
	case "darwin":
		loadProvider = load.NewDarwinProvider()
	default:
		loadProvider = load.NewLinuxProvider()
	}

	// --- Network providers ---
	var dnsProvider dns.Provider
	switch plat {
	case "debian":
		if platform.IsContainer() {
			dnsProvider = dns.NewDebianDockerProvider(log, appFs)
		} else {
			dnsProvider = dns.NewDebianProvider(log, execManager)
		}
	case "darwin":
		dnsProvider = dns.NewDarwinProvider(log, execManager)
	default:
		dnsProvider = dns.NewLinuxProvider()
	}

	var pingProvider ping.Provider
	switch plat {
	case "debian":
		pingProvider = ping.NewDebianProvider()
	case "darwin":
		pingProvider = ping.NewDarwinProvider()
	default:
		pingProvider = ping.NewLinuxProvider()
	}

	var netinfoProvider netinfo.Provider
	switch plat {
	case "darwin":
		netinfoProvider = netinfo.NewDarwinProvider(execManager)
	default:
		netinfoProvider = netinfo.NewLinuxProvider()
	}

	// --- Command provider ---
	commandProvider := command.New(log, execManager)

	// --- Docker provider ---
	var dockerProvider dockerProv.Provider
	dockerClient, err := dockerNewFn()
	if err == nil {
		if pingErr := dockerClient.Ping(ctx); pingErr == nil {
			dockerProvider = dockerClient
		} else {
			log.Info("Docker not available, docker operations disabled",
				slog.String("error", pingErr.Error()))
		}
	} else {
		log.Info("Docker client creation failed, docker operations disabled",
			slog.String("error", err.Error()))
	}

	// --- File provider ---
	hostname, _ := job.GetAgentHostname(appConfig.Agent.Hostname)
	fileProvider, fileStateKV := createFileProvider(ctx, log, b, namespace, hostname)

	// --- Cron provider ---
	cronProvider := createCronProvider(log, fileProvider, fileStateKV, hostname)

	// --- Sysctl provider ---
	sysctlProvider := createSysctlProvider(log, appFs, fileStateKV, execManager, hostname)

	// --- NTP provider ---
	ntpProvider := createNtpProvider(log, appFs, execManager)

	// --- Timezone provider ---
	timezoneProvider := createTimezoneProvider(log, execManager)

	// --- Power provider ---
	powerProvider := createPowerProvider(log, execManager)

	// --- Process provider ---
	processProvider := createProcessProvider(log)

	// --- User provider ---
	userProvider := createUserProvider(log, appFs, execManager)

	// --- Package provider ---
	packageProvider := createPackageProvider(log, execManager)

	// --- Log provider ---
	logProvider := createLogProvider(log, execManager)

	// --- Build registry ---
	registry := agent.NewProviderRegistry()

	registry.Register(
		"node",
		agent.NewNodeProcessor(
			hostProvider,
			diskProvider,
			memProvider,
			loadProvider,
			sysctlProvider,
			ntpProvider,
			timezoneProvider,
			powerProvider,
			processProvider,
			userProvider,
			packageProvider,
			logProvider,
			appConfig,
			log,
		),
		hostProvider,
		diskProvider,
		memProvider,
		loadProvider,
		sysctlProvider,
		ntpProvider,
		timezoneProvider,
		powerProvider,
		processProvider,
		userProvider,
		packageProvider,
		logProvider,
	)

	registry.Register("network",
		agent.NewNetworkProcessor(dnsProvider, pingProvider, log),
		dnsProvider, pingProvider, netinfoProvider,
	)

	registry.Register("command",
		agent.NewCommandProcessor(commandProvider, log),
		commandProvider,
	)

	registry.Register("file",
		agent.NewFileProcessor(fileProvider, log),
		fileProvider,
	)

	registry.Register("docker",
		agent.NewDockerProcessor(dockerProvider, log),
		dockerProvider,
	)

	registry.Register("schedule",
		agent.NewScheduleProcessor(cronProvider, log),
		cronProvider,
	)

	a := agent.New(
		appFs,
		appConfig,
		log,
		b.jobClient,
		streamName,
		hostProvider,
		diskProvider,
		memProvider,
		loadProvider,
		netinfoProvider,
		process.New(),
		registry,
		b.registryKV,
		b.factsKV,
	)

	enabledOrDisabled := func(enabled bool) string {
		if enabled {
			return "ok"
		}

		return "disabled"
	}

	subComponents := map[string]job.SubComponentInfo{
		"agent.facts":     {Status: "ok"},
		"agent.heartbeat": {Status: "ok"},
		"agent.metrics": {
			Status: enabledOrDisabled(appConfig.Agent.Metrics.Enabled),
			Address: fmt.Sprintf(
				"http://%s:%d",
				appConfig.Agent.Metrics.Host,
				appConfig.Agent.Metrics.Port,
			),
		},
	}
	a.SetSubComponents(subComponents)

	return a, b, subComponents
}

// createFileProvider creates a file provider if Object Store and file-state KV
// are configured. Returns nil provider and nil KV if either is unavailable.
// The stateKV is also returned so the cron provider can check managed files.
func createFileProvider(
	ctx context.Context,
	log *slog.Logger,
	b *natsBundle,
	namespace string,
	hostname string,
) (fileProv.Provider, jetstream.KeyValue) {
	if appConfig.NATS.Objects.Bucket == "" || appConfig.NATS.FileState.Bucket == "" {
		return nil, nil
	}

	objStoreName := job.ApplyNamespaceToInfraName(namespace, appConfig.NATS.Objects.Bucket)
	objStore, err := b.nc.ObjectStore(ctx, objStoreName)
	if err != nil {
		log.Warn("Object Store not available, file operations disabled",
			slog.String("bucket", objStoreName),
			slog.String("error", err.Error()),
		)
		return nil, nil
	}

	fileStateKVConfig := cli.BuildFileStateKVConfig(namespace, appConfig.NATS.FileState)
	fileStateKV, err := b.nc.CreateOrUpdateKVBucketWithConfig(ctx, fileStateKVConfig)
	if err != nil {
		log.Warn("file state KV not available, file operations disabled",
			slog.String("bucket", fileStateKVConfig.Bucket),
			slog.String("error", err.Error()),
		)
		return nil, nil
	}

	// Seed system templates into the object store (idempotent).
	if err := agent.SeedSystemTemplates(ctx, log, objStore); err != nil {
		log.Warn("failed to seed system templates",
			slog.String("error", err.Error()),
		)
	}

	return fileProv.New(log, appFs, objStore, fileStateKV, hostname), fileStateKV
}

// createSysctlProvider creates a platform-specific sysctl provider. On Debian,
// the sysctl provider writes conf files directly to /etc/sysctl.d/ and tracks
// state in the file-state KV. On other platforms, all operations return
// ErrUnsupported.
func createSysctlProvider(
	log *slog.Logger,
	fs avfs.VFS,
	fileStateKV jetstream.KeyValue,
	execManager exec.Manager,
	hostname string,
) sysctlProv.Provider {
	plat := platform.Detect()

	switch plat {
	case "debian":
		if fileStateKV == nil {
			log.Warn("file state KV not available, sysctl operations disabled")
			return sysctlProv.NewLinuxProvider()
		}
		return sysctlProv.NewDebianProvider(log, fs, fileStateKV, execManager, hostname)
	case "darwin":
		return sysctlProv.NewDarwinProvider()
	default:
		return sysctlProv.NewLinuxProvider()
	}
}

// createNtpProvider creates a platform-specific NTP provider. On Debian, the
// NTP provider manages chrony configuration. On other platforms, all operations
// return ErrUnsupported.
func createNtpProvider(
	log *slog.Logger,
	fs avfs.VFS,
	execManager exec.Manager,
) ntpProv.Provider {
	plat := platform.Detect()

	switch plat {
	case "debian":
		if platform.IsContainer() {
			log.Info("running in container, NTP operations disabled")
			return ntpProv.NewLinuxProvider()
		}
		return ntpProv.NewDebianProvider(log, fs, execManager)
	case "darwin":
		return ntpProv.NewDarwinProvider()
	default:
		return ntpProv.NewLinuxProvider()
	}
}

// createTimezoneProvider creates a platform-specific timezone provider. On
// Debian, the timezone provider manages the system timezone via timedatectl.
// In containers, timedatectl is not available — returns ErrUnsupported.
// On other platforms, all operations return ErrUnsupported.
func createTimezoneProvider(
	log *slog.Logger,
	execManager exec.Manager,
) timezoneProv.Provider {
	plat := platform.Detect()

	switch plat {
	case "debian":
		if platform.IsContainer() {
			log.Info("running in container, timezone operations disabled")
			return timezoneProv.NewLinuxProvider()
		}
		return timezoneProv.NewDebianProvider(log, execManager)
	case "darwin":
		return timezoneProv.NewDarwinProvider()
	default:
		return timezoneProv.NewLinuxProvider()
	}
}

// createPowerProvider creates a platform-specific power provider. On Debian, the
// power provider manages system reboot and shutdown via shutdown(8).
// In containers, shutdown is not meaningful — returns ErrUnsupported.
// On other platforms, all operations return ErrUnsupported.
func createPowerProvider(
	log *slog.Logger,
	execManager exec.Manager,
) powerProv.Provider {
	plat := platform.Detect()

	switch plat {
	case "debian":
		if platform.IsContainer() {
			log.Info("running in container, power operations disabled")
			return powerProv.NewLinuxProvider()
		}
		return powerProv.NewDebianProvider(log, execManager)
	case "darwin":
		return powerProv.NewDarwinProvider()
	default:
		return powerProv.NewLinuxProvider()
	}
}

// createProcessProvider creates a platform-specific process provider. On Debian,
// the process provider reads /proc and sends signals to processes. In containers,
// process visibility is limited so the provider is disabled. On other platforms,
// all operations return ErrUnsupported.
func createProcessProvider(
	log *slog.Logger,
) processProv.Provider {
	plat := platform.Detect()

	switch plat {
	case "debian":
		return processProv.NewDebianProvider(
			log,
			processProv.NewGopsutilLister(),
			processProv.NewSyscallSignaler(),
		)
	case "darwin":
		return processProv.NewDarwinProvider()
	default:
		return processProv.NewLinuxProvider()
	}
}

// createCronProvider creates a platform-specific cron provider. On Debian, the
// cron provider delegates file writes to the file provider. On other platforms,
// all operations return ErrUnsupported.
func createCronProvider(
	log *slog.Logger,
	fileProvider fileProv.Provider,
	fileStateKV jetstream.KeyValue,
	hostname string,
) cronProv.Provider {
	plat := platform.Detect()

	switch plat {
	case "debian":
		if fileProvider == nil {
			log.Warn("file provider not available, cron operations disabled")
			return cronProv.NewLinuxProvider()
		}
		return cronProv.NewDebianProvider(log, appFs, fileProvider, fileStateKV, hostname)
	case "darwin":
		return cronProv.NewDarwinProvider()
	default:
		return cronProv.NewLinuxProvider()
	}
}

// createUserProvider creates a platform-specific user provider. On Debian, the
// user provider manages system users and groups via useradd/usermod/groupadd.
// In containers, user management is not meaningful — returns ErrUnsupported.
// On other platforms, all operations return ErrUnsupported.
func createUserProvider(
	log *slog.Logger,
	fs avfs.VFS,
	execManager exec.Manager,
) userProv.Provider {
	plat := platform.Detect()

	switch plat {
	case "debian":
		return userProv.NewDebianProvider(log, fs, execManager)
	case "darwin":
		return userProv.NewDarwinProvider()
	default:
		return userProv.NewLinuxProvider()
	}
}

// createPackageProvider creates a platform-specific package provider. On Debian,
// the package provider manages apt packages. In containers, apt operations are
// disabled because package management is not meaningful. On other platforms, all
// operations return ErrUnsupported.
func createPackageProvider(
	log *slog.Logger,
	execManager exec.Manager,
) aptProv.Provider {
	plat := platform.Detect()

	switch plat {
	case "debian":
		return aptProv.NewDebianProvider(log, execManager)
	case "darwin":
		return aptProv.NewDarwinProvider()
	default:
		return aptProv.NewLinuxProvider()
	}
}

// createLogProvider creates a platform-specific log provider. On Debian, the
// log provider queries journal logs via journalctl. In containers, journalctl
// requires systemd which is not available, so the provider is disabled. On
// other platforms, all operations return ErrUnsupported.
func createLogProvider(
	log *slog.Logger,
	execManager exec.Manager,
) logProv.Provider {
	plat := platform.Detect()

	switch plat {
	case "debian":
		if platform.IsContainer() {
			return logProv.NewLinuxProvider()
		}
		return logProv.NewDebianProvider(log, execManager)
	case "darwin":
		return logProv.NewDarwinProvider()
	default:
		return logProv.NewLinuxProvider()
	}
}
