# Reboot

Schedule a reboot on the target host. The agent calls `shutdown -r` with the
configured delay. When `--delay` is 0, the reboot happens immediately:

```bash
$ osapi client node power reboot --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  ACTION  DELAY
  web-01    changed  true     reboot  0

  1 host: 1 changed
```

Reboot with a 60-second delay and a broadcast message:

```bash
$ osapi client node power reboot --target web-01 \
    --delay 60 --message "Scheduled maintenance reboot"

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  ACTION  DELAY
  web-01    changed  true     reboot  60

  1 host: 1 changed
```

Broadcast reboot to all hosts at once:

```bash
$ osapi client node power reboot --target _all --delay 30

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  ACTION  DELAY
  web-01    changed  true     reboot  30
  mac-01    skip

  2 hosts: 1 changed, 1 skipped

  Details:
  mac-01    unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node power reboot --target web-01 --json
{"results":[{"hostname":"web-01","action":"reboot","delay":0,"changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--delay`      | Seconds to wait before rebooting                         | `0`     |
| `--message`    | Optional message to broadcast before reboot              |         |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `-j, --json`   | Output raw JSON response                                 |         |
