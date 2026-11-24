# Management dashboard

osapi has an embedded React management dashboard. The controller binary serves
it. You need no different web server, container image or static file host. When
you deploy the `osapi` binary, the UI is ready.

## Overview

The dashboard provides a visual interface for fleet management:

- **Dashboard**: an overview of the fleet health, with agent cards (machine ID,
  fingerprint, version, scheduling state), the component status, JetStream
  metrics and node conditions
- **Behaviors**: what osapi manages on a host, by area. The page of each
  behavior shows its methods, their permissions, platforms and chains, and their
  examples.
- **Plans**: steps that run in order. Each step is a method of a behavior with
  its target and its input. A plan exports as YAML.
- **Admin**: the audit log with export, the job queue with retry and delete, PKI
  enrollment (accept, reject and remove agents) and an RBAC reference (admin
  only)

## Enabling the UI

The dashboard is enabled by default. To disable it, set:

```yaml
controller:
  ui:
    enabled: false
```

When disabled, the controller serves only the REST API.

## Authentication

The UI uses the same JWT-based authentication as the CLI and API. Generate a
token with `osapi controller token generate` and paste it on the sign-in page.
The token's `roles` claim determines what the user can see and do. The UI shows
and hides each control by the permissions of the token's roles.
[Roles and permissions](authentication.md#roles-and-permissions) lists the
built-in roles.

## Architecture

The controller serves the dashboard at `/`. For a path that it does not know, it
sends `index.html`, so that the UI can route the page. Thus each API endpoint
starts with `/api/`, to keep the two apart.

Refer to [UI Architecture](../../architecture/ui.md) for the configuration, the
roles and what each page shows. The
[specifications repository](https://github.com/osapi-io/specs/blob/main/components/osapi/ui.md)
tells how the UI is embedded, structured and generated.

## Development

Refer to [UI Development](../../development/ui-development.md) for the index of
the rules a contributor works under, and where the commands live.
