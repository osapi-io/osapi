---
sidebar_position: 4
---

# UI Development

The embedded management dashboard lives in the `ui/` directory. It is a React
single-page application whose API client is generated from the same OpenAPI
specification the Go SDK is generated from, and whose production build is
compiled into the controller binary.

The rules are not on this page. They are in the specifications repository,
stated once. This page is the index.

:::note

The links below are full GitHub addresses rather than relative links, because
the specifications repository is separate from this one and is not published as
part of this site. Every other cross-reference here is relative.

:::

## The rules

| What you need to know                                                            | Stated in                                                                    |
| -------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| What the UI is, and that it is one application rather than a frontend per domain | [FR-001](https://github.com/osapi-io/specs/blob/main/components/osapi/ui.md) |
| The setting that disables it, and what the controller serves when it is off      | [FR-002](https://github.com/osapi-io/specs/blob/main/components/osapi/ui.md) |
| The stack, stated as what each part is for                                       | [FR-003](https://github.com/osapi-io/specs/blob/main/components/osapi/ui.md) |
| The four kinds of component, and what decides which one you are writing          | [FR-004](https://github.com/osapi-io/specs/blob/main/components/osapi/ui.md) |
| That the API client is generated, and what editing it by hand costs              | [FR-005](https://github.com/osapi-io/specs/blob/main/components/osapi/ui.md) |
| The embedding mechanism, and the build order it forces                           | [FR-006](https://github.com/osapi-io/specs/blob/main/components/osapi/ui.md) |
| How the UI authenticates, and that it decodes the token without verifying it     | [FR-007](https://github.com/osapi-io/specs/blob/main/components/osapi/ui.md) |
| That the permission model is osapi's, cited rather than restated                 | [FR-008](https://github.com/osapi-io/specs/blob/main/components/osapi/ui.md) |
| Where the commands live, and why they are not listed as prose                    | [FR-009](https://github.com/osapi-io/specs/blob/main/components/osapi/ui.md) |
| Why `ui/` is excluded from the coverage gate                                     | [FR-010](https://github.com/osapi-io/specs/blob/main/components/osapi/ui.md) |

## The commands

They are in the justfile, not here. `just --list` shows every recipe; the ones
this directory uses are prefixed `react-`. Build through `just build`,
`just ready` or `just test` rather than a recipe directly — each builds the UI
before the binary, which is the order FR-006 explains.

## Related

- [Contributing](contributing.md) — setup, building, and the conventions code
  follows
- [UI Architecture](../architecture/ui.md) — what an operator configures and
  sees
- [Adding an API Domain](adding-an-api-domain.md) — the layers an endpoint
  touches
