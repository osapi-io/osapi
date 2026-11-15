<p align="center">
  <picture>
    <source srcset="asset/logo-dark.svg" media="(prefers-color-scheme: dark)">
    <source srcset="asset/logo-light.svg" media="(prefers-color-scheme: light)">
    <img src="asset/logo-dark.svg" alt="osapi" width="236">
  </picture>
</p>

<p align="center">Typed, repeatable operations for Linux hosts.</p>

<p align="center">
  <a href="https://github.com/osapi-io/osapi/releases/latest"><img alt="release" src="https://img.shields.io/github/release/osapi-io/osapi.svg?style=for-the-badge"></a>
  <a href="https://codecov.io/gh/osapi-io/osapi"><img alt="codecov" src="https://img.shields.io/codecov/c/github/osapi-io/osapi?token=NF0T86B1EP&style=for-the-badge"></a>
  <a href="LICENSE"><img alt="license" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge"></a>
  <a href="https://github.com/osapi-io/osapi/actions/workflows/go.yml"><img alt="build" src="https://img.shields.io/github/actions/workflow/status/osapi-io/osapi/go.yml?style=for-the-badge"></a>
  <a href="https://github.com/osapi-io/osapi/pkgs/container/osapi"><img alt="docker" src="https://img.shields.io/badge/ghcr.io-osapi-blue?style=for-the-badge&logo=docker&logoColor=white"></a>
  <a href="https://osapi-io.github.io/osapi/#docker"><img alt="cosign" src="https://img.shields.io/badge/signed-cosign-blueviolet?style=for-the-badge&logo=sigstore&logoColor=white"></a>
  <a href="https://github.com/osapi-io/osapi/actions/workflows/docker-publish.yml"><img alt="sbom" src="https://img.shields.io/badge/SBOM-attached-green?style=for-the-badge"></a>
  <a href="https://github.com/goreleaser"><img alt="powered by" src="https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=for-the-badge"></a>
  <a href="https://conventionalcommits.org"><img alt="conventional commits" src="https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge"></a>
  <img alt="openapi initiative" src="https://img.shields.io/badge/openapiinitiative-%23000000.svg?style=for-the-badge&logo=openapiinitiative&logoColor=white">
  <img alt="Linux" src="https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black">
  <img alt="gitHub commit activity" src="https://img.shields.io/github/commit-activity/m/osapi-io/osapi?style=for-the-badge">
  <a href="https://pkg.go.dev/github.com/osapi-io/osapi"><img alt="go reference" src="https://img.shields.io/badge/go-reference-00ADD8?style=for-the-badge&logo=go&logoColor=white"></a>
</p>

# Give your agent a fleet, not a shell

osapi lets an AI agent manage Linux hosts through typed operations, not through
a login. Each operation is a **behavior**: one thing on a host, defined one
time, with a fixed input and output. osapi checks each request before it queues
it. A second run changes nothing. Each change is recorded.

The agent does not guess commands. It asks osapi what exists, and osapi answers
with the behaviors that this binary has. No model runs at run time. The same
input always does the same thing.

## Why osapi

- **The agent asks the binary.** `osapi behavior get` and `osapi help --json`
  give each input, output and permission. The agent reads them before it acts.
- **A second run is safe.** A write reports `changed: true` one time. After
  that, it reports `changed: false`.
- **One host or the fleet.** Target a hostname, any host, each host, or a label
  such as `group:web`. Each host answers for itself.
- **A slow host does not stop you.** A request becomes a job. The agent on each
  host runs it.
- **You can see what occurred.** osapi keeps each job, an audit log, and
  OpenTelemetry traces.
- **Root is a decision, not a default.** Command exec is off until you turn it
  on. Work is under way so that the agent does not run as root: each behavior
  declares the steps that need root, and an admin installs the sudoers rules for
  the behaviors that a host needs.
- **New behaviors are small.** A behavior is one directory. osapi generates its
  API, CLI command, permission, docs and tests. An agent can build one with the
  `build-a-behavior` skill.

## What you can ask

Install the skills with `osapi init`, then ask your agent:

- _"Set vm.swappiness to 10 on each host in group:web, and tell me which hosts
  changed."_
