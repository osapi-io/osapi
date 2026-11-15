---
sidebar_position: 15
---

# Package management

osapi manages system packages on target hosts. It puts one API in front of the
package manager of the host (`apt` on Debian family systems). The API can list
installed packages, install, remove, refresh the sources and check for updates.

## How it works

The package provider uses the package manager of the host. On Debian family
systems, it uses `dpkg-query` for reads and `apt-get` for changes. All
operations run on the agent. The controller does not run package commands.

### List and get

The list operation queries the local package database for all installed
packages. Each entry includes the package name, version, status, and installed
size. The get operation fetches a single package by name.

### Install

Installs a package by name with `apt-get install`. If the package is already
installed, the operation returns `changed: false`. The agent runs
`apt-get update` automatically before it installs, so that the latest version is
available.

### Remove

Removes a package by name with `apt-get remove`. If the package is not
installed, the operation returns `changed: false`.

### Update (Refresh sources)

Refreshes the package source lists with `apt-get update`. This does not upgrade
packages. It only updates the local cache of available packages from the
configured repositories. When the refresh is successful, it returns
`changed: true`.

### List available updates

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

## CLI usage

```bash
# List all installed packages on a host
osapi client node package list --target web-01

# Get details for a specific package
osapi client node package get --target web-01 --name nginx

# Install a package
osapi client node package install --target web-01 --name nginx

# Remove a package
osapi client node package uninstall --target web-01 --name nginx

# Refresh package sources (apt-get update)
osapi client node package update --target web-01

# List available updates
osapi client node package list --updates --target web-01

# Broadcast install to all hosts
osapi client node package install --target _all --name htop
```

All commands accept `--json` to give the raw JSON output.

## Broadcast support

All package operations accept broadcast targets. Use `--target _all` to run on
all registered agents. Use a label selector, for example `--target group:web`,
to target a part of the fleet.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME    VERSION      STATUS      SIZE
  web-01    nginx   1.24.0-2     installed   1.2 MB
  web-02    nginx   1.24.0-2     installed   1.2 MB
```

Skipped and failed hosts appear with their error in the output.

## Supported platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On a platform that is not supported, package operations return
`status: skipped`. They do not fail. Refer to
[Platform Detection](../sdk/platform/detection.md) for how osapi finds the OS
family.

## Permissions

| Operation                       | Permission      |
| ------------------------------- | --------------- |
| List, Get, Updates              | `package:read`  |
| Install, Remove, Update Sources | `package:write` |

All built-in roles (`admin`, `write`, `read`) include `package:read`. The
`admin` and `write` roles also include `package:write`.

## Related

- [CLI Reference](../usage/cli/client/node/package/package.mdx): package commands
- [Platform Detection](../sdk/platform/detection.md): OS family detection
- [Configuration](../usage/configuration.md): full configuration reference
