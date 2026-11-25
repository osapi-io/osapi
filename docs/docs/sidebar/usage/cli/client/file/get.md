# Get

Get file metadata.

Get metadata for a specific file in the Object Store.

```bash
osapi client file get [flags]
```

## Examples

```bash
osapi client file get --name app.conf
osapi client file get --name app.conf --json
```

## Result

The command prints:

```ansi
$ osapi client file get --name app.conf
[1;38;2;0;213;255mFile:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;10;143;176mName:[0m          app.conf
  [38;2;10;143;176mSHA256:[0m        a3f1c2e4b5d6978812f3e4d5c6b7a8f9e0d1c2b3a4f5e6d7c8b9a0f1e2d3c4b5
  [38;2;10;143;176mSize:[0m          1.0 KB
  [38;2;10;143;176mContent type:[0m  raw
```

With `--json`, the API's response:

```json
{
  "name": "app.conf",
  "sha256": "a3f1c2e4b5d6978812f3e4d5c6b7a8f9e0d1c2b3a4f5e6d7c8b9a0f1e2d3c4b5",
  "size": 1024,
  "content_type": "raw"
}
```

## Permission

It needs the `file:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag     | Description                                         | Default |
| -------- | --------------------------------------------------- | ------- |
| `--name` | The name of the file in the Object Store (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `http://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
