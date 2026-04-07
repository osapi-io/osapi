---
sidebar_position: 24
---

# Certificate Management

OSAPI provides management of custom CA certificates on managed hosts. CA
certificates are deployed as PEM files via the Object Store and installed into
the system trust store using `update-ca-certificates`. Both system-provided and
OSAPI-managed custom certificates are visible through the list operation.

## How It Works

The certificate provider manages custom CA certificates in
`/usr/local/share/ca-certificates/` on Debian-family systems. Certificates are
uploaded to the Object Store as PEM content and deployed to agents via the file
provider's SHA-based idempotency tracking.

### List

Returns all CA certificates on the host, including both system-provided
certificates (from `/usr/share/ca-certificates/`) and OSAPI-managed custom
certificates (from `/usr/local/share/ca-certificates/`). Each entry includes a
`source` field indicating whether it is `system` or `custom`, and custom
certificates include the `object` reference used to deploy them.

### Create

Deploys a new custom CA certificate to the host. The PEM content must first be
uploaded to the Object Store. The agent writes the file to
`/usr/local/share/ca-certificates/{name}.crt` and runs `update-ca-certificates`
to rebuild the system trust store. Idempotent: returns `changed: false` if
already managed. Use `update` to replace it.

### Update

Replaces an existing custom CA certificate with a new Object Store reference.
The agent redeploys the PEM file and runs `update-ca-certificates`. If the
content has not changed (same SHA), `changed: false` is returned and the trust
store is not rebuilt. Fails if the certificate does not exist -- use `create`
first.

### Delete

Removes a custom CA certificate from the host. The agent deletes the file from
`/usr/local/share/ca-certificates/` and runs `update-ca-certificates` to rebuild
the trust store without the removed certificate.

## Operations

| Operation | Description                                      |
| --------- | ------------------------------------------------ |
| List      | List all CA certificates (system and custom)     |
| Create    | Deploy a custom CA certificate from Object Store |
| Update    | Redeploy a custom CA certificate                 |
| Delete    | Remove a custom CA certificate                   |

## CLI Usage

```bash
# Upload a CA certificate PEM to the Object Store
osapi client file upload --name internal-ca \
  --type raw --file /path/to/internal-ca.pem

# Deploy the certificate to a host
osapi client node certificate create --target web-01 \
  --name internal-ca --object internal-ca

# List all certificates on a host
osapi client node certificate list --target web-01

# Update a certificate with a new object
osapi client node certificate update --target web-01 \
  --name internal-ca --object internal-ca-v2

# Delete a certificate
osapi client node certificate delete --target web-01 \
  --name internal-ca

# Broadcast certificate deployment to all hosts
osapi client node certificate create --target _all \
  --name internal-ca --object internal-ca
```

All commands support `--json` for raw JSON output.

## Broadcast Support

All certificate operations support broadcast targeting. Use `--target _all` to
manage certificates on every registered agent, or use a label selector like
`--target group:web` to target a subset.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME         CHANGED
  web-01    internal-ca  true
  web-02    internal-ca  true
```

Skipped and failed hosts appear with their error in the output.

## Supported Platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On unsupported platforms, certificate operations return `status: skipped`
instead of failing. See [Platform Detection](../sdk/platform/detection.md) for
details on OS family detection.

## Container Behavior

Certificate operations work inside containers. The provider writes to
`/usr/local/share/ca-certificates/` and runs `update-ca-certificates`, which
functions normally in standard container environments.

## Permissions

| Operation              | Permission          |
| ---------------------- | ------------------- |
| List                   | `certificate:read`  |
| Create, Update, Delete | `certificate:write` |

Certificate listing requires `certificate:read`, included in all built-in roles
(`admin`, `write`, `read`). Mutation operations require `certificate:write`,
included in the `admin` and `write` roles.

## Related

- [CLI Reference](../usage/cli/client/node/certificate/certificate.md) --
  certificate commands
- [SDK Reference](../sdk/client/security/certificate.md) -- Certificate service
- [Platform Detection](../sdk/platform/detection.md) -- OS family detection
- [Configuration](../usage/configuration.md) -- full configuration reference
