# Create

Create a container.

Runs docker create, which pulls the image when the host does not have it. The
environment goes on docker's standard input, not in its arguments, which each
user of the host can read. auto_start starts the container once it is made. A
container of the name that the host has already is not made again, and reports
changed: false.

```bash
osapi client node docker create [flags]
```

## Examples

```bash
# Create a container
osapi client node docker create --target web-01 --name osapitest --image alpine:3.20 --command sleep --command 300

# A container that the host has
osapi client node docker create --target web-01 --name osapitest --image alpine:3.20 --command sleep --command 300
```

## Result

Create a container, on one host:

```ansi
$ osapi client node docker create --target web-01 --name osapitest --image alpine:3.20 --command sleep --command 300
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mID[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m  [38;2;138;138;138m3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node docker create --target _all --name osapitest --image alpine:3.20 --command sleep --command 300
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mID[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m  [38;2;138;138;138m3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1mosapitest[0m  [38;2;138;138;138m3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7[0m

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
      "id": "3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7",
      "name": "osapitest",
      "status": "ok"
    }
  ]
}
```

A container that the host has, on one host:

```ansi
$ osapi client node docker create --target web-01 --name osapitest --image alpine:3.20 --command sleep --command 300
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mID[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m  [38;2;138;138;138m3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node docker create --target _all --name osapitest --image alpine:3.20 --command sleep --command 300
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME       [0m[38;2;10;143;176mID[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m  [38;2;138;138;138m3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7[0m
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mosapitest[0m  [38;2;138;138;138m3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7[0m

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
      "id": "3f4e8a1b2c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7",
      "name": "osapitest",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example            | On the host                                                 |
| ------------------ | ----------------------------------------------------------- |
| Create a container | runs `docker create --name osapitest alpine:3.20 sleep 300` |

The command fails in these cases:

| Call                                        | Flags                                                                                                                                                                                                                                                                                                                      | Fails with     |
| ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------- |
| A name that is an option                    | `--name -f --image alpine:3.20`                                                                                                                                                                                                                                                                                            | container_id   |
| No image                                    | `--name osapitest`                                                                                                                                                                                                                                                                                                         | required       |
| An image that is an option                  | `--name osapitest --image --privileged`                                                                                                                                                                                                                                                                                    | image_ref      |
| A hostname that is an option                | `--name osapitest --image alpine:3.20 --container-hostname -x`                                                                                                                                                                                                                                                             | dns_name       |
| A hostname longer than 253 characters       | `--name osapitest --image alpine:3.20 --container-hostname aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | max            |
| A DNS server that is not an address         | `--name osapitest --image alpine:3.20 --dns nameserver`                                                                                                                                                                                                                                                                    | ip             |
| A port that is not host:container           | `--name osapitest --image alpine:3.20 --ports 80`                                                                                                                                                                                                                                                                          | port_mapping   |
| A bind mount that is not two absolute paths | `--name osapitest --image alpine:3.20 --volumes data:/data`                                                                                                                                                                                                                                                                | volume_mapping |
| No name                                     | `--image alpine:3.20`                                                                                                                                                                                                                                                                                                      | required       |

## Permission

It needs the `docker:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag                   | Description                                                                                      | Default |
| ---------------------- | ------------------------------------------------------------------------------------------------ | ------- |
| `--auto-start`         | Start the container once it is created                                                           |         |
| `--command`            | The command, in place of the image's own                                                         |         |
| `--container-hostname` | The hostname inside the container                                                                |         |
| `--dns`                | The DNS servers of the container                                                                 |         |
| `--env`                | The environment variables of the container (secrets: NAME=secret://&lt;vault&gt;/&lt;key&gt;)    |         |
| `--image`              | The image to run, such as alpine:3.20; docker pulls it when the host does not have it (required) |         |
| `--name`               | The container's name (required)                                                                  |         |
| `--ports`              | The published ports, each host:container, such as 8080:80                                        |         |
| `-T, --target`         | Target: _all, _any, hostname, or label (group:web.dev)                                           | `_all`  |
| `--volumes`            | The bind mounts, each host:container, such as /srv/data:/data                                    |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
