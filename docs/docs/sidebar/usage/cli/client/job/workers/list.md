# List

List active workers in the fleet:

```bash
$ osapi client job workers list


  Active Workers (2):

  ┏━━━━━━━━━━━━━━━━━━━━┓
  ┃ HOSTNAME           ┃
  ┣━━━━━━━━━━━━━━━━━━━━┫
  ┃ worker-node-1      ┃
  ┃ worker-node-2      ┃
  ┗━━━━━━━━━━━━━━━━━━━━┛
```

Discovers all active workers by broadcasting a hostname query and collecting
responses.
