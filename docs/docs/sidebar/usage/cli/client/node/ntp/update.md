# Update

Replace the NTP servers that osapi set.

The same as create, for servers that osapi set already. It fails with not found
when osapi set none. The servers that the file has already report changed:
false.

```bash
osapi client node ntp update [flags]
```

## Examples

```bash
# Replace the servers
osapi client node ntp update --target web-01 --servers time.cloudflare.com

# The servers that the file has
osapi client node ntp update --target web-01 --servers 0.pool.ntp.org
```

## Result

Replace the servers, on one host:

```ansi
$ osapi client node ntp update --target web-01 --servers time.cloudflare.com
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node ntp update --target _all --servers time.cloudflare.com
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m

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
      "status": "ok"
    }
  ]
}
```

The servers that the file has, on one host:

```ansi
$ osapi client node ntp update --target web-01 --servers 0.pool.ntp.org
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node ntp update --target _all --servers 0.pool.ntp.org
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m
  [1mweb-02[0m    [38;2;103;234;147mok[0m

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
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example             | On the host                                                                                                          |
| ------------------- | -------------------------------------------------------------------------------------------------------------------- |
| Replace the servers | writes `/etc/chrony/sources.d/osapi-ntp.sources`: `server time.cloudflare.com iburst`; runs `chronyc reload sources` |

The command fails in these cases:

| Call                               | Flags                                                                                                                                                                                                                                                                                                                                                                                                                                      | Fails with                         |
| ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------- |
| When osapi set no servers          | `--servers 0.pool.ntp.org`                                                                                                                                                                                                                                                                                                                                                                                                                 | osapi manages no servers           |
| No servers                         |                                                                                                                                                                                                                                                                                                                                                                                                                                            | required                           |
| An empty list of servers           |                                                                                                                                                                                                                                                                                                                                                                                                                                            | min                                |
| A server with a directive after it | `--servers "0.pool.ntp.org iburst\nallow all"`                                                                                                                                                                                                                                                                                                                                                                                             | is not a hostname or an IP address |
| More than 16 servers               | `--servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org --servers 0.pool.ntp.org` | max                                |

## Permission

It needs the `ntp:write` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--servers`    | The NTP servers, hostnames or IP addresses (required)  |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                                                 | Default                 |
| ------------------ | ----------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                |                         |
| `-j, --json`       | Enable JSON output                                          |                         |
| `-f, --osapi-file` | Path to config file                                         | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url) | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
