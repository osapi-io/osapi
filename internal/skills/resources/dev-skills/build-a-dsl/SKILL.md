---
name: build-a-dsl
description: >
  Add a DSL to osapi, or change what the three DSLs share: the behavior DSL
  (internal/behavior), the endpoint DSL (internal/controller/endpoint) and the
  command DSL (internal/cli/command). Use when asked to add a fourth DSL, or
  when a change touches the parts that each DSL has: the guide.Spec and
  guide.Checks tables and the tests that keep them complete, validate, typed
  values and the KnownTags list of tag keys, the permissions that the role rule
  reads (declarations, the SDK constants generator and RegistryOperations), the
  generators and genfile.Writer, the scaffold command and its check, the
  generated DSL pages, the dev skill of a DSL, records of calls at run time, or
  what extensions need. Not for a change inside one DSL: build-a-behavior,
  build-an-endpoint and build-a-command cover those.
---

# Build a DSL

A DSL of osapi is a Go package whose definitions osapi turns into routes,
commands, a spec, docs and tests. The three DSLs define different things, but
each has the same parts. A new DSL gets each part, or the change says why it
does not need it. A change to a part of one DSL that the others share changes
each DSL, or it says why not.

This skill gives a procedure and decisions. The binary tells you about each
DSL:

- `behavior spec get`, `behavior chain spec get`, `controller endpoint spec get`
  and `cli command spec get` tell each part of a DSL.
- `behavior validate`, `behavior chain validate`, `controller endpoint validate`
  and `cli command validate` tell each rule, with `--help` for the rules.

`references/parts.md` names each part, where it is, and what a new DSL does for
it. Read it before you start.

## Procedure

1. **Decide that it is a DSL.** A DSL is for a kind of definition that people
   write again and again, and that osapi turns into more than one thing. One
   more behavior, endpoint group or local command is not a DSL: use its skill.
2. **Write the spec and the checks first.** Declare a `guide.Spec` with a term
   for each exported part, and a `guide.Checks` with each rule. Copy the spec
   test and the checks test of an existing DSL, so a part without a term or a
   rule without a check fails `just test`.
3. **Give each value from a fixed set a type.** A named type, its constants, and
   a validate rule that refuses a value outside the set. Go converts an untyped
   string literal to the type, so the rule is the guard. Struct tags use the
   keys of `command.KnownTags`.
4. **Wire the permissions, if the DSL serves API routes.** Add the registry of
   the DSL to the three places that read permissions: `declarations()` in
   `internal/authtoken/permissions.go`, the generator of the SDK constants in
   `internal/authtoken/gen`, and `RegistryOperations` in
   `internal/authtoken/operations.go`. Give it an `AdminOnly()` option. A DSL
   with no API route declares none.
5. **Generate through `genfile.Writer`.** Record each change and `Flush` at the
   end, so a failed run changes nothing on disk. Write a named type with
   `jen.Lit(string(v))`. Add the generated files to `go-generate-check`.
6. **Add the `create` command and its scaffold check.** The command writes a
   new definition. The `*-scaffold-check` recipe runs it in a copy of the
   checkout and runs the tests of what it wrote.
7. **Add the docs page and the skill.** Add the spec to `dslPages` in
   `cmd/commands.go`, then run `just cli-docs-generate`. Add a dev skill in
   `internal/skills/resources/dev-skills/<name>/` and a pointer in
   `.claude/skills/<name>/SKILL.md`, the same shape as the others.
8. **Verify.** `just go-vet`, `just test`, `just go-generate-check`,
   `just cli-docs-gen-check` and the scaffold check of the DSL.

## Decisions

- **A part that one DSL does not need.** Say why in the change, as the command
  DSL declares no permissions because it calls no API.
- **A part that the DSLs share.** Change it in one place for each DSL, such as
  `command.KnownTags`. Do not copy it into the new DSL.
- **Extensions.** A DSL that extensions use must register a definition without
  a recompile. Keep its registry, spec and checks free of code that osapi
  generates. osapi-io/osapi#1064 has the design questions.
