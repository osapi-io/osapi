# Migrating a hand-written domain

To move a domain, copy the code that works. Do not write it again. sysctl moved
this way, and you can check each step:

1. **Build the behavior next to the domain.** Mark it `behavior.Next`. Then osapi
   serves it under `/api/next` and `osapi client next`, and it does not conflict
   with the hand-written routes and commands. Use the method names of the
   domain. Then its routes, operationIds and permissions are the same. Add a
   row for it to `migrating` in
   `internal/behavior/openapi/openapi_public_test.go`. That test fails until the
   generated spec describes the same API as the hand-written spec.
2. **Move the provider and its tests without changes.** Move them from
   `internal/provider/<category>/<domain>/` into `internal/behaviors/<name>/`,
   as `<platform>.go` and `<platform>_public_test.go`. Point the hand-written
   domain at them. If the move deletes nothing from the tests, it lost nothing.
   Change the operation in a later commit, not in the move.
3. **Wrap the provider.** `provider.go` declares its interface and the builder.
   `methods.go` changes the input of each method into provider calls. The types
   of the provider move into `types.go`.
4. **Find a replacement for each hand-written handler test case** in
   `internal/controller/api/node/<domain>/`. Each case becomes one of these:
   - a row that the generated HTTP tests already have: success, broadcast,
     skipped, failed, timeout, 401, 403, 404, 500, bad target, missing field,
     bad JSON or wrong type
   - an example
   - a refusal that `.Fails` declares

   List the cases that you mapped, and how. A case with no replacement is a
   gap. Add the example or the refusal before you continue.
5. **Examine the semantics of the domain before the cutover makes them
   permanent.** A hand-written domain can be incorrect, and its tests can agree
   with the error. The `list` and `get` of sysctl reported only the parameters
   that osapi wrote. Thus a new host listed nothing. The change made them read
   the kernel and mark the parameters that osapi manages. Its parity row went
   away, because its API was then different on purpose. Make such a change
   after the move, in its own commit. Say in the prose of the definition what
   changed.
6. **Keep the hand-written domain until the cutover.** For the cutover, remove
   `behavior.Next` and fix each failure:
   - The route check names the hand-written API domain. Delete it.
   - The CLI does not start until you delete the hand-written commands.
   - The compiler finds the case of the processor in `NewNodeProcessor`.

   Delete the hand-written spec, the CLI docs pages and the platform stubs that
   no build uses. Run `mise exec -- just generate`. It joins the spec of the
   behavior to the main spec, and it generates the SDK and UI clients again.
   Fix the code that does not compile against them.
