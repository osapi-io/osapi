# List

List each key and each certificate of osapi on this host.

Run this on the host. It reads what osapi pki key list and osapi pki cert list
read, in one list: the kind, the name, the component, the state, the algorithm,
the fingerprint, or the ID of a secret key, and the expiry of a certificate.

```bash
osapi pki list
```

## Examples

```bash
osapi pki list
```

## Result

The command prints:

```ansi
$ osapi pki list
[1;38;2;0;213;255mKeys and certificates (3):[0m
  [38;2;10;143;176mKIND  [0m[38;2;10;143;176mNAME        [0m[38;2;10;143;176mCOMPONENT   [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mALGORITHM    [0m[38;2;10;143;176mEXPIRES[0m
  key   [1mcontroller[0m  controller  ok     Ed25519
  key   [1mjetstream[0m   nats        ok     AES-256-GCM
  cert  [1mapi[0m         controller  ok     ECDSA P-256  2031-11-25 12:00:00
```

With `--json`, the command prints:

```json
{
  "items": [
    {
      "kind": "key",
      "name": "controller",
      "component": "controller",
      "state": "ok",
      "file": "/etc/osapi/pki/controller.key",
      "algorithm": "Ed25519",
      "fingerprint": "SHA256:9c1f2e3d4b5a69788796a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4"
    },
    {
      "kind": "key",
      "name": "jetstream",
      "component": "nats",
      "state": "ok",
      "file": "/etc/osapi/nats/jetstream.key",
      "algorithm": "AES-256-GCM",
      "fingerprint": "7f3a9c1e5b2d8f40"
    },
    {
      "kind": "cert",
      "name": "api",
      "component": "controller",
      "state": "ok",
      "file": "/etc/osapi/pki/api.crt",
      "algorithm": "ECDSA P-256",
      "fingerprint": "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
      "expires": "2031-11-25T12:00:00Z"
    }
  ]
}
```

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
