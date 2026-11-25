# Get

Get a machine's admission record.

Returns one machine's admission record by machine ID.

```bash
osapi client admission get [flags]
```

## Examples

```bash
osapi client admission get --machine-id d530664edc1c431eb5ef1e9dfd6617a5
```

## Result

The command prints:

```ansi
$ osapi client admission get --machine-id d530664edc1c431eb5ef1e9dfd6617a5
[1;38;2;0;213;255mAdmission:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;10;143;176mMachine ID:[0m      d530664edc1c431eb5ef1e9dfd6617a5
  [38;2;10;143;176mHostname:[0m        web-03
  [38;2;10;143;176mState:[0m           pending
  [38;2;10;143;176mOnline:[0m          yes
  [38;2;10;143;176mLast seen:[0m       8s
  [38;2;10;143;176mFingerprint:[0m     SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a
  [38;2;10;143;176mEncryption key:[0m  XA7NTLWJ2ECILFDLHORNVYROXOK65ONQ2SJYHRNAWRJZ3UXMVXIGUVJM
  [38;2;10;143;176mRequested at:[0m    20m
```

With `--json`, the API's response:

```json
{
  "machine_id": "d530664edc1c431eb5ef1e9dfd6617a5",
  "hostname": "web-03",
  "state": "pending",
  "online": true,
  "last_seen": "2026-01-15T09:59:52Z",
  "fingerprint": "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
  "encryption_key": "XA7NTLWJ2ECILFDLHORNVYROXOK65ONQ2SJYHRNAWRJZ3UXMVXIGUVJM",
  "requested_at": "2026-01-15T09:40:00Z"
}
```

## Permission

It needs the `admission:read` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                           | Default |
| -------------- | ------------------------------------- | ------- |
| `--machine-id` | The machine ID of the node (required) |         |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
