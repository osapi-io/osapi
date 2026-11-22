# Get

Get the memory usage of the host.

Reads the total, the free and the used memory, in bytes. Used is the total less
what is available, which includes the cache, as free(1) reports it.

```bash
osapi client node memory get [flags]
```

## Examples

```bash
# A host with 8 GiB, 2 GiB of it free
osapi client node memory get --target web-01
```

## Result

Each example prints this:

A host with 8 GiB, 2 GiB of it free:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mMEMORY[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      8.0 GB 2.0 GB 4.0 GB

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

## Permission

It needs the `memory:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
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
