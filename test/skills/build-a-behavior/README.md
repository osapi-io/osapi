# build-a-behavior evals

Each eval gives one request to a new agent session (`run.sh` starts
`claude -p`). An example request is "add a behavior that manages /etc/motd". The
session runs in a scratch copy of this checkout, with this skill and the osapi
CLI, as in each checkout. The session does not see the checks. When it stops,
`run.sh` checks what it left:

| Check                                         | How                                                |
| --------------------------------------------- | -------------------------------------------------- |
| The definition breaks no rule                 | `go run . behavior validate`                       |
| It is registered under its name               | `go run . behavior get --name <name> --json`       |
| Its own tests pass with 100% coverage         | `go test -cover ./internal/behaviors/<name>/`      |
| Every behavior is valid, mounted and attached | `go test ./internal/behaviors/all/`                |
| Generated files are current                   | `go generate` changes nothing                      |
| Lint is clean                                 | golangci-lint on the behavior's package            |
| It touched nothing it should not              | only its directory, its import and generated files |
| Per eval: sudo declared, refusals declared    | a grep of the behavior's package                   |
| Per eval: internal, requires, platforms       | `behavior get --json` and whether `gen/` exists    |

An eval of kind `ask` gives a request that is not clear enough to build. It
passes when the session changes nothing and asks the user two or more questions.

An eval of kind `refuse` asks for something that the platform cannot express
yet. It passes when the session changes nothing and its answer names what is
missing. This shows that the "Not yet" list of the skill works.

| Eval        | Asks for                                  | Proves                                              |
| ----------- | ----------------------------------------- | --------------------------------------------------- |
| `motd`      | a file's content, read and set            | the basic path                                      |
| `modules`   | kernel modules, loaded with modprobe      | a keyed method, sudo and refusals                   |
| `aptpkg`    | package installs for other behaviors only | an internal behavior: no routes, commands or `gen/` |
| `issue`     | a banner file osapi keeps a record of     | requiring `managedfile` rather than copying it      |
| `loadavg`   | load averages on any Linux                | the `linux` platform                                |
| `restart`   | restarting a service                      | stopping at the "Not yet" list                      |
| `swap`      | "manage swap", and nothing more           | asking what it should do instead of guessing        |
| `apptoken`  | a token written to a file, never shown    | a secret input                                      |
| `neighbors` | `ip neigh`, under the network group       | a grouped behavior, in its group's directory        |

```sh
test/skills/build-a-behavior/run.sh          # every eval in evals.json
test/skills/build-a-behavior/run.sh motd     # one
KEEP=1 OUT=/tmp/evals BUDGET=10 test/skills/build-a-behavior/run.sh
```

`run.sh` keeps the transcript of each session as `<id>.session.json`, with its
turns, cost and last message. `results.txt` has a PASS or FAIL for each check.
Fix a failure in the skill or the CLI, not in the output of the session. Then
run the eval again.
