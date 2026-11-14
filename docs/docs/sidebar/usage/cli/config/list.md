# List

List every setting osapi reads, with the value in effect and where it came from:
`env`, `file`, `default` or `unset`. The file and the environment are read the
way `osapi start` reads them. Without a config file, the values are the defaults
and the environment. Passwords, signing keys and tokens print as `********`.

```bash
$ osapi config list

Config:
  File:         /etc/osapi/osapi.yaml (not found: defaults and environment only)
  Environment:  OSAPI_, then the key in capitals with dots as underscores
  One key:      osapi config get --key <key>

Controller (26):
  KEY                                         VALUE                SOURCE   ENV
  controller.client.url                       http://0.0.0.0:8080  default  OSAPI_CONTROLLER_CLIENT_URL
  controller.client.security.bearer_token                          unset    OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN
  controller.api.port                         8080                 default  OSAPI_CONTROLLER_API_PORT
  ...
```

`--json` carries each key's type, default, rules and description as well:

```bash
$ osapi config list --json
[{"name":"controller.client.url","env":"OSAPI_CONTROLLER_CLIENT_URL","type":"string","doc":"Base URL the CLI client connects to.","value":"http://0.0.0.0:8080","source":"default"}, ...]
```
