# List

List all user accounts on a target host:

```bash
$ osapi client node user list --target web-01

  NAME     UID   GID   HOME            SHELL        GROUPS       LOCKED
  deploy   1001  1001  /home/deploy    /bin/bash     sudo,docker  no
  app      1002  1002  /home/app       /bin/sh       users        no
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node user list --target web-01 --json
{"results":[{"hostname":"web-01","users":[{"name":"deploy","uid":1001,"gid":1001,"home":"/home/deploy","shell":"/bin/bash","groups":["sudo","docker"],"locked":false}],"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `-j, --json`   | Output raw JSON response                                 |         |
