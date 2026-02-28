---
title: Declarative playbook engine (osapi-apply)
status: backlog
created: 2026-02-25
updated: 2026-02-25
---

## Objective

Build a declarative automation layer on top of the `osapi-sdk` that parses YAML
task files and executes steps via the SDK — similar to Ansible playbooks but
simpler. This is the planned `osapi-apply` orchestration tool.

## Motivation

The SDK provides composable primitives (`client.Command.Exec()`,
`client.Network.DNS.Get()`, etc.) for programmatic Go usage. A declarative
engine would let operators define desired system state in YAML and apply it
without writing Go code:

```yaml
tasks:
  - name: Set DNS servers
    network.dns.update:
      interface: eth0
      servers: [1.1.1.1, 8.8.8.8]
      target: _all

  - name: Verify connectivity
    command.exec:
      command: ping
      args: [-c, '1', '1.1.1.1']
      target: _all
```

The `changed` field added to mutation responses is the foundation for reporting
convergence status (e.g., "3 of 5 operations changed, 2 already converged").

## Design Considerations

- **Playbook format** — YAML task files with step names, module references, and
  parameters
- **Execution model** — sequential steps with optional conditionals and error
  handling
- **Change reporting** — aggregate `changed` status across steps to report
  convergence
- **Targeting** — inherit SDK target routing (`_any`, `_all`, hostname, label
  selectors)
- **Dry-run mode** — preview what would change without applying

## Notes

- Spun out from the completed Client SDK task which delivered the programmatic
  SDK layer
- The `changed` field (done) is a prerequisite for meaningful convergence
  reporting
