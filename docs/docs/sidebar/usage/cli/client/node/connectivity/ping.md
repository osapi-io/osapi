# Ping

Ping an address from the host.

Sends three ICMP echo requests from the host and reports how many came back, and
their round-trip times. A host that does not answer is 100% loss, not an error.
It changes nothing, so it reports changed: false. On Linux the agent pings
without root when net.ipv4.ping_group_range admits its group.

```bash
osapi client node connectivity ping [flags]
```

## Examples

```bash
# A host that answers
osapi client node connectivity ping --target web-01 --address 127.0.0.1

# A host that answers, on macOS
osapi client node connectivity ping --target web-01 --address 127.0.0.1

# A host that does not answer
osapi client node connectivity ping --target web-01 --address 192.0.2.1
```

## Result

A host that answers, on one host:

```ansi
$ osapi client node connectivity ping --target web-01 --address 127.0.0.1
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPACKETS SENT  [0m[38;2;10;143;176mPACKETS RECEIVED  [0m[38;2;10;143;176mPACKET LOSS  [0m[38;2;10;143;176mMIN RTT  [0m[38;2;10;143;176mAVG RTT  [0m[38;2;10;143;176mMAX RTT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      3             3                 0.00         0.04ms   0.05ms   0.06ms

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node connectivity ping --target _all --address 127.0.0.1
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPACKETS SENT  [0m[38;2;10;143;176mPACKETS RECEIVED  [0m[38;2;10;143;176mPACKET LOSS  [0m[38;2;10;143;176mMIN RTT  [0m[38;2;10;143;176mAVG RTT  [0m[38;2;10;143;176mMAX RTT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      3             3                 0.00         0.04ms   0.05ms   0.06ms
  [1mweb-02[0m    [38;2;103;234;147mok[0m      3             3                 0.00         0.04ms   0.05ms   0.06ms

  3 hosts: [38;2;103;234;147m2 ok[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "avg_rtt": "0.05ms",
      "changed": false,
      "hostname": "web-01",
      "max_rtt": "0.06ms",
      "min_rtt": "0.04ms",
      "packet_loss": 0,
      "packets_received": 3,
      "packets_sent": 3,
      "status": "ok"
    }
  ]
}
```

A host that answers, on macOS, on one host:

```ansi
$ osapi client node connectivity ping --target web-01 --address 127.0.0.1
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPACKETS SENT  [0m[38;2;10;143;176mPACKETS RECEIVED  [0m[38;2;10;143;176mPACKET LOSS  [0m[38;2;10;143;176mMIN RTT  [0m[38;2;10;143;176mAVG RTT  [0m[38;2;10;143;176mMAX RTT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      3             3                 0.00         0.08ms   0.14ms   0.17ms

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node connectivity ping --target _all --address 127.0.0.1
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPACKETS SENT  [0m[38;2;10;143;176mPACKETS RECEIVED  [0m[38;2;10;143;176mPACKET LOSS  [0m[38;2;10;143;176mMIN RTT  [0m[38;2;10;143;176mAVG RTT  [0m[38;2;10;143;176mMAX RTT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      3             3                 0.00         0.08ms   0.14ms   0.17ms
  [1mweb-02[0m    [38;2;103;234;147mok[0m      3             3                 0.00         0.08ms   0.14ms   0.17ms

  3 hosts: [38;2;103;234;147m2 ok[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "avg_rtt": "0.14ms",
      "changed": false,
      "hostname": "web-01",
      "max_rtt": "0.17ms",
      "min_rtt": "0.08ms",
      "packet_loss": 0,
      "packets_received": 3,
      "packets_sent": 3,
      "status": "ok"
    }
  ]
}
```

A host that does not answer, on one host:

```ansi
$ osapi client node connectivity ping --target web-01 --address 192.0.2.1
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPACKETS SENT  [0m[38;2;10;143;176mPACKETS RECEIVED  [0m[38;2;10;143;176mPACKET LOSS  [0m[38;2;10;143;176mMIN RTT  [0m[38;2;10;143;176mAVG RTT  [0m[38;2;10;143;176mMAX RTT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      3             0                 100.00

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node connectivity ping --target _all --address 192.0.2.1
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPACKETS SENT  [0m[38;2;10;143;176mPACKETS RECEIVED  [0m[38;2;10;143;176mPACKET LOSS  [0m[38;2;10;143;176mMIN RTT  [0m[38;2;10;143;176mAVG RTT  [0m[38;2;10;143;176mMAX RTT[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      3             0                 100.00
  [1mweb-02[0m    [38;2;103;234;147mok[0m      3             0                 100.00

  3 hosts: [38;2;103;234;147m2 ok[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "changed": false,
      "hostname": "web-01",
      "packet_loss": 100,
      "packets_received": 0,
      "packets_sent": 3,
      "status": "ok"
    }
  ]
}
```

The command fails in these cases:

| Call                   | Flags                   | Fails with |
| ---------------------- | ----------------------- | ---------- |
| A name, not an address | `--address example.com` | ip_or_fact |

## Permission

It needs the `connectivity:execute` permission.
`osapi controller role get --name <role>` shows what a role grants.

## Flags

| Flag           | Description                                                                                 | Default |
| -------------- | ------------------------------------------------------------------------------------------- | ------- |
| `--address`    | The IPv4 or IPv6 address to ping, or an @fact. reference that the agent resolves (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                                      | `_all`  |

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `http://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
