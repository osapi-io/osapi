# List

List active nodes in the fleet:

```bash
$ osapi client node list


  Active Agents (2):

  ┏━━━━━━━━━━━━━━━━━━━━┓
  ┃ HOSTNAME           ┃
  ┣━━━━━━━━━━━━━━━━━━━━┫
  ┃ node-1             ┃
  ┃ node-2             ┃
  ┗━━━━━━━━━━━━━━━━━━━━┛
```

Discovers all active node agents by broadcasting a hostname query and collecting
responses.
