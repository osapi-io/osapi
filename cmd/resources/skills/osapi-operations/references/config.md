# Configuration

```sh
osapi config list --json                      # every setting: value, source, env, type, default, description
osapi config get --key controller.api.port --json  # one setting
osapi config validate --json                  # the checks start runs, without starting
```

Precedence, highest first: an `OSAPI_` environment variable, the file, the
default. The variable for a key is `OSAPI_` and the key in capitals with dots as
underscores: `controller.api.port` is `OSAPI_CONTROLLER_API_PORT`.

`config list` and `config get` report `source` for each key: `env`, `file`,
`default` or `unset`. A value that is not what you set is usually an environment
variable overriding the file.

Two keys have no default: `controller.api.security.signing_key`, at least 32
characters, from `openssl rand -hex 32`, and
`controller.client.security.bearer_token`, from `osapi token generate`.

Durations are Go durations, `30s`, `5m`, `168h`. A `d` suffix is rejected.

`config validate` exits 1 with the field that failed. Fix that field and run it
again before starting anything.