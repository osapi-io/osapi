# Agent Hardening

This guide covers running the OSAPI agent as an unprivileged user with minimal
permissions. By default the agent runs as root, but production deployments
should use a dedicated `osapi` user with sudo rules and Linux capabilities.

## Overview

The agent executes system commands via its exec manager. Some commands need root
privileges (user management, service control), while others work unprivileged
(package queries, reading logs). The hardening strategy:

1. Run the agent as a dedicated `osapi` user
2. Grant sudo access only for specific commands
3. Set Linux capabilities on the binary for direct file operations

## Create the Service Account

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin osapi
```

## Sudoers Configuration

Create `/etc/sudoers.d/osapi-agent` with the following content. This grants
the `osapi` user passwordless sudo access to exactly the commands the agent
needs — nothing more.

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
osapi ALL=(root) NOPASSWD: /usr/bin/systemctl is-active *
osapi ALL=(root) NOPASSWD: /usr/bin/systemctl is-enabled *
osapi ALL=(root) NOPASSWD: /usr/bin/systemctl show *
osapi ALL=(root) NOPASSWD: /usr/bin/systemctl list-units *
osapi ALL=(root) NOPASSWD: /usr/bin/systemctl list-unit-files *

# Kernel parameters
osapi ALL=(root) NOPASSWD: /usr/sbin/sysctl -p *
osapi ALL=(root) NOPASSWD: /usr/sbin/sysctl --system
osapi ALL=(root) NOPASSWD: /usr/sbin/sysctl -n *

# Timezone
osapi ALL=(root) NOPASSWD: /usr/bin/timedatectl show *
osapi ALL=(root) NOPASSWD: /usr/bin/timedatectl set-timezone *

# Hostname
osapi ALL=(root) NOPASSWD: /usr/bin/hostnamectl hostname
osapi ALL=(root) NOPASSWD: /usr/bin/hostnamectl set-hostname *

# System logs (read-only, but journalctl often needs root for full access)
osapi ALL=(root) NOPASSWD: /usr/bin/journalctl *

# NTP
osapi ALL=(root) NOPASSWD: /usr/bin/chronyc tracking
osapi ALL=(root) NOPASSWD: /usr/bin/chronyc sources *
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

Set permissions:

```bash
sudo cp osapi-agent.sudoers /etc/sudoers.d/osapi-agent
sudo chmod 0440 /etc/sudoers.d/osapi-agent
sudo visudo -c  # validate syntax
```

## Linux Capabilities

Capabilities grant the agent binary specific privileges without full root
access. These cover direct file operations that don't go through external
commands.

```bash
sudo setcap \
  'cap_dac_read_search+ep cap_dac_override+ep cap_fowner+ep cap_kill+ep' \
  /usr/local/bin/osapi
```

| Capability             | Purpose                                          |
| ---------------------- | ------------------------------------------------ |
| `cap_dac_read_search`  | Read any file (e.g., `/etc/shadow` for user info)|
| `cap_dac_override`     | Write files regardless of ownership (unit files, |
|                        | cron files, sysctl conf, CA certs)               |
| `cap_fowner`           | Change file ownership (SSH key directories)      |
| `cap_kill`             | Send signals to any process                      |

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

Commands the agent executes, grouped by privilege requirement:

### Requires sudo

| Command                  | Domain      | Operation         |
| ------------------------ | ----------- | ----------------- |
| `systemctl start/stop/…` | Service     | Lifecycle control |
| `systemctl daemon-reload`| Service     | Unit file reload  |
| `sysctl -p`, `--system`  | Sysctl      | Apply parameters  |
| `timedatectl set-timezone`| Timezone   | Set timezone      |
| `hostnamectl set-hostname`| Hostname   | Set hostname      |
| `chronyc reload sources` | NTP         | Apply NTP config  |
| `useradd`, `usermod`     | User        | User management   |
| `userdel -r`             | User        | Delete user       |
| `groupadd`, `groupdel`   | Group       | Group management  |
| `gpasswd -M`             | Group       | Set members       |
| `chown -R`               | SSH Key     | Fix ownership     |
| `apt-get install/remove`  | Package    | Install/remove    |
| `apt-get update`         | Package     | Update index      |
| `update-ca-certificates` | Certificate | Rebuild trust     |
| `shutdown -r/-h`         | Power       | Reboot/shutdown   |
| `sh -c "echo … chpasswd"`| User       | Set password      |

### Runs unprivileged (or with capabilities)

| Command                  | Domain      | Operation         |
| ------------------------ | ----------- | ----------------- |
| `systemctl list-units`   | Service     | List services     |
| `systemctl list-unit-files`| Service   | List unit files   |
| `systemctl show`         | Service     | Get service info  |
| `systemctl is-active`    | Service     | Check status      |
| `systemctl is-enabled`   | Service     | Check enabled     |
| `sysctl -n`              | Sysctl      | Read parameter    |
| `timedatectl show`       | Timezone    | Read timezone     |
| `hostnamectl hostname`   | Hostname    | Read hostname     |
| `journalctl`             | Log         | Query logs        |
| `chronyc tracking`       | NTP         | Read NTP status   |
| `chronyc sources -c`     | NTP         | List sources      |
| `id -Gn`                 | User        | Get user groups   |
| `passwd -S`              | User        | Password status   |
| `dpkg-query`             | Package     | Query packages    |
| `apt list --upgradable`  | Package     | Check updates     |
| `date +%:z`              | Timezone    | Read UTC offset   |

## Verification

After setup, verify the agent can operate:

```bash
# Switch to the osapi user and test
sudo -u osapi osapi agent start --dry-run

# Verify sudo works for allowed commands
sudo -u osapi sudo systemctl list-units --type=service --no-pager

# Verify sudo is denied for non-whitelisted commands
sudo -u osapi sudo rm /etc/passwd  # should be denied
```
