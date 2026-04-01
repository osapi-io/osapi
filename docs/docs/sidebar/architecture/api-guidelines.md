---
sidebar_position: 4
---

# API Design Guidelines

1. **Top-Level Categories**

Group endpoints by functional domain. Each domain gets its own top-level path
prefix (e.g., `/node/`, `/job/`, `/health/`). Avoid nesting unrelated operations
under a shared prefix.

2. **Resource-Oriented Paths**

Use nouns for resources and let HTTP methods convey the action. Sub-resources
should be nested under their parent (e.g.,
`/node/{hostname}/network/dns/{interfaceName}`).

3. **Consistent Verb Mapping**

- `GET` — Read or list resources
- `POST` — Create a resource or trigger an action
- `PUT` — Replace or update a resource
- `DELETE` — Remove a resource

4. **Scalability and Future Needs**

If an area is expected to grow in complexity with more endpoints, separate it
into its own top-level category early — even if it only has a few operations
today. This avoids clutter and keeps each domain cohesive.

5. **Node as Top-Level Resource**

All operations that target a managed machine are nested under
`/node/{hostname}`. The `{hostname}` path segment identifies the target and
accepts literal hostnames, reserved routing values (`_any`, `_all`), or label
selectors (`key:value`).

Sub-resources represent distinct capabilities of the node:

| Path Pattern                                   | Domain      |
| ---------------------------------------------- | ----------- |
| `/node/{hostname}`                             | Status      |
| `/node/{hostname}/disk`                        | Node        |
| `/node/{hostname}/memory`                      | Node        |
| `/node/{hostname}/network/dns/{interfaceName}` | Network     |
| `/node/{hostname}/command/exec`                | Command     |
| `/node/{hostname}/schedule/cron`               | Schedule    |
| `/node/{hostname}/schedule/cron/{name}`        | Schedule    |
| `/node/{hostname}/sysctl`                      | Sysctl      |
| `/node/{hostname}/sysctl/{key}`                | Sysctl      |
| `/node/{hostname}/ntp`                         | NTP         |
| `/node/{hostname}/timezone`                    | Timezone    |
| `/node/{hostname}/power/reboot`                | Power       |
| `/node/{hostname}/power/shutdown`              | Power       |
| `/node/{hostname}/process`                     | Process     |
| `/node/{hostname}/process/{pid}`               | Process     |
| `/node/{hostname}/process/{pid}/signal`        | Process     |
| `/node/{hostname}/user`                        | User        |
| `/node/{hostname}/user/{name}`                 | User        |
| `/node/{hostname}/user/{name}/password`        | User        |
| `/node/{hostname}/user/{name}/ssh-key`         | User        |
| `/node/{hostname}/user/{name}/ssh-key/{fingerprint}` | User  |
| `/node/{hostname}/group`                       | Group       |
| `/node/{hostname}/group/{name}`                | Group       |
| `/node/{hostname}/package`                     | Package     |
| `/node/{hostname}/package/{name}`              | Package     |
| `/node/{hostname}/package/update`              | Package     |
| `/node/{hostname}/package/updates`             | Package     |
| `/node/{hostname}/log`                         | Log         |
| `/node/{hostname}/log/source`                  | Log         |
| `/node/{hostname}/log/unit/{name}`             | Log         |
| `/node/{hostname}/certificate/ca`              | Certificate |
| `/node/{hostname}/certificate/ca/{name}`       | Certificate |

6. **Path Parameters Over Query Parameters**

Use path parameters for **resource identification and targeting**. Use query
parameters only for **filtering and pagination** on collection endpoints (e.g.,
`/job?status=completed&limit=20`).

Never use query parameters to identify which resource to act on. Complex input
data belongs in request bodies.
