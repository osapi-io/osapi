---
sidebar_position: 2
---

# Contributing

Contributions to OSAPI are very welcome, but we ask that you read this document
before submitting a PR.

:::note

This document applies to the [OSAPI][] repository.

:::

## Before you start

- **Get familiar with the project** - Read through the docs in this order:
  1. [Development](development.md) - Prerequisites, setup, code style, testing,
     and commit conventions
  2. [Guiding Principles](../architecture/principles.md) - Design philosophy and
     project values
  3. [System Architecture](../architecture/system-architecture.md) - High-level
     component overview (REST API, NATS, CLI)
  4. [API Design Guidelines](../architecture/api-guidelines.md) - REST
     conventions and endpoint structure
  5. [Job System Architecture](../architecture/job-architecture.md) - KV-first
     job processing, subject routing, and agent pipeline
- **Check existing work** - Is there an existing PR? Are there issues discussing
  the feature/change you want to make? Please make sure you consider/address
  these discussions in your work.
- **Backwards compatibility** - Will your change break existing OSAPI files? It
  is much more likely that your change will merged if it backwards compatible.
  Is there an approach you can take that maintains this compatibility? If not,
  consider opening an issue first so that API changes can be discussed before
  you invest your time into a PR.

## Making changes

- **Code style** - Follow the conventions described in the
  [Development](development.md#code-style) guide.

- **Documentation** - Ensure that you add/update any relevant documentation.
- **Tests** - Ensure that you add/update any relevant tests and that all tests
  are passing before submitting the PR. See
  [Development](development.md#testing) for how to run tests.

## Submitting a PR

- **Describe your changes** - Ensure that you provide a comprehensive
  description of your changes.
- **Issue/PR links** - Link any previous work such as related issues or PRs.
  Please describe how your changes differ to/extend this work.
- **Examples** - Add any examples or screenshots that you think are useful to
  demonstrate the effect of your changes.
- **Draft PRs** - If your changes are incomplete, but you would like to discuss
  them, open the PR as a draft and add a comment to start a discussion. Using
  comments rather than the PR description allows the description to be updated
  later while preserving any discussions.

## FAQ

> I want to contribute, where do I start?

All kinds of contributions are welcome, whether its a typo fix or a shiny new
feature. You can also contribute by upvoting/commenting on issues or helping to
answer questions.

> I'm stuck, where can I get help?

If you have questions, feel free open a [Discussion][] on GitHub.

<!-- prettier-ignore-start -->
[OSAPI]: https://github.com/retr0h/osapi
[Discussion]: https://github.com/retr0h/go-gilt/discussions
<!-- prettier-ignore-end -->
