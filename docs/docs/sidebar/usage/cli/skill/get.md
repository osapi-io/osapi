# Get

Print the `SKILL.md` of a skill, or one of its other files:

```bash
$ osapi skill get --name build-a-behavior
$ osapi skill get --name build-a-behavior --reference references/services.md
```

## Flags

| Flag          | Description                                              |
| ------------- | -------------------------------------------------------- |
| `--name`      | The name of the skill. Necessary.                        |
| `--reference` | A reference file of the skill, relative to its directory |

With `--json`, it prints `name`, `reference` and `content`.
