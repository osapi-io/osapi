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

| What you need to know                                                                    | Stated in                                                                                                          |
| ---------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| What a domain consists of, and how to tell one is incomplete                             | [FR-001](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md)         |
| What to cite rather than restate — the provider contract, the job system, agent identity | [FR-002, FR-003](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md) |
| Which build orderings are forced by tooling, and which are convention                    | [FR-004, FR-005](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md) |
| Where a domain's code goes, and what decides it                                          | [FR-006](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md)         |
| The layers, the entry points, and the request path an operation takes                    | [FR-007, FR-008](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md) |
| Agent wiring, and what must not change in order to add a domain                          | [FR-009, FR-010](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md) |
| Validation: the three places a tag goes, and the one place it does nothing               | [FR-011, FR-012](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md) |
| Verb mapping, and why a combined upsert endpoint is forbidden                            | [FR-013](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md)         |
| The API design guidelines, and what `{hostname}` accepts                                 | [FR-014, FR-015](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md) |
| Broadcast support, and the collection shape both paths return                            | [FR-016, FR-017](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md) |
| Handler registration and startup wiring                                                  | [FR-018](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md)         |
| The SDK service's obligations                                                            | [FR-019, FR-020](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md) |
| The CLI's obligations                                                                    | [FR-021](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md)         |
| The eight design principles, and what each constrains                                    | [FR-022, FR-023](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md) |
| What verifies a finished domain                                                          | [FR-024](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/005-building-a-domain/spec.md)         |

Provider rules are the
[provider contract's](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/001-provider-contract/spec.md)
and job delivery is the
[job system's](https://github.com/osapi-io/specs/blob/main/components/osapi/specs/004-job-system/spec.md).
Neither is restated above.

Read FR-024 before running anything. A domain's last step writes eight
documentation files, and those are checked by `just docusaurus-fmt-check` and
`just docusaurus-build` — which run in `just test`, not in `just ready`.

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
