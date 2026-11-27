# Get

Show the rotation of the controller's key.

Shows the controller's key, and during a rotation, each agent that the switch
waits for. The controller switches when each one has the next key, or when
controller.pki.rotation_grace_period ends.

```bash
osapi client pki key get
```

## Examples

```bash
osapi client pki key get
```

## Result

The command prints:

```ansi
$ osapi client pki key get
[1;38;2;0;213;255mController key:[0m
  [38;2;10;143;176mCurrent:[0m     SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a
  [38;2;10;143;176mNext:[0m        SHA256:9c2e41b07d3a8f6e5c4b3a2918f7e6d5c4b3a29180f7e6d5c4b3a2918f7e6d5c
  [38;2;10;143;176mStarted at:[0m  2026-11-26 12:00:00
  [38;2;10;143;176mSwitch by:[0m   2026-11-27 12:00:00

[1;38;2;255;0;255mAgents (1):[0m
  [38;2;176;48;176mHOSTNAME  [0m[38;2;176;48;176mMACHINE ID                        [0m[38;2;176;48;176mHAS NEXT[0m
  [1mweb-01[0m    [38;2;138;138;138md530664edc1c431eb5ef1e9dfd6617a5[0m  yes

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "current": "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
  "next": "SHA256:9c2e41b07d3a8f6e5c4b3a2918f7e6d5c4b3a29180f7e6d5c4b3a2918f7e6d5c",
  "started_at": "2026-11-26T12:00:00Z",
  "switch_by": "2026-11-27T12:00:00Z",
  "agents": [
    {
      "hostname": "web-01",
      "machine_id": "d530664edc1c431eb5ef1e9dfd6617a5",
      "has_next": true
    }
  ]
}
```

## Permission

It needs the `pki:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
