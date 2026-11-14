---
sidebar_position: 20
sidebar_label: Agent Hardening
---

# Agent hardening

osapi can run the agent as a user without privileges. Then a configuration
setting gives privilege escalation for write operations. Reads run as the user
of the agent. Writes use `sudo` when you configure it. Linux capabilities give
access to files without a full sudo setup. When you enable one of these options,
the agent verifies the configuration automatically when it starts, before it
accepts jobs.

When you enable none of these options, the agent runs commands as its current
user, root or a different user.

## Configuration

```yaml
agent:
  privilege_escalation:
    # Activate least-privilege mode: sudo for write commands
    # and capability verification at startup.
    enabled: false
```

| Field     | Type | Default | Description                                                       |
| --------- | ---- | ------- | ----------------------------------------------------------------- |
| `enabled` | bool | false   | Activate sudo for write commands and capability checks at startup |

## How it works

The exec manager has two ways to run a command:

- **`RunCmd`** runs the command as the current user of the agent. All read
  operations use it, for example to list services, read kernel parameters or get
  the package state.
- **`RunPrivilegedCmd`** runs the command with `sudo` in front of it when
  `privilege_escalation.enabled: true`. When the setting is disabled, it is the
  same as `RunCmd`.

Providers call `RunCmd` for reads and `RunPrivilegedCmd` for writes. The
providers do not know if `sudo` is enabled. The exec manager controls it.

```go
// Read — always unprivileged
output, _ := d.execManager.RunCmd("systemctl", []string{"is-active", name})

// Write — elevated when configured
_, err := d.execManager.RunPrivilegedCmd(
    "systemctl", []string{"start", name})
```

## Sudoers drop-in

Make `/etc/sudoers.d/osapi-agent` with this content. Then the `osapi` system
user can run write commands without a password:

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

Validate the file with `sudo visudo -c -f /etc/sudoers.d/osapi-agent` before you
reload.

## Linux capabilities

To give access to files without `sudo`, give the agent binary specific Linux
capabilities:

```bash
sudo setcap \
  'cap_dac_read_search+ep cap_dac_override+ep cap_chown+ep cap_fowner+ep cap_kill+ep cap_net_admin+ep' \
  /usr/local/bin/osapi
```

When `privilege_escalation.enabled: true`, the agent reads `/proc/self/status`
when it starts, and checks the `CapEff` bitmask for the required bits:

| Capability            | Bit | Purpose                                           |
| --------------------- | --- | ------------------------------------------------- |
| `CAP_DAC_READ_SEARCH` | 2   | Read restricted files (`/etc/shadow`, proc)       |
| `CAP_DAC_OVERRIDE`    | 1   | Write files of all owners (netplan, sysctl)       |
| `CAP_CHOWN`           | 0   | Set file ownership to root:root (netplan configs) |
| `CAP_FOWNER`          | 3   | Skip permission checks on owned files             |
| `CAP_KILL`            | 5   | Signal any process                                |
| `CAP_NET_ADMIN`       | 12  | Network configuration (interface, routes, DNS)    |

If a required capability is missing, the agent writes a warning to the log. Only
the operations that need the capability fail.

## Systemd unit file

Use this unit file to run the agent as a user without privileges. The unit keeps
the capabilities when the agent restarts:

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
AmbientCapabilities=CAP_DAC_READ_SEARCH CAP_DAC_OVERRIDE CAP_CHOWN CAP_FOWNER CAP_KILL CAP_NET_ADMIN
CapabilityBoundingSet=CAP_DAC_READ_SEARCH CAP_DAC_OVERRIDE CAP_CHOWN CAP_FOWNER CAP_KILL CAP_NET_ADMIN
SecureBits=keep-caps
NoNewPrivileges=no
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

`AmbientCapabilities` gives the capabilities to the process. Then the binary
does not need `setcap`. If you also use `sudo`, set `NoNewPrivileges=no`, so
that `sudo` can increase the privileges.

## Preflight checks

When `enabled` is true, the agent runs a check during `agent start`, before it
subscribes to NATS.

The agent lists its sudo rules once with `sudo -n -l`. The `-n` flag makes sudo
fail at once if it needs a password. If that fails, sudo is not usable, and the
agent logs the error and exits.

Then the agent finds a rule for each program it runs with sudo, by the program's
name. It skips a program that is not installed. It does not check the rule's
arguments, so a narrow rule such as
`/usr/sbin/sysctl -p /etc/sysctl.d/osapi-*.conf` counts as a rule for `sysctl`.
A program with no rule is a warning, not an error: the agent starts, and only
the operations that run that program fail.

The agent also checks each capability above. A missing capability is a warning
too.

The agent checks the programs above, and each program that a behavior declares
with `behavior.Sudo`. The page of a behavior under
[Behaviors](../behaviors/behaviors.mdx) names the programs that it runs with
sudo. `osapi behavior get --name <name>` also lists them. Thus you find the
sudoers entry of a new behavior there.

Example log, with no rule for `useradd`:

```
WRN preflight check failed check=sudo:useradd error="no sudo rule runs useradd; operations that need it fail"
INF preflight checks done
```

## Command reference

### Write operations (use `RunPrivilegedCmd`)

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

## What does not change

- **Controller and NATS server.** They already run without privileges. They need
  no change.
- **`command exec` and `command shell`.** These endpoints run commands that the
  user gives, with all the privileges of the agent. `exec` needs the
  `command:execute` RBAC permission. `shell` also needs `command:shell`.
- **Docker provider.** It connects to the Docker API socket. It does not run
  system commands. The `osapi` user must be in the `docker` group.
