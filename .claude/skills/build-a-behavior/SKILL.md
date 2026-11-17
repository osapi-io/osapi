---
name: build-a-behavior
description: Build or change an osapi behavior, the typed definition under internal/behaviors/<name>/ that osapi's API routes, agent dispatch, CLI commands, OpenAPI spec, docs pages, settings, permissions and tests are generated from. Use when asked to add a behavior, manage something new on a host, add a method or a setting to a behavior, give a behavior a provider for another platform, migrate a hand-written domain to a behavior, or when a behavior fails osapi behavior validate, behavior-gen-check or its tests. Also use to decide whether something should be a behavior at all, whether it should be internal or served, which platforms it runs on, whether it should require another behavior or be split into an internal one, and whether it belongs in osapi-orchestrator instead. Not for the hand-written domains under internal/controller/api/node, which the add-a-domain skill in osapi-io/specs covers, and not for operations that run on the controller itself, such as the agents, the jobs and the audit log, which the build-an-endpoint skill covers.
---

# Build a behavior

The osapi binary has this skill. This file only points to it, so that the skill
has one source.

Run this command, and obey its output:

```bash
go run . skill get --name build-a-behavior
```

The skill names reference files, for example `references/services.md`. To read
one, run:

```bash
go run . skill get --name build-a-behavior --reference references/services.md
```

The binary embeds the UI. Thus, in a new checkout, run `mise exec -- just build`
one time before `go run .` works.
