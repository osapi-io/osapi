# List

List the dead letters.

Lists the jobs whose delivery to an agent used up its retries, the newest first.
Send one back to the queue with osapi client job retry --id.

```bash
osapi client job deadletter list [flags]
```

## Examples

```bash
osapi client job deadletter list
osapi client job deadletter list --limit 5
```

## Result

The command prints:

```ansi
$ osapi client job deadletter list
[1;38;2;0;213;255mDead letters (1 of 654):[0m
  [38;2;10;143;176mSEQ  [0m[38;2;10;143;176mTIME                 [0m[38;2;10;143;176mJOB ID                                [0m[38;2;10;143;176mTARGET  [0m[38;2;10;143;176mERROR                 [0m[38;2;10;143;176mDELIVERIES[0m
  654  2026-11-26 10:00:00  [38;2;138;138;138m550e8400-e29b-41d4-a716-446655440000[0m  _all    agent did not answer  5

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "total_items": 654,
  "items": [
    {
      "seq": 654,
      "time": "2026-11-26T10:00:00Z",
      "job_id": "550e8400-e29b-41d4-a716-446655440000",
      "operation": {
        "type": "node.hostname.get"
      },
      "target": "_all",
      "error": "agent did not answer",
      "deliveries": 5,
      "consumer": "jobs-agent"
    }
  ]
}
```

## Permission

It needs the `job:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag      | Description                                             | Default |
| --------- | ------------------------------------------------------- | ------- |
| `--limit` | The most dead letters to list, the newest first (1-100) | `20`    |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
