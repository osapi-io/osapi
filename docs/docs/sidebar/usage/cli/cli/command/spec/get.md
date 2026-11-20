# Get

Print the command DSL: each option, method, refinement, field tag and check.

Print each part of the command DSL, what it makes osapi do and when to use it.
The list comes from this binary. Thus it agrees with the build that you run.
Read it before you write a local group, then check the result with
`osapi cli command validate`.

```bash
osapi cli command spec get
```

## Examples

```bash
osapi cli command spec get
osapi cli command spec get --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
