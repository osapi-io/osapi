---
sidebar_position: 4
---

# Platform detection

The agent selects the provider of each operation from the OS family of its host.
osapi does not detect the family itself. It uses
[gohai](https://github.com/osapi-io/gohai), the library that osapi reads hosts
through: `family.Detect()` in `github.com/osapi-io/gohai/pkg/gohai/family`.

```go
import "github.com/osapi-io/gohai/pkg/gohai/family"

f := family.Detect()
// "darwin" on macOS; "debian", "rhel", "fedora", "amazon", ... on Linux
```

## OS families

gohai names the families as
[Chef's Ohai](https://github.com/chef/ohai/blob/main/lib/ohai/plugins/linux/platform.rb)
does. The platform is the `ID` of `/etc/os-release`, with a few IDs remapped
(`rhel` → `redhat`, `amzn` → `amazon`). The family follows from the platform.

| Family   | Platforms                                                        |
| -------- | ---------------------------------------------------------------- |
| `debian` | ubuntu, debian, raspbian, linuxmint, cumulus, kali, pop          |
| `rhel`   | redhat, centos, rocky, almalinux, oracle, scientific, and others |
| `fedora` | fedora, and other RPM distributions that do not track RHEL       |
| `amazon` | Amazon Linux                                                     |
| `darwin` | macOS                                                            |

gohai also names `suse`, `arch`, `gentoo`, `alpine` and others. A Linux host
whose platform has no family gets its platform name, such as `nixos`. On macOS,
the family is `darwin`, not Ohai's `mac_os_x`.

## Provider selection

The agent selects providers with the family, in `internal/agent/providers.go`:

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

A family with no case gets the Linux provider. A
[behavior](../../reference/behaviors/behaviors.mdx) names a family with
`behavior.Platform`, and can limit one method to a family with `.On`.
`behavior.PlatformDebian`, `behavior.PlatformSUSE` and `behavior.PlatformDarwin`
are gohai's values.

## Adding a new OS family

1. If gohai does not name the family yet, add it to gohai first.
2. Create provider files: `{family}_*.go`, for example `rhel_get_disk.go`. A
   provider can embed `Linux`, which has the reads that are the same on each
   distribution, and add or override only what differs.
3. Add a `case` to each provider switch in `internal/agent/providers.go`, or
   `behavior.Platform(family.RHEL, ...)` to a behavior.