- _"Which hosts have package updates?"_
- _"Find each host with disk or memory pressure."_
- _"Add a behavior that manages the message of the day."_

The agent finds the behavior, runs it with `--json`, and reads the result of
each host. You can run the same commands from the CLI.

## Core concepts

| Concept        | What it is                                                           |
| -------------- | -------------------------------------------------------------------- |
| Behavior       | One thing that osapi manages on a host, defined one time as typed Go |
| Method         | A read or a write of a behavior, with a typed input and output       |
| Building block | An internal behavior that other behaviors use, such as `managedfile` |
| Job            | A request on the queue. The agent on each target host runs it        |
| Target         | A hostname, `_any`, `_all`, or a label such as `group:web`           |
| Skill          | What `osapi init` installs, so that an agent knows how to use osapi  |

[Core concepts][concepts] tells more about each one.

## Where osapi goes next

osapi starts with behaviors on hosts. The next steps make it a platform for
automation:

- **Workflows.** A file with named, ordered behavior calls. osapi checks the
  file before it runs. A step can run only where an earlier step changed
  something. A second run reports `changed: false`.
- **Behaviors on one system.** A behavior that runs on the controller, or on
  your machine, and not on a fleet host. For example, it gets data from AWS. A
  workflow then uses that data to configure the hosts.
- **Scheduled workflows.** A workflow declares a schedule, and osapi runs it.
- **More behaviors.** Each hand-written domain moves to a behavior, one at a
  time.

The [design pages][specs] in osapi-io/specs have the plans.

## Alternatives

[Cockpit] and [webmin] give a person a web console for one server. [Ansible]
pushes configuration over SSH. osapi is built for an agent:

- The agent asks osapi what exists, and gets typed answers for each host.
- The agent can build a new behavior, and each osapi user then gets it.
- There is no SSH. A request is a job, and a slow host does not block the
  others.
- Workflows of behaviors are next.

[Why osapi][why] has the full comparison, and when to use the others.

## Getting started

The [docs][getting started] give these steps:

1. Start osapi.
2. Give your agent its skills with `osapi init`.
3. Ask for what you want.
4. Run the same operations from the CLI.

The docs also tell how to build a behavior, by hand or with an agent.

## Documentation

- [Getting Started]
- [Core concepts][concepts]
- [Working with an agent]
- [Behaviors]
- [API]
- [Usage]
- [SDK]

## Sister projects

| Project              | Description                                                          |
| -------------------- | -------------------------------------------------------------------- |
| [gohai]              | A Go-based system fact collector inspired by Chef Ohai               |
| [nats-client]        | A Go package for connecting to and interacting with a NATS server    |
| [nats-server]        | A Go package for running an embedded NATS server                     |
| [osapi-justfiles]    | Shared justfiles this repository fetches its recipes from            |
| [osapi-orchestrator] | A Go package for orchestrating operations across OSAPI-managed hosts |
| [specs]              | The design record for this repository, under `components/osapi/`     |

## Contributing

See the [Contributing](CONTRIBUTING.md) guide for prerequisites, setup,
conventions, and the PR workflow.

## License

The [MIT] License.

[ansible]: https://www.ansible.com/
[api]: https://osapi-io.github.io/osapi/category/api
[behaviors]: https://osapi-io.github.io/osapi/sidebar/behaviors/
[cockpit]: https://cockpit-project.org/
[concepts]: https://osapi-io.github.io/osapi/sidebar/concepts
[getting started]: https://osapi-io.github.io/osapi/
[gohai]: https://github.com/osapi-io/gohai
[mit]: LICENSE
[nats-client]: https://github.com/osapi-io/nats-client
[nats-server]: https://github.com/osapi-io/nats-server
[osapi-justfiles]: https://github.com/osapi-io/osapi-justfiles
[osapi-orchestrator]: https://github.com/osapi-io/osapi-orchestrator
[sdk]: https://osapi-io.github.io/osapi/sidebar/sdk
[specs]: https://github.com/osapi-io/specs
[usage]: https://osapi-io.github.io/osapi/sidebar/usage
[webmin]: https://webmin.com/
[why]: https://osapi-io.github.io/osapi/#why-osapi
[working with an agent]: https://osapi-io.github.io/osapi/sidebar/usage/agents
