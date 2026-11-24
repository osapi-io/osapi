# Certificate management

osapi reads the certificate authorities (CAs) that a host trusts, and adds and
removes the CAs that it deploys. The
[certificate behavior](../../reference/behaviors/security/certificate/certificate.md)
does this work.

## How it works

`list` and `get` read the trust store through gohai's `ca_certificates`
collector, on Linux and on macOS. Each entry has its subject, its SHA-256
fingerprint and its expiry. Its `source` is `system` for a CA of the
distribution, and `local` for a CA that an administrator added.

| Family         | System CAs                                         | Local CAs                            |
| -------------- | -------------------------------------------------- | ------------------------------------ |
| Debian, Ubuntu | `/usr/share/ca-certificates`                       | `/usr/local/share/ca-certificates`   |
| SUSE           | `/usr/share/pki/trust`                             | `/etc/pki/trust/anchors`             |
| RHEL, Fedora   | `/usr/share/pki/ca-trust-source`                   | `/etc/pki/ca-trust/source/anchors`   |
| macOS          | `/System/Library/Keychains/SystemRootCertificates` | `/Library/Keychains/System.keychain` |

`create` writes the PEM of an object from the Object Store as a file in the
local directory of the family. The file name has the `osapi-` prefix: the CA
`internal-ca` is `osapi-internal-ca.crt`. Then osapi rebuilds the trust store,
with `update-ca-certificates` on the Debian family and SUSE, and with
`update-ca-trust` on the RHEL family. `managedfile` keeps a record of each file,
so `list` shows a CA that osapi deployed by its name, `internal-ca`, with
`managed: true` and the object that it came from.

Each write is idempotent:

- `create` of a CA that osapi deployed with the same content reports
  `changed: false`, and does not rebuild the trust store.
- `create` refuses a file of the same name that osapi did not write.
- `update` writes the CA again from a new object, or from its object when you do
  not give one. A CA that osapi did not deploy is not found.
- `delete` removes the file and rebuilds the trust store. On the Debian family
  and SUSE, it runs `update-ca-certificates --fresh`, which removes the links of
  the CA from `/etc/ssl/certs`. A CA that osapi did not deploy, such as a system
  CA, does not change and reports `changed: false`.
- When the rebuild of the trust store fails, a retry with the same input
  rebuilds it again. After `create` or `update`, the file stays, and its record
  has a mark that tells that the rebuild did not finish. After `delete`, the
  file comes back with the mark. The next write of the CA reports a change and
  removes the mark.

On macOS, `create`, `update` and `delete` report `skipped`.

`get` takes a name that a path segment can hold. A name with a slash, such as
the Debian system CA `mozilla/ISRG_Root_X1`, is in `list` only.

## CLI usage

```bash
# Upload a CA certificate PEM to the Object Store
osapi client file upload --name internal-ca \
  --content-type raw --file /path/to/internal-ca.pem

# Deploy the certificate to a host
osapi client node certificate create --target web-01 \
  --name internal-ca --object internal-ca

# Get one certificate
osapi client node certificate get --target web-01 --name internal-ca

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

| OS family    | list, get | create, update, delete |
| ------------ | --------- | ---------------------- |
| Debian       | Yes       | Yes                    |
| SUSE         | Yes       | Yes                    |
| RHEL, Fedora | Yes       | Yes                    |
| Darwin       | Yes       | Skipped                |

Refer to [Platform Detection](../../sdk/library/platform-detection.md) for how
osapi finds the OS family.

## Container behavior

The writes operate in a container. They write a file and run
`update-ca-certificates` or `update-ca-trust`, which operate in a container as
on a host.

## Permissions

The [certificate](../../reference/behaviors/security/certificate/certificate.md)
reference page gives the permission of each method.

## Related

- [Behavior reference](../../reference/behaviors/security/certificate/certificate.md):
  the certificate behavior
- [CLI Reference](../../usage/cli/client/node/certificate/certificate.mdx):
  certificate commands
- [SDK Reference](../../sdk/security/certificate.md): Certificate service
- [Platform Detection](../../sdk/library/platform-detection.md): OS family
  detection
- [Configuration](../../usage/configuration.md): full configuration reference
