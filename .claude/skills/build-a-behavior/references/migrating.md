# Migrating a hand-written domain

Moving a domain is copying what works, not rewriting it. sysctl was moved this
way, and every step is checkable:

1. **Build the behavior beside the domain**, with the domain's method names, so
   its routes, operationIds and permissions match. Add a row for it to
   `migrating` in `internal/behavior/openapi/openapi_public_test.go`, which fails
   until the generated spec describes the same API as the hand-written one.
2. **Move the provider and its tests unchanged** from
   `internal/provider/<category>/<domain>/` into `internal/behaviors/<name>/`, as
   `<platform>.go` and `<platform>_public_test.go`, and point the hand-written
   domain at them. A move with nothing deleted from the tests is the proof
   nothing was lost; change behavior in a later commit, not in the move.
3. **Wrap the provider**: `provider.go` declares its interface and the builder,
   `methods.go` translates each method's input to provider calls. The provider's
   own types move into `types.go`.
4. **Account for every hand-written handler test case** in
   `internal/controller/api/node/<domain>/`. Each becomes one of: a row the
   generated HTTP tests already have (success, broadcast, skipped, failed,
   timeout, 401, 403, 404, 500, bad target, missing field, bad JSON, wrong
   type), an example, or a refusal declared with `.Fails`. List the ones you
   mapped and how. A case with none of these is a gap: add the example or
   refusal before going on.
5. **Question the domain's semantics before the cutover locks them in.** A
   hand-written domain can be wrong in ways its tests agree with. sysctl's
   `list` and `get` reported only the parameters osapi wrote, so a fresh host
   listed nothing; they were changed to read the kernel and mark what osapi
   manages, and its parity row was dropped because its API then differed on
   purpose. Make such a change after the move, in its own commit, and say in
   the definition's prose what changed.
6. **Leave the hand-written domain running** until the cutover, which moves the
   behavior to `/api` and deletes the handler, processor, commands and spec in
   one change.
