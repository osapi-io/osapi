---
sidebar_position: 15
---

# Package Management

OSAPI manages system packages on target hosts. It wraps the native package
manager (`apt` on Debian-family systems) behind a consistent API that supports
listing installed packages, installing, removing, refreshing sources, and
checking for available updates.

## How It Works

The package provider delegates to the host's native package manager. On
Debian-family systems this means `dpkg-query` for reads and `apt-get` for
mutations. All operations run on the agent -- the controller never executes
package commands directly.

### List and Get

The list operation queries the local package database for all installed packages.
Each entry includes the package name, version, status, and installed size. The
get operation fetches a single package by name.

### Install

Installs a package by name using `apt-get install`. If the package is already
installed, the operation returns `changed: false`. The agent runs
`apt-get update` automatically before installing to ensure the latest version is
available.

### Remove

Removes a package by name using `apt-get remove`. If the package is not
installed, the operation returns `changed: false`.

### Update (Refresh Sources)

Refreshes the package source lists using `apt-get update`. This does not upgrade
any packages -- it only updates the local cache of available packages from
configured repositories. Returns `changed: true` when sources are refreshed
successfully.

### List Available Updates

Queries for packages that have newer versions available in the configured
repositories. Returns each package name along with the current installed version
and the new available version.

## Operations

| Operation | Description                          |
| --------- | ------------------------------------ |
| List      | List all installed packages          |
| Get       | Get a specific package by name       |
| Install   | Install a package                    |
| Remove    | Remove a package                     |
| Update    | Refresh package sources              |
| Updates   | List packages with available updates |

## CLI Usage

```bash
# List all installed packages on a host
osapi client node package list --target web-01

# Get details for a specific package
osapi client node package get --target web-01 --name nginx

# Install a package
osapi client node package install --target web-01 --name nginx

# Remove a package
osapi client node package remove --target web-01 --name nginx

# Refresh package sources (apt-get update)
osapi client node package update --target web-01

# List available updates
osapi client node package updates --target web-01

# Broadcast install to all hosts
osapi client node package install --target _all --name htop
```

All commands support `--json` for raw JSON output.

## Broadcast Support

All package operations support broadcast targeting. Use `--target _all` to run
on every registered agent, or use a label selector like `--target group:web` to
target a subset.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME    VERSION      STATUS      SIZE
  web-01    nginx   1.24.0-2     installed   1.2 MB
  web-02    nginx   1.24.0-2     installed   1.2 MB
```

Skipped and failed hosts appear with their error in the output.

## Supported Platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On unsupported platforms, package operations return `status: skipped` instead of
failing. See [Platform Detection](../sdk/platform/detection.md) for details on
OS family detection.

## Permissions

| Operation                       | Permission      |
| ------------------------------- | --------------- |
| List, Get, Updates              | `package:read`  |
| Install, Remove, Update Sources | `package:write` |

All built-in roles (`admin`, `write`, `read`) include `package:read`. The
`admin` and `write` roles also include `package:write`.

## Related

- [CLI Reference](../usage/cli/client/node/package/package.md) -- package
  commands
- [Platform Detection](../sdk/platform/detection.md) -- OS family detection
- [Configuration](../usage/configuration.md) -- full configuration reference
