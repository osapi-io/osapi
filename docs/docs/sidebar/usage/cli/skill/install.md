# Install

Install the skills of this binary into a repository.

The command writes the skills in this binary for each `--tool`. It also writes a
short section, with links to the skills, in the instructions file of that tool:
`CLAUDE.md` or `AGENTS.md`. It does not change text outside the osapi section.

Run it again after you upgrade osapi. It writes the skills again to agree with
the binary. It removes the files that an older version installed. Thus the agent
does not read a skill about operations that the binary does not have now.

```bash
osapi skill install [flags]
```

## Examples

```bash
osapi skill install
osapi skill install --tool claude --tool codex
osapi skill install --path ./ops-repo --json
```

## Result

The command prints:

```ansi
$ osapi skill install
[1;38;2;0;213;255mSkills:[0m
  [38;2;10;143;176mVersion:[0m  v1.4.0
  [38;2;10;143;176mTools:[0m    claude
  [38;2;10;143;176mChanged:[0m  5

[1;38;2;255;0;255mFiles (5):[0m
  [38;2;176;48;176mFILE                                                  [0m[38;2;176;48;176mSTATUS[0m
  .claude/skills/osapi/SKILL.md                         created
  .claude/skills/osapi/references/behaviors.md          created
  .claude/skills/osapi-operations/SKILL.md              created
  .claude/skills/osapi-operations/references/config.md  created
  CLAUDE.md                                             updated
```

With `--json`, the command prints:

```json
{
  "version": "v1.4.0",
  "tools": ["claude"],
  "changed": 5,
  "files": [
    {
      "path": ".claude/skills/osapi/SKILL.md",
      "status": "created"
    },
    {
      "path": ".claude/skills/osapi/references/behaviors.md",
      "status": "created"
    },
    {
      "path": ".claude/skills/osapi-operations/SKILL.md",
      "status": "created"
    },
    {
      "path": ".claude/skills/osapi-operations/references/config.md",
      "status": "created"
    },
    {
      "path": "CLAUDE.md",
      "status": "updated"
    }
  ]
}
```

## Flags

| Flag     | Description                                                                              | Default |
| -------- | ---------------------------------------------------------------------------------------- | ------- |
| `--path` | The repository to install into; the current directory without it                         | `.`     |
| `--tool` | Agent tool to install for: claude, codex (repeatable); claude without it (claude, codex) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
