# Tokens

A token is a JWT signed with the first key of
`controller.api.security.signing_keys`, with that key's `id` as its `kid`. The
controller verifies a token with the key that its `kid` names, and refuses a
token with no `kid` or an unknown one, so a token made on one install does not
work on another, and a token stops when its key leaves the list.
`token validate` shows the `kid`.

```sh
osapi controller token generate --roles admin --subject ops@example.com --json
osapi controller token generate --roles read --subject ci --permissions status:read --json
osapi controller token validate --token <jwt> --json
```

Give each person or program the smallest role that does its work:

| Role         | For                                               | May                                                  | May not                              |
| ------------ | ------------------------------------------------- | ---------------------------------------------------- | ------------------------------------ |
| `admin`      | the owner of the fleet                            | everything                                           |                                      |
| `operator`   | a person on call                                  | read; restart, signal, reboot; retry or delete a job | change a configuration, admit a host |
| `automation` | a pipeline that applies configuration, such as CI | read; change configuration; use a secret in a change | run an action, such as a restart     |
| `read`       | a person who looks into a problem                 | read each thing that is not admin only               | change anything                      |
| `auditor`    | a person who reviews what occurred                | the audit log, traces, jobs and health               | read a host                          |
| `observer`   | a dashboard or a monitor                          | health, status and the list of agents                | read a host or a job                 |

```sh
osapi controller role list --json                 # each role, built-in or custom
osapi controller role get --name operator --json  # each permission it grants
```

Only `admin` has the permissions that are the same as root on a host or that
decide trust: `command:execute`, `command:shell`, `docker:exec`,
`schedule:write`, `user:write`, `secret:reveal`, and the admission and agent
writes. A custom role in the config can grant one of them to fewer people.

`--targets` limits a token to some hosts, by hostname or label, such as
`--targets group:web`: a request for any other host is refused. `--permissions`
grants exactly the listed permissions instead of a role's;
`osapi help token generate --json` lists them. Prefer a role and `--targets`
over a list of permissions: a role gets the permissions of each new behavior by
its rule.

The CLI sends `controller.client.security.bearer_token`, usually set as
`OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN`.

A request refused with `unauthorized` has no valid token: missing, expired, or
signed with a different key. One refused with `forbidden` has a valid token
without the permission; `controller token validate` shows what it carries.
