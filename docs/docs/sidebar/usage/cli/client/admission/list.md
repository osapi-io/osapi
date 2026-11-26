# List

List admission records.

Each node that asked to join, each node osapi admitted, and each decision
against a node. An admitted node that is not running is Offline, with when it
was last seen.

```bash
osapi client admission list [flags]
```

## Examples

```bash
osapi client admission list
osapi client admission list --state pending
```

## Result

The command prints:

```ansi
$ osapi client admission list
[1;38;2;0;213;255mAdmissions (2):[0m
  [38;2;10;143;176mMACHINE ID                        [0m[38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATE     [0m[38;2;10;143;176mONLINE  [0m[38;2;10;143;176mLAST SEEN  [0m[38;2;10;143;176mFINGERPRINT[0m
  [38;2;138;138;138md530664edc1c431eb5ef1e9dfd6617a5[0m  [1mweb-03[0m    pending   yes     8s         SHA256:1f53dea6b0c3e...
  [38;2;138;138;138m8f14e45fceea167a5a36dedd4bea2543[0m  [1mweb-01[0m    admitted  yes     8s         SHA256:9a7c3e1b5d2f4...

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "total_items": 2,
  "items": [
    {
      "machine_id": "d530664edc1c431eb5ef1e9dfd6617a5",
      "hostname": "web-03",
      "state": "pending",
      "online": true,
      "last_seen": "2026-01-15T09:59:52Z",
      "fingerprint": "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
      "encryption_key": "XA7NTLWJ2ECILFDLHORNVYROXOK65ONQ2SJYHRNAWRJZ3UXMVXIGUVJM",
      "requested_at": "2026-01-15T09:40:00Z"
    },
    {
      "machine_id": "8f14e45fceea167a5a36dedd4bea2543",
      "hostname": "web-01",
      "state": "admitted",
      "online": true,
      "last_seen": "2026-01-15T09:59:52Z",
      "fingerprint": "SHA256:9a7c3e1b5d2f4a6c8e0b1d3f5a7c9e2b4d6f8a0c2e4b6d8f0a1c3e5b7d9f1a3c",
      "encryption_key": "XBYF3LNXM7EBLLEZY3ANJUQIF5G2WZ6ICP3UDHLG42RMQHM6Y4QQBN75",
      "admitted_at": "2026-01-15T09:40:00Z"
    }
  ]
}
```

## Permission

It needs the `admission:read` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag      | Description                                                                                                   | Default |
| --------- | ------------------------------------------------------------------------------------------------------------- | ------- |
| `--state` | Only the records in this state: pending, admitted, rejected or revoked (pending, admitted, rejected, revoked) |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
