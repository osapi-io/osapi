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

## Quickstart

Install OSAPI and start all three processes:

```bash
# Start the embedded NATS server
osapi nats server start &

# Start the API server
osapi api server start &

# Start a node agent
osapi node agent start &
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
  (NATS, API server, worker) fit together
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
<!-- prettier-ignore-end -->
