# Get

Print the config file: each key, what it does, its type, default and rules.

The list comes from this binary. Thus it agrees with the build that you run. A
section for each top word of a key, such as controller.
`osapi config get --key <key>` shows one key with its value in effect.

```bash
osapi config spec get
```

## Examples

```bash
osapi config spec get
osapi config spec get --json
```

## Result

The command prints:

```ansi
$ osapi config spec get
[1;38;2;0;213;255mcontroller:[0m
  [1mcontroller.api.port[0m
    Port the REST API server listens on. Type: int. Default: 8080.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_API_PORT, which wins over[0m
    [38;2;138;138;138mthe file. Rules: min=1,max=65535.[0m

  [1mcontroller.ui.enabled[0m
    Serve the embedded UI. Type: bool. Default: true.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_UI_ENABLED, which wins over[0m
    [38;2;138;138;138mthe file.[0m
```

With `--json`, the command prints:

```json
{
  "kinds": [
    {
      "kind": "controller",
      "title": "controller",
      "terms": [
        {
          "kind": "controller",
          "name": "controller.api.port",
          "does": "Port the REST API server listens on. Type: int. Default: 8080.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_API_PORT, which wins over the file. Rules: min=1,max=65535."
        },
        {
          "kind": "controller",
          "name": "controller.ui.enabled",
          "does": "Serve the embedded UI. Type: bool. Default: true.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_UI_ENABLED, which wins over the file."
        }
      ]
    }
  ]
}
```

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
