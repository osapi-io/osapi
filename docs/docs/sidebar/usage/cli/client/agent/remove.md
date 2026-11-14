# Remove

Remove an enrolled agent's stored key:

```bash
$ osapi client agent remove --hostname web-01

  Status: Removed
  Message: agent web-01 (3f2a9c...) removed
```

When PKI is enabled, the controller keeps each accepted agent's public key. This
command deletes it. From that moment the controller does not verify the agent's
responses or registration, so targeting no longer resolves it and a broadcast
gives it no row.

The agent is not told. It keeps the controller key it was given at acceptance
and does not enroll again by itself. To enroll it again, delete `controller.pub`
from its key directory on the host (`/etc/osapi/pki` by default) and restart the
agent. It then sends a new request, which `agent accept` accepts.

Two machines may share a hostname, because the agent's identity is its machine
ID. When more than one accepted agent has the hostname, the command fails and
names their machine IDs. Run it again with `--machine-id`:

```bash
$ osapi client agent remove --hostname web-01 --machine-id 3f2a9c...
```

## Flags

| Flag           | Description                                                     | Required |
| -------------- | --------------------------------------------------------------- | -------- |
| `--hostname`   | Hostname of the accepted agent to remove                        | Yes      |
| `--machine-id` | Machine ID to remove, when more than one agent has the hostname | No       |
| `--json`       | Output raw JSON                                                 | No       |
