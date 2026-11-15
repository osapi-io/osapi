# List

List the behaviors this binary has.

An internal behavior, one that exists only to be built on by others, is marked
`(internal)`.

```bash
osapi behavior list
```

## Examples

```bash
osapi behavior list
osapi behavior list --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
