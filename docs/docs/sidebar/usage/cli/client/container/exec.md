# Exec

Execute a command inside a running container on the target node:

```bash
$ osapi client container exec --id my-nginx --command "ls,-la,/"

  Job ID:    550e8400-e29b-41d4-a716-446655440000
  Hostname:  server1
  Exit Code: 0
  Stdout:    total 80
             drwxr-xr-x   1 root root 4096 Jan 15 10:30 .
             drwxr-xr-x   1 root root 4096 Jan 15 10:30 ..
             ...
```

Execute with environment variables and a working directory:

```bash
$ osapi client container exec \
    --id my-app \
    --command "python,-c,import os; print(os.environ['MY_VAR'])" \
    --env "MY_VAR=hello" \
    --working-dir /app
```

Target a specific host:

```bash
$ osapi client container exec \
    --id my-nginx \
    --command "nginx,-t" \
    --target web-01
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client container exec --id my-nginx --command "ls" --json
```

## Flags

| Flag            | Description                                              | Default |
| --------------- | -------------------------------------------------------- | ------- |
| `--id`          | Container ID or name to exec in (**required**)           |         |
| `--command`     | Command to execute, comma-separated (**required**)       |         |
| `--env`         | Environment variable in `KEY=VALUE` format (repeatable)  | `[]`    |
| `--working-dir` | Working directory inside the container                   |         |
| `-T, --target`  | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`    | Output raw JSON response                                 |         |
