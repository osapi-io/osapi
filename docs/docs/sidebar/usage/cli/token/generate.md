# Generate

Generate a new token with role-based permissions:

```bash
$ osapi token generate --subject user123 --roles=read
8:51AM INF generated token token=eyJhbG...u2E roles=read subject=user123
```

Generate a token with direct permissions (overrides role expansion):

```bash
$ osapi token generate --subject svc@example.com --roles=read \
  -p system:read -p health:read
8:51AM INF generated token token=eyJhbG...x3Q roles=read subject=svc@example.com
8:51AM INF token permissions permissions=system:read,health:read
```

## Flags

| Flag                | Description                                    | Default  |
| ------------------- | ---------------------------------------------- | -------- |
| `-r, --roles`       | Roles for the token (`admin`, `write`, `read`) | required |
| `-u, --subject`     | Subject for the token (e.g., user ID)          | required |
| `-p, --permissions` | Direct permissions (overrides role expansion)  | optional |

Available permissions: `system:read`, `network:read`, `network:write`,
`job:read`, `job:write`, `health:read`.
