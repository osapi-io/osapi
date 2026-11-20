# Create

Write a new endpoint group that builds and passes its tests.

It writes `internal/controller/endpoints/<name>/` in the osapi checkout in the
current directory, and adds its import to `internal/controller/endpoints/all`.
The group has a list and a get, an empty `Deps` and the tests of its methods.
Then run the steps that it prints. While its `Deps` is empty, the controller
needs no wiring. `osapi controller endpoint spec get` lists each part of the
DSL.

```bash
osapi controller endpoint create [flags]
```

## Examples

```bash
osapi controller endpoint create --name trace --description "The traces of the controller"
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
