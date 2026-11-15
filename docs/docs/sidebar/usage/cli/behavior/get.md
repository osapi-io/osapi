# Get

Show a behavior: its surfaces, chain, methods, types and telemetry.

**Uses** lists the behaviors this one is built on. The agent builds their
providers first, and the methods of this behavior call theirs on the same host,
in the same job. If a behavior that it uses cannot run on a platform, this one
reports `skipped` there too. **Used by** lists the behaviors built on this one.
A behavior with `Internal: yes` has no route, no command and no permission. Use
`--method` to show one method in full: its fields with their constraints, its
output and its examples.

```bash
osapi behavior get [flags]
```

## Examples

```bash
osapi behavior get --name sysctl
osapi behavior get --name sysctl --method create
osapi behavior get --name sysctl --method create --json
```

## Flags

| Flag       | Description                                                      | Default |
| ---------- | ---------------------------------------------------------------- | ------- |
| `--method` | One method to show in full: its fields, constraints and examples |         |
| `--name`   | The behavior, from 'osapi behavior list' (required)              |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
