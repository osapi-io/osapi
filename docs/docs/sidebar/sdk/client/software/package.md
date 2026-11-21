---
sidebar_position: 4
---

# Package

The `Package` service has methods that manage system packages on target hosts.
Use `client.Package.List()` and `client.Package.Install()`, for example.

## Methods

| Method                                  | Description                                  |
| --------------------------------------- | -------------------------------------------- |
| `List(ctx, hostname)`                   | List all installed packages                  |
| `Get(ctx, hostname, name)`              | Get a package by name                        |
| `Install(ctx, hostname, name, version)` | Install a package, at version when not empty |
| `Remove(ctx, hostname, name)`           | Remove a package                             |
| `Update(ctx, hostname)`                 | Refresh the package index                    |
| `ListUpdates(ctx, hostname)`            | List available package updates               |

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
resp, err := c.Package.Install(ctx, "web-01", "nginx", "")
for _, r := range resp.Data.Results {
    fmt.Printf("changed=%v\n", r.Changed)
}

// Remove a package
resp, err := c.Package.Remove(ctx, "web-01", "nginx")

// Refresh the package index (apt-get update)
resp, err := c.Package.Update(ctx, "web-01")

// List available updates
resp, err := c.Package.ListUpdates(ctx, "web-01")
for _, r := range resp.Data.Results {
    fmt.Printf("%s: %s -> %s\n", r.Name, r.CurrentVersion, r.NewVersion)
}
```

## Example

- [`examples/sdk/client/package.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/package.go)

## Permissions

| Operation       | Permission             |
| --------------- | ---------------------- |
| List, Get       | `package:read`         |
| Install, Remove | `package:write`        |
| ListUpdates     | `updatepackages:read`  |
| Update          | `updatepackages:write` |

List and Get run on every platform: dpkg, rpm or Homebrew. Install, Remove,
ListUpdates and Update run on the Debian family. On other hosts they return
`status: skipped`.
