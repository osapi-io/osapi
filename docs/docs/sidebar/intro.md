---
slug: /
sidebar_position: 1
title: Home
---

<div style={{textAlign: 'center', marginBottom: '2rem'}}>
  <img src="img/logo.png" style={{maxWidth: '200px'}} />
</div>

# Give your agent a fleet, not a shell

osapi manages Linux hosts through typed operations instead of a login. Every
change is checked before it is queued, safe to run twice, and recorded. An agent
drives it through skills `osapi init` installs; a person drives it through the
same CLI.

## Install

```bash
go install github.com/osapi-io/osapi@latest
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

Attestations are also visible on the [package page][ghcr].

## Get started

### 1. Start osapi

`osapi start` runs the queue, the controller and an agent in one process. The
controller signs tokens with a key, and the CLI sends one:

```bash
export OSAPI_API_SERVER_SECURITY_SIGNING_KEY=$(openssl rand -hex 32)
osapi start &

osapi token generate -r admin -u admin@example.com
export OSAPI_API_CLIENT_SECURITY_BEARER_TOKEN=<token from above>
```

`osapi config list` shows every setting and where its value came from, and
[Configuration](usage/configuration.md) is the full reference.

### 2. Give your agent osapi

In the repository you work from:

```bash
osapi init                  # Claude Code, the default
osapi init --tool codex     # Codex
```

It installs two skills: `osapi`, which operates hosts, and `osapi-operations`,
which runs osapi itself. Run it again after upgrading osapi.
[Working with an agent](usage/agents.md) says what each one does.

### 3. Ask for what you want

- _"Set vm.swappiness to 10 on every host in group:web and tell me which ones
  changed"_
- _"Which hosts have package updates waiting?"_
- _"Find every host under disk or memory pressure"_

The agent looks up the operation, runs it with `--json`, and reads every host's
result. Nothing logs in to a host, and the agent can only do what osapi has an
operation for.

### 4. Or run it yourself

The agent runs the same commands you can. Each operation is a
[behavior](behaviors/behaviors.mdx): a typed definition whose API route,
command, permission, docs and tests are generated from it. There is no model
anywhere at run time; the same input does the same thing every time.

```bash
$ osapi client next node sysctl get --key vm.swappiness --target _all

Results:
  HOSTNAME  STATUS  KEY            VALUE  MANAGED
  debian-1  ok      vm.swappiness  60     false
  ubuntu-1  ok      vm.swappiness  60     false
```

A write reports whether it changed the host, and running it again reports
`changed: false`. sysctl is the first behavior and is served under
`osapi client next` while it replaces the older sysctl commands.

### 5. See what exists

```bash
osapi behavior list                        # every behavior and its methods
osapi behavior get --name sysctl           # routes, inputs, platforms, permissions
```

The agent reads the same thing as JSON before it runs anything, which is how it
knows what each field accepts.

### 6. Build a new behavior

Ask an agent in a checkout of osapi, _"Add a behavior that manages the message
of the day"_, and its `build-a-behavior` skill writes one: it asks what it needs
to decide, writes the definition and the provider, and runs the checks. Or start
one yourself with `osapi behavior new`. Either way the result is one directory,
and its API, command, docs and tests are generated.
[Building a behavior with an agent](development/building-with-an-agent.md) has
the prompts and what to expect;
[Your first behavior](development/your-first-behavior.md) builds one by hand.

### 7. Build on what is there

A behavior can be built on another. sysctl requires `managedfile`, an internal
behavior that writes files and keeps a record of each, so sysctl gets drift
detection without writing it. An internal behavior has no route or command of
its own; only other behaviors call it, on the same host, in the same job.

### 8. Safe by default

Every method has a permission, every request is in the audit log, a secret input
never reaches a log or a trace, and the agent checks at startup that it may run
what each behavior declares it runs with sudo.

### 9. One host or the fleet

`--target` takes a hostname, `_any`, `_all`, or a label such as `group:web`.
Every host answers for itself: `ok`, `failed`, or `skipped` where the behavior
does not run on its platform.

## Explore the Docs

- [Behaviors](behaviors/behaviors.mdx) — every behavior, generated from its
  definition
- [Working with an agent](usage/agents.md) — `osapi init`, the skills, and the
  prompts
- [Features](features/node-management.md) — what OSAPI can manage and how each
  feature works
- [Architecture](architecture/architecture.md) — how the three processes (NATS,
  API server, agent) fit together
- [Configuration](usage/configuration.md) — full `osapi.yaml` reference
- [API](/category/api) — OpenAPI documentation for all endpoints
- [CLI Usage](usage/usage.mdx) — command reference with examples

## Alternatives

- [Cockpit][]
- [webmin][]

<!-- prettier-ignore-start -->
[Cockpit]: https://cockpit-project.org/
[webmin]: https://webmin.com/
[releases]: https://github.com/osapi-io/osapi/releases
[ghcr]: https://github.com/osapi-io/osapi/pkgs/container/osapi
[cosign]: https://github.com/sigstore/cosign
[SBOM]: https://en.wikipedia.org/wiki/Software_supply_chain
<!-- prettier-ignore-end -->
