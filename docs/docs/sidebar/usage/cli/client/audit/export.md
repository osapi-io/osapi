# Export

Show all audit log entries, with no pages.

To save them, send the JSON to a file: osapi client audit export --json &gt;
audit.json

```bash
osapi client audit export
```

## Examples

```bash
osapi client audit export
osapi client audit export --json
```

## Result

The command prints:

```ansi
$ osapi client audit export
[1;38;2;0;213;255mAudit entries (2):[0m
  [38;2;10;143;176mID                                    [0m[38;2;10;143;176mTIMESTAMP            [0m[38;2;10;143;176mUSER             [0m[38;2;10;143;176mMETHOD  [0m[38;2;10;143;176mPATH                                         [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mDURATION  [0m[38;2;10;143;176mTRACE ID[0m
  [38;2;138;138;138m550e8400-e29b-41d4-a716-446655440000[0m  [38;2;138;138;138m2026-01-15 09:58:12[0m  ops@example.com  [1mPUT[0m     /api/node/web-01/sysctl/net.ipv4.ip_forward  200     184ms     [38;2;138;138;138m26e448c9dbe1c4a8...[0m
  [38;2;138;138;138m6ba7b810-9dad-11d1-80b4-00c04fd430c8[0m  [38;2;138;138;138m2026-01-15 09:41:03[0m  ci@example.com   [1mGET[0m     /api/agent                                   200     12ms      [38;2;138;138;138m[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "total_items": 2,
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "timestamp": "2026-01-15T09:58:12Z",
      "user": "ops@example.com",
      "roles": ["admin"],
      "method": "PUT",
      "path": "/api/node/web-01/sysctl/net.ipv4.ip_forward",
      "response_code": 200,
      "duration_ms": 184,
      "source_ip": "10.0.0.5",
      "operation_id": "UpdateSysctl",
      "trace_id": "26e448c9dbe1c4a8d216a98c675079c9",
      "job_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "steps": [
        {
          "hostname": "web-01",
          "steps": [
            {
              "behavior": "managedfile",
              "method": "update",
              "changed": true,
              "duration_ms": 31
            }
          ]
        }
      ],
      "chain": ["web-01: managedfile update (changed)"]
    },
    {
      "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "timestamp": "2026-01-15T09:41:03Z",
      "user": "ci@example.com",
      "roles": ["read"],
      "method": "GET",
      "path": "/api/agent",
      "response_code": 200,
      "duration_ms": 12,
      "source_ip": "10.0.0.9",
      "operation_id": "GetAgents"
    }
  ]
}
```

## Permission

It needs the `audit:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `http://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
