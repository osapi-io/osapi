# List

List stale file deployments.

The list shows the deployments whose object changed after the last deploy. Those
are the files that you must deploy again.

```bash
osapi client file stale list
```

## Examples

```bash
osapi client file stale list
osapi client file stale list --json
```

## Result

The command prints:

```ansi
$ osapi client file stale list
[1;38;2;0;213;255mStale deployments (1):[0m
  [38;2;10;143;176mOBJECT    [0m[38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mPROVIDER  [0m[38;2;10;143;176mDEPLOYED             [0m[38;2;10;143;176mDEPLOYED SHA     [0m[38;2;10;143;176mCURRENT SHA[0m
  app.conf  [1mweb-01[0m    file      2026-01-14 16:20:00  0b1c2d3e4f5a...  a3f1c2e4b5d6...

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "items": [
    {
      "object_name": "app.conf",
      "hostname": "web-01",
      "provider": "file",
      "deployed_at": "2026-01-14T16:20:00Z",
      "path": "/etc/app/app.conf",
      "deployed_sha": "0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c",
      "current_sha": "a3f1c2e4b5d6978812f3e4d5c6b7a8f9e0d1c2b3a4f5e6d7c8b9a0f1e2d3c4b5"
    }
  ],
  "total_items": 1
}
```

## Permission

It needs the `file:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
