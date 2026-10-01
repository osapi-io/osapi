<p align="center">
  <picture>
    <source srcset="asset/logo-dark.svg" media="(prefers-color-scheme: dark)">
    <source srcset="asset/logo-light.svg" media="(prefers-color-scheme: light)">
    <img src="asset/logo-dark.svg" alt="osapi" width="236">
  </picture>
</p>

<p align="center">A CRUD API for managing Linux systems.</p>

<p align="center">
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

<p align="center">
<b>Built for agents, there to empower humans.</b>
</p>

<p align="center">
Install one binary, point it at a config file, and get a REST API, a CLI,
a Go SDK and an embedded dashboard over a fleet of Linux hosts. Work reaches
a host by being queued rather than called, so a request becomes a job and an
agent runs it.
</p>

## Documentation

- [Getting Started]
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

[api]: https://osapi-io.github.io/osapi/category/api
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
