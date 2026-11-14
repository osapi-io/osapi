# Init

Install osapi's agent skills into a repository, so an agent working there knows
how to use osapi:

```bash
$ osapi init                          # Claude Code, into the current directory
$ osapi init --tool codex             # Codex
$ osapi init --tool claude --tool codex
$ osapi init ./ops-repo --json        # another directory, with a JSON summary
```

## What it writes

The skills compiled into this binary, for each `--tool`, and a short section in
that tool's instructions file pointing at them: `CLAUDE.md` for Claude Code,
`AGENTS.md` for Codex. Text outside osapi's section is never touched.

| Skill              | What the agent uses it for                                                                                  |
| ------------------ | ----------------------------------------------------------------------------------------------------------- |
| `osapi`            | Operating hosts: reading and changing kernel parameters, packages, services, users, files, DNS and the rest |
| `osapi-operations` | Running osapi itself: starting it, its configuration, tokens, and why it is unhealthy                       |

## Upgrading

Run it again after upgrading osapi. It rewrites the skills to match the binary
and removes files an older version shipped, so the agent never reads a skill
that describes operations the binary no longer has.

## Flags

| Flag     | Default  | Description                                                          |
| -------- | -------- | -------------------------------------------------------------------- |
| `--tool` | `claude` | The agent tool to install for: `claude` or `codex`. Repeat for both. |

[Working with an agent](../agents.md) covers what to ask and how the agent finds
what osapi can do.
