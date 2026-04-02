---
sidebar_position: 20
sidebar_label: Agent Hardening
---

# Agent Hardening

OSAPI supports running the agent as an unprivileged user with config-driven
privilege escalation for write operations. Reads run as the agent's own
user. Writes use `sudo` when configured. Linux capabilities provide an
alternative for file-level access without a full sudo setup. When either
option is enabled, the agent automatically verifies the configuration at
startup before accepting any jobs.

When none of these options are enabled, the agent behaves as before —
commands run as the current user, root or otherwise.

## Configuration

```yaml
agent:
  privilege_escalation:
    # Prepend "sudo" to write commands.
    sudo: false
    # Verify Linux capabilities at startup.
    capabilities: false
```

| Field          | Type | Default | Description                          |
| -------------- | ---- | ------- | ------------------------------------ |
| `sudo`         | bool | false   | Prepend `sudo` to write commands     |
| `capabilities` | bool | false   | Verify Linux capabilities at startup |

## How It Works

The exec manager exposes two execution paths:

- **`RunCmd`** — runs the command as the agent's current user. Used for all
  read operations (listing services, reading kernel parameters, querying
  package state, etc.).
- **`RunPrivilegedCmd`** — runs the command with `sudo` prepended when
  `privilege_escalation.sudo: true`. When `sudo` is false, this is
  identical to `RunCmd`.

Providers call `RunCmd` for reads and `RunPrivilegedCmd` for writes. The
providers themselves have no knowledge of whether `sudo` is enabled — the
exec manager handles it transparently.

```go
// Read — always unprivileged
output, _ := d.execManager.RunCmd("systemctl", []string{"is-active", name})

// Write — elevated when configured
_, err := d.execManager.RunPrivilegedCmd(
    "systemctl", []string{"start", name})
```

## Sudoers Drop-In

Create `/etc/sudoers.d/osapi-agent` with the following content to allow
the `osapi` system user to run write commands without a password:

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

Validate the file with `sudo visudo -c -f /etc/sudoers.d/osapi-agent`
before reloading.

## Linux Capabilities

As an alternative to `sudo` for file-level access, grant the agent binary
specific Linux capabilities:

```bash
sudo setcap \
  'cap_dac_read_search+ep cap_dac_override+ep cap_fowner+ep cap_kill+ep' \
  /usr/local/bin/osapi
```

When `privilege_escalation.capabilities: true`, the agent reads
`/proc/self/status` at startup and checks the `CapEff` bitmask for the
required bits:

| Capability            | Bit | Purpose                         |
| --------------------- | --- | ------------------------------- |
| `CAP_DAC_READ_SEARCH` | 2   | Read restricted files           |
| `CAP_DAC_OVERRIDE`    | 1   | Write files regardless of owner |
| `CAP_FOWNER`          | 3   | Change file ownership           |
| `CAP_KILL`            | 5   | Signal any process              |

If any required capability is missing the agent logs the failure and
exits with a non-zero status.

## Systemd Unit File

The recommended way to run the agent as an unprivileged user with
capabilities preserved across restarts:

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

`AmbientCapabilities` grants the capabilities to the process without
requiring `setcap` on the binary. `NoNewPrivileges=no` is required so
that `sudo` (if also used) can elevate correctly.

## Preflight Checks

When `sudo` or `capabilities` is enabled, the agent automatically runs a
verification pass during `agent start` before subscribing to NATS. Checks
are sequential: sudo first, then capabilities. If any check fails, the
agent logs the failure and exits with a non-zero status.

The sudo check runs `sudo -n <command> --version` (or `sudo -n which
<command>` for commands that do not support `--version`). The `-n` flag
makes `sudo` fail immediately if a password prompt would be required,
confirming that the sudoers entry is present and correct.

Example output:

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

## Command Reference

### Write Operations (use `RunPrivilegedCmd`)

| Command                    | Domain      |
| -------------------------- | ----------- |
| `systemctl start/stop/…`   | Service     |
| `systemctl daemon-reload`  | Service     |
| `sysctl -p`, `--system`    | Sysctl      |
| `timedatectl set-timezone` | Timezone    |
| `hostnamectl set-hostname` | Hostname    |
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

### Read Operations (use `RunCmd`)

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

## What Is Not Changed

- **Controller and NATS server** — already run unprivileged, no changes
  needed.
- **`command exec` and `command shell`** — these endpoints execute
  arbitrary user-provided commands and inherit whatever privileges the
  agent has. They are gated by the `command:execute` RBAC permission.
- **Docker provider** — talks to the Docker API socket, not system
  commands. The `osapi` user needs to be in the `docker` group.
