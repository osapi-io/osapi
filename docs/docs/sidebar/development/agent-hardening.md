# Agent Hardening

This guide covers running the OSAPI agent as an unprivileged user with minimal
permissions. By default the agent runs as root, but production deployments
should use a dedicated `osapi` user with sudo rules and Linux capabilities.

## Overview

The agent executes system commands via its exec manager. Some commands need root
privileges (user management, service control), while others work unprivileged
(package queries, reading logs). The hardening strategy:

1. Run the agent as a dedicated `osapi` user
2. Enable privilege escalation in the config
3. Grant sudo access only for specific commands
4. Set Linux capabilities on the binary for direct file operations
5. The agent verifies its privileges at startup before accepting jobs

## Configuration

```yaml
agent:
  privilege_escalation:
    # Enable sudo for write operations. When true, the exec manager
    # prepends "sudo" to commands that modify system state. Read
    # operations are never elevated.
    sudo: true
    # Enable capability verification at startup. When true, the agent
    # checks that the required Linux capabilities are set on the
    # binary before accepting jobs.
    capabilities: true
    # Run preflight checks at startup. Verifies sudo access and
    # capabilities are correctly configured. The agent refuses to
    # start if checks fail.
    preflight: true
```

When `privilege_escalation` is not set or all fields are false, the agent
behaves as before — commands run as the current user with no elevation.

## Design: Exec Manager Changes

The exec `Manager` interface gains a `RunPrivilegedCmd` method. Providers call
it for write operations and `RunCmd` for reads:

```go
type Manager interface {
    // RunCmd executes a command as the current user. Use for read
    // operations that don't modify system state.
    RunCmd(
        name string,
        args []string,
    ) (string, error)

    // RunPrivilegedCmd executes a command with privilege escalation.
    // When sudo is enabled in config, prepends "sudo" to the command.
    // When sudo is disabled, behaves identically to RunCmd.
    // Use for write operations that modify system state.
    RunPrivilegedCmd(
        name string,
        args []string,
    ) (string, error)

    // RunCmdFull executes a command with separate stdout/stderr
    // capture, an optional working directory, and a timeout.
    RunCmdFull(
        name string,
        args []string,
        cwd string,
        timeout int,
    ) (*CmdResult, error)
}
```

The `Exec` struct gains a `sudo bool` field set from config:

```go
type Exec struct {
    logger *slog.Logger
    sudo   bool
}

func New(
    logger *slog.Logger,
    sudo bool,
) *Exec {
    return &Exec{
        logger: logger.With(slog.String("subsystem", "exec")),
        sudo:   sudo,
    }
}

func (e *Exec) RunPrivilegedCmd(
    name string,
    args []string,
) (string, error) {
    if e.sudo {
        args = append([]string{name}, args...)
        name = "sudo"
    }
    return e.RunCmdImpl(name, args, "")
}
```

### Provider changes

Each provider's write operations change from `RunCmd` to `RunPrivilegedCmd`.
Read operations stay on `RunCmd`. Example from the service provider:

```go
// Read — no elevation
output, _ := d.execManager.RunCmd(
    "systemctl",
    []string{"is-active", unitName},
)

// Write — elevated when configured
_, err := d.execManager.RunPrivilegedCmd(
    "systemctl",
    []string{"start", unitName},
)
```

The mock `Manager` interface gains `RunPrivilegedCmd`. Tests for write
operations expect `RunPrivilegedCmd`, tests for reads expect `RunCmd`. This
enforces the read/write distinction at the test level — using the wrong method
causes a mock expectation failure.

## Design: Preflight Checks

When `preflight: true`, the agent runs checks during `agent start` before
accepting jobs. If any check fails, the agent logs the failure and exits.

### Sudo verification

For each command in the sudo whitelist, run `sudo -n <command> --help` (or
equivalent no-op) to verify the sudoers rule is configured. The `-n` flag
makes sudo fail immediately if a password would be required.

```go
func (p *Preflight) CheckSudo(
    commands []string,
) []PreflightResult {
    // For each command, run "sudo -n <command> --version" or
    // similar no-op to verify access without side effects.
}
```

### Capability verification

Read `/proc/self/status` and parse the `CapEff` line to check that required
capability bits are set:

```go
func (p *Preflight) CheckCapabilities(
    required []capability,
) []PreflightResult {
    // Parse /proc/self/status CapEff bitmask
    // Check each required capability bit
}
```

