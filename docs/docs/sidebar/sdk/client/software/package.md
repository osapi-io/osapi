---
sidebar_position: 4
---

# Package

The `Package` service has methods that manage system packages on target hosts.
Use `client.Package.List()` and `client.Package.Install()`, for example.

## Methods

| Method                                                      | Returns                                         | Permission             | Description                                       |
| ----------------------------------------------------------- | ----------------------------------------------- | ---------------------- | ------------------------------------------------- |
| `c.Package.List(ctx, hostname)`                             | `Response[Collection[PackageInstalled]]`        | `package:read`         | List the packages installed on the host           |
| `c.Package.Create(ctx, hostname, input PackageCreateInput)` | `Response[Collection[PackageResult]]`           | `package:write`        | Install a package                                 |
| `c.Package.Get(ctx, hostname, name)`                        | `Response[Collection[PackageInstalled]]`        | `package:read`         | Get one installed package                         |
| `c.Package.Delete(ctx, hostname, name)`                     | `Response[Collection[PackageResult]]`           | `package:write`        | Remove a package                                  |
| `c.UpdatePackages.List(ctx, hostname)`                      | `Response[Collection[UpdatepackagesAvailable]]` | `updatepackages:read`  | List the packages the index has newer versions of |
| `c.UpdatePackages.Update(ctx, hostname)`                    | `Response[Collection[UpdatepackagesResult]]`    | `updatepackages:write` | Refresh the package index                         |

## Types

The fields of each type are the properties of its schema in the API's spec.

### PackageInstalled

| Field          | Type     | Description                                                                            |
| -------------- | -------- | -------------------------------------------------------------------------------------- |
| `Architecture` | `string` | The architecture it is built for, such as amd64.                                       |
| `Description`  | `string` | Its one-line summary.                                                                  |
| `Error`        | `string` | Error message if the agent failed.                                                     |
| `Hostname`     | `string` | Hostname of the agent that produced this result.                                       |
| `Manager`      | `string` | The package database that has it: dpkg, rpm or brew.                                   |
| `Name`         | `string` | The package.                                                                           |
| `Size`         | `int64`  | Its installed size in bytes.                                                           |
| `Status`       | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Version`      | `string` | The installed version.                                                                 |

### PackageCreateInput

| Field     | Type     | Description                                                             |
| --------- | -------- | ----------------------------------------------------------------------- |
| `Name`    | `string` | The package to install, such as curl.                                   |
| `Version` | `string` | The version to install, such as 8.5.0-2ubuntu10; the newest when empty. |

### PackageResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Name`     | `string` | The package.                                                                           |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### UpdatepackagesAvailable

| Field            | Type     | Description                                                                            |
| ---------------- | -------- | -------------------------------------------------------------------------------------- |
| `CurrentVersion` | `string` | The installed version.                                                                 |
| `Error`          | `string` | Error message if the agent failed.                                                     |
| `Hostname`       | `string` | Hostname of the agent that produced this result.                                       |
| `Name`           | `string` | The package.                                                                           |
| `NewVersion`     | `string` | The version the index has.                                                             |
| `Status`         | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### UpdatepackagesResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the refresh got new index data.                                                |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all installed packages: one row per package per host
resp, err := c.Package.List(ctx, "web-01")
for _, r := range resp.Data.Results {
    fmt.Printf("%s %s (%s)\n", r.Name, r.Version, r.Manager)
}

// Get a specific package
resp, err := c.Package.Get(ctx, "web-01", "nginx")

// Install a package
resp, err := c.Package.Create(ctx, "web-01", client.PackageCreateInput{Name: "nginx"})
for _, r := range resp.Data.Results {
    fmt.Printf("changed=%v\n", r.Changed)
}

// Remove a package
resp, err := c.Package.Delete(ctx, "web-01", "nginx")

// Refresh the package index (apt-get update)
resp, err := c.UpdatePackages.Update(ctx, "web-01")

// List available updates
resp, err := c.UpdatePackages.List(ctx, "web-01")
for _, r := range resp.Data.Results {
    fmt.Printf("%s: %s -> %s\n", r.Name, r.CurrentVersion, r.NewVersion)
}
```

## Example

- [`examples/sdk/client/package.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/package.go)

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.

Package.List and Package.Get run on every platform: dpkg, rpm or Homebrew.
Package.Create, Package.Delete and the UpdatePackages methods run on the Debian
family. On other hosts they return `status: skipped`.
