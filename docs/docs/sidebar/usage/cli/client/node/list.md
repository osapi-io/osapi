# List

List active nodes in the fleet:

```bash
$ osapi client node list


  Active Workers (2):

  ┏━━━━━━━━━━━━━━━━━━━━┓
  ┃ HOSTNAME           ┃
  ┣━━━━━━━━━━━━━━━━━━━━┫
  ┃ worker-node-1      ┃
  ┃ worker-node-2      ┃
  ┗━━━━━━━━━━━━━━━━━━━━┛
```

Discovers all active node agents by broadcasting a hostname query and collecting
responses.
