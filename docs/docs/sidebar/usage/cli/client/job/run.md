# Run

Submit a job and wait for completion:

```bash
$ osapi client job run \
    --json-file operation.json \
    --target-hostname _any \
    --timeout 60

  Job ID: 550e8400-e29b-41d4-a716-446655440000    Status: completed
  Hostname: server1
  Created: 2026-02-16T13:21:06Z
  Updated At: 2026-02-16T13:21:06Z

  Job Request:

  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
  ┃ DATA                            ┃
  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
  ┃ {                               ┃
  ┃   "data": {},                   ┃
  ┃   "type": "system.hostname.get" ┃
  ┃ }                               ┃
  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

  Job Result:

  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
  ┃ DATA                      ┃
  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
  ┃ {                         ┃
  ┃   "hostname": "server1"   ┃
  ┃ }                         ┃
  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

## Flags

| Flag                  | Description                                        | Default |
| --------------------- | -------------------------------------------------- | ------- |
| `--json-file`         | Path to the JSON file containing operation data    |         |
| `--target-hostname`   | Target hostname (`_any`, `_all`, or specific host) |         |
| `-t, --timeout`       | Timeout in seconds                                 | 60      |
| `-p, --poll-interval` | Poll interval in seconds                           | 2       |

This combines job submission and retrieval into a single command. It submits the
job, polls for completion, and displays the results.
