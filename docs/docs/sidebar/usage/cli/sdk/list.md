# List

List the types for values and the services of the client.

```bash
osapi sdk list
```

## Examples

```bash
osapi sdk list
osapi sdk list --json
```

## Result

The command prints:

```ansi
$ osapi sdk list
[1;38;2;0;213;255mSDK (2):[0m
  [38;2;10;143;176mKIND     [0m[38;2;10;143;176mNAME         [0m[38;2;10;143;176mCOUNT  [0m[38;2;10;143;176mDOC[0m
  type     [1mContentType[0m  2      ContentType is how a deploy treats a file.
  service  [1mHostname[0m     2      Read and set the hostname of each node.
```

With `--json`, the command prints:

```json
{
  "items": [
    {
      "kind": "type",
      "name": "ContentType",
      "count": 2,
      "doc": "ContentType is how a deploy treats a file."
    },
    {
      "kind": "service",
      "name": "Hostname",
      "count": 2,
      "doc": "Read and set the hostname of each node."
    }
  ]
}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
