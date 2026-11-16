---
name: build-a-behavior
description: Build or change an osapi behavior, the typed definition under internal/behaviors/<name>/ that osapi's API routes, agent dispatch, CLI commands, OpenAPI spec, docs pages, settings, permissions and tests are generated from. Use when asked to add a behavior, manage something new on a host, add a method or a setting to a behavior, give a behavior a provider for another platform, migrate a hand-written domain to a behavior, or when a behavior fails osapi behavior validate, behavior-gen-check or its tests. Also use to decide whether something should be a behavior at all, whether it should be internal or served, which platforms it runs on, whether it should require another behavior or be split into an internal one, and whether it belongs in osapi-orchestrator instead. Not for the hand-written domains under internal/controller/api/node, which the add-a-domain skill in osapi-io/specs covers.
---

# Build a behavior

A behavior is one directory, `internal/behaviors/<name>/`, plus one import in
`internal/behaviors/all`. Everything else is generated or inferred from its
definition.

This skill gives a procedure and decisions. It is not a catalog. The platform tells about itself:

- `behavior spec` tells each part of the DSL, what it does and when to use it, the generated paths and what the platform cannot do yet.
- `behavior validate --help` tells each rule and each test check, with its fix. Each failure also ends with its fix.
- `behavior list` and `behavior get` tell what each behavior accepts and does. This includes the building blocks.

Ask these tools. Do not use your memory. If a tool and this skill do not agree, the tool is correct. If the platform does not have what you need, stop and say what is missing. That is a platform change, and a person decides it. Do not add code around it.

## Before anything

The binary embeds the UI, so `go run .` fails until it is built once:
`mise exec -- just build`. After that, `go run .` works for every command
below. Run tools through `mise exec --`, as the repository's AGENTS.md says.

## Find what you need

Ask the code and the binary before you read prose. Each one answers one question:

| Question                                         | Ask                                                       |
| ------------------------------------------------ | --------------------------------------------------------- |
| What options, methods, example parts and tags exist, and when to use each | `go run . behavior spec`                                  |
| What rules and test checks a behavior must pass, and the fix of each | `go run . behavior validate --help`                       |
| What a provider builder gets from the agent      | `go doc ./internal/behavior/host Host`                         |
| What behaviors exist already                     | `go run . behavior list`                                  |
| Which building blocks do generic work, such as a managed file, and when to make a new one | `go run . behavior spec`, under "Building on other behaviors" and "Building blocks" |
| What one accepts and does: routes, fields, platforms, sudo, what it uses | `go run . behavior get --name <name> --json`, building blocks included |
| Whether a definition breaks a rule               | `go run . behavior validate`                              |
| A complete example of every part                 | `internal/behaviors/sysctl`                                     |
| How tests fake each host service, and which pattern to copy for each kind of state | [references/services.md](references/services.md) |
| The walkthrough, with a worked behavior          | `docs/docs/sidebar/development/your-first-behavior.md`    |

## Decide first

Decide these with the user before you write code. Each answer becomes a declaration. Thus an incorrect guess becomes an incorrect API, and other code then uses that API. Ask only what the request does not tell you. Give your recommendation with each question. Say which declaration the answer becomes.

