# Status

Get the system status:

```bash
$ osapi client system status

  Hostname: server1
  Load Average (1m, 5m, 15m): 1.83, 1.96, 2.02
  Memory: 19 GB used / 31 GB total / 10 GB free
  OS: Ubuntu 24.04


  Disks:

  ┏━━━━━━━━━━━┳━━━━━━━┳━━━━━━━┳━━━━━━━┓
  ┃ DISK NAME ┃ TOTAL ┃ USED  ┃ FREE  ┃
  ┣━━━━━━━━━━━╋━━━━━━━╋━━━━━━━╋━━━━━━━┫
  ┃ /         ┃ 97 GB ┃ 56 GB ┃ 36 GB ┃
  ┃ /boot     ┃ 1 GB  ┃ 0 GB  ┃ 1 GB  ┃
  ┗━━━━━━━━━━━┻━━━━━━━┻━━━━━━━┻━━━━━━━┛
```

When targeting all hosts, a summary table is shown:

```bash
$ osapi client system status --target _all

  ┏━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━┳━━━━━━━━━━━━━━━┓
  ┃ HOSTNAME ┃ UPTIME                         ┃ LOAD (1m) ┃ MEMORY USED   ┃
  ┣━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━╋━━━━━━━━━━━━━━━┫
  ┃ server1  ┃ 64 days, 11 hours, 20 minutes  ┃ 1.83      ┃ 19 GB / 31 GB ┃
  ┃ server2  ┃ 12 days, 3 hours, 45 minutes   ┃ 0.45      ┃ 8 GB / 16 GB  ┃
  ┗━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━┻━━━━━━━━━━━━━━━┛
```
