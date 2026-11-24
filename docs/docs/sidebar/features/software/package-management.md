# Package management

osapi reads the packages installed on each host, and installs and removes them
on the Debian family. It is two behaviors:

- [package](../../reference/behaviors/software/package/package.md): the
  installed packages.
- [package updates](../../reference/behaviors/software/package/updates/updates.md):
  the updates that the package index has, and the refresh of the index.

## How it works

### List and get

gohai reads the installed packages on every platform: `dpkg-query` on the Debian
family, `rpm` on the other Linux families, and Homebrew on macOS. Each package
has its name, version, architecture, package database, summary and installed
size. Get reads one package from the same list. A package that is not installed
is not found.

### Install and remove

`apt-get install` and `apt-get remove` run on the Debian family only. Install
takes an optional `version`, such as `--version 8.5.0-2ubuntu10`; with none, it
installs the newest that the index has. On other hosts they report skipped. A
package that is installed does not install again, and a package that is not
installed does not remove: each reports `changed: false`. A package that is
installed at another version than the one asked for installs that version.
Install does not refresh the index first. Run the index refresh for the newest
version.

### The package index

The index is not a fact about the host. It is what the repositories have, so
`apt` reads it, not gohai. It runs on the Debian family. Its commands are under
the packages, `osapi client node package updates`, at
`/api/node/{hostname}/package/updates`, and they need `package:read` and
`package:write`. The route of the updates comes before the route of one package.
Thus `get` cannot read a package with the name `updates`.

- The list runs `apt list --upgradable`: each package with its installed version
  and the version that the index has.
- The update runs `apt-get update`. It reports `changed: true` when the Release
  file of a repository under `/var/lib/apt/lists` changed, and `changed: false`
  when each repository was unchanged. apt-get prints `Get:` lines on some hosts
  even when nothing changed, so its output does not tell.

## Permissions and platforms

The generated page of each behavior gives the permission, the platforms and the
route of each method:
[package](../../reference/behaviors/software/package/package.md),
[package updates](../../reference/behaviors/software/package/updates/updates.md).
The pages come from the definitions, so they stay true when a method changes,
and this page does not copy them.
[Authentication](../platform/authentication.md) tells which permissions each
role has.

## CLI usage

```bash
# List all installed packages on a host
osapi client node package list --target web-01

# Get one package
osapi client node package get --target web-01 --name nginx

# Install a package
osapi client node package create --target web-01 --name nginx

# Remove a package
osapi client node package delete --target web-01 --name nginx

# Refresh the package index (apt-get update)
osapi client node package updates update --target web-01

# List available updates
osapi client node package updates list --target web-01

# Broadcast install to all hosts
osapi client node package create --target _all --name htop
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

## Related

- [CLI Reference](../../usage/cli/client/node/package/package.mdx): package
  commands
- [Platform Detection](../../sdk/library/platform-detection.md): OS family
  detection
- [Configuration](../../usage/configuration.md): full configuration reference
