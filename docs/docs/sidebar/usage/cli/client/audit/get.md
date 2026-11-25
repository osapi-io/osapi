# Get

Get a single audit log entry.

Returns a single audit log entry by ID. An entry of a request to a target
outside the target scope of the token is not found.

```bash
osapi client audit get [flags]
```

## Examples

```bash
osapi client audit get --id 550e8400-e29b-41d4-a716-446655440000
osapi client audit get --id 550e8400-e29b-41d4-a716-446655440000 --json
```

## Result

The command prints:

```ansi
$ osapi client audit get --id 550e8400-e29b-41d4-a716-446655440000
[1;38;2;0;213;255mAudit entry:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;10;143;176mID:[0m         550e8400-e29b-41d4-a716-446655440000
  [38;2;10;143;176mTimestamp:[0m  2026-01-15 09:58:12
  [38;2;10;143;176mUser:[0m       ops@example.com
  [38;2;10;143;176mRoles:[0m      admin
  [38;2;10;143;176mMethod:[0m     PUT
  [38;2;10;143;176mPath:[0m       /api/node/web-01/sysctl/net.ipv4.ip_forward
  [38;2;10;143;176mStatus:[0m     200
  [38;2;10;143;176mDuration:[0m   184ms
  [38;2;10;143;176mSource IP:[0m  10.0.0.5
  [38;2;10;143;176mOperation:[0m  UpdateSysctl
  [38;2;10;143;176mTrace ID:[0m   26e448c9dbe1c4a8d216a98c675079c9
  [38;2;10;143;176mJob ID:[0m     7c9e6679-7425-40de-944b-e07fc1f90ae7
  [38;2;10;143;176mChain:[0m      web-01: managedfile update (changed)
```

With `--json`, the API's response:

```json
{
  "entry": {
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
  }
}
```

## Permission

It needs the `audit:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag   | Description                   | Default |
| ------ | ----------------------------- | ------- |
| `--id` | The audit entry ID (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `http://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
