# Delete

Delete a file from the Object Store.

Delete a file from the Object Store.

```bash
osapi client file delete [flags]
```

## Examples

```bash
osapi client file delete --name app.conf
osapi client file delete --name app.conf --json
```

## Result

The command prints:

```ansi
$ osapi client file delete --name app.conf
[ok] app.conf deleted

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the command prints:

```json
{ "name": "app.conf", "result": "deleted" }
```

## Permission

It needs the `file:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag     | Description                                         | Default |
| -------- | --------------------------------------------------- | ------- |
| `--name` | The name of the file in the Object Store (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
