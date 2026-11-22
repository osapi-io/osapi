# build-an-endpoint evals

Each eval gives one request to a new agent session (`run.sh` starts
`claude -p`). The session runs in a scratch copy of this checkout, with this
skill and the osapi CLI, as in each checkout. The session does not see the
checks. When it stops, `run.sh` checks what it left:

| Check                                       | How                                                         |
| ------------------------------------------- | ----------------------------------------------------------- |
| The definition breaks no rule               | `go run . controller endpoint validate`                     |
| It is registered under its name             | `go run . controller endpoint get --name <group> --json`    |
| Its own tests pass with 100% coverage       | `go test -cover ./internal/controller/endpoints/<group>/`   |
| Every group is valid and mounted            | `go test ./internal/controller/endpoints/all/`              |
| Generated files are current                 | `just generate` and `just cli-docs-generate` change nothing |
| Lint is clean                               | golangci-lint on the group's package                        |
| It touched nothing it should not            | only its directory, its registration, its wiring and generated files |
| Per eval: refusals declared, the permission | a grep of the group's package, and `get --json`            |

An eval of kind `ask` gives a request that is not clear enough to build. It
passes when the session changes nothing and asks the user two or more questions.

An eval of kind `refuse` asks for work that is not an endpoint. It passes when
the session changes nothing and its answer names what the work is instead. This
shows that the skill's "endpoint or behavior" decision works.

| Eval         | Asks for                                       | Proves                                       |
| ------------ | ---------------------------------------------- | -------------------------------------------- |
| `buildinfo`  | the controller's version and commit, one GET   | the basic path                               |
| `notes`      | list, create and delete of notes in a store    | Deps with a mock, 201 and 204, a permission  |
| `reports`    | "an endpoint for reports", and nothing more    | asking what it should do instead of guessing |
| `hostkernel` | the kernel version of each host                | work on a host is a behavior                 |

```sh
internal/skills/resources/dev-skills/build-an-endpoint/evals/run.sh          # every eval in evals.json
internal/skills/resources/dev-skills/build-an-endpoint/evals/run.sh notes    # one
KEEP=1 OUT=/tmp/evals BUDGET=10 internal/skills/resources/dev-skills/build-an-endpoint/evals/run.sh
```

`run.sh` keeps the transcript of each session as `<id>.session.json`, with its
turns, cost and last message. `results.txt` has a PASS or FAIL for each check.
Fix a failure in the skill or the CLI, not in the output of the session. Then
run the eval again.
