# Create

Deploy a service.

Gets the object from the object store and writes it as the unit file of the
service, such as /etc/systemd/system/osapi-app.service for app, then reloads
systemd. A unit file at that path that osapi did not write, such as the agent's
own osapi-agent.service, is refused. A unit file that already has the content
does not change. The service does not start: start and enable are their own
actions. On a Mac, the object is the plist of the daemon,
/Library/LaunchDaemons/io.osapi.app.plist for app, with the label io.osapi.app,
and launchd does not load it until start.

```bash
osapi client node service create [flags]
```

## Examples

```bash
# Deploy a unit file
osapi client node service create --target web-01 --name app --object app.service

# A unit file that has the content
osapi client node service create --target web-01 --name app --object app.service

# Deploy a plist on a Mac
osapi client node service create --target web-01 --name app --object app.plist

# A plist that has the content
osapi client node service create --target web-01 --name app --object app.plist
```

## Result

Deploy a unit file, on one host:

```ansi
$ osapi client node service create --target web-01 --name app --object app.service
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mapp[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node service create --target _all --name app --object app.service
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mapp[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1mapp[0m

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
      "name": "app",
      "status": "ok"
    }
  ]
}
```

A unit file that has the content, on one host:

```ansi
$ osapi client node service create --target web-01 --name app --object app.service
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mapp[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node service create --target _all --name app --object app.service
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mapp[0m
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mapp[0m

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
      "name": "app",
      "status": "ok"
    }
  ]
}
```

Deploy a plist on a Mac, on one host:

```ansi
$ osapi client node service create --target web-01 --name app --object app.plist
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mapp[0m

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node service create --target _all --name app --object app.plist
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  [1mapp[0m
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  [1mapp[0m

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
      "name": "app",
      "status": "ok"
    }
  ]
}
```

A plist that has the content, on one host:

```ansi
$ osapi client node service create --target web-01 --name app --object app.plist
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mapp[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node service create --target _all --name app --object app.plist
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mapp[0m
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mapp[0m

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
      "name": "app",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example                 | On the host                                                                                                                                                                                                                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Deploy a unit file      | writes `/etc/systemd/system/osapi-app.service`: `[Unit]\nDescription=Example app\n\n[Service]\nExecStart=/bin/sleep infinity\n\n[Install]\nWantedBy=multi-user.target`; runs `systemctl daemon-reload`                                                                               |
| Deploy a plist on a Mac | writes `/Library/LaunchDaemons/io.osapi.app.plist`: `<?xml version="1.0" encoding="UTF-8"?>\n<plist version="1.0"><dict><key>Label</key><string>io.osapi.app</string>\n<key>ProgramArguments</key><array><string>/bin/sleep</string><string>86400</string></array>\n</dict></plist>` |

The command fails in these cases:

| Call                                             | Flags                                                                                                                                                                                                                                                                                  | Fails with               |
| ------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------ |
| An object the store does not have                | `--name app --object nosuch`                                                                                                                                                                                                                                                           | not found                |
| A unit file or a plist that osapi did not deploy | `--name agent --object app.service`                                                                                                                                                                                                                                                    | osapi did not deploy it  |
| No object                                        | `--name app`                                                                                                                                                                                                                                                                           | Object                   |
| An object name over 255 characters               | `--name app --object aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa` | Object                   |
| A name with a space                              | `--name "a b" --object app.service`                                                                                                                                                                                                                                                    | Name                     |
| A name that is an option                         | `--name --now --object app.service`                                                                                                                                                                                                                                                    | not a valid service name |

## Permission

It needs the `service:write` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                                                   | Default |
| -------------- | ----------------------------------------------------------------------------- | ------- |
| `--name`       | The service, such as app for osapi-app.service (required)                     |         |
| `--object`     | The object that holds the unit file, from osapi client file upload (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                        | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
