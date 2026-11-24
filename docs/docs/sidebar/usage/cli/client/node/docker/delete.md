# Delete

Remove a container.

Runs docker rm -f, which also removes a container that runs: a delete leaves no
container of the name. A container that the host does not have reports changed:
false.

```bash
osapi client node docker delete [flags]
```

## Examples

```bash
# Remove a container
osapi client node docker delete --target web-01 --name osapitest

# A container that the host does not have
osapi client node docker delete --target web-01 --name osapitest
```

## Result

Remove a container, on one host:

```ansi
$ osapi client node docker delete --target web-01 --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mID[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m  [38;2;138;138;138m[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node docker delete --target _all --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mID[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m  [38;2;138;138;138m[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m  [38;2;138;138;138m[0m

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
      "hostname": "web-01",
      "name": "osapitest",
      "status": "ok"
    }
  ]
}
```

A container that the host does not have, on one host:

```ansi
$ osapi client node docker delete --target web-01 --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mID[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m  [38;2;138;138;138m[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node docker delete --target _all --name osapitest
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mID[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m  [38;2;138;138;138m[0m
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m  [38;2;138;138;138m[0m

  3 hosts: [38;2;103;234;147m2 ok[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "changed": false,
      "hostname": "web-01",
      "name": "osapitest",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example            | On the host                   |
| ------------------ | ----------------------------- |
| Remove a container | runs `docker rm -f osapitest` |

The command fails in these cases:

| Call                     | Flags       | Fails with   |
| ------------------------ | ----------- | ------------ |
| A name that is an option | `--name -f` | container_id |

## Permission

It needs the `docker:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--name`       | The container's name or ID (required)                  |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
