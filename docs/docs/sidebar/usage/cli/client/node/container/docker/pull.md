# Pull

Pull a container image on the target node:

```bash
$ osapi client node container docker pull --image nginx:latest

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  STATUS  CHANGED  ERROR  IMAGE ID            TAG     SIZE
  ok      true            sha256:a1b2c3d4...  latest  187.8 MiB
```

Pull a specific image version:

```bash
$ osapi client node container docker pull --image alpine:3.18
```

Pull from a custom registry:

```bash
$ osapi client node container docker pull \
    --image registry.example.com/myapp:v1.2.3
```

Target a specific host:

```bash
$ osapi client node container docker pull --image redis:7 --target web-01
```

Pull on all hosts:

```bash
$ osapi client node container docker pull --image nginx:latest --target _all
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node container docker pull --image nginx:latest --json
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--image`      | Image reference to pull (**required**)                   |         |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
