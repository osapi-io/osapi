# Update

Deploy a new unit file for a service.

Writes the object as the unit file of a service that osapi deployed, then
reloads systemd when the file changed. A service that osapi did not deploy is
not found.

```bash
osapi client node service update [flags]
```

## Examples

```bash
# A new unit file
osapi client node service update --target web-01 --name app --object app-v2.service
```

## Result

Each example prints this:

A new unit file:

```ansi
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mapp[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With the default settings, each example does this on the host:

| Example         | On the host                                                                                                                                                                                                            |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| A new unit file | writes `/etc/systemd/system/osapi-app.service`: `[Unit]\nDescription=Example app\n\n[Service]\nExecStart=/bin/sleep infinity\n\n[Install]\nWantedBy=multi-user.target\nRestart=always`; runs `systemctl daemon-reload` |

The command fails in these cases:

| Call                                  | Flags                                                                                                                                                                                                                                                                                  | Fails with               |
| ------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------ |
| A service that osapi did not deploy   | `--name ssh --object app.service`                                                                                                                                                                                                                                                      | not found                |
| A unit file that osapi did not deploy | `--name agent --object app.service`                                                                                                                                                                                                                                                    | not found                |
| No object                             | `--name app`                                                                                                                                                                                                                                                                           | Object                   |
| An object name over 255 characters    | `--name app --object aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | Object                   |
| A name with a space                   | `--name "a b" --object app.service`                                                                                                                                                                                                                                                    | Name                     |
| A name that is an option              | `--name --now --object app.service`                                                                                                                                                                                                                                                    | not a valid service name |

## Permission

It needs the `service:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                                   | Default |
| -------------- | ----------------------------------------------------------------------------- | ------- |
| `--name`       | The service, such as app for osapi-app.service (required)                     |         |
| `--object`     | The object that holds the unit file, from osapi client file upload (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                        | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
