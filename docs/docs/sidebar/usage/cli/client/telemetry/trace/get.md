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

## Result

The command prints:

```ansi
$ osapi client telemetry trace get --id 26e448c9dbe1c4a8d216a98c675079c9
[1;38;2;0;213;255mTrace:[0m
  [38;2;10;143;176mID:[0m                   26e448c9dbe1c4a8d216a98c675079c9
  [38;2;10;143;176mStarted:[0m              2026-01-15 09:58:12
  [38;2;10;143;176mDuration:[0m             186ms
  [38;2;10;143;176mServices:[0m             osapi-controller, osapi-agent
  [38;2;10;143;176mNATS calls left out:[0m  2
  [38;2;10;143;176mNATS call time:[0m       4ms

[1;38;2;255;0;255mSpans (1):[0m
  [38;2;176;48;176mSPAN                                   [0m[38;2;176;48;176mON                [0m[38;2;176;48;176mDURATION  [0m[38;2;176;48;176mSTATUS[0m
  PUT /api/node/{hostname}/sysctl/{key}  [1mosapi-controller[0m  186ms     [38;2;103;234;147mok[0m
  [38;2;138;138;138m└─ [0mbehavior.sysctl.update              [1mweb-01[0m            171ms     [38;2;103;234;147mok[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "trace_id": "26e448c9dbe1c4a8d216a98c675079c9",
  "started": "2026-01-15T09:58:12Z",
  "duration": "186ms",
  "services": ["osapi-controller", "osapi-agent"],
  "nats_calls": 2,
  "nats_time": "4ms",
  "spans": [
    {
      "name": "PUT /api/node/{hostname}/sysctl/{key}",
      "on": "osapi-controller",
      "duration": "186ms",
      "status": "ok",
      "span": {
        "span_id": "a1b2c3d4e5f60718",
        "name": "PUT /api/node/{hostname}/sysctl/{key}",
        "service": "osapi-controller",
        "kind": "server",
        "start_time": "2026-01-15T09:58:12Z",
        "duration_ms": 186,
        "status": "ok",
        "attributes": {
          "http.response.status_code": "200"
        }
      },
      "children": [
        {
          "name": "behavior.sysctl.update",
          "on": "web-01",
          "duration": "171ms",
          "status": "ok",
          "span": {
            "span_id": "b2c3d4e5f6071829",
            "parent_span_id": "a1b2c3d4e5f60718",
            "name": "behavior.sysctl.update",
            "service": "osapi-agent",
            "kind": "internal",
            "start_time": "2026-01-15T09:58:12.009Z",
            "duration_ms": 171,
            "status": "ok",
            "attributes": {
              "osapi.host": "web-01"
            }
          }
        }
      ]
    }
  ]
}
```

## Permission

It needs the `trace:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag    | Description                                                                                 | Default |
| ------- | ------------------------------------------------------------------------------------------- | ------- |
| `--all` | Show the NATS calls too, the reads and writes of the job queue                              |         |
| `--id`  | The trace ID, which a command prints after its job ID, 32 hexadecimal characters (required) |         |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
