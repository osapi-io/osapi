# Container DNS Provider + Container Detection

## Problem

When the OSAPI agent runs inside a Docker container on a Debian-based image,
the DNS provider fails because `resolvectl` is not available. Containers use
`/etc/resolv.conf` directly — there is no systemd-resolved. DNS writes are
managed by the container runtime (Docker, Kubernetes), not the agent.

## Design

Two changes: a new `DebianDocker` DNS provider that reads `/etc/resolv.conf`
instead of calling `resolvectl`, and a `containerized` built-in fact so
providers and consumers can detect container environments.

### 1. Container Detection — `platform.IsContainer()`

**File:** `pkg/sdk/platform/container.go`

Add an `IsContainer() bool` function that checks for `/.dockerenv` file
existence. Use an injectable function variable (`ContainerCheckFn`) following
the existing `HostInfoFn` pattern for testability.

**File:** `pkg/sdk/platform/container_public_test.go`

Test both container and non-container paths by overriding `ContainerCheckFn`.

### 2. DebianDocker DNS Provider

Three new files in `internal/provider/network/dns/`:

**`debian_docker.go`** — Provider struct and constructor.

```go
type DebianDocker struct {
    provider.FactsAware
    logger *slog.Logger
    fs     avfs.VFS
}

func NewDebianDockerProvider(
    logger *slog.Logger,
    fs avfs.VFS,
) *DebianDocker
```

No exec manager — this provider only reads files via avfs.

Compile-time check: `var _ Provider = (*DebianDocker)(nil)`

**`debian_docker_get_resolv_conf_by_interface.go`** — Get implementation.

`GetResolvConfByInterface` reads `/etc/resolv.conf` via avfs and parses
`nameserver` and `search` lines. The `interfaceName` parameter is accepted
but ignored — containers have a single global DNS configuration.

Returns `GetResult` with `DNSServers` and `SearchDomains` populated from
the file contents. Returns `["."]` for search domains if none are found
(matching the Debian provider convention).

**`debian_docker_update_resolv_conf_by_interface.go`** — Update implementation.

`UpdateResolvConfByInterface` returns `provider.ErrUnsupported`. DNS in
containers is managed by the container runtime, not the agent.

### 3. Containerized Built-in Fact

**`internal/facts/keys.go`** — Add `KeyContainerized = "containerized"` constant
with description "Whether the agent is running inside a container".

**`internal/job/types.go`** — Add `Containerized bool` field to
`FactsRegistration`.

**`internal/agent/facts.go`** — Call `platform.IsContainer()` during facts
collection and set `Containerized` on the registration.

**`internal/agent/factref.go`** — Add resolver case for `@fact.containerized`
that returns the boolean value from `FactsRegistration.Containerized`.

### 4. Agent Setup Wiring

**`cmd/agent_setup.go`** — Update the DNS provider switch:

```go
case "debian":
    if platform.IsContainer() {
        dnsProvider = dns.NewDebianDockerProvider(log, appFs)
    } else {
        dnsProvider = dns.NewDebianProvider(log, execManager)
    }
```

All other providers remain unchanged — host, disk, mem, load, ping all work
inside containers already since they read `/proc` directly.

### 5. Test Files

Each new production file gets a matching `*_public_test.go`:

- `pkg/sdk/platform/container_public_test.go`
- `internal/provider/network/dns/debian_docker_public_test.go`
- `internal/provider/network/dns/debian_docker_get_resolv_conf_by_interface_public_test.go`
- `internal/provider/network/dns/debian_docker_update_resolv_conf_by_interface_public_test.go`

All tests use testify/suite with table-driven patterns. DNS tests use
`memfs.New()` for filesystem mocking.

Facts-related changes are covered by updating existing test files:
- `internal/facts/keys_public_test.go` — add `containerized` to key list
- `internal/agent/facts_public_test.go` — verify `Containerized` is set
- `internal/agent/factref_public_test.go` — test `@fact.containerized` resolution

## Out of Scope

- Non-Debian container images (no `redhat_docker` etc. until needed)
- Container detection for podman, LXC, or other runtimes (Docker only for now)
- Changes to other providers (host, disk, mem, load work in containers already)
- SDK or CLI changes (the fact surfaces automatically through existing paths)
