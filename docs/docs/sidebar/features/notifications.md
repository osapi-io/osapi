---
sidebar_position: 6
---

# Notifications

OSAPI monitors component health through condition evaluation and notifies when
conditions change. The notification system watches the registry KV bucket for
condition transitions and dispatches events through a pluggable backend.

## How It Works

Every component (agent, controller, NATS server) evaluates conditions on each
heartbeat and writes them to the registry KV bucket. A watcher on the controller
detects transitions:

- **Fired**: a condition becomes active (e.g., `DiskPressure` crosses threshold)
- **Resolved**: a condition becomes inactive (e.g., disk usage drops below
  threshold)
- **Unreachable**: a component's heartbeat expires (TTL timeout)

Active conditions are re-fired at a configurable interval so they remain visible
in logs and alerts.

## Conditions

| Condition               | Components       | Description                                               |
| ----------------------- | ---------------- | --------------------------------------------------------- |
| `MemoryPressure`        | agent            | Host memory usage exceeds threshold (default 90%)         |
| `HighLoad`              | agent            | Load average exceeds CPU count × multiplier (default 2.0) |
| `DiskPressure`          | agent            | Any disk usage exceeds threshold (default 90%)            |
| `ProcessMemoryPressure` | agent, api, nats | Process RSS exceeds configured byte threshold             |
| `ProcessHighCPU`        | agent, api, nats | Process CPU usage exceeds configured percent threshold    |
| `ComponentUnreachable`  | agent, api, nats | Heartbeat expired (TTL timeout)                           |

Host-level conditions are evaluated on agents only. Process-level conditions are
evaluated on all components. `ComponentUnreachable` is emitted by the watcher
when a heartbeat TTL expires — it does not appear on the component's
registration because the component is already gone.

## Notifier Backends

The notification system uses a pluggable `Notifier` interface. Currently one
backend is available:

### Log (default)

Writes condition events to the structured log. Fired conditions log at WARN
level, resolved conditions at INFO:

```
WRN condition fired   component=agent hostname=web-01 condition=DiskPressure active=true reason="/ 92% used"
INF condition resolved component=agent hostname=web-01 condition=DiskPressure active=false
WRN condition fired   component=nats hostname=nats-01 condition=ComponentUnreachable active=true reason="heartbeat expired"
```

### Future Backends

- **Slack** — post to a webhook URL
- **Email** — send via SMTP
- **Webhook** — POST to a configurable URL

## Re-notification

By default, a condition fires once when it becomes active and once when it
resolves. To keep active conditions visible, configure `renotify_interval`:

```yaml
controller:
  notifications:
    enabled: true
    notifier: 'log'
    renotify_interval: '5m'
```

With `renotify_interval: '5m'`, an active `DiskPressure` condition re-fires
every 5 minutes until resolved. Uses Go duration format (`1m`, `5m`, `1h`). Set
to `'0'` to disable re-notification.

## Configuration

| Key                 | Env Variable                                       | Description                           |
| ------------------- | -------------------------------------------------- | ------------------------------------- |
| `enabled`           | `OSAPI_CONTROLLER_NOTIFICATIONS_ENABLED`           | Enable the watcher (default: `false`) |
| `notifier`          | `OSAPI_CONTROLLER_NOTIFICATIONS_NOTIFIER`          | Backend: `"log"` (default)            |
| `renotify_interval` | `OSAPI_CONTROLLER_NOTIFICATIONS_RENOTIFY_INTERVAL` | Re-fire interval (default: `"0"`)     |

See [Configuration](../usage/configuration.md) for the full reference.

## Architecture

The watcher runs as a background goroutine in the controller. It monitors the
registry KV bucket using NATS KV Watch. On each update it compares the previous
condition set to the current one and emits `ConditionEvent`s for transitions.

The watcher is designed to be extractable into a separate process — its only
dependency is NATS KV access and a `Notifier` implementation.

## Permissions

No specific permissions are required for notifications. The watcher reads the
same registry KV bucket that the health status endpoint uses.
