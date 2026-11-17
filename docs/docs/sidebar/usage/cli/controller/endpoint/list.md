# List

List the endpoint groups this binary has.

```bash
osapi controller endpoint list
```

## Examples

```bash
osapi controller endpoint list
osapi controller endpoint list --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |
| `-p, --port`       | Port the server will bind to | `8080`                  |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
