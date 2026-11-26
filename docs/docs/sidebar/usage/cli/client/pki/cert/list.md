# List

List the certificates of the fleet.

Lists the certificates of the controller and of each agent, with the days that
are left until each expires. osapi pki cert list reads only the host that it
runs on.

```bash
osapi client pki cert list [flags]
```

## Examples

```bash
osapi client pki cert list
osapi client pki cert list --hostname web-01
```

## Result

The command prints:

```ansi
$ osapi client pki cert list --hostname web-01
[1;38;2;0;213;255mCerts (2):[0m
  [38;2;10;143;176mHOSTNAME          [0m[38;2;10;143;176mNAME           [0m[38;2;10;143;176mSTATE  [0m[38;2;10;143;176mCOMPONENT   [0m[38;2;10;143;176mSOURCE       [0m[38;2;10;143;176mEXPIRES              [0m[38;2;10;143;176mDAYS LEFT  [0m[38;2;10;143;176mKEY          [0m[38;2;10;143;176mFIPS APPROVED[0m
  [1mosapi-controller[0m  [1mapi[0m            ok     controller  self-signed  2027-02-12 19:40:01  78         ECDSA P-256  yes
  [1mweb-01[0m            [1magent-metrics[0m  ok     agent       self-signed  2027-02-12 19:40:01  78         ECDSA P-256  yes

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "items": [
    {
      "hostname": "osapi-controller",
      "name": "api",
      "state": "ok",
      "component": "controller",
      "source": "self-signed",
      "file": "",
      "expires": "2027-02-12T19:40:01Z",
      "days_left": 78,
      "key": "ECDSA P-256",
      "fips_approved": true
    },
    {
      "hostname": "web-01",
      "name": "agent-metrics",
      "state": "ok",
      "component": "agent",
      "source": "self-signed",
      "file": "/etc/osapi/pki/agent/agent-metrics.crt",
      "names": ["web-01", "10.0.0.11"],
      "expires": "2027-02-12T19:40:01Z",
      "days_left": 78,
      "fingerprint": "SHA256:3f1c0d9e4b2a7c6d5e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d",
      "key": "ECDSA P-256",
      "signature": "ECDSA-SHA256",
      "fips_approved": true
    }
  ],
  "total_items": 2
}
```

## Permission

It needs the `pki:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag         | Description                                                                  | Default |
| ------------ | ---------------------------------------------------------------------------- | ------- |
| `--hostname` | Only the certificates of this host: the controller's hostname, or an agent's |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
