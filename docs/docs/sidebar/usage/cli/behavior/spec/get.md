# Get

Print the behavior DSL: each option, method, example part and field tag.

Print each part of the behavior DSL, what it makes osapi do and when to use it.
The list comes from this binary. Thus it agrees with the build that you run.
Read it before you write a behavior, then check the result with
`osapi behavior validate`.

```bash
osapi behavior spec get
```

## Examples

```bash
osapi behavior spec get
osapi behavior spec get --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
