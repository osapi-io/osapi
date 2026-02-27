# List

List jobs with queue summary and status breakdown:

```bash
$ osapi client job list

  Total: 86               Showing: All jobs
  Submitted: 2            Completed: 83           Failed: 1               Partial: 0
  Filter: limit 10

  Jobs:

  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━┓
  ┃ JOB ID                               ┃ STATUS    ┃ CREATED          ┃ TARGET    ┃ OPERATION           ┃ AGENTS ┃
  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━┫
  ┃ 550e8400-e29b-41d4-a716-446655440000 ┃ completed ┃ 2026-02-16 13:21 ┃ server1   ┃ node.hostname.get ┃         ┃
  ┃ 661f9511-f30c-41d4-a716-557766551111 ┃ completed ┃ 2026-02-16 13:43 ┃ server1   ┃ network.dns.get     ┃         ┃
  ┃ 772a0622-a41d-52e5-b827-668877662222 ┃ failed    ┃ 2026-02-16 15:48 ┃ server1   ┃ node.status.get   ┃         ┃
  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━┛
```

Filter by status:

```bash
$ osapi client job list --status completed --limit 3

  Total: 86                 Showing: completed (3)
  Submitted: 2            Completed: 83           Failed: 1               Partial: 0
  Filter: limit 3

  Jobs:

  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━┳━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━┓
  ┃ JOB ID                               ┃ STATUS    ┃ CREATED          ┃ TARGET  ┃ OPERATION           ┃ AGENTS ┃
  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━╋━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━┫
  ┃ 550e8400-e29b-41d4-a716-446655440000 ┃ completed ┃ 2026-02-16 13:21 ┃ server1 ┃ node.hostname.get ┃         ┃
  ┃ 661f9511-f30c-41d4-a716-557766551111 ┃ completed ┃ 2026-02-16 13:43 ┃ server1 ┃ network.dns.get     ┃         ┃
  ┃ 772a0622-a41d-52e5-b827-668877662222 ┃ completed ┃ 2026-02-16 15:48 ┃ server1 ┃ node.status.get   ┃         ┃
  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━┻━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━┛
```

## Flags

| Flag       | Description                                     | Default |
| ---------- | ----------------------------------------------- | ------- |
| `--status` | Filter by status (submitted, processing, etc.)  |         |
| `--limit`  | Limit number of jobs displayed (0 for no limit) | 10      |
| `--offset` | Skip the first N jobs (for pagination)          | 0       |
