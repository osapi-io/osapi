---
slug: /
sidebar_position: 1
title: Home
---

<img src="img/logo.png" align="left" />

A CRUD API for managing Linux systems.

This project provides basic management capabilities to Linux systems, enabling
them to be used as appliances. You install a single binary, point it at a config
file, and get a REST API and CLI for querying and changing system configuration
— hostname, DNS, disk usage, memory, load averages, and more.

<br clear="left"/>

## Install

```bash
go install github.com/retr0h/osapi@latest
```

Or download a prebuilt binary from the [releases][] page.

### Docker

A multi-arch distroless image is published to [GitHub Container Registry][ghcr]
on every commit to main. Images are signed with [cosign][] (keyless, via GitHub
OIDC) and include an [SBOM][] attestation.

```bash
docker pull ghcr.io/osapi-io/osapi:latest
docker run ghcr.io/osapi-io/osapi:latest --help
```

Verify the image signature:

```bash
cosign verify ghcr.io/osapi-io/osapi:latest \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp github.com/osapi-io/osapi
```

Verify build provenance and SBOM attestations via the GitHub CLI:

```bash
gh attestation verify oci://ghcr.io/osapi-io/osapi:latest \
  --owner osapi-io
```

## Quickstart

Install OSAPI and start all three components in a single process:

```bash
osapi start
```

Or start each component separately:

```bash
osapi nats server start &
osapi api server start &
osapi agent start &
```

Generate a token and configure the CLI:

```bash
# Generate a signing key
export OSAPI_API_SERVER_SECURITY_SIGNING_KEY=$(openssl rand -hex 32)

# Generate a bearer token
osapi token generate -r admin -u admin@example.com

# Set the token for CLI use
export OSAPI_API_CLIENT_SECURITY_BEARER_TOKEN=<token from above>
```

Query the system:

```bash
# Get the hostname
osapi client node hostname

# Check node status
osapi client node status

# View health
osapi client health
```

## Explore the Docs

- [Features](sidebar/features/node-management.md) — what OSAPI can manage and
  how each feature works
- [Architecture](sidebar/architecture/architecture.md) — how the three processes
  (NATS, API server, agent) fit together
- [Configuration](sidebar/usage/configuration.md) — full `osapi.yaml` reference
- [API](category/api) — OpenAPI documentation for all endpoints
- [CLI Usage](sidebar/usage/usage.mdx) — command reference with examples
- [Roadmap](sidebar/development/roadmap.md) — current capabilities and what's
  next

## Alternatives

- [Cockpit][]
- [webmin][]

<!-- prettier-ignore-start -->
[Cockpit]: https://cockpit-project.org/
[webmin]: https://webmin.com/
[releases]: https://github.com/retr0h/osapi/releases
[ghcr]: https://github.com/osapi-io/osapi/pkgs/container/osapi
[cosign]: https://github.com/sigstore/cosign
[SBOM]: https://en.wikipedia.org/wiki/Software_supply_chain
<!-- prettier-ignore-end -->
