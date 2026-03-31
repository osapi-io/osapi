# Get

Get a specific user account by name:

```bash
$ osapi client node user get --target web-01 --name deploy

  NAME     UID   GID   HOME            SHELL        GROUPS       LOCKED
  deploy   1001  1001  /home/deploy    /bin/bash     sudo,docker  no
```

## JSON Output

```bash
$ osapi client node user get --target web-01 --name deploy --json
{"results":[{"hostname":"web-01","users":[{"name":"deploy","uid":1001,"gid":1001,"home":"/home/deploy","shell":"/bin/bash","groups":["sudo","docker"],"locked":false}],"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Username to look up (required)                           |         |
| `-j, --json`   | Output raw JSON response                                 |         |
