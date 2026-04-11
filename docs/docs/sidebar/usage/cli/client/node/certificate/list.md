# List

List all CA certificates on a target host, including both system-provided and
OSAPI-managed custom certificates:

```bash
$ osapi client node certificate list --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME                    SOURCE
  web-01    ok      mozilla/DigiCert.crt    system
  web-01    ok      mozilla/GlobalSign.crt  system
  web-01    ok      internal-ca             custom

  1 host: 1 ok
```

Target all hosts to list certificates across the fleet:

```bash
$ osapi client node certificate list --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME                    SOURCE
  web-01    ok      mozilla/DigiCert.crt    system
  web-01    ok      mozilla/GlobalSign.crt  system
  web-01    ok      internal-ca             custom
  web-02    ok      mozilla/DigiCert.crt    system
  web-02    ok      mozilla/GlobalSign.crt  system

  2 hosts: 2 ok
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
