# build-a-command evals

Each eval gives one request to a new agent session (`run.sh` starts
`claude -p`). The session runs in a scratch copy of this checkout, with this
skill and the osapi CLI, as in each checkout. The session does not see the
checks. When it stops, `run.sh` checks what it left:

| Check                                 | How                                                    |
| ------------------------------------- | ------------------------------------------------------ |
| The definition breaks no rule         | `go run . cli command validate`                        |
| It is registered under its name       | `go run . cli command get --name <group> --json`       |
| Its own tests pass with 100% coverage | `go test -cover ./internal/cli/commands/<group>/`      |
| Every local group is valid            | `go test ./internal/cli/commands/all/`                 |
| Generated files are current           | `go generate` and `just cli-docs-generate` change nothing |
| Lint is clean                         | golangci-lint on the group's package                   |
| It touched nothing it should not      | only its directory, its registration, its wiring and generated files |
| Per eval: refusals declared           | a grep of the group's package                          |

An eval of kind `ask` gives a request that is not clear enough to build. It
passes when the session changes nothing and asks the user two or more questions.

An eval of kind `refuse` asks for work that is not a local command. It passes
when the session changes nothing and its answer names what the work is instead.

| Eval       | Asks for                                        | Proves                                       |
| ---------- | ----------------------------------------------- | -------------------------------------------- |
| `greet`    | a greeting for a required name                  | the basic path, with a refusal               |
| `checksum` | the SHA-256 of a file, with the file system as a dependency | Deps that a test replaces           |
| `logs`     | "a command for the logs", and nothing more      | asking what it should do instead of guessing |
| `jobcount` | a count of failed jobs from the controller      | a command that calls the API is an endpoint  |

```sh
internal/skills/resources/dev-skills/build-a-command/evals/run.sh          # every eval in evals.json
internal/skills/resources/dev-skills/build-a-command/evals/run.sh greet    # one
KEEP=1 OUT=/tmp/evals BUDGET=10 internal/skills/resources/dev-skills/build-a-command/evals/run.sh
```

`run.sh` keeps the transcript of each session as `<id>.session.json`, with its
turns, cost and last message. `results.txt` has a PASS or FAIL for each check.
Fix a failure in the skill or the CLI, not in the output of the session. Then
run the eval again.
