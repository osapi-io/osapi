# Validate

Validate a JSON Web Token (JWT) by checking its signature, expiration, and
claims:

```bash
$ osapi token validate --token eyJhbGciOiJIUzI1NiIs...

  Subject: user123            Roles: admin
  Effective Permissions: system:read, network:read, network:write, job:read, job:write, health:read
  Audience: osapi
  Issued: 2026-01-15T08:00:00Z    Expires: 2026-07-15T08:00:00Z
```

When a token carries direct permissions, they are shown separately and override
role expansion:

```bash
$ osapi token validate --token eyJhbGciOiJIUzI1NiIs...

  Subject: svc@example.com    Roles: read
  Permissions: system:read, health:read
  Effective Permissions: system:read, health:read
  Audience: osapi
  Issued: 2026-01-15T08:00:00Z    Expires: 2026-07-15T08:00:00Z
```

## Flags

| Flag          | Description      | Default  |
| ------------- | ---------------- | -------- |
| `-t, --token` | The token string | required |
