---
name: build-a-behavior
description: Build or change an osapi behavior, the typed definition under internal/behaviors/<name>/ that osapi's API routes, agent dispatch, CLI commands, OpenAPI spec, docs pages, settings, permissions and tests are generated from. Use when asked to add a behavior, manage something new on a host, add a method or a setting to a behavior, give a behavior a provider for another platform, migrate a hand-written domain to a behavior, or when a behavior fails osapi behavior validate, behavior-gen-check or its tests. Also use to decide whether something should be a behavior at all, whether it should be internal or served, which platforms it runs on, whether it should require another behavior or be split into an internal one, and whether it belongs in osapi-orchestrator instead. Not for the hand-written domains under internal/controller/api/node, which the add-a-domain skill in osapi-io/specs covers.
---

# Build a behavior

A behavior is one directory, `internal/behaviors/<name>/`, plus one import in
`internal/behaviors/all`. Everything else is generated or inferred from its
definition.

This skill is procedure and judgment, not a catalog. The platform describes
itself: `behavior list` and `behavior get` say what every behavior, building
blocks included, accepts and does; `go doc` says what every option and host
service is. Ask them rather than remembering, and believe them over this skill.
When what you need is not there, stop and say what is missing; that is a
platform change for a person to decide, not something to wire around.

## Before anything

The binary embeds the UI, so `go run .` fails until it is built once:
`mise exec -- just build`. After that, `go run .` works for every command
below. Run tools through `mise exec --`, as the repository's AGENTS.md says.

## Find what you need

Ask the code and the binary before reading prose. Each answers one question:

| Question                                         | Ask                                                       |
| ------------------------------------------------ | --------------------------------------------------------- |
| What options, tags and checks exist              | `go doc ./internal/behavior`, then `go doc ./internal/behavior <Name>` |
| What a provider builder gets from the agent      | `go doc ./internal/behavior/host Host`                         |
| What behaviors exist already                     | `go run . behavior list`                                  |
| What one accepts and does: routes, fields, platforms, sudo, what it uses | `go run . behavior get --name <name> --json`, building blocks included |
| Whether a definition breaks a rule               | `go run . behavior validate`                              |
| A complete example of every part                 | `internal/behaviors/sysctl`                                     |
| How tests fake each host service, and which pattern to copy for each kind of state | [references/services.md](references/services.md) |
| The walkthrough, with a worked behavior          | `docs/docs/sidebar/development/your-first-behavior.md`    |

## Decide first

Settle these with the user before writing anything. Each answer is a
declaration, so a wrong guess becomes a wrong API that someone then depends on.
Ask only what the request leaves open, give your recommendation with each
question, and say which declaration the answer becomes.

| Ask | Why it matters | Answer becomes |
| --- | --- | --- |
| Is there one close already? Check `go run . behavior list --json`. | Two behaviors managing the same thing disagree about its state | Extend that one: a method, a field or a platform |
| Who calls it: operators, or only other behaviors? | An operator needs routes, commands and permissions; a building block must not have them | `behavior.API` and `behavior.CLI`, or `behavior.Internal` |
| Which hosts does it run on? | Every other platform reports `skipped`, never fails | `behavior.Platform("debian", ...)` per distribution, `"linux"` for every Linux host without its own, `"darwin"` for macOS |
| Does part of the work exist as an internal behavior? `behavior list` marks them `(internal)`. | Copying it means two copies of the hard part | `behavior.Requires(that.Behavior)`, and call its methods through `host.Require` |
| Would a second behavior need part of what this one does? | The rule of two: the second need is when shared work becomes its own behavior | Build that part as an internal behavior first, then require it |
| Does a value change per call, or stay the same on a host? | A setting is agent config; an input is part of each request | An input field, or a field in `settings.go` |
| Which programs does it run, which as root, and does it take a secret? | The agent checks each is installed and sudo allows it; secrets are kept out of the audit log and traces | `behavior.Runs("program")`, or `behavior.Sudo("program")` for root; `secret:"true"` on the field |
| Is it one host, one step? | A behavior acts on one host inside one job | If it spans hosts or orders several steps, it is an osapi-orchestrator workflow, not a behavior |
| Does it need anything under "Not yet" below? | The platform cannot express it, and wiring around that breaks the generated layers | Stop and tell the user what is missing |

Then settle how it behaves, which decides the provider and its examples. Read
[references/services.md](references/services.md) first, and pick the pattern
its answers point to:

| Ask | Answer becomes |
| --- | --- |
| Where is the truth about what it manages: a file osapi writes, a file something else owns, a program's output, or `/proc`? | Which pattern on that page the provider follows, and which host options its examples use |
| Does a read show the host as it is, or only what osapi wrote? | The host as it is, the way Ansible, Salt and Puppet read: every item, with whether osapi manages it as a field. A read limited to osapi's own records hides the host from the operator |
| What does "already done" look like, so a second run changes nothing? | The comparison the provider makes before acting, and the example's `changed: false` second run |
| Which commands does it run, and which need root? | `h.Exec` calls; `behavior.Runs` for each program, `behavior.Sudo` instead for a privileged one |
| Which inputs reach a path, a line in a file or a program's arguments? | Their `validate` patterns and the refusals declared with `.Fails` |
| What should it refuse, and what does it answer when the thing is not there? | Refusals, and `behavior.ErrNotFound` for a 404 |

Tests never run anything real: the fake host's filesystem is in memory and its
commands answer with what the examples declare, so a behavior for Debian is
built and tested on any machine. Do not skip an example because the program is
not installed where you are.

