# Setup

Create the JetStream streams and key-value buckets osapi needs on a NATS server
it does not run, such as an existing cluster:

```bash
$ osapi nats setup -f /etc/osapi/osapi.yaml
```

It connects as the controller does, with `controller.nats`: its host, port,
namespace and auth. Run it once before starting the controller and the agents
against that server, and again after changing a stream's or a bucket's settings.
Each stream and bucket is created or updated to match the config, so running it
again changes nothing.

`osapi nats server start` does the same for the embedded server when it starts,
so a deployment that runs osapi's own NATS does not need this.

## A server in operator mode

Against a NATS server in operator (decentralized JWT) mode, the controller
authenticates with a credentials file, and so does this command:

```yaml
controller:
  nats:
    host: nats.example.com
    port: 4222
    namespace: osapi
    auth:
      type: creds
      creds_file: /etc/osapi/controller.creds
```

The account the credentials belong to needs JetStream enabled, and the user
needs permission to create streams and key-value buckets. Agents need only their
own credentials; [Configuration](../../configuration.md#authentication) has the
auth types.
