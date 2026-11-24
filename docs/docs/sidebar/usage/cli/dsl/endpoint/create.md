# Create

Write a new endpoint group that builds and passes its tests.

It writes `internal/controller/endpoints/<name>/` in the osapi checkout in the
current directory, and adds its import to `internal/controller/endpoints/all`.
The group has a list and a get, an empty `Deps` and the tests of its methods.
Then run the steps that it prints. While its `Deps` is empty, the controller
needs no wiring. `osapi dsl endpoint spec get` lists each part of the DSL.

```bash
osapi dsl endpoint create [flags]
```

## Examples

```bash
osapi dsl endpoint create --name trace --description "The traces of the controller"
```

## Result

The command prints:

```ansi
$ osapi dsl endpoint create --name trace --description "The traces of the controller"
[1;38;2;0;213;255mEndpoint group:[0m
  [38;2;10;143;176mName:[0m  trace

[1;38;2;255;0;255mFiles (5):[0m
  [38;2;176;48;176mPATH[0m
  internal/controller/endpoints/trace/trace.go
  internal/controller/endpoints/trace/deps.go
  internal/controller/endpoints/trace/methods.go
  internal/controller/endpoints/trace/methods_public_test.go
  internal/controller/endpoints/trace/types.go

[1;38;2;255;209;0mNext (4):[0m
  [38;2;176;143;0mCOMMAND[0m
  go generate ./internal/controller/endpoints/all
  mise exec -- just generate
  go test ./internal/controller/endpoints/trace/...
  go run . dsl endpoint validate
```

With `--json`, the command prints:

```json
{
  "name": "trace",
  "files": [
    {
      "path": "internal/controller/endpoints/trace/trace.go"
    },
    {
      "path": "internal/controller/endpoints/trace/deps.go"
    },
    {
      "path": "internal/controller/endpoints/trace/methods.go"
    },
    {
      "path": "internal/controller/endpoints/trace/methods_public_test.go"
    },
    {
      "path": "internal/controller/endpoints/trace/types.go"
    }
  ],
  "next": [
    {
      "command": "go generate ./internal/controller/endpoints/all"
    },
    {
      "command": "mise exec -- just generate"
    },
    {
      "command": "go test ./internal/controller/endpoints/trace/..."
    },
    {
      "command": "go run . dsl endpoint validate"
    }
  ]
}
```

## Flags

| Flag            | Description                                                 | Default |
| --------------- | ----------------------------------------------------------- | ------- |
| `--description` | What the group is, in one line (required)                   |         |
| `--name`        | The group's name: its path segment and CLI group (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
