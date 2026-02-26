---
title: Client SDK for programmatic automation
status: done
created: 2026-02-22
updated: 2026-02-25
---

## Outcome

Created `osapi-sdk` as a separate sibling repo with generated HTTP client,
service wrappers (`Command`, `Network`, `Health`, `Job`, `Audit`, `System`,
`Metrics`), and connection management. Migrated all CLI commands from
`internal/client/` to the SDK and deleted the internal package entirely.
Declarative engine/playbook support tracked in
[backlog](../backlog/2026-02-25-feat-declarative-playbook-engine.md).

## Objective

Extract a public Go SDK from `internal/client/` so that developers can
programmatically automate system management without going through the CLI. The
SDK would be the foundation layer that both the CLI and declarative automation
tools build on top of.

## Motivation

Today OSAPI is CLI-driven. A real automation developer would want to:

- Write Go code that imports an OSAPI SDK package
- Define system configurations in YAML/TOML and apply them via the SDK
- Build custom automation pipelines (CI/CD, GitOps, fleet management)
- Compose operations programmatically (conditionals, loops, error handling)

This is the Ansible/Terraform model: SDK at the bottom, CLI and declarative
configs on top.

## Design Considerations

- **`pkg/sdk/`** — public, importable, semver'd Go module
- **Composable primitives** — `sdk.System.Hostname()`, `sdk.Command.Exec()`,
  `sdk.Network.DNS.Get()`, etc.
- **Structured config** — accept Go structs or parse YAML playbooks
- **Connection management** — handle auth, retries, timeouts
- **CLI refactor** — rebuild CLI commands on top of the SDK
- **Declarative engine** — parse YAML task files, execute steps via SDK (similar
  to Ansible playbooks)

## Example Usage

```go
client, _ := sdk.New(sdk.Config{
    URL:         "http://localhost:8080",
    BearerToken: os.Getenv("OSAPI_TOKEN"),
})

// Programmatic usage
result, _ := client.Command.Exec(ctx, sdk.ExecParams{
    Command: "uptime",
    Target:  "_all",
})

// Declarative usage
tasks, _ := sdk.LoadPlaybook("configure.yaml")
results := client.Apply(ctx, tasks)
```

## Notes

- Breaking this out as a separate module (`go.sum`-tracked) enables independent
  versioning
- Current `internal/client/` can evolve into the SDK with public API surface
- Consider whether the SDK should wrap the REST API or also support direct NATS
  communication for lower latency
- Playbook format could be inspired by Ansible but simpler
