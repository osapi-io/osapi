---
sidebar_position: 4
---

# Detection

The `platform` package detects the OS family of the running system. The agent
uses it to select the correct provider implementation (Debian, Darwin, or
generic Linux). SDK consumers can use it to make platform-aware decisions.

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/platform"

family := platform.Detect()
// Returns: "debian", "darwin", or "" (unknown/unsupported)
```

## OS Families

OSAPI follows
[Ansible's OS family naming](https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_conditionals.html#ansible-facts-distribution):

| Family   | Distributions            | Provider Files |
| -------- | ------------------------ | -------------- |
| `debian` | Ubuntu, Debian, Raspbian | `debian_*.go`  |
| `darwin` | macOS                    | `darwin_*.go`  |

Distributions within the same family share the same provider implementations.
For example, Ubuntu 24.04 and Debian 12 both use the `debian_*.go` providers
because they use the same system interfaces (`/proc`, `/sys`, `ip`).

## How It Works

`Detect()` calls `gopsutil/host.Info()` to get the platform name, then maps it
to an OS family:

- `"ubuntu"`, `"debian"`, `"raspbian"` → `"debian"`
- `"darwin"` (or empty platform with `OS=darwin`) → `"darwin"`
- Anything else → returned as-is (falls through to generic Linux providers)

## Provider Selection

The agent uses the family name to select providers, in
`internal/agent/providers.go`:

```go
switch e.platform {
case "debian":
    return disk.NewDebianProvider(e.log)
case "darwin":
    return disk.NewDarwinProvider(e.log)
default:
    return disk.NewLinuxProvider()
}
```

A [behavior](../../behaviors/behaviors.mdx) names the family with
`behavior.Platform`, and can limit one method to a family with `.On`.

## Adding a New OS Family

1. Add the family to `supportedFamilies` in `internal/cli/validate.go`
2. Add the distribution mapping to `debianFamily` (or create a new map) in
   `pkg/sdk/platform/platform.go`
3. Create provider files: `{family}_*.go` (for example `redhat_get_disk.go`). A
   provider can embed `Linux`, which has the reads that are the same on each
   distribution, and add or override only what differs
4. Add a `case` to each provider switch in `internal/agent/providers.go`
