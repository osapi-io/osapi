# Validate

Validate a token for authenticity and claims.

```bash
osapi controller token validate [flags]
```

## Examples

```bash
osapi controller token validate --token -
```

## Result

The command prints:

```ansi
$ osapi controller token validate --token -
[1;38;2;0;213;255mToken:[0m
  [38;2;10;143;176mKey ID:[0m                 2026-11
  [38;2;10;143;176mSubject:[0m                user123
  [38;2;10;143;176mRoles:[0m                  read
  [38;2;10;143;176mEffective permissions:[0m  file:read, health:read, status:read
  [38;2;10;143;176mAudience:[0m               osapi
  [38;2;10;143;176mIssued at:[0m              2026-01-15 09:00:00
  [38;2;10;143;176mExpires at:[0m             2026-04-15 09:00:00
```

With `--json`, the command prints:

```json
{
  "kid": "2026-11",
  "subject": "user123",
  "roles": ["read"],
  "effective_permissions": ["file:read", "health:read", "status:read"],
  "audience": ["osapi"],
  "issued_at": "2026-01-15T09:00:00Z",
  "expires_at": "2026-04-15T09:00:00Z"
}
```

## Flags

| Flag          | Description                                                                                              | Default |
| ------------- | -------------------------------------------------------------------------------------------------------- | ------- |
| `-t, --token` | The token (a secret: secret://&lt;vault&gt;/&lt;key&gt;, or - to read it from standard input) (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
