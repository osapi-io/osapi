# Spec

Print the endpoint DSL: each option, method, refinement, error and field tag.

Print each part of the endpoint DSL, what it makes osapi do and when to use it,
and what it cannot do yet. The list comes from this binary. Thus it agrees with
the build that you run. Read it before you write a group, then check the result
with `osapi controller endpoint validate`.

```bash
osapi controller endpoint spec
```

## Examples

```bash
osapi controller endpoint spec
osapi controller endpoint spec --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
