# List

List jobs from the NATS KV store:

```bash
$ osapi client job list


  Jobs:

  ┏━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━┳━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━┓
  ┃ JOB ID             ┃ STATUS      ┃ TARGET     ┃ OPERATION               ┃
  ┣━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━━╋━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━━━━━┫
  ┃ 550e8400-e29b-...  ┃ completed   ┃ _any       ┃ system.hostname.get     ┃
  ┃ 661f9511-f30c-...  ┃ processing  ┃ _all       ┃ network.dns.get         ┃
  ┗━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━┻━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

## Flags

| Flag       | Description                                     | Default |
| ---------- | ----------------------------------------------- | ------- |
| `--status` | Filter by status (submitted, processing, etc.)  |         |
| `--limit`  | Limit number of jobs displayed (0 for no limit) | 10      |
| `--offset` | Skip the first N jobs (for pagination)          | 0       |
