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

<!-- osapi:output hostname get web-01 -->

```ansi
$ osapi client node hostname get --target web-01
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mweb-01[0m

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

<!-- /osapi:output -->

`osapi client job get --id <job>` shows the job. Put the trace ID in the search
box of your tracing backend, for example Jaeger, to see the request on the
controller and on each agent.
[Distributed tracing](../../features/platform/distributed-tracing.md) tells how
to read it. The trace ID is not there when the controller does not trace.

## More than one host

A target such as `_all` or `group:web` gives one row for each host that
answered, in the order of the names:

<!-- osapi:output hostname get db-01 web-01 web-02 -->

```ansi
$ osapi client node hostname get --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mNAME[0m
  [1mdb-01[0m     [38;2;103;234;147mok[0m      [1mdb-01[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      [1mweb-01[0m
  [1mweb-02[0m    [38;2;103;234;147mok[0m      [1mweb-02[0m

  3 hosts: [38;2;103;234;147m3 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

<!-- /osapi:output -->

## A host that skips or fails

A host gives `skipped` when the operation does not run on its platform, and
`failed` when the operation failed. Such a host has no row in the table. The
summary line counts it, and the **Details** section gives the reason:

<!-- osapi:output sysctl list web-01 mac-01:skipped web-02:failed -->

```ansi
$ osapi client node sysctl list --target _all
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mKEY                        [0m[38;2;10;143;176mVALUE         [0m[38;2;10;143;176mMANAGED[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      kernel.panic               0             no
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      net.ipv4.ip_forward        1             yes
  [38;2;138;138;138mweb-01[0m    [38;2;103;234;147mok[0m      net.ipv4.ping_group_range  0 2147483647  no

  3 hosts: [38;2;103;234;147m1 ok[0m, [38;5;222m1 skipped[0m, [38;2;239;68;68m1 failed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
  [1mweb-02[0m  [38;2;239;68;68mfailed [0m  permission denied
```

<!-- /osapi:output -->

## A write

A write that changed the host gives the status `changed`. When you run it again
with the same input, it gives `ok`, because nothing changed:

<!-- osapi:output sysctl create web-01 -->

```ansi
$ osapi client node sysctl create --target web-01 --key net.ipv4.ip_forward --value 1
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mKEY[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  net.ipv4.ip_forward

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

<!-- /osapi:output -->

## JSON

With `--json`, a command prints the response body of the API, with no table and
no color. A script reads it. The trace ID is not in the body:

<!-- osapi:output sysctl list web-01 --json -->

```ansi
$ osapi client node sysctl list --target web-01 --json
{"job_id":"550e8400-e29b-41d4-a716-446655440000","results":[{"hostname":"web-01","key":"kernel.panic","managed":false,"status":"ok","value":"0"},{"hostname":"web-01","key":"net.ipv4.ip_forward","managed":true,"status":"ok","value":"1"},{"hostname":"web-01","key":"net.ipv4.ping_group_range","managed":false,"status":"ok","value":"0 2147483647"}]}
```

<!-- /osapi:output -->
