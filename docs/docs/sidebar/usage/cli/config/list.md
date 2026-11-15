# List

List every setting, its value in effect and where it came from.

Each value shows where it came from: `env`, `file`, `default` or `unset`. The
file and the environment are read the way `osapi start` reads them. Without a
config file, the values are the defaults and the environment. Passwords, signing
keys and tokens print as `********`.

```bash
osapi config list
```

## Examples

```bash
osapi config list
osapi config list -f ./osapi.yaml --json
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
