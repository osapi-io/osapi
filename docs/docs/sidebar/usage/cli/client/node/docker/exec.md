# Exec

Run a command in a container.

Runs docker exec, and answers what the command wrote to standard output and to
standard error, and its exit code. A command that exits with a code other than 0
is an answer, not a failure. It changes what the command changes, which osapi
cannot tell, so it reports changed: true. Its permission is docker:exec, which
only the admin role gets, apart from the docker:execute of start and stop. An
agent with agent.command.enabled false refuses it, as it refuses a command.

```bash
osapi client node docker exec [flags]
```

## Examples

```bash
# Run a command
osapi client node docker exec --target web-01 --name osapitest --command echo --command hello
```

## Result

Run a command, on one host:

```ansi
$ osapi client node docker exec --target web-01 --name osapitest --command echo --command hello
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mSTDOUT  [0m[38;2;10;143;176mSTDERR  [0m[38;2;10;143;176mEXIT CODE[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  hello           0

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node docker exec --target _all --name osapitest --command echo --command hello
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mSTDOUT  [0m[38;2;10;143;176mSTDERR  [0m[38;2;10;143;176mEXIT CODE[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  hello           0
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  hello           0

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
      "exit_code": 0,
      "hostname": "web-01",
      "status": "ok",
      "stderr": "",
      "stdout": "hello\n"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example       | On the host                             |
| ------------- | --------------------------------------- |
| Run a command | runs `docker exec osapitest echo hello` |

The command fails in these cases:

| Call                                     | Flags                                              | Fails with   |
| ---------------------------------------- | -------------------------------------------------- | ------------ |
| An empty command                         | `--name osapitest`                                 | min          |
| No command                               | `--name osapitest`                                 | required     |
| A name that is an option                 | `--name -f --command true`                         | container_id |
| A working directory that is not absolute | `--name osapitest --command pwd --working-dir tmp` | startswith   |

## Permission

It needs the `docker:exec` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag            | Description                                                                              | Default |
| --------------- | ---------------------------------------------------------------------------------------- | ------- |
| `--command`     | The command and its arguments (required)                                                 |         |
| `--env`         | Environment variables for the command (secrets: NAME=secret://&lt;vault&gt;/&lt;key&gt;) |         |
| `--name`        | The container's name or ID (required)                                                    |         |
| `-T, --target`  | Target: _all, _any, hostname, or label (group:web.dev)                                   | `_all`  |
| `--working-dir` | The directory the command runs in                                                        |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
