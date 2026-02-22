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

Explore the docs:

- [Architecture](sidebar/architecture/architecture.md) — how the three processes
  (NATS, API server, worker) fit together
- [Configuration](sidebar/configuration.md) — full `osapi.yaml` reference
- [API](category/api) — OpenAPI documentation for all endpoints
- [CLI Usage](sidebar/usage/usage.md) — command reference with examples
- [Roadmap](sidebar/roadmap.md) — current capabilities and what's next

## Alternatives

- [Cockpit][]
- [webmin][]

<!-- prettier-ignore-start -->
[Cockpit]: https://cockpit-project.org/
[webmin]: https://webmin.com/
<!-- prettier-ignore-end -->
