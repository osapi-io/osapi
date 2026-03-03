# Demo Recording Design

## Overview

Create a VHS-scripted terminal recording (GIF) for the README that sells OSAPI's
value in ~30-60 seconds. Narrative: "zero to managed system in 30 seconds."

## Tool

[VHS](https://github.com/charmbracelet/vhs) — scriptable `.tape` files that
render to GIF. Reproducible, version-controllable.

## Output

- Format: GIF
- Location: `asset/demo.gif` (embedded in README)
- Duration: ~36 seconds

## Demo Flow

### Scene 1: Start (~5s)

```
$ osapi start
```

One command boots NATS, API server, and agent. Brief pause on startup output.

### Scene 2: Health check (~5s)

```
$ osapi client health status
```

Full system health — component status, agent metrics, job stats. "Everything's
green."

### Scene 3: Node discovery (~8s)

```
$ osapi client agent list
$ osapi client node status
```

Agent registered with OS info, load, memory. Rich node status with uptime, disk,
memory, load averages.

### Scene 4: Run a command (~8s)

```
$ osapi client node command exec --command "uname" --args "-a"
```

Async job submission + result. Demonstrates the job system implicitly.

### Scene 5: Audit trail (~5s)

```
$ osapi client audit list
```

Audit log showing all API calls we just made. "Everything is tracked."

### Scene 6: JSON output (~5s)

```
$ osapi client node status --json
```

Structured JSON output — shows automation-friendliness.

## Out of Scope (future multi-host recording)

- `--target _all` / `--target hostname` broadcasting
- Agent labels and label-based routing
- Job lifecycle (add/list/get/retry/delete)
- DNS get/update, ping
- Token generation / RBAC

## Implementation Notes

- VHS `.tape` file lives at project root (`demo.tape`)
- Need VHS installed (`brew install vhs`)
- Uses tmux with a top/bottom split: top pane runs `osapi start` (server logs
  visible), bottom pane runs client commands
- VHS drives tmux via keystrokes — starts tmux, splits pane, types commands into
  each pane
- GIF stored in `asset/demo.gif`, referenced from README
