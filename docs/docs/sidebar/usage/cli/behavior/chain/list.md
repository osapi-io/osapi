# List

List each behavior in a chain: what it calls, and what calls it.

\*\*Chain\*\* is the tree of the behaviors that a behavior calls, down to the
last one. A building block, with `Internal`, is in the list for the behaviors
that call it.

```bash
osapi behavior chain list
```

## Examples

```bash
osapi behavior chain list
osapi behavior chain list --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
