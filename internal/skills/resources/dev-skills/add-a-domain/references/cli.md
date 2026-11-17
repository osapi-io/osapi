# CLI commands

Anything the API can do has a CLI equivalent. The CLI is thin: parse flags, call
the SDK, print.

## The rules are in the corpus, not here

The obligations this layer carries are stated in
[005-building-a-domain](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md).

| What you need to know | Where |
| --- | --- |
| One parent command per domain, one subcommand per endpoint | [one command per endpoint](https://github.com/osapi-io/specs/blob/main/components/osapi/cli.md#one-command-per-endpoint) |
| Flags rather than positional arguments for identifiers | [values arrive as flags](https://github.com/osapi-io/specs/blob/main/components/osapi/cli.md#values-arrive-as-flags-never-as-positional-arguments) |
| What `--target` accepts, and why no domain redeclares it | [two inherited flags](https://github.com/osapi-io/specs/blob/main/components/osapi/cli.md#two-flags-every-command-inherits) |
| `--json` returning raw bytes before anything is formatted | [json returns first](https://github.com/osapi-io/specs/blob/main/components/osapi/cli.md#json-returns-before-anything-is-formatted) |
| Verbs from `list`, `get`, `create`, `update`, `delete`, `status`; `--id`, `--name`, `--key` for identifiers; no synonyms | [command and flag names](https://github.com/osapi-io/specs/blob/main/components/osapi/cli.md#command-and-flag-names) |
| `cli.PrintHeading` over every block, `cli.PrintKV` for labelled values, `cli.PrintJobID` for the job, `cli.PrintCompactTable` for rows with `Empty` set | [shared helpers](https://github.com/osapi-io/specs/blob/main/components/osapi/cli.md#rendering-is-shared-helpers-not-per-command-formatting) |
| No colors or `fmt` printing in `cmd/`: style comes from `internal/cli/theme.go`; one layout for one host or forty | [output style](https://github.com/osapi-io/specs/blob/main/components/osapi/cli.md#output-style) |
| Every response code the spec declares handled the same way | [one error handler](https://github.com/osapi-io/specs/blob/main/components/osapi/cli.md#errors-go-through-one-handler) |
| A domain appears everywhere an existing domain appears | [every layer, or not done](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#a-domain-is-in-every-layer-or-it-is-not-done) |

## Files

```
cmd/client_node_{domain}.go             parent command, registered under clientNodeCmd
cmd/client_node_{domain}_{operation}.go one per endpoint
```

Controller-only domains drop the `node` segment: `cmd/client_{domain}_*.go`. One
file per command, named for the command it implements.

## Scaffolding

```go
switch resp.StatusCode() {
case http.StatusOK:
    // print
case http.StatusBadRequest, http.StatusNotFound,
     http.StatusInternalServerError:
    handleUnknownError(...)
case http.StatusUnauthorized, http.StatusForbidden:
    handleAuthError(...)
}
```

A code the spec declares and the switch omits becomes a silent success.

A result row has the hostname. When the operation changes something, it also
tells if it changed. Use the columns of the reference domain. Do not invent new
ones, because operators compare these outputs.

## Three rules the corpus does not yet hold

Stated here because they are real and nothing else states them, not the corpus,
not osapi's `CONTRIBUTING.md`. Recorded as unstated rather than left to be
discovered.

**A command whose remote work failed exits with a value that is not zero.** For
`command exec` and `command shell`, the exit code is the code of the remote
command, through `cli.MaxExitCode(results)`. Apply it on each output path, also
`--json` and the default table. Scripts use it to decide.

**`cmd/` is the only place that may call `os.Exit`**, and only at the top of a
command. Below that, return an error. Nothing panics.

**Assert the exit code, not just the output.** The exit-code rule above was
broken once and invisible to tests that read only stdout.

## Tests

`cmd/` is excluded from the coverage gate by `.coverignore`, so unit tests are
not
required there, which makes the integration suite the check:

```
test/integration/{domain}_test.go
```

It has the build tag `integration`. It starts a real binary and runs the
commands from start to end. Guard each test that changes things with
`skipWrite(s.T())`. Thus CI runs only reads by default, and
`OSAPI_INTEGRATION_WRITES=1` enables the others.

```bash
mise exec -- just go-unit-int
```
