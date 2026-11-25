# Shell

Run a command line through /bin/sh -c.

Runs the command line through /bin/sh -c, so pipes, redirects and variables
work. It answers as exec does. Its permission is command:shell, apart from
exec's command:execute: a token that may run a program may not open a shell.
Only the admin role gets it.

```bash
osapi client node command shell [flags]
```

## Examples

```bash
# A pipe
osapi client node command shell --target web-01 --command "echo hello | tr a-z A-Z"

# A command line that exits non-zero
osapi client node command shell --target web-01 --command "exit 3"
```

## Result

A pipe, on one host:

```ansi
$ osapi client node command shell --target web-01 --command "echo hello | tr a-z A-Z"
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mSTDOUT  [0m[38;2;10;143;176mSTDERR  [0m[38;2;10;143;176mEXIT CODE  [0m[38;2;10;143;176mDURATION MS[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  HELLO           0          0

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node command shell --target _all --command "echo hello | tr a-z A-Z"
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mSTDOUT  [0m[38;2;10;143;176mSTDERR  [0m[38;2;10;143;176mEXIT CODE  [0m[38;2;10;143;176mDURATION MS[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  HELLO           0          0
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  HELLO           0          0

  3 hosts: [38;2;103;234;147m2 changed[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "changed": true,
      "duration_ms": 0,
      "exit_code": 0,
      "hostname": "web-01",
      "status": "ok",
      "stderr": "",
      "stdout": "HELLO\n"
    }
  ]
}
```

A command line that exits non-zero, on one host:

```ansi
$ osapi client node command shell --target web-01 --command "exit 3"
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mSTDOUT  [0m[38;2;10;143;176mSTDERR  [0m[38;2;10;143;176mEXIT CODE  [0m[38;2;10;143;176mDURATION MS[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m                  3          0

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node command shell --target _all --command "exit 3"
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mSTDOUT  [0m[38;2;10;143;176mSTDERR  [0m[38;2;10;143;176mEXIT CODE  [0m[38;2;10;143;176mDURATION MS[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m                  3          0
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m                  3          0

  3 hosts: [38;2;103;234;147m2 changed[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "changed": true,
      "duration_ms": 0,
      "exit_code": 3,
      "hostname": "web-01",
      "status": "ok",
      "stderr": "",
      "stdout": ""
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example                            | On the host                                |
| ---------------------------------- | ------------------------------------------ |
| A pipe                             | runs `/bin/sh -c echo hello \| tr a-z A-Z` |
| A command line that exits non-zero | runs `/bin/sh -c exit 3`                   |

The command fails in these cases:

| Call                             | Flags                          | Fails with    |
| -------------------------------- | ------------------------------ | ------------- |
| No command line                  |                                | required      |
| A command line that is an option | `--command -x`                 | startsnotwith |
| A timeout above 300 seconds      | `--command true --timeout 301` | max           |
| A negative timeout               | `--command true --timeout -1`  | min           |

## Permission

It needs the `command:shell` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                                                          | Default |
| -------------- | ------------------------------------------------------------------------------------ | ------- |
| `--command`    | The command line, which /bin/sh -c reads, such as ls -la /tmp \| grep log (required) |         |
| `--cwd`        | The directory to run it in; the agent's own when empty                               |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                               | `_all`  |
| `--timeout`    | Seconds to let it run before it is stopped, up to 300; 30 when 0                     |         |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