### Preflight output

```
OSAPI Agent Preflight Check
─────────────────────────────
Sudo access:
  ✓ systemctl    ✓ sysctl       ✓ timedatectl
  ✓ hostnamectl  ✓ chronyc      ✓ useradd
  ✓ usermod      ✓ userdel      ✓ groupadd
  ✓ groupdel     ✓ gpasswd      ✓ chown
  ✓ apt-get      ✓ shutdown     ✓ update-ca-certificates
  ✗ sh (sudoers rule missing)

Capabilities:
  ✓ CAP_DAC_READ_SEARCH   ✓ CAP_DAC_OVERRIDE
  ✓ CAP_FOWNER            ✓ CAP_KILL

Result: FAILED (1 error)
  - sudo: sh not configured in /etc/sudoers.d/osapi-agent
```

## Create the Service Account

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin osapi
```

## Sudoers Configuration

Create `/etc/sudoers.d/osapi-agent` with the following content. This grants the
`osapi` user passwordless sudo access to exactly the commands the agent needs.

```sudoers
# /etc/sudoers.d/osapi-agent
# OSAPI agent privilege escalation rules.
# Only the commands listed here can be run as root.

# Service management
osapi ALL=(root) NOPASSWD: /usr/bin/systemctl start *
osapi ALL=(root) NOPASSWD: /usr/bin/systemctl stop *
osapi ALL=(root) NOPASSWD: /usr/bin/systemctl restart *
osapi ALL=(root) NOPASSWD: /usr/bin/systemctl enable *
osapi ALL=(root) NOPASSWD: /usr/bin/systemctl disable *
osapi ALL=(root) NOPASSWD: /usr/bin/systemctl daemon-reload

# Kernel parameters
osapi ALL=(root) NOPASSWD: /usr/sbin/sysctl -p *
osapi ALL=(root) NOPASSWD: /usr/sbin/sysctl --system

# Timezone
osapi ALL=(root) NOPASSWD: /usr/bin/timedatectl set-timezone *

# Hostname
osapi ALL=(root) NOPASSWD: /usr/bin/hostnamectl set-hostname *

# NTP
osapi ALL=(root) NOPASSWD: /usr/bin/chronyc reload sources

# User and group management
osapi ALL=(root) NOPASSWD: /usr/sbin/useradd *
osapi ALL=(root) NOPASSWD: /usr/sbin/usermod *
osapi ALL=(root) NOPASSWD: /usr/sbin/userdel *
osapi ALL=(root) NOPASSWD: /usr/sbin/groupadd *
osapi ALL=(root) NOPASSWD: /usr/sbin/groupdel *
osapi ALL=(root) NOPASSWD: /usr/bin/gpasswd *
osapi ALL=(root) NOPASSWD: /usr/bin/chown *
osapi ALL=(root) NOPASSWD: /bin/sh -c echo *

# Package management
osapi ALL=(root) NOPASSWD: /usr/bin/apt-get install *
osapi ALL=(root) NOPASSWD: /usr/bin/apt-get remove *
osapi ALL=(root) NOPASSWD: /usr/bin/apt-get update

# Certificate trust store
osapi ALL=(root) NOPASSWD: /usr/sbin/update-ca-certificates

# Power management
osapi ALL=(root) NOPASSWD: /sbin/shutdown *
```

Note: read-only commands (`systemctl list-units`, `sysctl -n`, `journalctl`,
etc.) are NOT in the sudoers file — they run unprivileged via `RunCmd`.

Set permissions:

```bash
sudo cp osapi-agent.sudoers /etc/sudoers.d/osapi-agent
sudo chmod 0440 /etc/sudoers.d/osapi-agent
sudo visudo -c  # validate syntax
```

## Linux Capabilities

Capabilities grant the agent binary specific privileges without full root
access. These cover direct file operations that don't go through external
commands (reads via `avfs`, writes to config directories).

```bash
sudo setcap \
  'cap_dac_read_search+ep cap_dac_override+ep cap_fowner+ep cap_kill+ep' \
  /usr/local/bin/osapi
