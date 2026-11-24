---
sidebar_position: 7.1
---

# Building an endpoint

An endpoint is one operation of the controller itself, such as `GET /api/agent`
or `POST /api/job/{id}/retry`. It runs on the controller, in the request. It has
no target and no result for each host. An operation that changes or reads a host
is a [behavior](behaviors.md), not an endpoint. Work that spans hosts or orders
several steps is an osapi-orchestrator workflow.

A group of endpoints is one directory, `internal/controller/endpoints/<group>/`,
and one import in `internal/controller/endpoints/all`. The DSL is in
`internal/controller/endpoint`. From the definition, osapi makes the routes, the
request decoders, the part of the OpenAPI spec, and the CLI commands of the
group. The handlers are typed Go functions.
`internal/controller/endpoints/agent` is a complete group.

## Steps

1. Write the layout: `go run . dsl endpoint create --name <group>`.
2. Read the [Endpoint DSL](dsl/endpoint.md) reference, or
   `go run . dsl endpoint spec get`. It lists each part of the DSL, what it does
   and when to use it.
3. Change the definition, the types and the handlers.
4. Run `just generate`.
5. Run `go run . dsl endpoint validate`. Each problem ends with its fix, and
   `--help` lists each rule.
6. Run the tests of the group.

`go run . dsl endpoint get --name <group>` shows the routes, permissions and
operationIds of a group. `go run . dsl endpoint openapi get --name <group>`
shows the spec that the definition makes.

The `build-an-endpoint` [skill](skills.md) gives the same steps to an agent.
