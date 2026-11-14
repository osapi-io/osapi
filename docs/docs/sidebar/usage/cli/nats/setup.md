# Setup

Make the JetStream streams and key-value buckets that osapi needs, on a NATS
server that osapi does not run, for example a cluster that exists:

```bash
$ osapi nats setup -f /etc/osapi/osapi.yaml
```

It connects as the controller does, with `controller.nats`: its host, port,
namespace and auth. Run it one time before you start the controller and the
agents with that server. Run it again after you change the settings of a stream
or a bucket. Each stream and bucket is created or updated to match the config,
so running it again changes nothing.

`osapi nats server start` does the same for the embedded server when the server
starts. Thus a deployment that runs the NATS server of osapi does not need this
command.

## A server in operator mode

With a NATS server in operator (decentralized JWT) mode, the controller
authenticates with a credentials file. This command does the same:

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

The account of the credentials must have JetStream enabled. The user must have
permission to make streams and key-value buckets. Agents need only their own
credentials. [Configuration](../../configuration.md#authentication) gives the
authentication types.