```

| Capability            | Purpose                                           |
| --------------------- | ------------------------------------------------- |
| `cap_dac_read_search` | Read any file (e.g., `/etc/shadow` for user info) |
| `cap_dac_override`    | Write files regardless of ownership (unit files,  |
|                       | cron files, sysctl conf, CA certs)                |
| `cap_fowner`          | Change file ownership (SSH key directories)       |
| `cap_kill`            | Send signals to any process                       |

Verify capabilities are set:

```bash
getcap /usr/local/bin/osapi
# Expected: /usr/local/bin/osapi cap_dac_override,cap_dac_read_search,cap_fowner,cap_kill=ep
```

## Systemd Unit File

Run the agent as the `osapi` user with ambient capabilities:

```ini
# /etc/systemd/system/osapi-agent.service
[Unit]
Description=OSAPI Agent
After=network.target

[Service]
Type=simple
User=osapi
Group=osapi
ExecStart=/usr/local/bin/osapi agent start
Restart=always
RestartSec=5

# Capabilities
AmbientCapabilities=CAP_DAC_READ_SEARCH CAP_DAC_OVERRIDE CAP_FOWNER CAP_KILL
CapabilityBoundingSet=CAP_DAC_READ_SEARCH CAP_DAC_OVERRIDE CAP_FOWNER CAP_KILL
SecureBits=keep-caps

# Hardening
NoNewPrivileges=no
ProtectSystem=false
ProtectHome=false
ReadWritePaths=/etc/systemd/system /etc/sysctl.d /etc/cron.d
ReadWritePaths=/usr/local/share/ca-certificates
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

Note: `NoNewPrivileges=no` is required because the agent uses `sudo` for
privileged commands. If `NoNewPrivileges=yes` were set, `sudo` would be
blocked.

## Command Privilege Reference

Commands the agent executes, grouped by privilege requirement. This determines
whether a provider calls `RunCmd` (read) or `RunPrivilegedCmd` (write).

### Write operations (use `RunPrivilegedCmd`)

| Command                   | Domain      | Operation         |
| ------------------------- | ----------- | ----------------- |
| `systemctl start/stop/…`  | Service     | Lifecycle control |
| `systemctl daemon-reload` | Service     | Unit file reload  |
| `sysctl -p`, `--system`   | Sysctl      | Apply parameters  |
| `timedatectl set-timezone` | Timezone   | Set timezone      |
| `hostnamectl set-hostname` | Hostname   | Set hostname      |
| `chronyc reload sources`  | NTP         | Apply NTP config  |
| `useradd`, `usermod`      | User        | User management   |
| `userdel -r`              | User        | Delete user       |
| `groupadd`, `groupdel`    | Group       | Group management  |
| `gpasswd -M`              | Group       | Set members       |
| `chown -R`                | SSH Key     | Fix ownership     |
| `apt-get install/remove`  | Package     | Install/remove    |
| `apt-get update`          | Package     | Update index      |
| `update-ca-certificates`  | Certificate | Rebuild trust     |
| `shutdown -r/-h`          | Power       | Reboot/shutdown   |
| `sh -c "echo … chpasswd"` | User       | Set password      |

### Read operations (use `RunCmd`)

| Command                      | Domain  | Operation       |
| ---------------------------- | ------- | --------------- |
| `systemctl list-units`       | Service | List services   |
| `systemctl list-unit-files`  | Service | List unit files |
| `systemctl show`             | Service | Get service info|
| `systemctl is-active`        | Service | Check status    |
| `systemctl is-enabled`       | Service | Check enabled   |
| `sysctl -n`                  | Sysctl  | Read parameter  |
| `timedatectl show`           | Timezone| Read timezone   |
| `hostnamectl hostname`       | Hostname| Read hostname   |
| `journalctl`                 | Log     | Query logs      |
| `chronyc tracking`           | NTP     | Read NTP status |
| `chronyc sources -c`         | NTP     | List sources    |
| `id -Gn`                     | User    | Get user groups |
| `passwd -S`                  | User    | Password status |
| `dpkg-query`                 | Package | Query packages  |
| `apt list --upgradable`      | Package | Check updates   |
| `date +%:z`                  | Timezone| Read UTC offset |

## Verification

After setup, verify the agent can operate:

```bash
# Switch to the osapi user and test
sudo -u osapi osapi agent start --dry-run

# Verify sudo works for allowed commands
sudo -u osapi sudo -n systemctl list-units --type=service --no-pager

# Verify sudo is denied for non-whitelisted commands
sudo -u osapi sudo -n rm /etc/passwd  # should be denied
```
