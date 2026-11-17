---
name: build-an-endpoint
description: Build or change an endpoint group, the typed definition under internal/controller/endpoints/<group>/ of operations that run on the osapi controller itself, such as the agents, the jobs, the audit log, the files, health, facts and telemetry, and not on a host. Use when asked to add or change a controller endpoint, to move a hand-written controller group under internal/controller/api to the endpoint DSL, when osapi controller endpoint validate, TestEndpointParity or TestJoinedSpec fails, or to decide whether an operation is an endpoint or a behavior. Not for work on a host, which the build-a-behavior skill covers.
---

# Build an endpoint

The osapi binary has this skill. This file only points to it, so that the skill
has one source.

Run this command, and obey its output:

```bash
go run . skill get --name build-an-endpoint
```

The binary embeds the UI. Thus, in a new checkout, run `mise exec -- just build`
one time before `go run .` works.
