# build-a-behavior evals

Each eval gives a fresh Claude Code session one request, such as "add a
behavior that manages /etc/motd", in a scratch copy of this checkout, with this
skill and the osapi CLI available as in any checkout. The session never sees
the checks. When it stops, `run.sh` checks what it left:

| Check                                        | How                                                  |
| -------------------------------------------- | ---------------------------------------------------- |
| The definition breaks no rule                | `go run . behavior validate`                         |
| It is registered under its name              | `go run . behavior get --name <name> --json`         |
| Its own tests pass with 100% coverage        | `go test -cover ./internal/behaviors/<name>/`              |
| Every behavior is valid, mounted and attached | `go test ./internal/behaviors/all/`                       |
| Generated files are current                  | `go generate` changes nothing                        |
| Lint is clean                                | golangci-lint on the behavior's package              |
| It touched nothing it should not             | only its directory, its import and generated files   |
| Per eval: sudo declared, refusals declared   | a grep of the behavior's package                     |
| Per eval: internal, requires, platforms      | `behavior get --json` and whether `gen/` exists      |

An eval of kind `ask` gives a request too vague to build from. It passes when
the session changes nothing and asks the user at least two questions.

An eval of kind `refuse` asks for something the platform cannot express yet. It
passes when the session changes nothing and its answer names what is missing,
which is the skill's "Not yet" list doing its job.

| Eval      | Asks for                                         | Proves                                              |
| --------- | ------------------------------------------------ | --------------------------------------------------- |
| `motd`    | a file's content, read and set                    | the basic path                                      |
| `modules` | kernel modules, loaded with modprobe              | a keyed method, sudo and refusals                   |
| `aptpkg`  | package installs for other behaviors only         | an internal behavior: no routes, commands or `gen/` |
| `issue`   | a banner file osapi keeps a record of             | requiring `managedfile` rather than copying it      |
| `loadavg` | load averages on any Linux                        | the `linux` platform                                |
| `restart` | restarting a service                              | stopping at the "Not yet" list                      |
| `swap`    | "manage swap", and nothing more                   | asking what it should do instead of guessing        |
| `apptoken` | a token written to a file, never shown           | a secret input                                      |
| `neighbors` | `ip neigh`, under the network group             | a grouped behavior, in its group's directory        |

```sh
.claude/skills/build-a-behavior/evals/run.sh          # every eval in evals.json
.claude/skills/build-a-behavior/evals/run.sh motd     # one
KEEP=1 OUT=/tmp/evals BUDGET=10 .claude/skills/build-a-behavior/evals/run.sh
```

Each session's transcript is kept as `<id>.session.json`, with its turns, cost
and final message, and `results.txt` has a PASS or FAIL per check. A failure is
fixed in the skill or the CLI, never in the session's output, and the eval is
run again.
