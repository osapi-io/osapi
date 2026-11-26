# Get

Show one certificate in full: its names, algorithms and fingerprint.

```bash
osapi pki cert get [flags]
```

## Examples

```bash
osapi pki cert get --name api
```

## Result

The command prints:

```ansi
$ osapi pki cert get --name api
[1;38;2;0;213;255mCert:[0m
  [38;2;10;143;176mName:[0m           api
  [38;2;10;143;176mState:[0m          ok
  [38;2;10;143;176mComponent:[0m      controller
  [38;2;10;143;176mSource:[0m         self-signed
  [38;2;10;143;176mFile:[0m           /etc/osapi/pki/api.crt
  [38;2;10;143;176mNames:[0m          localhost, controller.example.com, 127.0.0.1
  [38;2;10;143;176mExpires:[0m        2031-11-25 12:00:00
  [38;2;10;143;176mDays left:[0m      1825
  [38;2;10;143;176mFingerprint:[0m    SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a
  [38;2;10;143;176mKey:[0m            ECDSA P-256
  [38;2;10;143;176mSignature:[0m      ECDSA-SHA256
  [38;2;10;143;176mFIPS approved:[0m  yes
```

With `--json`, the command prints:

```json
{
  "name": "api",
  "state": "ok",
  "component": "controller",
  "source": "self-signed",
  "file": "/etc/osapi/pki/api.crt",
  "names": ["localhost", "controller.example.com", "127.0.0.1"],
  "expires": "2031-11-25T12:00:00Z",
  "days_left": 1825,
  "fingerprint": "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
  "key": "ECDSA P-256",
  "signature": "ECDSA-SHA256",
  "fips_approved": true
}
```

## Flags

| Flag     | Description                                                                   | Default |
| -------- | ----------------------------------------------------------------------------- | ------- |
| `--name` | The name of the certificate, as list shows it, such as api or nats (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
