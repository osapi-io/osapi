# Get

Get job details and status.

Retrieve details of a specific job by its ID. A job whose target is outside the
target scope of the token is not found.

```bash
osapi client job get [flags]
```

## Examples

```bash
osapi client job get --id 550e8400-e29b-41d4-a716-446655440000
```

## Result

The command prints:

```ansi
$ osapi client job get --id 550e8400-e29b-41d4-a716-446655440000
[1;38;2;0;213;255mJob:[0m
  [38;2;176;48;176mJob ID:[0m    550e8400-e29b-41d4-a716-446655440000
  [38;2;176;48;176mStatus:[0m    [38;2;103;234;147mcompleted[0m
  [38;2;176;48;176mCreated:[0m   2026-01-15 09:58:12
  [38;2;176;48;176mTarget:[0m    web-01
  [38;2;176;48;176mHostname:[0m  web-01
  [38;2;176;48;176mUpdated:[0m   2026-01-15 09:58:13
  [38;2;176;48;176mChanged:[0m   yes
  [38;2;176;48;176mChain:[0m     web-01: managedfile update (changed)

[1;38;2;255;0;255mOperation:[0m
  [38;2;176;143;0mType:[0m  node.sysctl.update

[1;38;2;255;209;0mOperation data:[0m
  [38;2;176;143;0mkey:[0m    net.ipv4.ip_forward
  [38;2;176;143;0mvalue:[0m  1

[1;38;2;69;174;23mResponses (1):[0m
  [38;2;58;143;20mHOSTNAME  [0m[38;2;58;143;20mSTATUS     [0m[38;2;58;143;20mCHANGED  [0m[38;2;58;143;20mDATA                                  [0m[38;2;58;143;20mSTEPS[0m
  [1mweb-01[0m    [38;2;103;234;147mcompleted[0m  yes      changed:yes, key:net.ipv4.ip_forward  managedfile

[1;38;2;107;92;255mAgent states (1):[0m
  [38;2;89;76;216mHOSTNAME  [0m[38;2;89;76;216mSTATUS     [0m[38;2;89;76;216mDURATION[0m
  [1mweb-01[0m    [38;2;103;234;147mcompleted[0m  184ms

[1;38;2;0;213;255mTimeline (3):[0m
  [38;2;10;143;176mTIMESTAMP            [0m[38;2;10;143;176mEVENT         [0m[38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mMESSAGE[0m
  [38;2;138;138;138m2026-01-15 09:58:12[0m  submitted     [1m[0m          job created
  [38;2;138;138;138m2026-01-15 09:58:12[0m  acknowledged  [1mweb-01[0m
  [38;2;138;138;138m2026-01-15 09:58:13[0m  completed     [1mweb-01[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "created": "2026-01-15T09:58:12Z",
  "target": "web-01",
  "operation": {
    "type": "node.sysctl.update",
    "data": {
      "key": "net.ipv4.ip_forward",
      "value": "1"
    }
  },
  "hostname": "web-01",
  "updated_at": "2026-01-15T09:58:13Z",
  "changed": true,
  "responses": [
    {
      "hostname": "web-01",
      "status": "completed",
      "changed": true,
      "data": {
        "changed": true,
        "key": "net.ipv4.ip_forward"
      },
      "steps": [
        {
          "behavior": "managedfile",
          "method": "update",
          "changed": true,
          "duration_ms": 1
        }
      ]
    }
  ],
  "chain": ["web-01: managedfile update (changed)"],
  "agent_states": [
    {
      "hostname": "web-01",
      "status": "completed",
      "duration": "184ms"
    }
  ],
  "timeline": [
    {
      "timestamp": "2026-01-15T09:58:12Z",
      "event": "submitted",
      "message": "job created"
    },
    {
      "timestamp": "2026-01-15T09:58:12Z",
      "event": "acknowledged",
      "hostname": "web-01"
    },
    {
      "timestamp": "2026-01-15T09:58:13Z",
      "event": "completed",
      "hostname": "web-01"
    }
  ]
}
```

## Permission

It needs the `job:read` permission. `osapi token roles get --name <role>` shows
what a role grants.

## Flags

| Flag   | Description                            | Default |
| ------ | -------------------------------------- | ------- |
| `--id` | UUID of the job to retrieve (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
