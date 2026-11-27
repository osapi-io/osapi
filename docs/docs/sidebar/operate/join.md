---
sidebar_position: 4
title: 4. Join hosts
---

# 4. Join hosts

Each managed host runs an agent. An agent gets no job until the controller
admits it. There are two ways to admit an agent:

- **A join token.** You make the token on the controller, and give the host its
  join string. The token admits the host at once. The host needs the join string
  and nothing else. Use it for each fleet, and for hosts that come up on their
  own.
- **An admin admits each host.** The agent has its NATS settings in its config,
  and asks to join. It waits as `pending` until an admin admits it, with the
  fingerprint that the host shows.

## Join with a token

### 1. Make a join token

On the host of the controller:

```bash
osapi controller join token create --expires 24h --uses 50 --labels group=web
```

The command prints the join string one time. Each agent that joins with it gets
the labels of the token. The string has the address of NATS, the namespace, the
fingerprint of the server's certificate, and a bootstrap user that can send the
join request and nothing else. The address is `controller.join.url`.

```bash
osapi controller join token list                # expiry and uses left
osapi controller join token revoke --id <id>    # it admits no more hosts
```

`--transport https` makes a string for a join over the controller's API in place
of NATS. That agent needs its NATS settings in its config, and the API needs
TLS.

### 2. Give the host the join string

On the host, `/etc/osapi/osapi.yaml` needs only the join string:

```yaml
agent:
  join: tls://nats.example.com:4222?creds=...&fp=SHA256%3A...&ns=osapi&token=osapi_jt_...
```

`agent.labels` in this file adds labels to the labels of the token. A label of
the file has priority over a label of the token with the same key.

### 3. Start the agent

The agent keeps its keys in `agent.pki.key_dir`, `/etc/osapi/pki` by default.
Thus it runs as root, or as a user that owns that directory.

```bash
sudo osapi agent start
```

At its first start, the agent:

1. makes its signing key and its encryption key;
1. connects to NATS as the bootstrap user, and checks the server's certificate
   against the fingerprint of the string;
1. sends the join request, signed with its key;
1. gets the controller's key and its own NATS user, and connects as that user.

Then it takes jobs. A join that fails, such as with an expired token, stops the
agent with the reason.

### What the host keeps

The join string is used one time. After the join, the agent keeps what it needs
in `agent.pki.key_dir`:

| File                      | What it is                                                           |
| ------------------------- | -------------------------------------------------------------------- |
| `agent.key`, `agent.pub`  | The signing key of the agent. The controller admitted it.            |
| `agent.ekey`              | The encryption key. Secrets sealed for this host open with it.       |
| `enrolled-controller.pub` | The controller's key. The agent checks each job with it.             |
| `nats.creds`              | Its NATS user, in operator mode. It renews it before it expires.     |
| `nats-join.json`          | Its NATS server and user, for static users.                          |
| `nats-join-ca.pem`        | The certificate that the server showed, when no public CA issued it. |

After the join, the agent checks the NATS server's certificate in one of two
ways:

- A CA that the host trusts, such as Let's Encrypt, issued it. The agent keeps
  the name of the certificate as `agent.nats.tls.server_name`, and checks the
  certificate against the host's roots. Thus a renewal of the certificate
  changes nothing. The agent can dial an address that the certificate does not
  name, such as a private address of the server.
- No such CA issued it, as for osapi's self-signed certificate. The agent trusts
  `nats-join-ca.pem`, the certificate that the join string pinned. A new
  certificate needs a join again of each agent.

A `server_name` or a `ca_file` in the agent's own config wins over both.

Each file with a secret has mode 0600. At each later start, the agent reads
these files and connects. It does not join again, and it uses no token. Thus a
reboot needs nothing from you, when the agent runs as a service. Do not delete
this directory: without it, the host has no key that the controller admitted.

## Run the agent as a service

One command, as root, writes the config with the join string, installs the
service and starts it, then waits up to a minute for the join:

```bash
sudo osapi agent service install --join '<the join string>' --labels group=web
sudo osapi agent service uninstall            # stop it and remove the service
sudo osapi agent service uninstall --purge    # also the config and the keys
```

It writes `/etc/osapi/osapi.yaml` (mode 0600) and a systemd unit on Linux, or a
launch daemon on macOS, as below. It does not replace a config that exists,
unless `--force`. To write them yourself, install the binary in
`/usr/local/bin`, then add one of these.

On Linux, `/etc/systemd/system/osapi-agent.service`:

