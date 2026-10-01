---
sidebar_position: 5
---

# Adding an API Domain

A domain is a coherent area of system behaviour exposed as API endpoints. Adding
one touches every layer: a provider that does the work on the agent, an OpenAPI
specification, a handler, registration, an SDK service, CLI commands, and
documentation.

A domain is complete when it appears everywhere an existing domain appears. The
check is to pick a finished domain — `sysctl` or `cron` — and search the
codebase for it. Anything that exists for `sysctl` and not for yours is missing.

The rules are not on this page. They are in the specifications repository,
stated once, and the `add-a-domain` skill cites them. This page is the index.

:::note

The links below are full GitHub addresses rather than relative links, because
the specifications repository is separate from this one and is not published as
part of this site. Every other cross-reference here is relative.

:::

## The rules

Everything a domain has to satisfy is in the design docs, one document per
subject.

| Subject                                                                                  | Document                                                                                     |
| ---------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| What a domain consists of, which layers it touches, and how to tell one is half-finished | [Building a domain](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md) |
| Validation tags: the three places one goes and the one place it does nothing             | [Building a domain](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md) |
| Verb mapping, and why one endpoint never both creates and updates                        | [Building a domain](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md) |
| What a provider returns, and the three idempotency outcomes                              | [Providers](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md)       |
| Why input is revalidated in the provider rather than trusted from the handler            | [Providers](https://github.com/osapi-io/specs/blob/main/components/osapi/providers.md)       |
| Broadcast targeting, and the collection shape single and broadcast both return           | [The job system](https://github.com/osapi-io/specs/blob/main/components/osapi/job-system.md) |
| What a new SDK service owes: four files, a client field, an example, a page              | [The Go SDK](https://github.com/osapi-io/specs/blob/main/components/osapi/sdk.md)            |
| Where a new permission goes, and why missing one makes the endpoint unreachable          | [Permissions](https://github.com/osapi-io/specs/blob/main/components/osapi/permissions.md)   |
| Running a command: the ceiling, the stdin path for secrets, the dash-leading rule        | [Running commands](https://github.com/osapi-io/specs/blob/main/components/osapi/exec.md)     |

The nine steps below are the procedure. The documents above are the rules the
procedure applies, and the short form of the ones you are most likely to break
is in
[CONTRIBUTING.md](https://github.com/osapi-io/osapi/blob/main/CONTRIBUTING.md).

## The sequence

The `add-a-domain` skill carries the build order, the file layouts and the
scaffolding to copy, and it cites the same requirements. Invoke it by name in
Claude Code:

```
add-a-domain
```

It lives in the specifications repository under `.claude/skills/add-a-domain/`,
with one reference per layer.

## Related

- [Contributing](contributing.md) — setup, building, and the conventions code
  follows
- [Testing](testing.md) — the unit and integration suites and the coverage gate
- [System Architecture](../architecture/system-architecture.md) — health checks,
  authentication, and external dependencies
