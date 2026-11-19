---
name: build-a-command
description: Build or change a group of local commands, the typed definition under internal/cli/commands/<group>/ of osapi commands that run in the CLI process and call no API, such as osapi token generate. Use when asked to add or change a local command, to move a hand-written command under cmd/ to the command DSL, when osapi cli command validate, TestRefusals or TestCommandShape fails, or to decide whether a command is a local command, an endpoint or a behavior. Not for a command that calls the API: the build-an-endpoint and build-a-behavior skills cover those.
---

# Build a command

The osapi binary has this skill. This file only points to it, so that the skill
has one source.

Run this command, and obey its output:

```bash
go run . skill get --name build-a-command
```

The binary embeds the UI. Thus, in a new checkout, run `mise exec -- just build`
one time before `go run .` works.
