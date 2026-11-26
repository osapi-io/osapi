# List

List the certificates, with their algorithms, expiry and fingerprint.

Run this on the host. It reads each certificate that the API, the NATS server
and the metrics servers serve: the self-signed one in the key directory of the
component, or the cert_file of the config. A certificate that the host does not
have is not in the list. FIPS mode tells whether this binary runs its crypto in
FIPS 140-3 mode (GODEBUG=fips140=on).

```bash
osapi pki cert list
```

## Examples

```bash
osapi pki cert list
```

## Result

The command prints:

```ansi
$ osapi pki cert list
[1;38;2;0;213;255mCerts (1):[0m
  [38;2;10;143;176mNAME  [0m[38;2;10;143;176mCOMPONENT   [0m[38;2;10;143;176mSOURCE       [0m[38;2;10;143;176mEXPIRES              [0m[38;2;10;143;176mDAYS LEFT  [0m[38;2;10;143;176mKEY          [0m[38;2;10;143;176mFIPS[0m
  [1mapi[0m   controller  self-signed  2031-11-25 12:00:00  1825       ECDSA P-256  yes
  [38;2;10;143;176mFIPS mode:[0m  no
```

With `--json`, the command prints:

```json
{
  "fips_mode": false,
  "items": [
    {
      "name": "api",
      "component": "controller",
      "source": "self-signed",
      "file": "/etc/osapi/pki/api.crt",
      "names": ["localhost", "controller.example.com", "127.0.0.1"],
      "expires": "2031-11-25T12:00:00Z",
      "days_left": 1825,
      "fingerprint": "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
      "key": "ECDSA P-256",
      "signature": "ECDSA-SHA256",
      "fips": true
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

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
