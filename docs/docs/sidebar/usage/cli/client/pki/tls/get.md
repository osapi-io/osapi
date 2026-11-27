# Get

Get the TLS of one component.

Shows one component's TLS in full: its FIPS 140-3 mode and why it is not
approved, each certificate with its names, issuer and validity, each TLS
listener with its suites and groups, and its algorithms.

```bash
osapi client pki tls get [flags]
```

## Examples

```bash
osapi client pki tls get --component agent --hostname web-01
osapi client pki tls get --component controller
```

## Result

The command prints:

```ansi
$ osapi client pki tls get --component controller
[1;38;2;0;213;255mTLS:[0m
  [38;2;10;143;176mComponent:[0m  controller
  [38;2;10;143;176mHostname:[0m   osapi-controller
  [38;2;10;143;176mSummary:[0m    FIPS off · cert 78d
  [38;2;10;143;176mFIPS:[0m       off
  [38;2;10;143;176mExpiry:[0m     [38;2;103;234;147mok[0m
  [38;2;10;143;176mReasons:[0m    FIPS 140-3 mode is off: set GODEBUG=fips140=on
  [38;2;10;143;176mGo:[0m         go1.27.3
  [38;2;10;143;176mModule:[0m     v1.0.0

[1;38;2;255;0;255mCertificates (1):[0m
  [38;2;176;48;176mNAME  [0m[38;2;176;48;176mSTATE  [0m[38;2;176;48;176mSUBJECT              [0m[38;2;176;48;176mSAN                          [0m[38;2;176;48;176mISSUER               [0m[38;2;176;48;176mSOURCE       [0m[38;2;176;48;176mVALID FROM           [0m[38;2;176;48;176mVALID TO             [0m[38;2;176;48;176mDAYS LEFT  [0m[38;2;176;48;176mEXPIRY  [0m[38;2;176;48;176mKEY          [0m[38;2;176;48;176mFIPS APPROVED[0m
  [1mapi[0m   ok     CN=osapi-controller  osapi-controller, 127.0.0.1  CN=osapi-controller  self-signed  2026-02-12 19:40:01  2027-02-12 19:40:01  78         [38;2;103;234;147mok[0m      ECDSA P-256  yes

[1;38;2;255;209;0mListeners (1):[0m
  [38;2;176;143;0mNAME  [0m[38;2;176;143;0mTLS  [0m[38;2;176;143;0mMIN VERSION  [0m[38;2;176;143;0mSUITES                                                           [0m[38;2;176;143;0mGROUPS                        [0m[38;2;176;143;0mAPPROVED[0m
  [1mapi[0m   yes  TLS 1.2      TLS_AES_128_GCM_SHA256, TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256  SecP256r1MLKEM768, CurveP256  yes

[1;38;2;69;174;23mAlgorithms (2):[0m
  [38;2;58;143;20mNAME                      [0m[38;2;58;143;20mALGORITHM                             [0m[38;2;58;143;20mAPPROVED  [0m[38;2;58;143;20mONLY MODE[0m
  [1msecrets sealed to agents[0m  ECDH P-256, HKDF-SHA256, AES-256-GCM  yes       yes
  [1mcontroller vault[0m          AES-256-GCM                           yes       yes

  [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

With `--json`, the API's response:

```json
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
```

## Permission

It needs the `pki:read` permission. `osapi controller role get --name <role>`
shows what a role grants.

## Flags

| Flag          | Description                                                                                   | Default |
| ------------- | --------------------------------------------------------------------------------------------- | ------- |
| `--component` | The component: controller, nats or agent (controller, nats, agent) (required)                 |         |
| `--hostname`  | The host of the component. Empty is the controller's host, which an agent's section must name |         |

## Global flags

| Flag               | Description                                                                      | Default                  |
| ------------------ | -------------------------------------------------------------------------------- | ------------------------ |
| `-d, --debug`      | Enable or disable debug mode                                                     |                          |
| `-j, --json`       | Enable JSON output                                                               |                          |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml`  |
| `--url`            | Base URL the CLI client connects to (controller.client.url)                      | `https://localhost:8080` |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
