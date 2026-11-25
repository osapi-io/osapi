# Get

Print the config file: each key, what it does, its type, default and rules.

The list comes from this binary. Thus it agrees with the build that you run. A
section for each top word of a key, such as controller.
`osapi config get --key <key>` shows one key with its value in effect.

```bash
osapi config spec get
```

## Examples

```bash
osapi config spec get
osapi config spec get --json
```

## Result

The command prints:

```ansi
$ osapi config spec get
[1;38;2;0;213;255mcontroller:[0m
  [1mcontroller.client.url[0m
    Base URL the CLI client connects to. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_CLIENT_URL, which wins over[0m
    [38;2;138;138;138mthe file.[0m

  [1mcontroller.client.security.bearer_token[0m
    JWT bearer token for client requests. Generate with osapi controller
    token generate. It can be a reference, secret://<vault>/<key>, which the
    command resolves from the vaults of its config. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with[0m
    [38;2;138;138;138mOSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN, which wins over the file.[0m
    [38;2;138;138;138mRules: omitempty. A secret: osapi prints it as ********.[0m

  [1mcontroller.client.tls.ca_file[0m
    A PEM certificate that the CLI trusts on an https:// URL, such as
    api.crt of a controller with a self-signed certificate, in its
    controller.pki.key_dir. Without it, the system's trusted roots. Type:
    string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_CLIENT_TLS_CA_FILE, which[0m
    [38;2;138;138;138mwins over the file.[0m

  [1mcontroller.api.host[0m
    Host the REST API server, and the UI on its listener, binds to. Empty
    binds every interface. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_API_HOST, which wins over[0m
    [38;2;138;138;138mthe file.[0m

  [1mcontroller.api.url[0m
    Where a person reaches the API, shown by health beside the bind address.
    Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_API_URL, which wins over[0m
    [38;2;138;138;138mthe file. Rules: omitempty,http_url.[0m

  [1mcontroller.api.port[0m
    Port the REST API server listens on. Type: int.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_API_PORT, which wins over[0m
    [38;2;138;138;138mthe file. Rules: min=1,max=65535.[0m

  [1mcontroller.api.security.cors.allow_origins[0m
    Origins allowed to make cross-origin requests. An empty list disables
    CORS headers entirely. Type: list.
    [38;2;138;138;138mUse: In osapi.yaml, or with[0m
    [38;2;138;138;138mOSAPI_CONTROLLER_API_SECURITY_CORS_ALLOW_ORIGINS, which wins over the[0m
    [38;2;138;138;138mfile.[0m

  [1mcontroller.api.security.signing_keys[0m
    HS256 keys for JWTs, each an id and a key. Without one, the controller
    makes a key in controller.pki.key_dir at its first start, which
    controller token generate on its host also uses. The first key signs;
    each key verifies the tokens with its id as kid. The environment
    variable takes id=key pairs with commas between them. Generate a key
    with openssl rand -hex 32. A key can be a reference,
    secret://<vault>/<key>, which the controller resolves at startup. Type:
    list.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_API_SECURITY_SIGNING_KEYS,[0m
    [38;2;138;138;138mwhich wins over the file. Rules: omitempty,unique=ID,dive. A secret:[0m
    [38;2;138;138;138mosapi prints it as ********.[0m

  [1mcontroller.api.security.roles[0m
    Custom roles with fine-grained permissions. Type: map.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_API_SECURITY_ROLES, which[0m
    [38;2;138;138;138mwins over the file. Rules: dive.[0m

  [1mcontroller.api.tls.enabled[0m
    Serve HTTPS on the API's port, in place of HTTP. Without cert_file and
    key_file, the controller makes a self-signed certificate in
    controller.pki.key_dir. Type: bool. Default: false.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_API_TLS_ENABLED, which wins[0m
    [38;2;138;138;138mover the file.[0m

  [1mcontroller.api.tls.cert_file[0m
    The PEM certificate to serve, with its chain, such as one from tailscale
    cert or certbot. Set it with key_file, or neither for the self-signed
    certificate. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_API_TLS_CERT_FILE, which[0m
    [38;2;138;138;138mwins over the file. Rules: required_with=KeyFile.[0m

  [1mcontroller.api.tls.key_file[0m
    The PEM private key of cert_file. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_API_TLS_KEY_FILE, which[0m
    [38;2;138;138;138mwins over the file. Rules: required_with=CertFile.[0m

  [1mcontroller.api.job_timeout[0m
    How long to wait for agent responses before returning partial results.
    Go duration format. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_API_JOB_TIMEOUT, which wins[0m
    [38;2;138;138;138mover the file. Rules: omitempty,go_duration.[0m

  [1mcontroller.nats.host[0m
    NATS server hostname to connect to. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_NATS_HOST, which wins over[0m
    [38;2;138;138;138mthe file.[0m

  [1mcontroller.nats.port[0m
    NATS server port to connect to. Type: int.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_NATS_PORT, which wins over[0m
    [38;2;138;138;138mthe file. Rules: min=1,max=65535.[0m

  [1mcontroller.nats.client_name[0m
    Client name sent to NATS for identification. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_NATS_CLIENT_NAME, which[0m
    [38;2;138;138;138mwins over the file.[0m

  [1mcontroller.nats.namespace[0m
    Subject namespace prefix. Must match nats.server.namespace. Type:
    string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_NATS_NAMESPACE, which wins[0m
    [38;2;138;138;138mover the file. Rules: omitempty,nats_name.[0m

  [1mcontroller.nats.auth.type[0m
    Authentication type: 'none', 'user_pass', 'nkey', or 'creds'. Type:
    string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_NATS_AUTH_TYPE, which wins[0m
    [38;2;138;138;138mover the file.[0m

  [1mcontroller.nats.auth.username[0m
    Username for user_pass auth. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_NATS_AUTH_USERNAME, which[0m
    [38;2;138;138;138mwins over the file.[0m

  [1mcontroller.nats.auth.password[0m
    Password for user_pass auth. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_NATS_AUTH_PASSWORD, which[0m
    [38;2;138;138;138mwins over the file. A secret: osapi prints it as ********.[0m

  [1mcontroller.nats.auth.nkey_file[0m
    Path to the NKey seed file for nkey auth. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_NATS_AUTH_NKEY_FILE, which[0m
    [38;2;138;138;138mwins over the file.[0m

  [1mcontroller.nats.auth.creds_file[0m
    Path to a NATS credentials file (user JWT and seed) for creds auth.
    Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_NATS_AUTH_CREDS_FILE, which[0m
    [38;2;138;138;138mwins over the file.[0m

  [1mcontroller.metrics.enabled[0m
    Enable the metrics endpoint. Type: bool. Default: false.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_METRICS_ENABLED, which wins[0m
    [38;2;138;138;138mover the file.[0m

  [1mcontroller.metrics.host[0m
    Host the metrics server binds to. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_METRICS_HOST, which wins[0m
    [38;2;138;138;138mover the file.[0m

  [1mcontroller.metrics.port[0m
    Port the metrics server listens on. Type: int.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_METRICS_PORT, which wins[0m
    [38;2;138;138;138mover the file. Rules: omitempty,min=1,max=65535.[0m

  [1mcontroller.metrics.url[0m
    Where a person opens the metrics endpoint, shown by health beside the
    bind address. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_METRICS_URL, which wins[0m
    [38;2;138;138;138mover the file. Rules: omitempty,http_url.[0m

  [1mcontroller.notifications.enabled[0m
    Enable the condition watcher and notifier. Type: bool. Default: false.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_NOTIFICATIONS_ENABLED,[0m
    [38;2;138;138;138mwhich wins over the file.[0m

  [1mcontroller.notifications.notifier[0m
    Notifier backend: 'log' (default). Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_NOTIFICATIONS_NOTIFIER,[0m
    [38;2;138;138;138mwhich wins over the file.[0m

  [1mcontroller.notifications.renotify_interval[0m
    How often to re-fire active conditions (Go duration). Zero disables
    re-notification. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with[0m
    [38;2;138;138;138mOSAPI_CONTROLLER_NOTIFICATIONS_RENOTIFY_INTERVAL, which wins over the[0m
    [38;2;138;138;138mfile. Rules: omitempty,go_duration.[0m

  [1mcontroller.ui.enabled[0m
    Serve the embedded UI. Type: bool. Default: false.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_UI_ENABLED, which wins over[0m
    [38;2;138;138;138mthe file.[0m

  [1mcontroller.ui.url[0m
    Where a person opens the UI, shown by health beside the bind address.
    Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_UI_URL, which wins over the[0m
    [38;2;138;138;138mfile. Rules: omitempty,http_url.[0m

  [1mcontroller.pki.key_dir[0m
    Directory for controller keypair storage. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_PKI_KEY_DIR, which wins[0m
    [38;2;138;138;138mover the file.[0m

  [1mcontroller.pki.auto_admit[0m
    Admit each machine that asks to join, with no operator (dev/test only).
    Type: bool. Default: false.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_PKI_AUTO_ADMIT, which wins[0m
    [38;2;138;138;138mover the file.[0m

  [1mcontroller.pki.rotation_grace_period[0m
    Grace period during key rotation where both old and new keys are
    accepted. Go duration format. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_PKI_ROTATION_GRACE_PERIOD,[0m
    [38;2;138;138;138mwhich wins over the file. Rules: omitempty,go_duration.[0m

  [1mcontroller.pki.delete_after[0m
    Delete an admitted machine that has not reported for this long, such as
    720h. Empty is never. Go duration format. Type: string.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_PKI_DELETE_AFTER, which[0m
    [38;2;138;138;138mwins over the file. Rules: omitempty,go_duration.[0m

  [1mcontroller.secrets.vaults[0m
    The vaults that a reference such as secret://default/registry-token
    names, each with name, type, targets, key_file or key_env, and config.
    Type: list.
    [38;2;138;138;138mUse: In osapi.yaml, or with OSAPI_CONTROLLER_SECRETS_VAULTS, which wins[0m
    [38;2;138;138;138mover the file. Rules: dive.[0m
```

