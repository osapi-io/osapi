---
sidebar_position: 1
---

# Output

Each command that sends a job prints its result in the same form. Thus this page
shows the form one time, and the page of each command does not show it again.

## One host

The job ID is first. Then a table has one row for each host, and a line counts
the hosts by status:

```bash
$ osapi client node hostname get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS
  web-01    ok

  1 host: 1 ok
```

## More than one host

A target such as `_all` or `group:web` gives one row for each host that
answered:

```bash
$ osapi client node hostname get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS
  server1   ok
  server2   ok

  2 hosts: 2 ok
```

## A host that skips or fails

A host gives `skip` when the operation does not run on its platform, and
`failed` when the operation failed. The **Details** section after the table
gives the reason for each one:

```bash
$ osapi client node hostname get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS
  server1   ok
  server2   skip

  2 hosts: 1 ok, 1 skipped

  Details:
  server2   unsupported platform
```

## A write

A write has a `CHANGED` column. A write that changed the host gives the status
`changed` and `true`. When you run it again with the same input, it gives
`false`:

```bash
$ osapi client node schedule update --target web-01 \
    --name backup-daily \
    --schedule "0 3 * * *"

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME          CHANGED
  web-01    changed  backup-daily  true

  1 host: 1 changed
```

## JSON

With `--json`, a command prints the response body of the API, with no table and
no color. A script reads it:

```bash
$ osapi client node schedule list --target web-01 --json
{"results":[{"name":"backup-daily","schedule":"0 2 * * *","user":"root","object":"backup-script"}],"job_id":"..."}
```
