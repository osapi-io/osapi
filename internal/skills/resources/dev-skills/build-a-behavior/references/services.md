# The services a behavior uses

osapi is a platform, and a behavior is a client of it. The agent gives each
provider a `host.Host` with the services that it can use. The platform does the
other work: messages, storage, telemetry, permissions and audit. If a behavior
uses these services as this page shows, the fake host tests it. You do not write
a fake of your own. If a behavior goes around these services, the fake host
cannot test it.

In tests, no real program runs. The filesystem of the fake host is in memory.
Its command runner is a generated mock. The mock records each command line and
gives the output that the example declares. Thus a behavior for Debian that runs
`dpkg-query` and writes `/etc` gets tests on macOS and in CI, where neither
exists. Only CI runs the real commands: its `Integration` step runs the write
examples as root on a throwaway Linux runner.

`osapi dsl behavior spec get` lists each service of the host under "The host a
provider gets", and a test fails when a service has no entry there.
`go doc ./internal/behavior/host Host` tells what each service is.
`dsl behavior get --name <name>` tells what a building block accepts. This page
tells the two things that they do not tell: how the tests replace each service,
and what to copy.

## How the tests fake each service

| A provider uses                                                                                | The fake host gives it                                                           | An example says                                                                                |
| ---------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `h.FS`, never the `os` package                                                                 | An in-memory filesystem                                                          | `Exists` before the call; `Wrote`, `Removed` after                                             |
| `h.Exec`, with each program declared by `behavior.Runs`, or `behavior.Sudo` if it runs as root | A mock that records each command line and answers with what the example declares | `Prints` for an output, `PrintsAfter` for one that changes once a command has run; `Ran` after |
| A gohai collector, with `h.Sources()`, for each read of a fact                                 | The collector, reading the fake filesystem and the mock commands                 | `Exists` and `Prints` for what gohai reads                                                     |
| `h.Objects`, through the `object` building block                                               | The objects that the example stores                                              | `Stored` before the call                                                                       |
| A building block, through `behavior.Requires` and `host.Require`                               | The building block's real provider, on the same fake host                        | Whatever that provider's own calls need; its `dsl behavior get` lists them                     |
| `h.Time()`, never `time.Now`                                                                   | `behaviortest.Now`, 2026-01-02 03:04:05 UTC                                      | An output computed from that time, such as an uptime from a boot time                          |
| `host.Settings[Settings](h)`                                                                   | The defaults                                                                     |                                                                                                |
| `h.Platform`, `h.Hostname`                                                                     | Each platform the behavior declares; `behaviortest.Hostname`                     |                                                                                                |

The platform does these without code in the behavior. Thus do not write code for
them:

- a span and metrics for each method
- messages
- `@fact.<name>` references
- permissions and audit A secret input is tagged `secret:"true"` and reaches a
  program on stdin, never as an argument.

## Where a behavior's state lives

The shape of the provider comes from the location of the truth about what it
manages. Find that location first.

**A file that osapi writes.** Require `managedfile` and write through it.
`dsl behavior get --name managedfile` shows what it accepts, for example the
owner and the mode of the file. It compares before it writes. Thus a second run
reports `changed: false`. It also keeps the record that the controller uses to
report drift. sysctl is this case. Copy `internal/behaviors/sysctl`.

**A file that something else owns.** Read and write `h.FS`. Compare before you
write. An example uses `Exists` for the file as it was and `Wrote` for the
result.

**What a program changes.** A write runs the program with `h.Exec`: a package
manager, `systemctl`, `ip`. Read the state through gohai, as the next case says,
and query before you act. Act only when the query shows that the host is
different. An example gives the answer of the query with `Prints`. It gives the
answer after the change with `PrintsAfter`. Thus the second run sees the change
and reports `changed: false`.

**A fact about the host, read only.** Call the gohai collector that has it, with
`h.Sources()`: `memory.NewWith(h.Sources()).Collect(ctx, nil)`. gohai reads the
host through the host's filesystem, commands, clock and resolver, so the fake
host answers it. Declare each program gohai runs with `behavior.Runs`, and give
the examples gohai's files with `Exists` and its commands with `Prints`. Do not
parse `/proc` or `sysctl` in the behavior. When gohai does not have the read,
add it to gohai first. When an answer is the platform's own, such as the name of
the operating system, limit each example with `behavior.OnPlatform`.

The tests enforce this. A read that runs a command outside gohai fails the
`gohai` check. The service behavior reads with `services.NewWith(h.Sources())`
and writes with `systemctl`. Only a read of what is not a fact of the host runs
a command itself, and it says why with `.NotAFact("...")`: the updates of the
package index are what the repositories have, and the definition that netplan
has of an interface is the configuration of netplan.

## Input that reaches the host

Each string in a request is dangerous until the provider checks it. The probes
check three ways that a string can become an injection:

- **A path.** A field that becomes part of a path gets a `validate` pattern that
  cannot go out of its directory, for example `^[a-zA-Z0-9][a-zA-Z0-9._-]*$`,
  and the provider checks it again.
- **A line in a file.** A field that goes into a line of a file refuses `\r` and
  `\n`. If it does not, one value becomes two settings.
- **A program's argument.** Pass `--` before a value that can start with `-`, or
  refuse a value that starts with `-`. Do not build a `sh -c` script from input.
  Run the program with its arguments.

Declare each refusal with `.Fails`. Then the CLI page of the method lists it,
and the tests check it.
