# Delete

Delete a job.

The job goes from storage permanently.

```bash
osapi client job delete [flags]
```

## Examples

```bash
osapi client job delete --id 550e8400-e29b-41d4-a716-446655440000
```

## Result

The command prints:

```ansi
$ osapi client job delete --id 550e8400-e29b-41d4-a716-446655440000
[ok] 550e8400-e29b-41d4-a716-446655440000 deleted

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the command prints:

```json
{ "id": "550e8400-e29b-41d4-a716-446655440000", "result": "deleted" }
```

## Permission

It needs the `job:write` permission. `osapi token roles get --name <role>` shows
what a role grants.

## Flags

| Flag   | Description                          | Default |
| ------ | ------------------------------------ | ------- |
| `--id` | UUID of the job to delete (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
