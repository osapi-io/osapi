# Joining a host

A host joins with a join string, which a join token makes on the controller's
host. The token admits the host at once, and gives it the token's labels.

```sh
osapi controller join token create --expires 24h --uses 50 --labels group=web --json
osapi controller join token list --json
osapi controller join token revoke --id <id>
```

On the host, as root, one command writes the config with the join string,
installs the agent as a service that starts at boot (systemd on Linux, launchd
on macOS), starts it, and waits up to a minute for the join:

```sh
sudo osapi agent service install --join '<join string>' --labels group=web --json
sudo osapi agent service uninstall --json           # stop and remove the service
sudo osapi agent service uninstall --purge --json   # also the config and the keys
```

`joined: false` means the host did not join in a minute: read the service's log
(`journalctl -u osapi-agent`, or `/var/log/osapi-agent.log` on macOS). After the
join, the host keeps its keys and its NATS user in `agent.pki.key_dir`, and each
start reads them: the join string is used one time.

A host with no join string asks to be admitted, and waits as `pending`:

```sh
osapi pki key get --name agent --json                    # on the host
osapi client admission list --state pending --json
osapi client admission admit --machine-id <id> --fingerprint <fp>
```

Remove a host with `osapi client admission revoke --machine-id <id>` (it is not
trusted) or `delete` (it goes on purpose, such as before a rebuild). The docs'
page Operate → Join hosts has the rest.
