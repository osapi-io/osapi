# Create

Write a new local group that builds and passes its tests.

It writes `internal/cli/commands/<name>/` in the osapi checkout in the current
directory, and adds its import to `internal/cli/commands/all`. The group has a
list and a get, an empty `Deps` and the tests of its methods. Then run the steps
that it prints. While its `Deps` is empty, the CLI needs no wiring.
`osapi dsl command spec get` lists each part of the DSL.

```bash
osapi dsl command create [flags]
```

## Examples

```bash
osapi dsl command create --name widget --description "The widgets of this host"
```

## Result

The command prints:

```ansi
$ osapi dsl command create --name widget --description "The widgets of this host"
[1;38;2;0;213;255mLocal group:[0m
  [38;2;10;143;176mName:[0m  widget

[1;38;2;255;0;255mFiles (6):[0m
  [38;2;176;48;176mPATH[0m
  internal/cli/commands/widget/widget.go
  internal/cli/commands/widget/deps.go
  internal/cli/commands/widget/methods.go
  internal/cli/commands/widget/methods_public_test.go
  internal/cli/commands/widget/types.go
  internal/cli/commands/all/all.go

[1;38;2;255;209;0mNext (4):[0m
  [38;2;176;143;0mCOMMAND[0m
  go generate ./internal/cli/commands/all
  go test ./internal/cli/commands/widget/...
  go run . dsl command validate
  mise exec -- just cli-docs-generate
```

With `--json`, the command prints:

```json
{
  "name": "widget",
  "files": [
    {
      "path": "internal/cli/commands/widget/widget.go"
    },
    {
      "path": "internal/cli/commands/widget/deps.go"
    },
    {
      "path": "internal/cli/commands/widget/methods.go"
    },
    {
      "path": "internal/cli/commands/widget/methods_public_test.go"
    },
    {
      "path": "internal/cli/commands/widget/types.go"
    },
    {
      "path": "internal/cli/commands/all/all.go"
    }
  ],
  "next": [
    {
      "command": "go generate ./internal/cli/commands/all"
    },
    {
      "command": "go test ./internal/cli/commands/widget/..."
    },
    {
      "command": "go run . dsl command validate"
    },
    {
      "command": "mise exec -- just cli-docs-generate"
    }
  ]
}
```

## Flags

| Flag            | Description                                          | Default |
| --------------- | ---------------------------------------------------- | ------- |
| `--description` | What the group is, in one line (required)            |         |
| `--name`        | The group's name: its command under osapi (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
