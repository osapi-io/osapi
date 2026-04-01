# List

List all CA certificates on a target host, including both system-provided and
OSAPI-managed custom certificates:

```bash
$ osapi client node certificate list --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME                    SOURCE
  mozilla/DigiCert.crt    system
  mozilla/GlobalSign.crt  system
  internal-ca             custom
```

Target all hosts to list certificates across the fleet:

```bash
$ osapi client node certificate list --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  NAME                    SOURCE
  mozilla/DigiCert.crt    system
  mozilla/GlobalSign.crt  system
  internal-ca             custom

  web-02
  NAME                    SOURCE
  mozilla/DigiCert.crt    system
  mozilla/GlobalSign.crt  system
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node certificate list --target web-01 --json
{"results":[{"hostname":"web-01","status":"ok","certificates":[
{"name":"mozilla/DigiCert.crt","source":"system"},
{"name":"internal-ca","source":"custom","object":"internal-ca"}
]}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
