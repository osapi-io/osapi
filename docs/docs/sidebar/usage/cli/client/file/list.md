# List

List stored files.

List all files stored in the Object Store.

```bash
osapi client file list
```

## Examples

```bash
osapi client file list
osapi client file list --json
```

## Result

The command prints:

```ansi
$ osapi client file list
[1;38;2;0;213;255mFiles (2):[0m
  [38;2;10;143;176mNAME           [0m[38;2;10;143;176mSHA256                                                            [0m[38;2;10;143;176mSIZE    [0m[38;2;10;143;176mSOURCE[0m
  [1mapp.conf[0m       a3f1c2e4b5d6978812f3e4d5c6b7a8f9e0d1c2b3a4f5e6d7c8b9a0f1e2d3c4b5  1.0 KB  user
  [1mapp.conf.tmpl[0m  7d8e9fa0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8  512 B   user

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "items": [
    {
      "name": "app.conf",
      "sha256": "a3f1c2e4b5d6978812f3e4d5c6b7a8f9e0d1c2b3a4f5e6d7c8b9a0f1e2d3c4b5",
      "size": 1024,
      "content_type": "raw",
      "source": "user"
    },
    {
      "name": "app.conf.tmpl",
      "sha256": "7d8e9fa0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8",
      "size": 512,
      "content_type": "template",
      "source": "user"
    }
  ],
  "total_items": 2
}
```

## Permission

It needs the `file:read` permission. `osapi token roles get --name <role>` shows
what a role grants.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