When the request leaves open what the behavior manages, where its truth lives
or which methods it has, ask before writing anything: a guess there becomes an
API. The defaults below are for a user who has answered that they have no
preference, not for a question nobody asked. When the user has no opinion,
the defaults are: a new behavior only if nothing
close exists; `API` and `CLI`; `debian`; inputs rather than settings; and no
`Requires` unless an internal behavior already does part of the work.

## Build a new one

1. Write it: `go run . behavior new --name <name> --description "<one line>"`,
   with `--group <group>` for a behavior nested under one, such as `network`,
   then `go generate ./internal/behaviors/<name>/mocks ./internal/behaviors/all`
   (`<group>/<name>` for a grouped one). The
   result builds, validates and passes every test before you change it.
2. Change it in this order, because each depends on the one before:
   `settings.go`, `types.go`, `provider.go` and `<platform>.go`, `methods.go`,
   then `behavior.go`. Reach the host only through `host.Host`, so the fake host
   can test it, and compare before acting, so running twice reports
   `changed: false`. A building block's own fields, such as what `managedfile`
   accepts, come from `behavior get` on it.
3. Give every method an example a reader would want, with the host it runs on
   and what it changes there, and declare what it refuses with `.Fails`.
   `go doc ./internal/behavior Given` and `Effect` list the options. The examples run
   the provider on a fake host and become the generated HTTP and CLI tests.
4. Keep a test for every file with code, named for it:
   `methods_public_test.go` on the mocked provider, and
   `<platform>_public_test.go` for what an example cannot show. Run
   `go generate ./internal/behaviors/<name>/mocks` after changing `provider.go`.
5. Run the checks below until they pass.

## Design docs

AGENTS.md puts design in `osapi-io/specs` before building. For a behavior, the
definition is the design: its methods, types, settings, examples and refusals
are what the generated docs page and `behavior get` publish, so a new behavior
needs no specs page of its own. Change `components/osapi/behaviors.md` in specs
only when the work changes the platform itself: a new option, rule, surface or
check that every behavior gets.

## Migrate a hand-written domain

Moving a domain is copying what works, not rewriting it. Follow
[references/migrating.md](references/migrating.md): build the behavior beside
the domain, move the provider and its tests unchanged, and account for every
hand-written handler test case before the cutover deletes it.

## Check it

```bash
go generate ./internal/behaviors/all
go run . behavior validate
go test ./internal/behaviors/<name>/... ./internal/behaviors/all/
mise exec -- just test
mise exec -- just behavior-int-linux   # its examples against a real agent on Debian
```

Read a failure for what it names:

| Failure names                         | It means                                                                 |
| ------------------------------------- | ------------------------------------------------------------------------ |
| `<name> <method> on <platform>, "..."` | An example's output or effect is not what the provider did on the fake host |
| `a write declares no effects`         | Add `Wrote`, `Removed` or `Ran` to that example                          |
| `running it again reports changed`    | The provider writes without comparing first                              |
| `with call N failing: panicked`       | The provider ignores an error and uses its zero value                    |
| `with store read N returning ...`     | The provider trusts a state record without checking it                   |
| `refusal "...": succeeded with`       | The provider accepts what the refusal says it rejects                    |
| `with path traversal in <Field>`      | A path built from that field leaves its directory: add a `validate` pattern, or check it in the provider and declare the refusal |
| `with a new line in <Field>`          | That field can add a line to a file: refuse `\r` and `\n`              |
| `with an option in <Field>`           | That field reaches a command as an option: refuse a leading `-`, or pass `--` before it |
| `it reached a shell script`           | Input reaches `sh -c`: run the program directly with its arguments instead |
| `runs <program> but does not declare it` | Declare it with `behavior.Runs`, or `behavior.Sudo` for one run as root |
| `<name> runs <program>, which this image lacks` | `behavior-int-linux` only: add the program's package to that recipe's `apt-get install` line |
| `TestAPI<Method>PublicTestSuite/...`  | The API answer for that case in `gen/api_<method>_public_test.go`: the status, the body, or the spec it breaks |
| `is not mounted` or `is not attached` | Run `go generate ./internal/behaviors/all`; the gen package is missing          |
| `behavior output is stale`            | Run `go generate ./internal/behaviors/all` and keep the result                 |
| `a behavior is below 100% coverage`   | A branch no test reaches: add an example, or a case to the test named for the file |

Never edit `gen/`, `mocks/`, `internal/behaviors/all/gen`,
`internal/controller/api/next/gen` or `docs/docs/sidebar/behaviors/<name>/`;
change the definition and generate. A generated test that fails is fixed in the
definition or in `internal/behavior/gen`, never in `gen/`.

## Not yet

These cannot be expressed yet. A request that needs one is answered by saying
so and stopping, without building around it:

- **Action methods**, such as restart or reboot. `behavior validate` refuses a
  write that is not create, update or delete.
- **Running what the caller supplies**, a command or a shell: there is no
  `execute` permission.
- **A resource under another**, with two keys in the path, such as a user's SSH
  keys.
- **Streaming and uploads**, and the object store they would deploy from.
- **A step on the controller** before the job is queued, such as hashing a
  password.

This list is the one catalog here, kept because finding these by building costs
a session. It shrinks as `behavior validate` learns to say them.

Nothing outside the behavior's directory needs editing for a behavior, except
a package `behavior-int-linux` says its image lacks. If you find you need to, such as in `cmd/agent_setup.go`, `internal/validation` or
`internal/authtoken`, the platform is missing a declaration. Say so instead of
wiring it by hand.
