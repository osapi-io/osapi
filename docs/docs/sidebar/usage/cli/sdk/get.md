# Get

Show a type with its values, or a service with its methods.

A name of a type, such as ContentType, shows its doc and its values. A name of a
service, such as Hostname, shows each method with the permission that it needs,
its route, its input and its result.

```bash
osapi sdk get [flags]
```

## Examples

```bash
osapi sdk get --name Hostname
osapi sdk get --name ContentType --json
```

## Result

The command prints:

```ansi
$ osapi sdk get --name Hostname
[1;38;2;0;213;255mSDK:[0m

[1;38;2;255;0;255mService:[0m
  [38;2;176;48;176mName:[0m  Hostname
  [38;2;176;48;176mDoc:[0m   Read and set the hostname of each node.
  [38;2;176;48;176mArea:[0m  system
  [38;2;176;48;176mNode:[0m  yes

[1;38;2;255;209;0mMethods (1):[0m
  [38;2;176;143;0mNAME    [0m[38;2;176;143;0mDOC                           [0m[38;2;176;143;0mPERMISSION      [0m[38;2;176;143;0mHTTP                               [0m[38;2;176;143;0mINPUT                [0m[38;2;176;143;0mRESULT                      [0m[38;2;176;143;0mEXAMPLES[0m
  [1mUpdate[0m  Set the hostname of the host  [38;2;138;138;138mhostname:write[0m  PUT /api/node/{hostname}/hostname  HostnameUpdateInput  Collection[HostnameResult]  2
```

With `--json`, the command prints:

```json
{
  "service": {
    "name": "Hostname",
    "doc": "Read and set the hostname of each node.",
    "area": "system",
    "node": true,
    "methods": [
      {
        "name": "Update",
        "doc": "Set the hostname of the host",
        "permission": "hostname:write",
        "http": "PUT /api/node/{hostname}/hostname",
        "input": "HostnameUpdateInput",
        "result": "Collection[HostnameResult]",
        "examples": 2
      }
    ]
  }
}
```

## Flags

| Flag     | Description                                                            | Default |
| -------- | ---------------------------------------------------------------------- | ------- |
| `--name` | A type, such as ContentType, or a service, such as Hostname (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
