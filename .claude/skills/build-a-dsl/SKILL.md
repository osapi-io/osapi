---
name: build-a-dsl
description: >-
  Add a DSL to osapi, or change what the three DSLs share: the behavior DSL
  (internal/behavior), the endpoint DSL (internal/controller/endpoint) and the
  command DSL (internal/cli/command). Use when asked to add a fourth DSL, or
  when a change touches the parts that each DSL has: the guide.Spec and
  guide.Checks tables and the tests that keep them complete, validate, typed
  values and the KnownTags list of tag keys, the permissions that the role rule
  reads (declarations, the SDK constants generator and Operations), the
  generators and genfile.Writer, the scaffold command and its check, the
  generated DSL pages, the dev skill of a DSL, records of calls at run time, or
  what extensions need. Not for a change inside one DSL: build-a-behavior,
  build-an-endpoint and build-a-command cover those.
---

# Build a DSL

The osapi binary has this skill. This file only points to it, so that the skill
has one source.

Run this command, and obey its output:

```bash
go run . skill get --name build-a-dsl
```

The binary embeds the UI. Thus, in a new checkout, run `mise exec -- just build`
one time before `go run .` works.
