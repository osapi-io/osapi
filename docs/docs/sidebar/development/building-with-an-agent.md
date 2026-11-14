---
sidebar_position: 5
---

# Building a behavior with an agent

A behavior is one directory, and an agent can write all of it. The osapi
repository carries a skill, `build-a-behavior`, that holds the procedure and the
judgment: what to ask before writing anything, which existing behavior to build
on, which checks prove it works. This is how most behaviors are meant to be
built. [Your first behavior](your-first-behavior.md) builds one by hand, and
[Behaviors](behaviors.md) is the reference for every part.

## Set it up

Open a checkout of osapi in Claude Code. The skill is in
`.claude/skills/build-a-behavior/` and loads when the request is about a
behavior. `osapi init` does not install it: it is for changing osapi, not for
using it. Build the binary once, since it embeds the UI:

```bash
mise exec -- just build
```

## What to ask

Say what should be managed, not how:

- _"Add a behavior that manages the message of the day in /etc/motd."_
- _"Add a behavior that lists the host's ARP and NDP neighbors, under network."_
- _"Add an internal behavior that installs and removes Debian packages, for
  other behaviors to require."_
- _"Let sysctl's list filter by whether osapi manages a parameter."_
- _"Give the sysctl behavior a provider for Fedora."_
- _"Migrate the timezone domain to a behavior."_

## What it will ask you

A wrong guess becomes an API someone depends on, so the agent settles these
before it writes code, with a recommendation for each:

- whether a behavior close to this already exists, to extend instead
- whether operators call it, or only other behaviors (an internal one)
- which platforms it runs on; every other platform reports `skipped`
- whether part of the work already exists as an internal behavior to require
- whether a value changes per call (an input) or per host (a setting)
- which programs it runs, which with sudo, and whether an input is a secret
- where the truth about the thing lives: a file osapi writes, a file something
  else owns, a program's output, or `/proc`
- what "already done" looks like, so a second run changes nothing

## What it writes, and how it proves it

It runs `osapi behavior new`, then writes the settings, types, provider, methods
and definition, with examples that are also the tests. Then it runs the same
checks a person would:

```bash
go generate ./internal/behaviors/all
mise exec -- just generate        # the spec, SDK and UI clients, API docs
go run . behavior validate
mise exec -- just test            # every check, 100% coverage per behavior
mise exec -- just behavior-int-linux
```

Each failure names what to fix: an example whose output is wrong, a write that
changes the host on its second run, a field that lets a path or an option
through, a program run without being declared. The skill's table maps each one
to its fix.

To review the result, read the definition in `internal/behaviors/<name>/`, then
`osapi behavior get --name <name>` and its generated page under
[Behaviors](../behaviors/behaviors.mdx). Everything else in the change is
generated from those.

## What it will refuse

Some things the platform cannot express yet, such as a method that is an action
rather than a create, update or delete (a restart), running a command the caller
supplies, or a resource nested under another. For those the agent stops and says
what is missing instead of building around it. Each is a change to the platform,
decided by a person.

## Moving a hand-written domain

A domain that is still hand-written moves in two steps. First its behavior is
built beside it, marked `behavior.Next`, so it is served under `/api/next` and
`osapi client next` while both run. Then the cutover removes the mark, and the
checks name what is left to delete: the route check names the hand-written
domain's directory, and the CLI refuses to start while a hand-written command
has the behavior's name. The skill's `references/migrating.md` is the procedure.
