# New

Write a new behavior's standard layout in an osapi checkout.

Before you change anything, the result builds, validates and passes every test.
Then change the types, the provider and the examples to what the behavior
manages. Writes nothing when the name is taken. A behavior that belongs to a
group, such as `network`, is nested under it: its routes are
`/api/node/{hostname}/network/<name>`, its commands
`osapi client node network <name>`, and its directory
`internal/behaviors/network/<name>/`.
[Your first behavior](../../../development/your-first-behavior.md) walks through
the rest.

```bash
osapi behavior new [flags]
```

## Examples

```bash
osapi behavior new --name motd --description "The message shown at login"
osapi behavior new --name motd --description "The message shown at login" --root ~/src/osapi
osapi behavior new --group network --name neighbors --description "The host's ARP neighbors"
```

## Flags

| Flag            | Description                                                                | Default |
| --------------- | -------------------------------------------------------------------------- | ------- |
| `--description` | What it manages, in one line (required)                                    |         |
| `--group`       | Nest it under a group, such as network: its routes, commands and directory |         |
| `--name`        | The behavior's name: lowercase letters and digits (required)               |         |
| `--root`        | The osapi checkout to write into                                           | `.`     |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
