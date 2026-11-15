# Skill

Show the agent skills that this binary has. An agent reads a skill and obeys it:

```bash
$ osapi skill list
$ osapi skill get --name build-a-behavior
$ osapi skill get --name build-a-behavior --reference references/services.md
```

## Which skills it has

| Skill              | Installed by `osapi init` | What the agent uses it for                         |
| ------------------ | ------------------------- | -------------------------------------------------- |
| `osapi`            | yes                       | Read and change hosts                              |
| `osapi-operations` | yes                       | Run osapi itself                                   |
| `build-a-behavior` | no                        | Build or change a behavior, in a checkout of osapi |

`osapi init` does not install `build-a-behavior`, because only a checkout of
osapi can build a behavior. In a checkout, `.claude/skills/build-a-behavior/` is
a short pointer. It tells the agent to run `go run . skill get` and obey the
output. Thus the binary is the only source of the skill.

## Commands

| Command | What it does                                                           |
| ------- | ---------------------------------------------------------------------- |
| `list`  | Lists each skill, if `osapi init` installs it, and its reference count |
| `get`   | Prints the `SKILL.md` of a skill, or one reference file                |

## Flags of `get`

| Flag          | Description                                              |
| ------------- | -------------------------------------------------------- |
| `--name`      | The name of the skill. Necessary.                        |
| `--reference` | A reference file of the skill, relative to its directory |

Both commands take `--json`. `get --json` prints `name`, `reference` and
`content`.

The command needs no config file and no controller.