| Ask | Why it matters | Answer becomes |
| --- | --- | --- |
| Is there one close already? Check `go run . behavior list --json`. | Two behaviors managing the same thing disagree about its state | Extend that one: a method, a field or a platform |
| Who calls it: operators, or only other behaviors? Would an operator ever call it from the CLI or the API? Is it only a piece of work other behaviors share, such as writing a file or reloading a service? | An operator needs routes, commands and permissions; a building block must not have them. `behavior validate` fails an internal behavior that no behavior requires | `behavior.API` and `behavior.CLI` when an operator calls it, or `behavior.Internal` when only other behaviors do |
| Which hosts does it run on? | Every other platform reports `skipped`, never fails | `behavior.Platform("debian", ...)` per distribution, `"linux"` for every Linux host without its own, `"darwin"` for macOS |
| Does part of the work exist as an internal behavior? `behavior list` marks them `(internal)`. | Copying it means two copies of the hard part | `behavior.Requires(that.Behavior)`. Get its provider with `host.Require`, and call each of its methods with `behavior.Call(ctx, that.Behavior, that.Method, provider, in)`, so the trace shows the call as its own step. A direct call fails the tests |
| Would a second behavior need part of what this one does? | The rule of two: the second need is when shared work becomes its own behavior | Build that part as an internal behavior first, then require it |
| Does a value change per call, or stay the same on a host? | A setting is agent config; an input is part of each request | An input field, or a field in `settings.go` |
| Which input names the thing a call acts on, such as a key or a path? | A trace shows it, so a reader sees what each call changed. `behavior validate` fails a write that records nothing | `.Key("key")`, or `telemetry:"attr"` on a method without a key |
| Which programs does it run, which as root, and does it take a secret? | The agent checks each is installed and sudo allows it; secrets are kept out of the audit log and traces | `behavior.Runs("program")`, or `behavior.Sudo("program")` for root; `secret:"true"` on the field |
| Which files does it write, remove or change the mode of, and does it keep a record in the file-state bucket? | Each one is a span in a trace, and the examples fail on a path or a record that is not declared. Declare the paths even when a required behavior, such as managedfile, does the write | `behavior.Writes("/etc/x.d/osapi-*.conf")`, and `behavior.KeepsState` for a record |
| Is it one host, one step? | A behavior acts on one host inside one job | If it spans hosts or orders several steps, it is an osapi-orchestrator workflow, not a behavior |
| Does it need anything that `behavior spec` lists under "Not supported yet"? | The platform cannot express it, and wiring around that breaks the generated layers | Stop and tell the user what is missing |

Then decide how it operates. This decides the provider and its examples. Read [references/services.md](references/services.md) first. Then select the pattern that the answers identify:

| Ask | Answer becomes |
| --- | --- |
| Where is the truth about what it manages: a file osapi writes, a file something else owns, a program's output, or `/proc`? | Which pattern on that page the provider follows, and which host options its examples use |
| Does a read show the host as it is, or only what osapi wrote? | The host as it is, the way Ansible, Salt and Puppet read: every item, with whether osapi manages it as a field. A read limited to osapi's own records hides the host from the operator |
| What does "already done" look like, so a second run changes nothing? | The comparison the provider makes before acting, and the example's `changed: false` second run |
| Which commands does it run, and which need root? | `h.Exec` calls; `behavior.Runs` for each program, `behavior.Sudo` instead for a privileged one |
| Which inputs reach a path, a line in a file or a program's arguments? | Their `validate` patterns and the refusals declared with `.Fails` |
| What should it refuse, and what does it answer when the thing is not there? | Refusals, and `behavior.ErrNotFound` for a 404 |

Tests do not run real programs. The filesystem of the fake host is in memory. Its commands give the output that the examples declare. Thus you can build and test a behavior for Debian on any machine. Do not skip an example because the program is not installed on your machine.

The request can leave these open: what the behavior manages, where its truth is, or which methods it has. Then ask before you write code, because a guess becomes an API. Use the defaults below only when the user says that they have no preference. Do not use them in place of a question. The defaults are:

- a new behavior only if no close behavior exists
- `API` and `CLI`
- `debian`
- inputs, not settings
- no `Requires`, unless an internal behavior already does part of the work

## Build a new one

1. Write it: `go run . behavior new --name <name> --description "<one line>"`,
   with `--group <group>` for a behavior nested under one, such as `network`,
   then `go generate ./internal/behaviors/<name>/mocks ./internal/behaviors/all`
   (`<group>/<name>` for a grouped one). The
   result builds, validates and passes every test before you change it.
