# Get

Get job details and status:

```bash
$ osapi client job get --job-id 550e8400-e29b-41d4-a716-446655440000

  Job ID: 550e8400-e29b-41d4-a716-446655440000    Status: completed
  Hostname: server1
  Created: 2026-02-16T13:21:06Z
  Updated At: 2026-02-16T13:21:06Z

  Job Request:

  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
  ┃ DATA                          ┃
  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
  ┃ {                             ┃
  ┃   "data": {},                 ┃
  ┃   "type": "node.status.get" ┃
  ┃ }                             ┃
  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

  Agent States:

  ┏━━━━━━━━━━┳━━━━━━━━━━━┳━━━━━━━━━━┳━━━━━━━┓
  ┃ HOSTNAME ┃ STATUS    ┃ DURATION ┃ ERROR ┃
  ┣━━━━━━━━━━╋━━━━━━━━━━━╋━━━━━━━━━━╋━━━━━━━┫
  ┃ server1  ┃ completed ┃ 0s       ┃       ┃
  ┗━━━━━━━━━━┻━━━━━━━━━━━┻━━━━━━━━━━┻━━━━━━━┛

  Job Result:

  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
  ┃ DATA                                ┃
  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
  ┃ {                                   ┃
  ┃   "hostname": "server1",            ┃
  ┃   "uptime": 231685,                 ┃
  ┃   "os": {                           ┃
  ┃     "distribution": "Ubuntu",       ┃
  ┃     "version": "24.04"              ┃
  ┃   }                                 ┃
  ┃ }                                   ┃
  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

## Flags

| Flag       | Description        | Default  |
| ---------- | ------------------ | -------- |
| `--job-id` | Job ID to retrieve | required |

Job status is retrieved from the API and reflects the current state of the job.
