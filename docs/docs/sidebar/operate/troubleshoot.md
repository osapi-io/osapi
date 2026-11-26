---
sidebar_position: 9
title: Troubleshoot
---

# Troubleshoot

Find the symptom. Then run the command that shows the cause.

## A host

| Symptom                                                | Cause                                                                           | Do this                                                                                                     |
| ------------------------------------------------------ | ------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| A red dot next to the host on the Admission page       | The host is offline: its agent has not reported for some time.                  | On the host: `systemctl status osapi-agent`, or `sudo launchctl print system/io.osapi.agent`. Read its log. |
| `mkdir /etc/osapi: permission denied` at start         | The agent cannot write `agent.pki.key_dir`.                                     | Run the agent as root, or set `agent.pki.key_dir` to a directory that its user owns.                        |
| The host stays `pending`                               | It asked to join with no token, and no admin admitted it.                       | `osapi client admission admit`, or give it a join string.                                                   |
| The agent stops: the token expired, or has no use left | The token of its join string cannot admit it.                                   | `osapi controller join token create`, and give the host the new string.                                     |
| The agent stops: the certificate does not match        | The NATS server has a new certificate since the token was made.                 | Make a new token. Its string has the fingerprint of the new certificate.                                    |
| `verified: false` in `agent list`                      | A registration did not pass the check of the host's key.                        | `osapi client agent get --hostname <host>`. Compare its fingerprint with `admission get`.                   |
| A host was deleted, and its agent still runs           | The agent has the keys of its old admission.                                    | Stop it, empty its `agent.pki.key_dir`, and give it a join string.                                          |
| A job shows `skipped` for a host                       | The host's OS cannot do the operation, such as power management in a container. | `osapi dsl behavior get --name <behavior>` lists the platforms of each method.                              |

## NATS

| Symptom                                                 | Cause                                                                                           | Do this                                                                                     |
| ------------------------------------------------------- | ----------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| `no responders` at a join                               | No controller answers the join, or the bootstrap user cannot send the request.                  | Start the controller. Make the NATS credentials again after an upgrade.                     |
| `Authorization Violation`                               | The NATS user of a process does not have a subject that the process uses.                       | `osapi nats permissions get --role <role>`. Make the credentials again, or add the subject. |
| `authentication error` for an agent, with auth callout  | The callout refused it: the host is revoked or not admitted, or it joined again with a new key. | `osapi client audit list` has the `nats.connect` entry with the reason.                     |
| `nkeys not supported by the server` at an agent's start | The server offers no nonce: it has no nkey user.                                                | Make the callout user an nkey ([2. NATS](nats.md#auth-callout)).                            |
| `maximum control line exceeded`                         | Your NATS server refuses a long user JWT.                                                       | Set `max_control_line: 32KB` in the config of your server.                                  |
| `osapi client health status` shows a stream as missing  | The streams and buckets were not made on your server.                                           | `osapi nats setup`.                                                                         |

## The controller and the CLI

| Symptom                                        | Cause                                                           | Do this                                                                                     |
| ---------------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| A config error at start                        | A key breaks its rule.                                          | `osapi config validate -f <file>` gives the key, the rule and the fix.                      |
| `401` from the API                             | The token is not signed by a key of `signing_keys`, or expired. | `osapi controller token validate --token -` on the controller, with the token on its input. |
| `403` from the API                             | The roles of the token do not grant the permission.             | `osapi controller role get --name <role>`.                                                  |
| A setting does not have the value that you set | An environment variable has priority over the file.             | `osapi config get --key <key>` shows the source of the value.                               |