2. Change it in this order, because each depends on the one before:
   `settings.go`, `types.go`, `provider.go` and `<platform>.go`, `methods.go`,
   then `behavior.go`. Use the host only through `host.Host`. Then the fake host can test it. Compare before you act. Then a second run reports `changed: false`. A building block's own fields, such as what `managedfile`
   accepts, come from `behavior get` on it.
3. Give each method a useful example. The example names the host that it runs on and the changes on that host. Declare what the method refuses with `.Fails`.
   `behavior spec` lists the parts of an example. The examples run
   the provider on a fake host and become the generated HTTP and CLI tests.
4. Keep a test for every file with code, named for it:
   `methods_public_test.go` on the mocked provider, and
   `<platform>_public_test.go` for what an example cannot show. After you change `provider.go`, run `go generate ./internal/behaviors/<name>/mocks`.
5. Run the checks below until they pass.

## Design docs

AGENTS.md puts the design in `osapi-io/specs` before you build. For a behavior, the definition is the design. The generated pages and `behavior get` publish its methods, types, settings, examples and refusals. The page of the behavior has its platforms, settings and the behaviors it uses. The CLI page of each method has its examples, what each example does on the host, its refusals and its permission. Thus a new behavior does not need its own specs page. Change `components/osapi/behaviors.md` in specs
only when the work changes the platform itself: a new option, rule, surface or
check that every behavior gets.

## Migrate a hand-written domain

To move a domain, copy the code that works. Do not write it again. Follow [references/migrating.md](references/migrating.md):

1. Build the behavior next to the domain.
1. Move the provider and its tests without changes.
1. Find a replacement for each hand-written handler test case before the cutover deletes it.

## Check it

```bash
go generate ./internal/behaviors/all
mise exec -- just generate   # joins its spec into the main one: SDK, UI, API docs and CLI pages
go run . behavior validate
go test ./internal/behaviors/<name>/... ./internal/behaviors/all/
mise exec -- just test
mise exec -- just behavior-int-linux   # its examples against a real agent on Debian
```

A failure of `behavior validate` or of a behavior's tests ends with its fix.
`go run . behavior validate --help` lists each rule and each test check before
you break one. The failures below come from elsewhere:

| Failure names                         | It means                                                                 |
| ------------------------------------- | ------------------------------------------------------------------------ |
| `<name> runs <program>, which this image lacks` | `behavior-int-linux` only: add the program's package to that recipe's `apt-get install` line |
| `TestAPI<Method>PublicTestSuite/...`  | The API answer for that case in `gen/api_<method>_public_test.go`: the status, the body, or the spec it breaks |
| `is not mounted` or `is not attached` | Run `go generate ./internal/behaviors/all`; the gen package is missing          |
| `behavior output is stale`            | Run `go generate ./internal/behaviors/all` and keep the result                 |
| `the joined spec is stale`            | Run `mise exec -- just generate` and keep the result                           |
| `cli-docs-gen-check` fails            | Run `mise exec -- just cli-docs-generate` and keep the result                  |
| `is served by the <name> behavior and by <dir>` | A route a hand-written domain still serves: this is a cutover, so delete that domain, or mark the behavior `behavior.Next` |
| `is both a hand-written command and a behavior's` | The same, for a command under `osapi client node` |
| `a behavior is below 100% coverage`   | A branch no test reaches: add an example, or a case to the test named for the file |

Never edit a path that `behavior spec` lists under "Generated, never edited".
Change the definition and generate. Do not write CLI docs for a behavior.

## Not yet

`behavior spec` lists what the platform cannot express yet, under "Not
supported yet". If a request needs one, say so and stop. Do not build around it.

A behavior needs no edits outside its directory. The only exception is a package that `behavior-int-linux` says is missing from its image. If you must edit other code, for example `cmd/agent_setup.go`, `internal/validation` or `internal/authtoken`, the platform does not have a declaration that you need. Say so. Do not add the code by hand.
