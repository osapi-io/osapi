# Version

Display the version of osapi.

The build of this binary: the version that each component reports, the commit
and the tree it was built from, and the Go that built it.

```bash
osapi version
```

## Examples

```bash
osapi version
osapi version --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](output.md) tells how the result is printed, and what `--json` gives.
