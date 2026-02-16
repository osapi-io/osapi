# Status

Display the job queue status with live updates using a BubbleTea TUI:

```bash
$ osapi client job status --poll-interval-seconds 5


  Queue Status:

  ┏━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━┓
  ┃ STATUS             ┃ COUNT       ┃
  ┣━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━━┫
  ┃ Total              ┃ 42          ┃
  ┃ Unprocessed        ┃ 5           ┃
  ┃ Processing         ┃ 2           ┃
  ┃ Completed          ┃ 30          ┃
  ┃ Failed             ┃ 5           ┃
  ┗━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━┛
```

## Flags

| Flag                      | Description                         | Default |
| ------------------------- | ----------------------------------- | ------- |
| `--poll-interval-seconds` | Interval between polling operations | 30      |

The status view auto-refreshes at the configured interval, showing job counts by
status and operation type.
