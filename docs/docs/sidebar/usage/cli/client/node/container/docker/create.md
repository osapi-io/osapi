# Create

Create a new container on the target node from the specified image:

```bash
$ osapi client node container docker create --image nginx:latest

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED  ERROR  ID            NAME          IMAGE         STATE
  server1   ok      true            a1b2c3d4e5f6  eager_turing  nginx:latest  running
```

Create a named container with environment variables, port mappings, and volume
mounts:

```bash
$ osapi client node container docker create \
    --image nginx:latest \
    --name my-nginx \
    --env "PORT=8080" --env "DEBUG=true" \
    --port "8080:80" --port "8443:443" \
    --volume "/data:/var/lib/data"
```

Create a container without starting it immediately:

```bash
$ osapi client node container docker create \
    --image alpine:latest \
    --name my-alpine \
    --auto-start=false
```

Target a specific host:

```bash
$ osapi client node container docker create \
    --image redis:7 \
    --name cache \
    --target web-01
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node container docker create --image nginx:latest --json
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--image`      | Container image reference (**required**)                 |         |
| `--name`       | Optional name for the container                          |         |
| `--env`        | Environment variable in `KEY=VALUE` format (repeatable)  | `[]`    |
| `--port`       | Port mapping in `host:container` format (repeatable)     | `[]`    |
| `--volume`     | Volume mount in `host:container` format (repeatable)     | `[]`    |
| `--auto-start` | Start the container immediately after creation           | `true`  |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
