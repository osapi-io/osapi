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

osapi lets an AI agent manage Linux hosts through typed operations instead of a
login. Every change is checked before it is queued, safe to run twice, and
recorded. `osapi init` teaches your agent what osapi can do. Built for agents,
there to empower humans.

Why it's better:

- It works in your harness. `osapi init` installs osapi's skills for Claude Code
  or Codex.
- Nothing is guessed. Every operation declares its input and output, and the
  agent reads them from `osapi behavior get` and `osapi help --json`.
- It is safe to run again. A repeated operation changes nothing the second time,
  and every result says whether the host changed.
- One host or the fleet: a hostname, any host, every host, or a label like
  `group:web`.
- A slow host never blocks you. A request becomes a job, and the agent on each
  host runs it.
- You can see exactly what happened. Every job is kept, with an audit log and
  OpenTelemetry traces.

## Getting started

The [docs][getting started] give these steps:

1. Start osapi.
2. Give your agent its skills with `osapi init`.
3. Ask for what you want.
4. Run the same operations from the CLI.

The docs also tell how to build a behavior, by hand or with an agent.

## Documentation

- [Getting Started]
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
| [osapi-lab]          | A fleet of real machines that runs a checkout of osapi               |
| [osapi-orchestrator] | A Go package for orchestrating operations across OSAPI-managed hosts |
| [specs]              | The design record for this repository, under `components/osapi/`     |

## Contributing

See the [Contributing](CONTRIBUTING.md) guide for prerequisites, setup,
conventions, and the PR workflow.

## License

The [MIT] License.

[api]: https://osapi-io.github.io/osapi/category/api
[behaviors]: https://osapi-io.github.io/osapi/sidebar/behaviors/
[getting started]: https://osapi-io.github.io/osapi/
[gohai]: https://github.com/osapi-io/gohai
[mit]: LICENSE
[nats-client]: https://github.com/osapi-io/nats-client
[nats-server]: https://github.com/osapi-io/nats-server
[osapi-justfiles]: https://github.com/osapi-io/osapi-justfiles
[osapi-lab]: https://github.com/osapi-io/osapi-lab
[osapi-orchestrator]: https://github.com/osapi-io/osapi-orchestrator
[sdk]: https://osapi-io.github.io/osapi/sidebar/sdk
[specs]: https://github.com/osapi-io/specs
[usage]: https://osapi-io.github.io/osapi/sidebar/usage
[working with an agent]: https://osapi-io.github.io/osapi/sidebar/usage/agents
