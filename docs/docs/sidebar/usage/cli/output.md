---
sidebar_position: 1
---

# Output

Each command that sends a job prints its result in the same form. Thus this page
shows the form one time, and the page of each command does not show it again.

## One host

A table has one row for each host. A summary line under it counts the hosts by
status. The line ends with the ID of the job and the ID of the trace of the
request:

```text
$ osapi client node hostname get --target web-01

Results:
  HOSTNAME  STATUS  HOSTNAME
  web-01    ok      web-01

  1 host: 1 ok   job 550e8400-e29b-41d4-a716-446655440000   trace 26e448c9dbe1c4a8d216a98c675079c9
```

`osapi client job get --id <job>` shows the job. Put the trace ID in the search
box of your tracing backend, for example Jaeger, to see the request on the
controller and on each agent.
[Distributed tracing](../../features/distributed-tracing.md) tells how to read
it. The trace ID is not there when the controller does not trace.

## More than one host

A target such as `_all` or `group:web` gives one row for each host that
answered, in the order of the names:

```text
$ osapi client node hostname get --target _all

Results:
  HOSTNAME  STATUS  HOSTNAME
  db-01     ok      db-01
  web-01    ok      web-01
  web-02    ok      web-02

  3 hosts: 3 ok   job 550e8400-e29b-41d4-a716-446655440000   trace 26e448c9dbe1c4a8d216a98c675079c9
```

## A host that skips or fails

A host gives `skipped` when the operation does not run on its platform, and
`failed` when the operation failed. Such a host has no row in the table. The
summary line counts it, and the **Details** section gives the reason:

```text
$ osapi client node sysctl list --target _all

Results:
  HOSTNAME  STATUS  KEY            VALUE
  web-01    ok      vm.swappiness  10

  3 hosts: 1 ok, 1 skipped, 1 failed   job 550e8400-e29b-41d4-a716-446655440000   trace 26e448c9dbe1c4a8d216a98c675079c9

Details:
  mac-01  skipped  operation not supported on this OS family
  web-02  failed   permission denied
```

## A write

A write that changed the host gives the status `changed`. When you run it again
with the same input, it gives `ok`, because nothing changed:

```text
$ osapi client node sysctl create --target web-01 \
    --key net.ipv4.ip_forward --value 1

Results:
  HOSTNAME  STATUS   KEY
  web-01    changed  net.ipv4.ip_forward

  1 host: 1 changed   job 550e8400-e29b-41d4-a716-446655440000   trace 26e448c9dbe1c4a8d216a98c675079c9
```

## JSON

With `--json`, a command prints the response body of the API, with no table and
no color. A script reads it. The trace ID is not in the body:

```bash
$ osapi client node schedule list --target web-01 --json
{"results":[{"name":"backup-daily","schedule":"0 2 * * *","user":"root","object":"backup-script"}],"job_id":"..."}
```