With `--json`, the command prints:

```json
{
  "kinds": [
    {
      "kind": "controller",
      "title": "controller",
      "terms": [
        {
          "kind": "controller",
          "name": "controller.client.url",
          "does": "Base URL the CLI client connects to. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_CLIENT_URL, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.client.security.bearer_token",
          "does": "JWT bearer token for client requests. Generate with osapi controller token generate. It can be a reference, secret://\u003cvault\u003e/\u003ckey\u003e, which the command resolves from the vaults of its config. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN, which wins over the file. Rules: omitempty. A secret: osapi prints it as ********."
        },
        {
          "kind": "controller",
          "name": "controller.client.tls.ca_file",
          "does": "A PEM certificate that the CLI trusts on an https:// URL, such as api.crt of a controller with a self-signed certificate, in its controller.pki.key_dir. Without it, the system's trusted roots. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_CLIENT_TLS_CA_FILE, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.api.host",
          "does": "Host the REST API server, and the UI on its listener, binds to. Empty binds every interface. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_API_HOST, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.api.url",
          "does": "Where a person reaches the API, shown by health beside the bind address. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_API_URL, which wins over the file. Rules: omitempty,http_url."
        },
        {
          "kind": "controller",
          "name": "controller.api.port",
          "does": "Port the REST API server listens on. Type: int.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_API_PORT, which wins over the file. Rules: min=1,max=65535."
        },
        {
          "kind": "controller",
          "name": "controller.api.security.cors.allow_origins",
          "does": "Origins allowed to make cross-origin requests. An empty list disables CORS headers entirely. Type: list.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_API_SECURITY_CORS_ALLOW_ORIGINS, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.api.security.signing_keys",
          "does": "HS256 keys for JWTs, each an id and a key. Without one, the controller makes a key in controller.pki.key_dir at its first start, which controller token generate on its host also uses. The first key signs; each key verifies the tokens with its id as kid. The environment variable takes id=key pairs with commas between them. Generate a key with openssl rand -hex 32. A key can be a reference, secret://\u003cvault\u003e/\u003ckey\u003e, which the controller resolves at startup. Type: list.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_API_SECURITY_SIGNING_KEYS, which wins over the file. Rules: omitempty,unique=ID,dive. A secret: osapi prints it as ********."
        },
        {
          "kind": "controller",
          "name": "controller.api.security.roles",
          "does": "Custom roles with fine-grained permissions. Type: map.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_API_SECURITY_ROLES, which wins over the file. Rules: dive."
        },
        {
          "kind": "controller",
          "name": "controller.api.tls.enabled",
          "does": "Serve HTTPS on the API's port, in place of HTTP. Without cert_file and key_file, the controller makes a self-signed certificate in controller.pki.key_dir. Type: bool. Default: false.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_API_TLS_ENABLED, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.api.tls.cert_file",
          "does": "The PEM certificate to serve, with its chain, such as one from tailscale cert or certbot. Set it with key_file, or neither for the self-signed certificate. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_API_TLS_CERT_FILE, which wins over the file. Rules: required_with=KeyFile."
        },
        {
          "kind": "controller",
          "name": "controller.api.tls.key_file",
          "does": "The PEM private key of cert_file. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_API_TLS_KEY_FILE, which wins over the file. Rules: required_with=CertFile."
        },
        {
          "kind": "controller",
          "name": "controller.api.job_timeout",
          "does": "How long to wait for agent responses before returning partial results. Go duration format. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_API_JOB_TIMEOUT, which wins over the file. Rules: omitempty,go_duration."
        },
        {
          "kind": "controller",
          "name": "controller.nats.host",
          "does": "NATS server hostname to connect to. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_NATS_HOST, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.nats.port",
          "does": "NATS server port to connect to. Type: int.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_NATS_PORT, which wins over the file. Rules: min=1,max=65535."
        },
        {
          "kind": "controller",
          "name": "controller.nats.client_name",
          "does": "Client name sent to NATS for identification. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_NATS_CLIENT_NAME, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.nats.namespace",
          "does": "Subject namespace prefix. Must match nats.server.namespace. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_NATS_NAMESPACE, which wins over the file. Rules: omitempty,nats_name."
        },
        {
          "kind": "controller",
          "name": "controller.nats.auth.type",
          "does": "Authentication type: 'none', 'user_pass', 'nkey', or 'creds'. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_NATS_AUTH_TYPE, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.nats.auth.username",
          "does": "Username for user_pass auth. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_NATS_AUTH_USERNAME, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.nats.auth.password",
          "does": "Password for user_pass auth. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_NATS_AUTH_PASSWORD, which wins over the file. A secret: osapi prints it as ********."
        },
        {
          "kind": "controller",
          "name": "controller.nats.auth.nkey_file",
          "does": "Path to the NKey seed file for nkey auth. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_NATS_AUTH_NKEY_FILE, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.nats.auth.creds_file",
          "does": "Path to a NATS credentials file (user JWT and seed) for creds auth. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_NATS_AUTH_CREDS_FILE, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.metrics.enabled",
          "does": "Enable the metrics endpoint. Type: bool. Default: false.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_METRICS_ENABLED, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.metrics.host",
          "does": "Host the metrics server binds to. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_METRICS_HOST, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.metrics.port",
          "does": "Port the metrics server listens on. Type: int.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_METRICS_PORT, which wins over the file. Rules: omitempty,min=1,max=65535."
        },
        {
          "kind": "controller",
          "name": "controller.metrics.url",
          "does": "Where a person opens the metrics endpoint, shown by health beside the bind address. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_METRICS_URL, which wins over the file. Rules: omitempty,http_url."
        },
        {
          "kind": "controller",
          "name": "controller.notifications.enabled",
          "does": "Enable the condition watcher and notifier. Type: bool. Default: false.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_NOTIFICATIONS_ENABLED, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.notifications.notifier",
          "does": "Notifier backend: 'log' (default). Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_NOTIFICATIONS_NOTIFIER, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.notifications.renotify_interval",
          "does": "How often to re-fire active conditions (Go duration). Zero disables re-notification. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_NOTIFICATIONS_RENOTIFY_INTERVAL, which wins over the file. Rules: omitempty,go_duration."
        },
        {
          "kind": "controller",
          "name": "controller.ui.enabled",
          "does": "Serve the embedded UI. Type: bool. Default: false.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_UI_ENABLED, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.ui.url",
          "does": "Where a person opens the UI, shown by health beside the bind address. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_UI_URL, which wins over the file. Rules: omitempty,http_url."
        },
        {
          "kind": "controller",
          "name": "controller.pki.key_dir",
          "does": "Directory for controller keypair storage. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_PKI_KEY_DIR, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.pki.auto_admit",
          "does": "Admit each machine that asks to join, with no operator (dev/test only). Type: bool. Default: false.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_PKI_AUTO_ADMIT, which wins over the file."
        },
        {
          "kind": "controller",
          "name": "controller.pki.rotation_grace_period",
          "does": "Grace period during key rotation where both old and new keys are accepted. Go duration format. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_PKI_ROTATION_GRACE_PERIOD, which wins over the file. Rules: omitempty,go_duration."
        },
        {
          "kind": "controller",
          "name": "controller.pki.delete_after",
          "does": "Delete an admitted machine that has not reported for this long, such as 720h. Empty is never. Go duration format. Type: string.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_PKI_DELETE_AFTER, which wins over the file. Rules: omitempty,go_duration."
        },
        {
          "kind": "controller",
          "name": "controller.secrets.vaults",
          "does": "The vaults that a reference such as secret://default/registry-token names, each with name, type, targets, key_file or key_env, and config. Type: list.",
          "when": "In osapi.yaml, or with OSAPI_CONTROLLER_SECRETS_VAULTS, which wins over the file. Rules: dive."
        }
      ]
    }
  ]
}
```

## Global flags

| Flag               | Description                  | Default                 |
| ------------------ | ---------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode |                         |
| `-j, --json`       | Enable JSON output           |                         |
| `-f, --osapi-file` | Path to config file          | `/etc/osapi/osapi.yaml` |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
