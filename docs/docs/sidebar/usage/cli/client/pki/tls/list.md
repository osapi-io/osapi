# List

List the TLS of each component.

Shows for the controller, the NATS server and each agent its FIPS 140-3 mode and
the reasons it is not approved, its certificates with the days left, its TLS
listeners with their suites and groups, and its crypto. The summary is the line
that the tables of health status and agent list link to.

```bash
osapi client pki tls list [flags]
```

## Examples

```bash
osapi client pki tls list --component nats
osapi client pki tls list --hostname web-01
osapi client pki tls list
```

## Result

The command prints:

```ansi
$ osapi client pki tls list
[1;38;2;0;213;255mTLS (1):[0m
  [38;2;10;143;176mCOMPONENT   [0m[38;2;10;143;176mHOSTNAME          [0m[38;2;10;143;176mSUMMARY              [0m[38;2;10;143;176mFIPS  [0m[38;2;10;143;176mEXPIRY[0m
  controller  [1mosapi-controller[0m  FIPS off · cert 78d  off   [38;2;103;234;147mok[0m

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
{
  "items": [
    {
      "component": "controller",
      "hostname": "osapi-controller",
      "summary": "FIPS off · cert 78d",
      "fips": "off",
      "expiry": "ok",
      "reasons": ["FIPS 140-3 mode is off: set GODEBUG=fips140=on"],
      "certificates": [
        {
          "name": "api",
          "state": "ok",
          "subject": "CN=osapi-controller",
          "names": ["osapi-controller", "127.0.0.1"],
          "issuer": "CN=osapi-controller",
          "source": "self-signed",
          "not_before": "2026-02-12T19:40:01Z",
          "expires": "2027-02-12T19:40:01Z",
          "days_left": 78,
          "expiry": "ok",
          "key": "ECDSA P-256",
          "fips_approved": true
        }
      ],
      "listeners": [
        {
          "name": "api",
          "tls": true,
          "min_version": "TLS 1.2",
          "suites": [
            "TLS_AES_128_GCM_SHA256",
            "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
          ],
          "groups": ["SecP256r1MLKEM768", "CurveP256"],
          "approved": true
        }
      ],
      "go": "go1.27.3",
      "module": "v1.0.0",
      "algorithms": [
        {
          "name": "secrets sealed to agents",
          "algorithm": "ECDH P-256, HKDF-SHA256, AES-256-GCM",
          "approved": true,
          "only_mode": true
        },
        {
          "name": "controller vault",
          "algorithm": "AES-256-GCM",
          "approved": true,
          "only_mode": true
        }
      ]
    }
  ],
  "total_items": 1
}
```

## Permission

It needs the `pki:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag          | Description                                                                | Default |
| ------------- | -------------------------------------------------------------------------- | ------- |
| `--component` | Only this component: controller, nats or agent (controller, nats, agent)   |         |
| `--hostname`  | Only the components of this host: the controller's hostname, or an agent's |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
