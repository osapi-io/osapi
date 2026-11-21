# Get

Show a chain: its tree, and what each call of each example sends and gets.

It runs each example of the behavior on the fake host that `go test` uses, and
records each `behavior.Call`: the behavior, the method, the input, and the
output or the error. The table cuts each input and output; `--json` has them in
full, under `runs`. An input field tagged `secret` shows as `[redacted]`.

```bash
osapi behavior chain get [flags]
```

## Examples

```bash
osapi behavior chain get --name dns
osapi behavior chain get --name dns --json
```

## Flags

| Flag     | Description                                             | Default |
| -------- | ------------------------------------------------------- | ------- |
| `--name` | The behavior, from osapi behavior chain list (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
