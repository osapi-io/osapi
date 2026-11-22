---
sidebar_position: 24
---

# Certificate management

osapi manages custom CA certificates on managed hosts. It deploys CA
certificates as PEM files through the Object Store, and adds them to the system
trust store with `update-ca-certificates`. The list operation shows the
certificates of the system and the custom certificates that osapi manages.

## How it works

The certificate provider manages custom CA certificates in
`/usr/local/share/ca-certificates/` on Debian family systems. You upload a
certificate to the Object Store as PEM content. The file provider deploys it to
the agents, and compares SHA checksums so that it does not write a file that did
not change.

### List

Returns all CA certificates on the host: the certificates of the system (from
`/usr/share/ca-certificates/`) and the custom certificates that osapi manages
(from `/usr/local/share/ca-certificates/`). Each entry has a `source` field,
`system` or `custom`. A custom certificate also has the `object` reference that
deployed it.

### Create

Deploys a new custom CA certificate to the host. First, upload the PEM content
to the Object Store. The agent writes the file to
`/usr/local/share/ca-certificates/{name}.crt`, and runs `update-ca-certificates`
to make the system trust store again. Create is idempotent **on the content**.
If a certificate on disk has the same PEM, it reports `changed: false`. If
somebody replaced or changed the certificate, the agent writes it again from the
Object Store. Use `update` to point the host to a different object.

### Update

Replaces a custom CA certificate with a new Object Store reference. The agent
deploys the PEM file again and runs `update-ca-certificates`. If the content did
not change (the same SHA), it returns `changed: false` and does not make the
trust store again. It fails if the certificate does not exist. Use `create`
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

## CLI usage

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

All commands accept `--json` to give the raw JSON output.

## Broadcast support

All certificate operations accept broadcast targets. Use `--target _all` to
manage the certificates on all registered agents. Use a label selector, for
example `--target group:web`, to target a part of the fleet.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME         CHANGED
  web-01    internal-ca  true
  web-02    internal-ca  true
```

Skipped and failed hosts appear with their error in the output.

## Supported platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On a platform that is not supported, certificate operations return
`status: skipped`. They do not fail. Refer to
[Platform Detection](../sdk/platform/detection.md) for how osapi finds the OS
family.

## Container behavior

Certificate operations work inside containers. The provider writes to
`/usr/local/share/ca-certificates/` and runs `update-ca-certificates`, which
functions normally in standard container environments.

## Permissions

| Operation              | Permission          |
| ---------------------- | ------------------- |
| List                   | `certificate:read`  |
| Create, Update, Delete | `certificate:write` |

Certificate listing requires `certificate:read`, which `admin`, `operator`,
`automation` and `read` have. Mutation operations require `certificate:write`,
which `admin` and `automation` have.

## Related

- [CLI Reference](../usage/cli/client/node/certificate/certificate.mdx):
  certificate commands
- [SDK Reference](../sdk/client/security/certificate.md): Certificate service
- [Platform Detection](../sdk/platform/detection.md): OS family detection
- [Configuration](../usage/configuration.md): full configuration reference
