# Generate

Generate a new token.

```bash
osapi token generate [flags]
```

## Examples

```bash
osapi token generate --subject user123 --roles=read
osapi token generate --subject svc@example.com --roles=read \
  -p status:read -p health:read
```

## Result

The command prints:

```ansi
$ osapi token generate --subject user123 --roles=read
[1;38;2;0;213;255mToken:[0m
  [38;2;10;143;176mToken:[0m    eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwicm9sZXMiOlsicmVhZCJdfQ.Xq3a8Zt0n2Jv5k7XqW1yR4sB6cD9eF0gH2iJ3kL5mN8
  [38;2;10;143;176mSubject:[0m  user123
  [38;2;10;143;176mRoles:[0m    read
```

With `--json`, the command prints:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwicm9sZXMiOlsicmVhZCJdfQ.Xq3a8Zt0n2Jv5k7XqW1yR4sB6cD9eF0gH2iJ3kL5mN8",
  "subject": "user123",
  "roles": ["read"]
}
```

## Flags

| Flag                | Description                                                                                     | Default |
| ------------------- | ----------------------------------------------------------------------------------------------- | ------- |
| `-p, --permissions` | Direct permissions, which override the roles; osapi token roles get --name admin lists each one |         |
| `-r, --roles`       | Roles for the token: built-in, or a custom role of the config (required)                        |         |
| `-u, --subject`     | Subject for the token (e.g., user ID or unique identifier) (required)                           |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
