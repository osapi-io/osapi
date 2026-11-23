# Validate

Validate a token for authenticity and claims.

```bash
osapi token validate [flags]
```

## Examples

```bash
osapi token validate --token eyJhbGciOiJIUzI1NiIs...
```

## Result

The command prints:

```ansi
$ osapi token validate --token eyJhbGciOiJIUzI1NiIs...
[1;38;2;0;213;255mToken:[0m
  [38;2;10;143;176mSubject:[0m                user123
  [38;2;10;143;176mRoles:[0m                  read
  [38;2;10;143;176mEffective permissions:[0m  file:read, health:read, node:read
  [38;2;10;143;176mAudience:[0m               osapi
  [38;2;10;143;176mIssued at:[0m              2026-01-15 09:00:00
  [38;2;10;143;176mExpires at:[0m             2026-04-15 09:00:00
```

With `--json`, the command prints:

```json
{
  "subject": "user123",
  "roles": ["read"],
  "effective_permissions": ["file:read", "health:read", "node:read"],
  "audience": ["osapi"],
  "issued_at": "2026-01-15T09:00:00Z",
  "expires_at": "2026-04-15T09:00:00Z"
}
```

## Flags

| Flag          | Description                 | Default |
| ------------- | --------------------------- | ------- |
| `-t, --token` | The Token string (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
