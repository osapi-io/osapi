# Get

Show one setting in full: its value in effect and where it came from, its
environment variable, type, default, rules and description.

```bash
$ osapi config get --key controller.api.port

Config:
  File:  /etc/osapi/osapi.yaml (not found: defaults and environment only)

Setting:
  Key:          controller.api.port
  Value:        8080
  Source:       default
  Env:          OSAPI_CONTROLLER_API_PORT
  Type:         int
  Default:      8080
  Validate:     min=1,max=65535
  Description:  Port the REST API server listens on.
```

An environment variable wins over the file:

```bash
$ OSAPI_CONTROLLER_API_PORT=9000 osapi config get --key controller.api.port --json
{"name":"controller.api.port","env":"OSAPI_CONTROLLER_API_PORT","type":"int","default":8080,"validate":"min=1,max=65535","doc":"Port the REST API server listens on.","value":9000,"source":"env"}
```

## Flags

| Flag    | Description                                                   |
| ------- | ------------------------------------------------------------- |
| `--key` | The setting to show, such as `controller.api.port` (required) |
