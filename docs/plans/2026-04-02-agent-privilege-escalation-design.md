# Agent Privilege Escalation Design

Run the OSAPI agent as an unprivileged user with config-driven sudo escalation
for write operations, Linux capabilities for direct file access, and preflight
verification at startup.

## Problem

The agent runs as root by default. This grants full system access to a
network-facing process that accepts jobs from NATS. A compromised agent (or a
malicious job) has unrestricted access to the host. The guiding principles call
for least-privilege mode.

## Solution

Split command execution into read and write paths. Reads run unprivileged.
Writes run through `sudo` when configured. The agent verifies its privileges at
startup and refuses to start if the configuration doesn't match the system
state.

## Config

```yaml
agent:
  privilege_escalation:
    sudo: true
    capabilities: true
    preflight: true
```

| Field          | Type | Default | Description                              |
| -------------- | ---- | ------- | ---------------------------------------- |
| `sudo`         | bool | false   | Prepend `sudo` to write commands         |
| `capabilities` | bool | false   | Verify Linux capabilities at startup     |
| `preflight`    | bool | false   | Run privilege checks before accepting jobs|

When all fields are false (or the section is absent), the agent behaves as
before — commands run as the current user.

## Exec Manager Interface

Add `RunPrivilegedCmd` to the `Manager` interface:

```go
type Manager interface {
    RunCmd(name string, args []string) (string, error)
    RunPrivilegedCmd(name string, args []string) (string, error)
    RunCmdFull(name string, args []string, cwd string, timeout int) (*CmdResult, error)
}
```

The `Exec` struct gains a `sudo bool` field:

```go
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

When `sudo` is false, `RunPrivilegedCmd` is identical to `RunCmd`.

## Provider Changes

Every provider write operation changes from `RunCmd` to `RunPrivilegedCmd`.
Read operations stay on `RunCmd`. The providers themselves don't know or care
whether sudo is enabled — the exec manager handles it.

```go
// Read — always unprivileged
output, _ := d.execManager.RunCmd("systemctl", []string{"is-active", name})

// Write — elevated when configured
_, err := d.execManager.RunPrivilegedCmd("systemctl", []string{"start", name})
```

Tests enforce this: the mock `Manager` has both methods. If a write operation
calls `RunCmd` instead of `RunPrivilegedCmd`, the mock expectation fails.

## Command Classification

### Write operations (use `RunPrivilegedCmd`)

| Command                    | Domain      |
| -------------------------- | ----------- |
| `systemctl start/stop/…`   | Service     |
| `systemctl daemon-reload`  | Service     |
| `sysctl -p`, `--system`    | Sysctl      |
| `timedatectl set-timezone`  | Timezone   |
| `hostnamectl set-hostname`  | Hostname   |
| `chronyc reload sources`   | NTP         |
| `useradd`, `usermod`       | User        |
| `userdel -r`               | User        |
| `groupadd`, `groupdel`     | Group       |
| `gpasswd -M`               | Group       |
| `chown -R`                 | SSH Key     |
| `apt-get install/remove`   | Package     |
| `apt-get update`           | Package     |
| `update-ca-certificates`   | Certificate |
| `shutdown -r/-h`           | Power       |
| `sh -c "echo … chpasswd"`  | User        |

### Read operations (use `RunCmd`)

| Command                     | Domain   |
| --------------------------- | -------- |
| `systemctl list-units`      | Service  |
| `systemctl list-unit-files` | Service  |
| `systemctl show`            | Service  |
| `systemctl is-active`       | Service  |
| `systemctl is-enabled`      | Service  |
| `sysctl -n`                 | Sysctl   |
| `timedatectl show`          | Timezone |
| `hostnamectl hostname`      | Hostname |
| `journalctl`                | Log      |
| `chronyc tracking`          | NTP      |
| `chronyc sources -c`        | NTP      |
| `id -Gn`                    | User     |
| `passwd -S`                 | User     |
| `dpkg-query`                | Package  |
| `apt list --upgradable`     | Package  |
| `date +%:z`                 | Timezone |

## Preflight Checks

Run during `agent start` before the agent subscribes to NATS. Checks are
sequential: sudo first, then capabilities. If any check fails, the agent logs
the failure and exits with a non-zero status.

### Sudo verification

For each write command, run `sudo -n <command> --version` (or equivalent no-op
flag). The `-n` flag makes sudo fail immediately if a password would be
required. If the command doesn't support `--version`, use `sudo -n which
<command>` as a fallback.

### Capability verification

Read `/proc/self/status`, parse the `CapEff` hexadecimal bitmask, and check
that required capability bits are set:

| Capability            | Bit | Purpose                      |
| --------------------- | --- | ---------------------------- |
| `CAP_DAC_READ_SEARCH` | 2   | Read restricted files        |
| `CAP_DAC_OVERRIDE`    | 1   | Write files regardless of owner |
| `CAP_FOWNER`          | 3   | Change file ownership        |
| `CAP_KILL`            | 5   | Signal any process           |

### Output format

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

## Deployment Artifacts

### Sudoers drop-in (`/etc/sudoers.d/osapi-agent`)

```sudoers
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

### Linux capabilities

```bash
sudo setcap \
  'cap_dac_read_search+ep cap_dac_override+ep cap_fowner+ep cap_kill+ep' \
  /usr/local/bin/osapi
```

### Systemd unit file

```ini
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
AmbientCapabilities=CAP_DAC_READ_SEARCH CAP_DAC_OVERRIDE CAP_FOWNER CAP_KILL
CapabilityBoundingSet=CAP_DAC_READ_SEARCH CAP_DAC_OVERRIDE CAP_FOWNER CAP_KILL
SecureBits=keep-caps
NoNewPrivileges=no
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

## Not Changing

- Controller and NATS server — already run unprivileged, no changes needed
- The `command exec` and `command shell` endpoints — these execute arbitrary
  user-provided commands, so they inherit whatever privileges the agent has.
  They are gated by the `command:execute` permission in RBAC.
- Docker provider — talks to the Docker API socket, not system commands. The
  `osapi` user needs to be in the `docker` group.

## Files Changed

- `internal/config/types.go` — add `PrivilegeEscalation` struct
- `internal/exec/manager.go` — add `RunPrivilegedCmd` to interface
- `internal/exec/types.go` — add `sudo bool` to `Exec` struct
- `internal/exec/run_privileged_cmd.go` — new file, implementation
- `internal/exec/run_privileged_cmd_public_test.go` — tests
- `internal/exec/mocks/` — regenerate
- `internal/agent/preflight.go` — new file, sudo + caps verification
- `internal/agent/preflight_public_test.go` — tests
- `internal/agent/agent.go` — call preflight during `Start()`
- `cmd/agent_setup.go` — pass `sudo` bool to exec manager
- Every provider `debian*.go` file — change write `RunCmd` to
  `RunPrivilegedCmd` (~37 call sites)
- Every provider `debian*_public_test.go` — update mock expectations
- `docs/docs/sidebar/usage/configuration.md` — add config reference
- `docs/docs/sidebar/features/` — add agent hardening feature page
