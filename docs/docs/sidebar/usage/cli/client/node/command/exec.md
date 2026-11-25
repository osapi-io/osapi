# Exec

Run a program with its arguments.

Runs the program with its arguments, each one word, with no shell between them:
a ; or a | in an argument is that character. It answers with standard output and
standard error apart, the exit code and the duration. A program that exits
non-zero is a result, not an error. Only the admin role may run it, and the
agent runs it only where agent.command.enabled is true.

```bash
osapi client node command exec [flags]
```

## Examples

```bash
# Print a word
osapi client node command exec --target web-01 --command echo --args hello

# A program that exits non-zero
osapi client node command exec --target web-01 --command false

# In a directory, for at most 5 seconds
osapi client node command exec --target web-01 --command true --cwd /tmp --timeout 5
```

## Result

Print a word, on one host:

```ansi
$ osapi client node command exec --target web-01 --command echo --args hello
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mSTDOUT  [0m[38;2;10;143;176mSTDERR  [0m[38;2;10;143;176mEXIT CODE  [0m[38;2;10;143;176mDURATION MS[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  hello           0          0

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node command exec --target _all --command echo --args hello
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mSTDOUT  [0m[38;2;10;143;176mSTDERR  [0m[38;2;10;143;176mEXIT CODE  [0m[38;2;10;143;176mDURATION MS[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  hello           0          0
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  hello           0          0

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
      "stdout": "hello\n"
    }
  ]
}
```

A program that exits non-zero, on one host:

```ansi
$ osapi client node command exec --target web-01 --command false
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mSTDOUT  [0m[38;2;10;143;176mSTDERR  [0m[38;2;10;143;176mEXIT CODE  [0m[38;2;10;143;176mDURATION MS[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m                  1          0

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node command exec --target _all --command false
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mSTDOUT  [0m[38;2;10;143;176mSTDERR  [0m[38;2;10;143;176mEXIT CODE  [0m[38;2;10;143;176mDURATION MS[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m                  1          0
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m                  1          0

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
      "exit_code": 1,
      "hostname": "web-01",
      "status": "ok",
      "stderr": "",
      "stdout": ""
    }
  ]
}
```

In a directory, for at most 5 seconds, on one host:

```ansi
$ osapi client node command exec --target web-01 --command true --cwd /tmp --timeout 5
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mSTDOUT  [0m[38;2;10;143;176mSTDERR  [0m[38;2;10;143;176mEXIT CODE  [0m[38;2;10;143;176mDURATION MS[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m                  0          0

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node command exec --target _all --command true --cwd /tmp --timeout 5
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mSTDOUT  [0m[38;2;10;143;176mSTDERR  [0m[38;2;10;143;176mEXIT CODE  [0m[38;2;10;143;176mDURATION MS[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m                  0          0
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m                  0          0

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
      "stdout": ""
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example                               | On the host       |
| ------------------------------------- | ----------------- |
| Print a word                          | runs `echo hello` |
| A program that exits non-zero         | runs `false`      |
| In a directory, for at most 5 seconds | runs `true`       |

The command fails in these cases:

| Call                        | Flags                          | Fails with    |
| --------------------------- | ------------------------------ | ------------- |
| No program                  |                                | required      |
| A program that is an option | `--command -x`                 | startsnotwith |
| A timeout above 300 seconds | `--command true --timeout 301` | max           |
| A negative timeout          | `--command true --timeout -1`  | min           |

## Permission

It needs the `command:execute` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                                              | Default |
| -------------- | ------------------------------------------------------------------------ | ------- |
| `--args`       | Its arguments, each one word, such as -la /tmp                           |         |
| `--command`    | The program, a name on the agent's PATH or a path, such as ls (required) |         |
| `--cwd`        | The directory to run it in; the agent's own when empty                   |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                   | `_all`  |
| `--timeout`    | Seconds to let it run before it is stopped, up to 300; 30 when 0         |         |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
