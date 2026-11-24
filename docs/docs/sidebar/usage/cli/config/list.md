# List

List every setting, its value in effect and where it came from.

Each value shows where it came from: `env`, `file`, `default` or `unset`. The
file and the environment are read the way `osapi start` reads them. Without a
config file, the values are the defaults and the environment. Passwords, signing
keys and tokens print as `********`. `--json` carries the type, the default, the
rules and the description of each key as well.

```bash
osapi config list
```

## Examples

```bash
osapi config list
osapi config list -f ./osapi.yaml --json
```

## Result

The command prints:

```ansi
$ osapi config list
[1;38;2;0;213;255mSettings (3):[0m
  [38;2;10;143;176mKEY                                   [0m[38;2;10;143;176mVALUE     [0m[38;2;10;143;176mSOURCE   [0m[38;2;10;143;176mENV[0m
  controller.api.port                   8080      file     OSAPI_CONTROLLER_API_PORT
  controller.api.security.signing_keys  ********  env      OSAPI_CONTROLLER_API_SECURITY_SIGNING_KEYS
  nats.server.port                      4222      default  OSAPI_NATS_SERVER_PORT
  [38;2;10;143;176mFile:[0m  /etc/osapi/osapi.yaml
```

With `--json`, the command prints:

```json
{
  "file": "/etc/osapi/osapi.yaml",
  "items": [
    {
      "name": "controller.api.port",
      "value": 8080,
      "source": "file",
      "env": "OSAPI_CONTROLLER_API_PORT",
      "type": "int",
      "default": 8080,
      "validate": "min=1,max=65535",
      "doc": "Port the REST API server listens on."
    },
    {
      "name": "controller.api.security.signing_keys",
      "value": "********",
      "source": "env",
      "env": "OSAPI_CONTROLLER_API_SECURITY_SIGNING_KEYS",
      "type": "list",
      "secret": true
    },
    {
      "name": "nats.server.port",
      "value": 4222,
      "source": "default",
      "env": "OSAPI_NATS_SERVER_PORT",
      "type": "int",
      "default": 4222
    }
  ]
}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../output.md) tells how the result is printed, and what `--json` gives.
