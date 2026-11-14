# Init

Install the agent skills of osapi into a repository. Then an agent in that
repository knows how to use osapi:

```bash
$ osapi init                          # Claude Code, into the current directory
$ osapi init --tool codex             # Codex
$ osapi init --tool claude --tool codex
$ osapi init ./ops-repo --json        # another directory, with a JSON summary
```

## What it writes

The command writes the skills in this binary for each `--tool`. It also writes a
short section, with links to the skills, in the instructions file of that tool:
`CLAUDE.md` or `AGENTS.md`. It does not change text outside the osapi section.

| Skill              | What the agent uses it for                                                                                  |
| ------------------ | ----------------------------------------------------------------------------------------------------------- |
| `osapi`            | Operating hosts: reading and changing kernel parameters, packages, services, users, files, DNS and the rest |
| `osapi-operations` | Running osapi itself: starting it, its configuration, tokens, and why it is unhealthy                       |

## Upgrading

Run it again after you upgrade osapi. It writes the skills again to agree with
the binary. It removes the files that an older version installed. Thus the agent
does not read a skill about operations that the binary does not have now.

## Flags

| Flag     | Default  | Description                                                          |
| -------- | -------- | -------------------------------------------------------------------- |
| `--tool` | `claude` | The agent tool to install for: `claude` or `codex`. Repeat for both. |

[Working with an agent](../agents.md) covers what to ask and how the agent finds
what osapi can do.
