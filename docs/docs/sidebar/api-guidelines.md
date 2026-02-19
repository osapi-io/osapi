---
sidebar_position: 9
---

# API Design Guidelines

1. **Top-Level Categories**

Group endpoints by functional domain. Each domain gets its own top-level path
prefix (e.g., `/system/`, `/network/`, `/job/`). Avoid nesting unrelated
operations under a shared prefix.

2. **Resource-Oriented Paths**

Use nouns for resources and let HTTP methods convey the action. Sub-resources
should be nested under their parent (e.g., `/network/dns/{interfaceName}`).

3. **Consistent Verb Mapping**

- `GET` — Read or list resources
- `POST` — Create a resource or trigger an action
- `PUT` — Replace or update a resource
- `DELETE` — Remove a resource

4. **Scalability and Future Needs**

If an area is expected to grow in complexity with more endpoints, separate it
into its own top-level category early — even if it only has a few operations
today. This avoids clutter and keeps each domain cohesive.
