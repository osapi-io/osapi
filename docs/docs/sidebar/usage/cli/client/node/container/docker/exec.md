# Exec

Execute a command in a container.

```bash
osapi client node container docker exec
```

## Examples

```bash
osapi client node container docker exec --id my-nginx --command "ls,-la,/"
osapi client node container docker exec \
  --id my-app \
  --command "python,-c,import os; print(os.environ['MY_VAR'])" \
  --env "MY_VAR=hello" \
  --working-dir /app
osapi client node container docker exec \
  --id my-nginx \
  --command "nginx,-t" \
  --target web-01
```

## Flags

| Flag            | Description                                           | Default |
| --------------- | ----------------------------------------------------- | ------- |
| `--command`     | Command to execute (comma-separated) (required)       |         |
| `--env`         | Environment variable in KEY=VALUE format (repeatable) |         |
| `--id`          | Container ID to exec in (required)                    |         |
| `--working-dir` | Working directory inside the container                |         |

## Global flags

| Flag               | Description                                            | Default                 |
| ------------------ | ------------------------------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                           |                         |
| `-j, --json`       | Enable JSON output                                     |                         |
| `-f, --osapi-file` | Path to config file                                    | `/etc/osapi/osapi.yaml` |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev) | `_all`                  |
| `--url`            | URL the client will connect to                         | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
