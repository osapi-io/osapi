# Get

Show the span tree of a trace.

Shows each span of a trace as a tree: the request on the controller, the job,
each agent that ran it with its host and status, and the commands and file
changes under each. The NATS calls, which record the job's progress, are left
out and counted; --all shows them. Each command prints the ID of its trace after
its job ID.

The controller reads the trace from the tracing backend at
`telemetry.tracing.url`, so this host needs no access to the backend. The
backend gets the spans a few seconds after the request: a trace that is not
found yet is there after a few seconds.

```bash
osapi client telemetry trace get [flags]
```

## Examples

```bash
osapi client telemetry trace get --id 26e448c9dbe1c4a8d216a98c675079c9
osapi client telemetry trace get --id 26e448c9dbe1c4a8d216a98c675079c9 --all
osapi client telemetry trace get --id 26e448c9dbe1c4a8d216a98c675079c9 --json
```

## Permission

It needs the `trace:read` permission. `osapi token roles list` lists the roles
that grant it.

## Flags

| Flag    | Description                                                                                 | Default |
| ------- | ------------------------------------------------------------------------------------------- | ------- |
| `--all` | Show the NATS calls too, the reads and writes of the job queue                              |         |
| `--id`  | The trace ID, which a command prints after its job ID, 32 hexadecimal characters (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