```ini
[Unit]
Description=osapi agent
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/osapi agent start
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable --now osapi-agent
systemctl status osapi-agent
```

On macOS, `/Library/LaunchDaemons/io.osapi.agent.plist`. A launch daemon runs as
root, at boot:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>io.osapi.agent</string>
  <key>ProgramArguments</key>
  <array>
    <string>/usr/local/bin/osapi</string>
    <string>agent</string>
    <string>start</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>/var/log/osapi-agent.log</string>
  <key>StandardErrorPath</key>
  <string>/var/log/osapi-agent.log</string>
</dict>
</plist>
```

```bash
sudo launchctl bootstrap system /Library/LaunchDaemons/io.osapi.agent.plist
sudo launchctl print system/io.osapi.agent
```

To run the agent as a user that is not root, with sudo for each write, use the
unit of [Agent hardening](secure/agent-hardening.md#systemd-unit-file).

## Admit each host by hand

An agent with no join string, and with its NATS settings in its config, asks to
join at its first start. It asks again at each heartbeat until an admin decides.

```bash
osapi pki key get --name agent                      # on the host: its fingerprint, and the admit command
osapi client admission list --state pending         # the hosts that wait
osapi client admission admit --machine-id <id> --fingerprint SHA256:...
osapi client admission reject --machine-id <id> --reason "unknown host"
```

Compare the fingerprint that the host shows with the fingerprint of the pending
record before you admit it. `controller.pki.auto_admit: true` admits each host
with no admin. Use it only on a test fleet.

## Hosts from an image

- Ship the image with an empty `/etc/machine-id`, so that systemd makes one at
  the first boot. The machine ID is the identity of the agent. Two clones with
  one machine ID are one host to osapi.
- Put the join string in the image or in the launch template, with a token that
  has enough uses and an expiry later than the life of the image.
- Target the hosts by their labels, such as `--target group:web`, not by their
  hostnames.
- Set `controller.pki.delete_after`, such as `24h`. Then a host that went away
  leaves no record.

## Remove a host

```bash
osapi client admission revoke --machine-id <id>   # you do not trust it: its key stops at once
osapi client admission delete --machine-id <id>   # it goes on purpose, such as before a rebuild
```

After `delete`, the host is new to osapi. A rebuilt host joins again with a join
string. On a host that you do not rebuild, stop the agent and empty its
`agent.pki.key_dir` first. Then it makes new keys, and uses the join string of
its config with a token that has a use left.

## Move hosts to a different NATS server

Give NATS a DNS name, and set `controller.join.url` to that name before the
hosts join. Each agent keeps the name of its join in `nats-join.json`. It
reconnects to NATS with no limit, and it looks up the name again at each
attempt. Thus NATS can move to a different host with no join, when the new host:

- has the same DNS name,
- serves the same certificate, or a certificate that a root of the agent's
  system verifies for that name, and
- has the JetStream store of the old host, from
  `nats.server.jetstream.store_dir`.

The certificate must have the DNS name in its subject alternative names. With a
self-signed certificate, the agent pins that certificate, so the new host must
serve it.

osapi does not cluster NATS servers. Thus you cannot add the new server beside
the old one and then remove the old one.

A move to a different name, or to a different self-signed certificate, needs a
join of each host. The controller does not push NATS settings to an agent:

1. Set `controller.join.url` to the new address of NATS, and make a token with a
   use for each host, and the labels of the hosts:

   ```bash
   osapi controller join token create --uses 10 --labels group=web
   ```

2. On each host, stop the agent.
3. Remove `nats-join.json` from `agent.pki.key_dir`. Do not remove the other
   files: with its keys, the host stays the same machine to osapi.
4. Set `agent.join` to the new join string, and start the agent.

The agent finds no `nats-join.json`, so it joins with the new string and the
keys that it has. The controller admits it again, and the token gives its
labels.

A new certificate on the same address does not need a join.
`osapi client pki nats rotate` gives it to each agent.

## Check it

```bash
osapi client admission list                      # each host, its state, online or not
osapi client admission get --machine-id <id>     # the token that admitted it, as join_token
osapi client agent list                          # each agent that runs, with its labels
osapi client agent get --hostname <host>         # its timeline: admitted, started
```

On the Admission page of the UI, a red dot next to a host means that it is
offline: it has not reported for some time. [Troubleshoot](troubleshoot.md) has
the causes.

Next: [Run the fleet](fleet.md).
