# Retry

Retry a failed or stuck job.

The new job has the same operation as the old job. The original job stays in the
history. Its timeline gets a `retried` event that links to the new job.

```bash
osapi client job retry [flags]
```

## Examples

```bash
osapi client job retry --id 550e8400-e29b-41d4-a716-446655440000
```

## Result

The command prints:

```ansi
$ osapi client job retry --id 550e8400-e29b-41d4-a716-446655440000
[1;38;2;0;213;255mJob:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;10;143;176mJob ID:[0m    9b2e4c6a-8d0f-4a1b-b3c5-d7e9f1a2b4c6
  [38;2;10;143;176mStatus:[0m    [38;5;222msubmitted[0m
  [38;2;10;143;176mRevision:[0m  42
```

With `--json`, the API's response:

```json
{
  "job_id": "9b2e4c6a-8d0f-4a1b-b3c5-d7e9f1a2b4c6",
  "status": "submitted",
  "revision": 42,
  "timestamp": "2026-01-15T10:00:00Z"
}
```

## Permission

It needs the `job:write` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                                             | Default |
| -------------- | --------------------------------------------------------------------------------------- | ------- |
| `--id`         | UUID of the job to retry (required)                                                     |         |
| `-T, --target` | The target of the retried job: _any, _all, a hostname, or a label such as group:web.dev | `_any`  |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `http://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
