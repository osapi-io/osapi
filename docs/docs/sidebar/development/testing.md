---
sidebar_position: 3
---

# Testing

Install dependencies:

```bash
$ just deps
```

## Unit Tests

Unit tests run with mocked dependencies and require no external services:

```bash
$ just go::unit       # Run unit tests
$ just go::unit-cov   # Run with coverage report
$ just test           # Run all checks (lint + unit + coverage)
```

Unit tests follow the Go convention of being located in `*_test.go` files in the
same package as the code being tested. Public API tests use the `_test` package
suffix in `*_public_test.go` files. Public test suites also include HTTP wiring
methods (`TestXxxHTTP`, `TestXxxRBACHTTP`) that send raw HTTP through the full
Echo middleware stack with mocked backends.

## Integration Tests

Integration tests build a real `osapi` binary, start all three components (NATS,
API server, agent), and exercise CLI commands end-to-end. They are guarded by a
`//go:build integration` tag and located in `test/integration/`:

```bash
$ just go::unit-int   # Run integration tests
```

The test harness allocates random ports, generates a JWT, and starts the server
automatically — no manual setup required. Tests validate JSON responses from CLI
commands with `--json` output.

## Formatting

Auto format code:

```bash
$ just go::fmt
```

## Listing Recipes

List helpful targets:

```bash
$ just --list --list-submodules
```
