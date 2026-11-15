# Undeploy

Remove a deployed file from a host.

The object store entry stays, so you can deploy the file again at any time. If
the file does not exist on disk, the operation changes nothing and returns
`Changed: false`. When you do not need the object, use
`osapi client file delete` to remove it from the store.

```bash
osapi client node file undeploy
```

## Examples

```bash
osapi client node file undeploy \
  --target server1 \
  --path /etc/app/app.conf
osapi client node file undeploy \
  --path /etc/nginx/nginx.conf \
  --target group:web
osapi client node file undeploy \
  --target server1 \
  --path /etc/app/app.conf \
  --json
```

## Flags

| Flag     | Description                                                      | Default |
| -------- | ---------------------------------------------------------------- | ------- |
| `--path` | Path of the file to undeploy on the target filesystem (required) |         |

## Global flags

| Flag               | Description                                            | Default                 |
| ------------------ | ------------------------------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                           |                         |
| `-j, --json`       | Enable JSON output                                     |                         |
| `-f, --osapi-file` | Path to config file                                    | `/etc/osapi/osapi.yaml` |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev) | `_all`                  |
| `--url`            | URL the client will connect to                         | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
