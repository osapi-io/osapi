# Upload

Upload a file to the Object Store.

An upload of the same content changes nothing (`Changed: false`). If the file
exists with different content, the command rejects the upload with 409 Conflict.
`file update` replaces the file.

```bash
osapi client file upload [flags]
```

## Examples

```bash
osapi client file upload --name app.conf --file /tmp/app.conf
osapi client file upload --name app.conf.tmpl --file /tmp/app.conf.tmpl --content-type template
```

## Result

The command prints:

```ansi
$ osapi client file upload --name app.conf --file /tmp/app.conf
[1;38;2;0;213;255mFile:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;10;143;176mName:[0m          app.conf
  [38;2;10;143;176mSHA256:[0m        a3f1c2e4b5d6978812f3e4d5c6b7a8f9e0d1c2b3a4f5e6d7c8b9a0f1e2d3c4b5
  [38;2;10;143;176mSize:[0m          1.0 KB
  [38;2;10;143;176mChanged:[0m       yes
  [38;2;10;143;176mContent type:[0m  raw
```

With `--json`, the API's response:

```json
{
  "name": "app.conf",
  "sha256": "a3f1c2e4b5d6978812f3e4d5c6b7a8f9e0d1c2b3a4f5e6d7c8b9a0f1e2d3c4b5",
  "size": 1024,
  "changed": true,
  "content_type": "raw"
}
```

The command prints:

```ansi
$ osapi client file upload --name app.conf.tmpl --file /tmp/app.conf.tmpl --content-type template
[1;38;2;0;213;255mFile:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;10;143;176mName:[0m          app.conf.tmpl
  [38;2;10;143;176mSHA256:[0m        7d8e9fa0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8
  [38;2;10;143;176mSize:[0m          512 B
  [38;2;10;143;176mChanged:[0m       yes
  [38;2;10;143;176mContent type:[0m  template
```

With `--json`, the API's response:

```json
{
  "name": "app.conf.tmpl",
  "sha256": "7d8e9fa0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8",
  "size": 512,
  "changed": true,
  "content_type": "template"
}
```

## Permission

It needs the `file:write` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag             | Description                                                                                                                                  | Default |
| ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `--content-type` | How the file should be treated during deploy. raw writes bytes as-is; template renders with Go text/template and agent facts (raw, template) | `raw`   |
| `--file`         | Path to the local file with the file content (required)                                                                                      |         |
| `--name`         | The name of the file in the Object Store (required)                                                                                          |         |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
