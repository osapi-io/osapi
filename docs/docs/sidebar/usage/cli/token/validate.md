# Validate

Validate a JSON Web Token (JWT) by checking its signature, expiration, and
claims:

```bash
$ osapi token validate --token eyJhbGciOiJIUzI1NiIs...

  Subject: user123            Roles: read, write
  Audience: osapi
  Issued: 2026-01-15T08:00:00Z    Expires: 2026-07-15T08:00:00Z
```

## Flags

| Flag          | Description      | Default  |
| ------------- | ---------------- | -------- |
| `-t, --token` | The token string | required |
