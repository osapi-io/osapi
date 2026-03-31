---
sidebar_position: 4
---

# Package

The `Package` service provides methods for managing system packages on target
hosts. Access via `client.Package.List()`, `client.Package.Install()`, etc.

## Methods

| Method                        | Description                     |
| ----------------------------- | ------------------------------- |
| `List(ctx, hostname)`         | List all installed packages     |
| `Get(ctx, hostname, name)`    | Get a package by name           |
| `Install(ctx, hostname, name)`| Install a package               |
| `Remove(ctx, hostname, name)` | Remove a package                |
| `Update(ctx, hostname)`       | Refresh package sources         |
| `ListUpdates(ctx, hostname)`  | List available package updates  |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all installed packages
resp, err := c.Package.List(ctx, "web-01")
for _, r := range resp.Data.Results {
    for _, p := range r.Packages {
        fmt.Printf("%s %s (%s)\n", p.Name, p.Version, p.Status)
    }
}

// Get a specific package
resp, err := c.Package.Get(ctx, "web-01", "nginx")

// Install a package
resp, err := c.Package.Install(ctx, "web-01", "nginx")
for _, r := range resp.Data.Results {
    fmt.Printf("changed=%v\n", r.Changed)
}

// Remove a package
resp, err := c.Package.Remove(ctx, "web-01", "nginx")

// Refresh package sources (apt-get update)
resp, err := c.Package.Update(ctx, "web-01")

// List available updates
resp, err := c.Package.ListUpdates(ctx, "web-01")
for _, r := range resp.Data.Results {
    for _, u := range r.Updates {
        fmt.Printf("%s: %s -> %s\n", u.Name, u.CurrentVersion, u.NewVersion)
    }
}
```

## Example

- [`examples/sdk/client/package.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/package.go)

## Permissions

| Operation                       | Permission      |
| ------------------------------- | --------------- |
| List, Get, ListUpdates          | `package:read`  |
| Install, Remove, Update Sources | `package:write` |

Package management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
