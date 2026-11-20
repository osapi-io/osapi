# Migrating a hand-written domain

A domain moves to a behavior in one change: the behavior goes to `/api` and
`osapi client node`, and the hand-written domain is deleted with it. The two
never run side by side. To move a domain, copy the code that works. Do not write
it again. sysctl moved this way, and you can check each step:

1. **Move the provider and its tests without changes.** Move them from
   `internal/provider/<category>/<domain>/` into `internal/behaviors/<name>/`,
   as `<platform>.go` and `<platform>_public_test.go`. If the move deletes
   nothing from the tests, it lost nothing. Change the operation in a later
   commit, not in the move.
2. **Wrap the provider.** `provider.go` declares its interface and the builder.
   `methods.go` changes the input of each method into provider calls. The types
   of the provider move into `types.go`. Use the method names of the domain.
   Then its routes, operationIds and permissions are the same, and the SDK keeps
   its methods.
3. **Find a replacement for each hand-written handler test case** in
   `internal/controller/api/node/<domain>/`. Each case becomes one of these:
   - a row that the generated HTTP tests already have: success, broadcast,
     skipped, failed, timeout, 401, 403, 404, 500, bad target, missing field,
     bad JSON or wrong type
   - a row that `internal/behavior/dispatch` tests for each behavior: a host
     that answers no data, a broadcast that reaches no host
   - an example
   - a refusal that `.Fails` declares

   List the cases that you mapped, and how. A case with no replacement is a
   gap. Add the example or the refusal before you continue.
4. **Examine the semantics of the domain before the cutover makes them
   permanent.** A hand-written domain can be incorrect, and its tests can agree
   with the error. The `list` and `get` of sysctl reported only the parameters
   that osapi wrote. Thus a new host listed nothing. The move made them read the
   kernel and mark the parameters that osapi manages. Say in the prose of the
   definition what changed.
5. **Delete the hand-written domain, and fix each failure:**
   - `TestNoRouteIsServedTwice` names the hand-written API directory. Delete it,
     with its spec, and its line in `internal/controller/handlers.go`.
   - `TestEveryCommandIsAttached` names each hand-written command that has the
     name of a behavior command. Delete the `cmd/client_node_<domain>*.go`
     files.
   - Delete the processor, `internal/agent/processor_<domain>.go`, its case in
     `NewNodeProcessor` and its provider in the agent's setup. The compiler
     then names each test that passes the provider.
   - Delete the job operations of the domain in `pkg/sdk/client/operations.go`
     and `internal/job/types.go`. A behavior's job is `behavior.<name>.<method>`,
     so the old names match no job, and an SDK user who filters by them finds
     nothing.
   - Delete the platform stubs that no build uses.
   - Delete the hand-written integration test of the domain:
     `TestBehaviorSmokeSuite` runs each behavior's commands from its definition.

   Run `mise exec -- just generate`. It joins the spec of the behavior to the
   main spec, and it generates the SDK and UI clients and the CLI pages again.
   Fix the code that does not compile against them, such as a list that gains a
   query parameter. Write the API contract again with
   `go test ./internal/apicontract -run TestContractPublicTestSuite -update`, and
   read its diff: it is the change to the API.
6. **Run the integration tests:** `go test -tags integration ./test/integration/`.
   They send each request and response through the joined spec, and fail on one
   that breaks it.
