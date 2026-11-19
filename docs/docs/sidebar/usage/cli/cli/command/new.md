# New

Write a new local group that builds and passes its tests.

It writes `internal/cli/commands/<name>/` in the osapi checkout in the current
directory, and adds its import to `internal/cli/commands/all`. The group has a
list and a get, an empty `Deps` and the tests of its methods. Then run
`go generate ./internal/cli/commands/all`. While its `Deps` is empty, the CLI
needs no wiring. `osapi cli command spec` lists each part of the DSL.

```bash
osapi cli command new [flags]
```

## Examples

```bash
osapi cli command new --name widget --description "The widgets of this host"
```

## Flags

| Flag            | Description                                          | Default |
| --------------- | ---------------------------------------------------- | ------- |
| `--description` | What the group is, in one line (required)            |         |
| `--name`        | The group's name: its command under osapi (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
