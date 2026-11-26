# Get

Get one certificate of a host.

Shows one certificate in full: its names, its fingerprint and its algorithms.
With no hostname, it is a certificate of the controller's host, such as api.

```bash
osapi client pki cert get [flags]
```

## Examples

```bash
osapi client pki cert get --name api
osapi client pki cert get --name agent-metrics --hostname web-01
```

## Result

The command prints:

```ansi
$ osapi client pki cert get --name agent-metrics --hostname web-01
[1;38;2;0;213;255mCert:[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
  [38;2;10;143;176mHostname:[0m       web-01
  [38;2;10;143;176mName:[0m           agent-metrics
  [38;2;10;143;176mState:[0m          ok
  [38;2;10;143;176mComponent:[0m      agent
  [38;2;10;143;176mSource:[0m         self-signed
  [38;2;10;143;176mFile:[0m           /etc/osapi/pki/agent/agent-metrics.crt
  [38;2;10;143;176mNames:[0m          web-01, 10.0.0.11
  [38;2;10;143;176mExpires:[0m        2027-02-12 19:40:01
  [38;2;10;143;176mDays left:[0m      78
  [38;2;10;143;176mFingerprint:[0m    SHA256:3f1c0d9e4b2a7c6d5e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d
  [38;2;10;143;176mKey:[0m            ECDSA P-256
  [38;2;10;143;176mSignature:[0m      ECDSA-SHA256
  [38;2;10;143;176mFIPS approved:[0m  yes
```

With `--json`, the API's response:

```json
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
```

## Permission

It needs the `pki:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag         | Description                                                                                               | Default |
| ------------ | --------------------------------------------------------------------------------------------------------- | ------- |
| `--hostname` | The host of the certificate. Empty is the controller's host, which has the certificates of the controller |         |
| `--name`     | The name of the certificate, such as api or agent-metrics (required)                                      |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
