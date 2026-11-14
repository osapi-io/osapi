# The services a behavior uses

osapi is a platform, and a behavior is a client of it. The agent hands every
provider a `host.Host` with the services it may use, and the platform does the
rest: messaging, storage, telemetry, permissions, audit. A behavior that uses
these services the way this page shows gets tested by the fake host without
writing a fake of its own, and a behavior that reaches around them cannot be.

In tests nothing real runs. The fake host's filesystem is in memory and its
command runner is a generated mock that records each command line and answers
with what the example declares, so a behavior for Debian that runs `dpkg-query`
and writes `/etc` is tested on macOS and in CI, where neither exists. Only
`just behavior-int-linux` runs the real commands, in a Debian container.

`go doc ./internal/behavior/host Host` says what each service is, and
`behavior get --name <name>` what a building block accepts. This page adds what
neither says: how the tests stand in for each service, and what to copy.

## How the tests fake each service

| A provider uses | The fake host gives it | An example says |
| --- | --- | --- |
| `h.FS`, never the `os` package | An in-memory filesystem | `Exists` before the call; `Wrote`, `Removed` after |
| `h.Exec`, with each program declared by `behavior.Runs`, or `behavior.Sudo` if it runs as root | A mock that records each command line and answers with what the example declares | `Prints` for an output, `PrintsAfter` for one that changes once a command has run; `Ran` after |
| A building block, through `behavior.Requires` and `host.Require` | The building block's real provider, on the same fake host | Whatever that provider's own calls need; its `behavior get` lists them |
| `host.Settings[Settings](h)` | The defaults | |
| `h.Platform`, `h.Hostname` | Each platform the behavior declares; `behaviortest.Hostname` | |

The platform does these without a line in the behavior, so write none: a span
and metrics per method, messaging, `@fact.<name>` references, permissions and
audit. A secret input is tagged `secret:"true"` and reaches a program on stdin,
never as an argument.

## Where a behavior's state lives

The provider's shape follows from where the truth about what it manages is. Find
that first.

**A file osapi writes.** Require `managedfile` and write through it; `behavior
get --name managedfile` shows what it accepts, such as the file's owner and
mode. It compares before writing, so a second run reports `changed: false`, and
keeps the record the controller reports drift from. sysctl is this case, and
`internal/behaviors/sysctl` the one to copy.

**A file something else owns, read or edited.** Read and write `h.FS`. Compare
before writing. An example uses `Exists` for the file as it was and `Wrote` for
the result.

**What a program reports.** Run it with `h.Exec` and parse its output: a package
manager, `systemctl`, `ip`. Query before acting, and act only when the query
says the host differs. An example gives the query's answer with `Prints`, and
its answer after the change with `PrintsAfter`, so the second run sees the
change and reports `changed: false`.

**The kernel's view, read only.** Read `/proc` or `/sys` through `h.FS`, declare
the `linux` platform, and give examples their files with `Exists`.

## Input that reaches the host

Every string a request carries is hostile until checked, and the probes check
three ways it turns into an injection:

- **A path.** A field that becomes part of a path gets a `validate` pattern that
  cannot climb out, such as `^[a-zA-Z0-9][a-zA-Z0-9._-]*$`, and the provider
  checks it again.
- **A line in a file.** A field written into a line of a file refuses `\r` and
  `\n`, or one value becomes two settings.
- **A program's argument.** Pass `--` before a value that could start with `-`,
  or refuse a leading `-`. Never build a `sh -c` script from input; run the
  program with its arguments.

Each refusal is declared with `.Fails`, so the docs page lists it and the tests
check it.
