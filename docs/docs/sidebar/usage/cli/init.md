# Init

Install osapi's agent skills into a repository.

The command writes the skills in this binary for each `--tool`. It also writes a
short section, with links to the skills, in the instructions file of that tool:
`CLAUDE.md` or `AGENTS.md`. It does not change text outside the osapi section.
Run it again after you upgrade osapi. It writes the skills again to agree with
the binary. It removes the files that an older version installed.

```bash
osapi init [path] [flags]
```

## Examples

```bash
osapi init
osapi init --tool claude --tool codex
osapi init ./ops-repo --json
```

## Flags

| Flag     | Description                                           | Default    |
| -------- | ----------------------------------------------------- | ---------- |
| `--tool` | Agent tool to install for: claude, codex (repeatable) | `[claude]` |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](output.md) tells how the result is printed, and what `--json` gives.
