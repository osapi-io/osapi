# Undeploy

Remove a deployed file from disk on the target node. The object store entry
stays, so you can deploy the file again at any time.

```bash
$ osapi client node file undeploy \
    --target server1 \
    --path /etc/app/app.conf

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  server1   changed  true

  1 host: 1 changed
```

If the file does not exist on disk, the operation changes nothing and returns
`Changed: false`.

Undeploy from all hosts in a label group:

```bash
$ osapi client node file undeploy \
    --path /etc/nginx/nginx.conf \
    --target group:web
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node file undeploy \
    --target server1 \
    --path /etc/app/app.conf \
    --json
{"results":[{"hostname":"server1","changed":true}],"job_id":"550e8400-e29b-41d4-a716-446655440000"}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--path`       | Path of the file to remove on the target (**required**)  |         |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |

:::note

The command does not delete the object store entry. When you do not need the
object, use `osapi client file delete` to remove it from the store.

:::
