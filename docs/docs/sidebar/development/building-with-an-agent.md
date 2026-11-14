---
sidebar_position: 5
---

# Building a behavior with an agent

A behavior is one directory, and an agent can write all of it. The osapi
repository has a skill, `build-a-behavior`, with the procedure and the
decisions: what to ask before you write, which behavior to build on, and which
checks prove that it works. Use this to build most behaviors.
[Your first behavior](your-first-behavior.md) builds one by hand.
[Behaviors](behaviors.md) is the reference for each part.

## Set it up

Open a checkout of osapi with a coding agent. The skill is in
`.claude/skills/build-a-behavior/`, and `.agents/skills` links to it. The agent
loads the skill when the request is about a behavior. `osapi init` does not
install it, because it is to change osapi, not to use it. Build the binary one
time, because the binary embeds the UI:

```bash
mise exec -- just build
```

## What to ask

Say what osapi must manage, not how:

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

- whether a similar behavior exists, to extend it instead
- whether operators call it, or only other behaviors (an internal behavior)
- which platforms it runs on. All other platforms report `skipped`
- whether part of the work already exists as an internal behavior to require
- whether a value changes with each call (an input) or with each host (a
  setting)
- which programs it runs, which with sudo, and whether an input is a secret
- where the real state is: a file that osapi writes, a file that something else
  owns, the output of a program, or `/proc`
- how "already done" looks, so that a second run changes nothing

## What it writes, and how it proves it

It runs `osapi behavior new`. Then it writes the settings, the types, the
provider, the methods and the definition, with examples that are also the tests.
Then it runs the same checks as a person:

```bash
go generate ./internal/behaviors/all
mise exec -- just generate        # the spec, SDK and UI clients, API docs
go run . behavior validate
mise exec -- just test            # every check, 100% coverage per behavior
mise exec -- just behavior-int-linux
```

Each failure names what to fix: an example with a wrong output, a write that
changes the host on its second run, a field that lets a path or an option
through, or a program that runs without a declaration. The table of the skill
gives the fix for each one.

To review the result, read the definition in `internal/behaviors/<name>/`. Then
read `osapi behavior get --name <name>` and the generated page under
[Behaviors](../behaviors/behaviors.mdx). osapi generates all of the rest of the
change from those.

## What it will refuse

The platform cannot do some things yet. Examples are a method that is an action,
for example a restart, and not a create, update or delete; a command that the
caller gives; and a resource under another resource. For those, the agent stops
and says what is missing. It does not build around the gap. Each one is a change
to the platform, and a person decides it.

## Moving a hand-written domain

A hand-written domain moves in two steps. First, you build its behavior next to
it, with the `behavior.Next` mark. Then osapi serves the behavior under
`/api/next` and `osapi client next` while both run. Then the cutover removes the
mark, and the checks name the code that is left to delete. The route check names
the directory of the hand-written domain. The CLI does not start while a
hand-written command has the name of the behavior. `references/migrating.md` of
the skill gives the procedure.
