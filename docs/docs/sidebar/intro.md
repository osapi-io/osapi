---
slug: /
sidebar_position: 1
title: Home
---

<div style={{textAlign: 'center', marginBottom: '2rem'}}>
  <img className="logo-light" src="img/logo-light.svg" alt="osapi" width="236" />
  <img className="logo-dark" src="img/logo-dark.svg" alt="osapi" width="236" />
</div>

# Give your agent a fleet, not a shell

osapi manages Linux and macOS hosts through typed operations, not through a
login. Each operation is a **behavior**: one thing on a host, defined one time,
with a fixed input and output. osapi checks each change before it puts the
change on the queue. A second run changes nothing, and osapi records each
change. Most behaviors run on Linux and on macOS. Some run only on Linux, such
as sysctl and the systemd journal. The
[reference page](reference/behaviors/behaviors.mdx) of each behavior gives its
platforms.

An agent uses osapi through the skills that `osapi init` installs. The agent
does not guess commands. It asks osapi which behaviors exist, and what each one
accepts. A person uses the same CLI.

osapi starts with behaviors on hosts. Next come workflows, behaviors that run on
one system, and scheduled automation. [Core concepts](concepts.md) tells what
each part is, and what comes next.

## Install

```bash
go install github.com/osapi-io/osapi@latest
```

Or download a prebuilt binary from the [releases][] page.

### Docker

Each commit to main publishes a multi-arch distroless image to the [GitHub
Container Registry][ghcr]. [cosign][] signs the images without a key, through
GitHub OIDC. Each image has an [SBOM][] attestation.

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

Verify the build provenance and the SBOM attestations with the GitHub CLI:

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
which runs osapi itself. Run it again after you upgrade osapi.
[Working with an agent](usage/agents.md) tells what each skill does.

### 3. Ask for what you want

- _"Set vm.swappiness to 10 on every host in group:web and tell me which ones
  changed"_
- _"Which hosts have package updates waiting?"_
- _"Find every host under disk or memory pressure"_

The agent looks up the operation, runs it with `--json`, and reads every host's
result. Nothing logs in to a host, and the agent can only do what osapi has an
operation for.

### 4. Or run it yourself

The agent runs the same commands that you can run. Each operation is a
[behavior](reference/behaviors/behaviors.mdx): a typed definition. osapi
generates its API route, command, permission, docs and tests from the
definition. No model runs at run time. The same input always does the same
thing.

```bash
$ osapi client node sysctl get --key vm.swappiness --target _all

Results:
  HOSTNAME  STATUS  KEY            VALUE  MANAGED
  debian-1  ok      vm.swappiness  60     false
  ubuntu-1  ok      vm.swappiness  60     false
```

A write reports if it changed the host. When you run it again, it reports
`changed: false`. sysctl is the first behavior.

### 5. See what exists

```bash
osapi behavior list                        # every behavior and its methods
osapi behavior get --name sysctl           # routes, inputs, platforms, permissions
```

The agent reads the same thing as JSON before it runs anything, which is how it
knows what each field accepts.

### 6. Build a new behavior

Ask an agent in a checkout of osapi: _"Add a behavior that manages the message
of the day"_. Its `build-a-behavior` skill writes the behavior. It asks about
what it must decide, writes the definition and the provider, and runs the
checks. Or start a behavior yourself with `osapi behavior create`. In both
cases, the result is one directory, and osapi generates its API, command, docs
and tests.
[Building a behavior with an agent](development/building-with-an-agent.md) gives
the prompts and what to expect.
[Your first behavior](development/your-first-behavior.md) builds one by hand.

### 7. Build on what is there

A behavior can use another behavior. sysctl requires `managedfile`, an internal
behavior that writes files and keeps a record of each file. Thus sysctl finds
drift without its own code for it. An internal behavior has no route or command
of its own. Only other behaviors call it, on the same host and in the same job.

### 8. Safe by default

Each method has a permission. Each request is in the audit log. A secret input
never goes to a log or a trace. When the agent starts, it checks that it can run
what each behavior declares it runs with sudo.

### 9. One host or the fleet

`--target` takes a hostname, `_any`, `_all`, or a label such as `group:web`.
Every host answers for itself: `ok`, `failed`, or `skipped` where the behavior
does not run on its platform.

## Explore the Docs

- [Core concepts](concepts.md): behaviors, building blocks, jobs, targets,
  skills, and what comes next
- [Behaviors](reference/behaviors/behaviors.mdx): every behavior, generated from
  its definition
- [Working with an agent](usage/agents.md): `osapi init`, the skills, and the
  prompts
- [Features](features/features.mdx): what OSAPI can manage and how each feature
  works
- [Architecture](architecture/architecture.md): how the three processes (NATS,
  API server, agent) work together
- [Configuration](usage/configuration.md): full `osapi.yaml` reference
- [API](/category/api): OpenAPI documentation for all endpoints
- [CLI Usage](usage/usage.mdx): command reference with examples

## Alternatives

- [Cockpit][] and [webmin][] give a person a web console for one server.
- [Ansible][] pushes configuration to hosts over SSH, from playbooks.

### Why osapi

- **Built for an agent.** The agent asks osapi which behaviors exist and what
  each one accepts. It does not read a web page or guess a module. Each answer
  is typed JSON, with a result for each host.
- **The agent can build a behavior.** In a checkout of osapi, the
  `build-a-behavior` skill writes a new behavior. A behavior is one directory.
  osapi generates its API, CLI command, permission, docs and tests.
- **Each behavior is shared.** When a behavior merges, each osapi user gets it:
  in the API, the CLI and the docs. The agent finds it with
  `osapi behavior get`.
- **A fleet without SSH.** An agent on each host connects out to the message
  bus. A request becomes a job. A slow host does not block the others.
- **Safe to run again.** Each write reports `changed`. A second run reports
  `changed: false`. Each host reports `ok`, `failed` or `skipped`.
- **One host or many, through one API.** A web console manages one server. osapi
  gives the same API, CLI, SDK and dashboard for one host or the fleet.
- **A record of each change.** osapi keeps each job, an audit log and
  OpenTelemetry traces. A secret input never goes to the log or a trace.
- **Workflows are next.** A workflow will be ordered behavior calls in a file,
  checked before it runs, and later on a schedule. Refer to
  [Core concepts](concepts.md).

### When to use the others

osapi has fewer behaviors than Ansible has modules. If you need a module that
osapi does not have yet, use Ansible for that task, or build the behavior. Use
Cockpit or webmin when one person manages one server through a browser.

<!-- prettier-ignore-start -->
[Ansible]: https://www.ansible.com/
[Cockpit]: https://cockpit-project.org/
[webmin]: https://webmin.com/
[releases]: https://github.com/osapi-io/osapi/releases
[ghcr]: https://github.com/osapi-io/osapi/pkgs/container/osapi
[cosign]: https://github.com/sigstore/cosign
[SBOM]: https://en.wikipedia.org/wiki/Software_supply_chain
<!-- prettier-ignore-end -->
