---
sidebar_position: 4
---

# UI Development

The embedded management dashboard is in the `ui/` directory. It is a React
single-page application. Its API client comes from the same OpenAPI
specification as the Go SDK. The controller binary has its production build.

The rules are not on this page. They are in the specifications repository,
stated once. This page is the index.

:::note

The links below are full GitHub addresses, not relative links. The
specifications repository is a different repository, and this site does not
publish it. All other links on this page are relative.

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

The recipes are in the justfile, not on this page. `just --list` shows each
recipe. The recipes of this directory start with `react-`, and the list shows
them under `[react]`. Build with `just build`, `just ready` or `just test`, not
with a `react-` recipe directly. Each of them builds the UI before the binary.
The binary embeds the UI, so the UI must exist first.

## Related

- [Contributing](contributing.md): the setup, the build and the code conventions
- [UI Architecture](../architecture/ui.md): what an operator configures and sees
- [Your first behavior](your-first-behavior.md): build a behavior, whose API the
  UI calls
