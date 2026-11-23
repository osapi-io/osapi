# Agent wiring

Two files connect a provider to the job pipeline. The registry handles the
dispatch and the facts.

## Two rules are in the corpus, not here

| What you need to know | Where |
| --- | --- |
| Delivery is at-least-once, and what the agent owes because of it | [delivery is at-least-once](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#delivery-is-at-least-once-and-the-agent-owes-idempotency) |
| What `_any`, `_all` and a label selector resolve to | [routing and targeting](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#routing-and-targeting) |
| Which failures terminate a message, and which redeliver | [which failures terminate](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#which-failures-terminate-a-message) |
| What happens when the response cannot be written | [when the response cannot be written](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#when-the-response-cannot-be-written) |
| The three limits, and what each one bounds | [three limits](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#three-limits-bounding-different-things) |
| The four result statuses, and why a row might not say `ok` | [why a row might not say ok](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#why-a-row-might-not-say-ok) |
| How a provider is selected for the host's platform | [platform selection](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#platform-selection-happens-outside-the-provider) |
| How the unsupported outcome becomes a skip | [unsupported is not failure](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#unsupported-is-not-failure-and-not-no-change) |

This file holds the shapes and the file names. The rules above are stated once,
in the corpus, so a change to either is a change in one place.

## 1. The processor

`internal/agent/processor_{domain}.go` turns a job request into a provider call
and its result into JSON.

```go
func process{Domain}Operation(
    provider {domain}.Provider,
    logger *slog.Logger,
    req job.Request,
) (json.RawMessage, error) {
    // switch on req.Operation, call the provider, marshal the result
}
```

There are two shapes. The shape depends on if the domain is its own category:

- **Its own category**, as `schedule` and `docker` are: add a
  `New{Domain}Processor` factory beside the helpers.
- **Inside an existing category**, usually `node`: add a `case` to that
  category's processor and delegate to helpers in the new file. Read
  `processor_node.go` for the current switch.

`pkg/sdk/client/operations.go` defines the operation constants one time, as
`JobOperation` strings with the shape `{category}.{domain}.{operation}`, for
example `"node.sysctl.list"`. `internal/job/types.go` gives each one an alias,
`Operation{Domain}{Operation}`, for use in the repository.

Add the constant to `operations.go`, and the alias next to its siblings. Never
write the string literal at a call site. That is how a handler and an agent
start to disagree about an operation name.

## What a domain does not touch

The sibling NATS projects are not part of this. Nothing under
`internal/provider/`, `internal/controller/` or `pkg/sdk/` imports
`osapi-io/nats-client`. Only the `cmd/` wiring, `internal/agent/consumer.go` and
`internal/cli` do. A new domain uses the subjects that already exist
(`jobs.query.*` and `jobs.modify.*`, routed for each host) and the buckets that
already exist. Thus it adds no stream, no consumer and no KV bucket.

`internal/job/config.go` declares streams and buckets one time. If a domain
really needs its own storage, that is a change to the job layer, with its own
spec. It is not part of a new domain. Meta providers write through the
file-state KV that the file provider already owns.

## 2. The registration

`agent.Setup` in `internal/agent/setup.go` makes each provider, and
`buildRegistry` in the same file registers it.

```go
// A new category, in buildRegistry
r.Register("{domain}", New{Domain}Processor({domain}Provider, e.log))

// An existing category: pass the provider into that category's
// processor factory. Read the current parameter list.
```

Platform selection happens in `internal/agent/providers.go`, not in the
provider. `Setup` reads the family from gohai's `family.Detect` and
whether it is in a container from gohai's virtualization `InContainer`, one time, and gives them to `hostEnv`. Add one method for the domain:

```go
func (e hostEnv) {domain}() {domain}.Provider {
    switch e.platform {
    case "debian":
        if e.container {
            return {domain}.NewDebianDockerProvider(...)
        }
        return {domain}.NewDebianProvider(e.exec, ...)
    case "darwin":
        return {domain}.NewDarwinProvider(...)
    default:
        return {domain}.NewLinuxProvider()
    }
}
```

Add the domain to the table in `providers_public_test.go`, which checks the
provider of each area on each platform.

A provider that uses an SDK has no switch. Make it and check availability. When
it is not available, leave it nil. Then the operation reports unsupported, and
does not panic.

## Delivery semantics worth knowing before you add an operation

The corpus specifies the guarantees of the job system. Read all of them there.
These are important when you add an operation, with the location of each rule:

| What you are relying on | Stated in |
| --- | --- |
| Delivery is at-least-once, and the agent checks for a recorded response before executing | [delivery is at-least-once](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#delivery-is-at-least-once-and-the-agent-owes-idempotency) |
| A job that has run is terminal, a failure is reported rather than retried by redelivery | [which failures terminate](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#which-failures-terminate-a-message) |
| What happens when the response cannot be written after the work ran | [when the response cannot be written](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#when-the-response-cannot-be-written) |
| The consumer's delivery settings, as defaults a deployment may override | [consumer defaults](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#consumer-defaults) |
| A long operation is kept alive while it runs, so it is not redelivered mid-flight | [three limits](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md#three-limits-bounding-different-things) |

Two consequences for a new operation, which are yours rather than the system's:

- **A new delivery does not protect you.** The idempotency of the provider makes
  a repeat safe. That is a requirement of the contract, not of this page.
- **`job retry` makes a new job.** It does not send the old message again. Thus
  no domain code handles it.

## Facts

A behavior's provider reads the agent's facts from the host it gets, as
`h.Facts()`. `buildRegistry` in `internal/agent/setup.go` sets `h.Facts` to the
facts of the agent. The provider does not embed a facts holder, and the registry
does not inject facts. What a provider does with facts is
[facts](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md#facts).

## Tests

- `internal/agent/processor_{domain}_public_test.go`: one suite method for each
  helper, with table rows for each sub-operation, an unknown operation and a
  provider error.
- Mock the provider from `internal/provider/{category}/{domain}/mocks/`.
- The wiring is in `internal/agent`, so the coverage gate holds it.
  `setup_public_test.go` runs `agent.Setup` against a real embedded NATS
  server.
