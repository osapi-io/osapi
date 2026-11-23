---
sidebar_position: 8
---

# Developing a skill

A skill tells an agent how to do one kind of work with osapi. The osapi binary
embeds each skill. The binary is the only source of a skill.

## Where the skills are

The skills are in `internal/skills/resources/`. Each skill is one directory:

```
internal/skills/resources/
  skills/                  the skills that osapi init installs
    osapi/
    osapi-operations/
  dev-skills/              the skills for a person who changes osapi
    build-a-behavior/
      SKILL.md             the frontmatter and the procedure
      references/          files that the procedure sends the agent to
      evals/               the requests that measure the skill, and run.sh
    build-a-command/
    build-a-dsl/           what the three DSLs share, and how to add a DSL
    build-an-endpoint/
```

`internal/skills` embeds all of `resources/`. It also has the code that installs
the skills (`osapi init`) and serves them (`osapi skill`).

- `osapi init` writes each skill in `skills/` into a repository.
- `osapi skill list` lists each skill. `osapi skill get --name <skill>` prints
  its `SKILL.md`, and `--reference <file>` prints one of its other files.

A skill in `dev-skills/` is only for a checkout of osapi. `osapi init` does not
install it. The checkout has a short pointer at
`.claude/skills/<skill>/SKILL.md`. The pointer has the frontmatter of the skill,
so that agents find it. Its body tells the agent to run
`go run . skill get --name <skill>` and obey the output.

## When a skill changes

The embedded skills are how a person and an agent learn osapi. Thus a change to
osapi that adds or changes a command, a flag, a DSL option or a check updates
each skill that tells about it, in the same pull request. That includes the
skills that `osapi init` installs, not only the skills in `dev-skills/`. Do not
wait until a feature is complete. Write in the skill what is not built yet, so
that an agent stops there and does not build around it.

## Change a skill

1. Edit the files under `internal/skills/resources/`. Do not edit the pointer in
   `.claude/skills/`, except when the frontmatter of the skill changes.
2. Keep `SKILL.md` a router. Put the details in `references/`. The agent loads a
   reference only when a step sends it there.
3. Write in Simplified Technical English, as each page of osapi-io is written.
4. Run the tests of the package:

   ```bash
   mise exec -- go test -cover ./internal/skills/
   ```

   A test fails on a `SKILL.md` that has no `name` and `description`
   frontmatter, or whose `name` is not the name of its directory.

## Add a skill

1. Make a directory in `skills/` or in `dev-skills/`. Its name is the name of
   the skill.
2. Write `SKILL.md`. It starts with YAML frontmatter that has `name` and
   `description`. The tool of the agent uses the description to decide when to
   load the skill.
3. For a skill in `dev-skills/`, add the pointer in `.claude/skills/<skill>/`.
4. Run the tests of the package.

Do not run mdformat on `internal/skills/resources/`. `just md-fmt` excludes it.
mdformat changes the frontmatter into a heading, and then agents do not find the
skill.

## Measure a skill

The evals of a skill are in its `evals/` directory. Each eval is a request that
`run.sh` gives to a new agent session, in a scratch copy of the checkout.
Commands check the result. Nobody reads the result to check it.

```bash
internal/skills/resources/dev-skills/build-a-behavior/evals/run.sh        # each eval
internal/skills/resources/dev-skills/build-a-behavior/evals/run.sh motd   # one
```

When an eval fails, fix the skill or the CLI. Do not fix the output of the
session. Then run the eval again. The `README.md` of the evals lists each eval
and what it proves.
