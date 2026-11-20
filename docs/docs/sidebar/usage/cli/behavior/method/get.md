# Get

Show one method of a behavior: its fields, constraints, output and examples.

Each input field is a flag of its command and a field of its request. `--json`
has the JSON Schema of the input and the output, with each constraint that osapi
checks before it queues the job.

```bash
osapi behavior method get [flags]
```

## Examples

```bash
osapi behavior method get --name sysctl --method update
osapi behavior method get --name sysctl --method update --json
```

## Flags

| Flag       | Description                                       | Default |
| ---------- | ------------------------------------------------- | ------- |
| `--method` | The method, from osapi behavior get (required)    |         |
| `--name`   | The behavior, from osapi behavior list (required) |         |

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
