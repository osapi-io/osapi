# Start

To start the server:

```bash
$ osapi controller start
2:24AM INF server configuration debug=false server.port=8080 server.security.cors.allow_origins="[http://localhost:3001 https://osapi-io.github.io]"
⇨ http server started on [::]:8080
```

## Flags

| Flag         | Description                  | Default |
| ------------ | ---------------------------- | ------- |
| `-p, --port` | Port the server will bind to | `8080`  |

## Least Privilege Mode

We aim to run this API service with the **least privilege mode** to maximize
security while reading data from Linux. This means that the API will only return
data that the running user has permission to access. The API is designed to
gracefully skip over partitions or other system resources where permission
errors occur (e.g., due to lack of root access).

To run the API with the fewest limits, run the API daemon as `root`.

When it runs as a user that is not root, the service does these things:

- Collect and return available disk usage statistics for partitions it has
  permission to access
- Skip partitions or system paths that result in "permission denied" errors
- Attempt to send an "unprivileged" ping via UDP

A regular user keeps the service in a safe, limited mode. Some functions are
limited, for example access to some system directories and files.

If you need full access to system resources, run the API daemon as `root`.
Examples are access to all disk partitions and privileged operations.

### ICMP Permissions

The API can send pings without root access. On Linux, this needs a change to a
system setting, with this sysctl command:

```bash
$ sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

Or run the API as root. Then it has full access to privileged operations, for
example pings on a raw socket.
