# Provider

The operations layer. Providers run on the agent, not the controller. A provider
receives parameters from the job payload, does the work, and returns a result.

```
CLI -> SDK -> REST API -> job client -> NATS -> agent -> provider
```

## The contract is in the corpus, not here

[Providers](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md) states each rule that a
provider obeys. Read it before you write one. This file has only what that page
does not have: where the files go, their names, and the scaffold to start from.

That split is deliberate. A second statement of a rule here becomes different
from the rule in the corpus. Then the copy that an agent loads first applies.

| What you need to know | Where |
| --- | --- |
| A provider is the operations layer, and what it returns | [what an operation returns](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#what-an-operation-returns) |
| The idempotency contract, as a table of operation outcomes | [the idempotency rule](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#the-idempotency-rule) |
| Somebody edited the host by hand, and the next run overwrites it | [osapi owns the state](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#osapi-owns-the-state-so-drift-gets-overwritten) |
| `ErrUnsupported` is a fourth outcome, not a failure | [unsupported is not failure](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#unsupported-is-not-failure-and-not-no-change) |
| The four implementation patterns, and how to choose | [the four patterns](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#the-four-implementation-patterns) |
| Platform variants, and how the agent selects one | [platform selection](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#platform-selection-happens-outside-the-provider) |
| How a provider obtains host facts | [facts](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#facts) |
| Why the provider validates what the API already validated | [validation does not discharge](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#validation-on-the-request-path-does-not-discharge-the-providers) |
| Secrets reach a command without appearing in it | [secrets are not arguments](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#a-secret-never-appears-in-a-commands-arguments) |
| A caller's value never becomes an option | [never parsed as an option](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#a-callers-value-is-never-parsed-as-an-option) |
| Filesystem access, the exec manager, and why a file is not written in place | [what it goes through](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#what-a-provider-goes-through-not-around) |
| How an error reads, and the two habits to avoid | [how an error reads](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#how-an-error-reads) |
| What a provider does not touch | [what it does not touch](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#what-a-provider-does-not-touch) |
| The testing obligations that belong to the provider | [what its tests owe](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#what-its-tests-owe) |

Read the provider package of the reference domain together with it. The
specification says what must be true. An existing domain shows it in use.

## Files

```
internal/provider/{category}/{domain}/
  types.go              Provider interface, domain types, compile-time checks
  debian.go             Debian family: Ubuntu, Debian, Raspbian
  debian_{operation}.go one file per large operation
  debian_docker.go      container-aware variant, when behaviour differs
  darwin.go             macOS stub
  linux.go              generic Linux stub
  mocks/generate.go     //go:generate mockgen directive
```

`types.go` holds only types. A function belongs in a file named for what it
does. Categorized domains live under `internal/provider/{category}/{domain}/`,
uncategorized ones directly under `internal/provider/{domain}/`.

## Naming

| Struct         | Constructor                  | Files                          |
| -------------- | ---------------------------- | ------------------------------ |
| `Debian`       | `NewDebianProvider(...)`     | `debian.go`, `debian_{op}.go`  |
| `DebianDocker` | `NewDebianDockerProvider(...)` | `debian_docker.go`           |
| `Darwin`       | `NewDarwinProvider(...)`     | `darwin.go`                    |
| `Linux`        | `NewLinuxProvider()`         | `linux.go`                     |
| `Client`       | `New()`, `NewWithClient(c)`  | `{domain}.go`                  |

`DebianDocker` embeds `Debian`, uses its reads and replaces its writes. Or it is
independent, as `network/dns` is: it reads `/etc/resolv.conf` directly.

## Scaffolding

The shapes to start from. What they have to satisfy is
[what an operation
returns](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#what-an-operation-returns)
and
[facts](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#facts).

```go
// types.go, package {domain}
type Provider interface {
    List(ctx context.Context) ([]Entry, error)
    Get(ctx context.Context, name string) (*Entry, error)
    Create(ctx context.Context, entry Entry) (*CreateResult, error)
    Update(ctx context.Context, entry Entry) (*UpdateResult, error)
    Delete(ctx context.Context, name string) (*DeleteResult, error)
}

var _ Provider = (*Debian)(nil)

// A template reads the agent's facts from the host, as h.Facts().
type Debian struct {
    logger *slog.Logger
    fs     avfs.VFS
}
```

A stub for a platform that does not support the domain:

```go
// darwin.go
func (d *Darwin) List(
    _ context.Context,
) ([]Entry, error) {
    return nil, fmt.Errorf("{domain}: %w", provider.ErrUnsupported)
}
```

A meta provider depends on the narrow interface rather than the file provider's
implementation:

```go
type Deployer interface {
    Deploy(ctx context.Context, req DeployRequest) (*DeployResult, error)
    Undeploy(ctx context.Context, req UndeployRequest) (*UndeployResult, error)
}
```

## Mocks

```
internal/provider/{category}/{domain}/mocks/generate.go
```

Use one `//go:generate go tool go.uber.org/mock/mockgen` directive for each
interface that the domain defines, and commit the output. Never write a double
by hand for an interface that this organization defines. `just generate` makes
them again. A double written by hand stops matching its interface, and nothing
reports it.
