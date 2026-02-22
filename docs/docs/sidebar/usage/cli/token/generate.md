# Generate

Generate a new token with role-based permissions:

```bash
$ osapi token generate --subject user123 --roles=read

  Token: eyJhbGciOiJI...u2E
  Subject: user123    Roles: read
```

Generate a token with direct permissions (overrides role expansion):

```bash
$ osapi token generate --subject svc@example.com --roles=read \
  -p system:read -p health:read

  Token: eyJhbGciOiJI...x3Q
  Subject: svc@example.com    Roles: read
  Permissions: system:read, health:read
```

## Flags

| Flag                | Description                                    | Default  |
| ------------------- | ---------------------------------------------- | -------- |
| `-r, --roles`       | Roles for the token (`admin`, `write`, `read`) | required |
| `-u, --subject`     | Subject for the token (e.g., user ID)          | required |
| `-p, --permissions` | Direct permissions (overrides role expansion)  | optional |

Available permissions: `system:read`, `network:read`, `network:write`,
`job:read`, `job:write`, `health:read`.
