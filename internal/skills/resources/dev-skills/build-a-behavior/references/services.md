# The services a behavior uses

osapi is a platform, and a behavior is a client of it. The agent gives each provider a `host.Host` with the services that it can use. The platform does the other work: messages, storage, telemetry, permissions and audit. If a behavior uses these services as this page shows, the fake host tests it. You do not write a fake of your own. If a behavior goes around these services, the fake host cannot test it.

In tests, no real program runs. The filesystem of the fake host is in memory. Its command runner is a generated mock. The mock records each command line and gives the output that the example declares. Thus a behavior for Debian that runs `dpkg-query` and writes `/etc` gets tests on macOS and in CI, where neither exists. Only
`just behavior-int-linux` runs the real commands, in a Debian container.

`go doc ./internal/behavior/host Host` tells what each service is. `behavior get --name <name>` tells what a building block accepts. This page tells the two things that they do not tell: how the tests replace each service, and what to copy.

## How the tests fake each service

| A provider uses | The fake host gives it | An example says |
| --- | --- | --- |
| `h.FS`, never the `os` package | An in-memory filesystem | `Exists` before the call; `Wrote`, `Removed` after |
| `h.Exec`, with each program declared by `behavior.Runs`, or `behavior.Sudo` if it runs as root | A mock that records each command line and answers with what the example declares | `Prints` for an output, `PrintsAfter` for one that changes once a command has run; `Ran` after |
| A building block, through `behavior.Requires` and `host.Require` | The building block's real provider, on the same fake host | Whatever that provider's own calls need; its `behavior get` lists them |
| `host.Settings[Settings](h)` | The defaults | |
| `h.Platform`, `h.Hostname` | Each platform the behavior declares; `behaviortest.Hostname` | |

The platform does these without code in the behavior. Thus do not write code for them:

- a span and metrics for each method
- messages
- `@fact.<name>` references
- permissions and audit A secret input is tagged `secret:"true"` and reaches a program on stdin,
never as an argument.

## Where a behavior's state lives

The shape of the provider comes from the location of the truth about what it manages. Find that location first.

**A file that osapi writes.** Require `managedfile` and write through it. `behavior get --name managedfile` shows what it accepts, for example the owner and the mode of the file. It compares before it writes. Thus a second run reports `changed: false`. It also keeps the record that the controller uses to report drift. sysctl is this case. Copy `internal/behaviors/sysctl`.

**A file that something else owns.** Read and write `h.FS`. Compare before you write. An example uses `Exists` for the file as it was and `Wrote` for
the result.

**What a program reports.** Run it with `h.Exec` and parse its output: a package
manager, `systemctl`, `ip`. Query before you act. Act only when the query shows that the host is different. An example gives the answer of the query with `Prints`. It gives the answer after the change with `PrintsAfter`. Thus the second run sees the change and reports `changed: false`.

**The kernel's view, read only.** Read `/proc` or `/sys` through `h.FS`, declare
the `linux` platform, and give examples their files with `Exists`.

## Input that reaches the host

Each string in a request is dangerous until the provider checks it. The probes check three ways that a string can become an injection:

- **A path.** A field that becomes part of a path gets a `validate` pattern that cannot go out of its directory, for example `^[a-zA-Z0-9][a-zA-Z0-9._-]*$`, and the provider
  checks it again.
- **A line in a file.** A field that goes into a line of a file refuses `\r` and `\n`. If it does not, one value becomes two settings.
- **A program's argument.** Pass `--` before a value that can start with `-`, or refuse a value that starts with `-`. Do not build a `sh -c` script from input. Run the program with its arguments.

Declare each refusal with `.Fails`. Then the CLI page of the method lists it, and the tests check it.
