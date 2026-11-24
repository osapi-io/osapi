# Create

Write a new behavior's standard layout in an osapi checkout.

Before you change anything, the result builds, validates and passes every test.
Then change the types, the provider and the examples to what the behavior
manages. Writes nothing when the name is taken. A behavior that belongs to a
group, such as `network`, is nested under it: its routes are
`/api/node/{hostname}/network/<name>`, its commands
`osapi client node network <name>`, and its directory
`internal/behaviors/network/<name>/`. \[Your first
behavior](../../../../development/your-first-behavior.md) walks through the
rest.

A new behavior is one an operator calls, with an API and commands. Work that two
behaviors share, such as writing a file, becomes an internal behavior that the
two require: add `behavior.Internal` to it then, and `behavior.Requires` to each
user. `osapi dsl behavior validate` fails an internal behavior that no behavior
requires.

`--requires` names the behaviors that the new one is built on. For each, it
writes `behavior.Requires`, the `host.Require` in the builder and a field of the
provider. You write each `behavior.Call`: until an example calls a required
behavior, the tests fail the rule `used`.

```bash
osapi dsl behavior create [flags]
```

## Examples

```bash
osapi dsl behavior create --name motd --description "The message shown at login"
osapi dsl behavior create --name motd --description "The message shown at login" --root ~/src/osapi
osapi dsl behavior create --group network --name neighbors --description "The host's ARP neighbors"
osapi dsl behavior create --name ntpconf --description "The NTP config" --requires managedfile
```

## Result

The command prints:

```ansi
$ osapi dsl behavior create --name motd --description "The message shown at login"
[1;38;2;0;213;255mBehavior:[0m
  [38;2;10;143;176mName:[0m  motd

[1;38;2;255;0;255mFiles (6):[0m
  [38;2;176;48;176mPATH[0m
  internal/behaviors/motd/behavior.go
  internal/behaviors/motd/behavior_public_test.go
  internal/behaviors/motd/debian.go
  internal/behaviors/motd/methods.go
  internal/behaviors/motd/provider.go
  internal/behaviors/all/all.go

[1;38;2;255;209;0mNext (5):[0m
  [38;2;176;143;0mSTEP[0m
  go generate ./internal/behaviors/motd/mocks ./internal/behaviors/all
  change the types, the provider and the examples in internal/behaviors/motd
  go generate ./internal/behaviors/all
  osapi dsl behavior validate
  just test
```

With `--json`, the command prints:

```json
{
  "name": "motd",
  "files": [
    {
      "path": "internal/behaviors/motd/behavior.go"
    },
    {
      "path": "internal/behaviors/motd/behavior_public_test.go"
    },
    {
      "path": "internal/behaviors/motd/debian.go"
    },
    {
      "path": "internal/behaviors/motd/methods.go"
    },
    {
      "path": "internal/behaviors/motd/provider.go"
    },
    {
      "path": "internal/behaviors/all/all.go"
    }
  ],
  "next": [
    {
      "step": "go generate ./internal/behaviors/motd/mocks ./internal/behaviors/all"
    },
    {
      "step": "change the types, the provider and the examples in internal/behaviors/motd"
    },
    {
      "step": "go generate ./internal/behaviors/all"
    },
    {
      "step": "osapi dsl behavior validate"
    },
    {
      "step": "just test"
    }
  ]
}
```

## Flags

| Flag            | Description                                                                                                          | Default |
| --------------- | -------------------------------------------------------------------------------------------------------------------- | ------- |
| `--description` | What it manages, in one line (required)                                                                              |         |
| `--group`       | Nest it under a group, such as network: its routes, commands and directory                                           |         |
| `--name`        | The behavior's name: lowercase letters and digits (required)                                                         |         |
| `--package`     | The Go package and directory, when it is not the name: package is a Go keyword, so --name package --package packages |         |
| `--requires`    | Behaviors it is built on, such as managedfile: it writes the Requires, the host.Require and a provider field of each |         |
| `--root`        | The osapi checkout to write into                                                                                     | `.`     |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
