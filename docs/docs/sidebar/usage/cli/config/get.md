# Get

Show one setting in full.

It shows the value in effect and where it came from, the environment variable,
the type, the default, the rules and the description. An environment variable
wins over the file.

```bash
osapi config get [flags]
```

## Examples

```bash
osapi config get --key controller.api.port
osapi config get --key agent.labels -f ./osapi.yaml --json
```

## Result

The command prints:

```ansi
$ osapi config get --key controller.api.port
[1;38;2;0;213;255mConfig:[0m
  [38;2;176;48;176mFile:[0m  /etc/osapi/osapi.yaml

[1;38;2;255;0;255mSetting:[0m
  [38;2;176;48;176mKey:[0m          controller.api.port
  [38;2;176;48;176mValue:[0m        8080
  [38;2;176;48;176mSource:[0m       file
  [38;2;176;48;176mEnv:[0m          OSAPI_CONTROLLER_API_PORT
  [38;2;176;48;176mType:[0m         int
  [38;2;176;48;176mDefault:[0m      8080
  [38;2;176;48;176mValidate:[0m     min=1,max=65535
  [38;2;176;48;176mDescription:[0m  Port the REST API server listens on.
```

With `--json`, the command prints:

```json
{
  "file": "/etc/osapi/osapi.yaml",
  "setting": {
    "name": "controller.api.port",
    "value": 8080,
    "source": "file",
    "env": "OSAPI_CONTROLLER_API_PORT",
    "type": "int",
    "default": 8080,
    "validate": "min=1,max=65535",
    "doc": "Port the REST API server listens on."
  }
}
```

## Flags

| Flag    | Description                                                 | Default |
| ------- | ----------------------------------------------------------- | ------- |
| `--key` | The setting to show, such as controller.api.port (required) |         |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
