# Docs and tables

The Docusaurus site is for users. It tells what a feature does, how to call it,
and what the SDK gives. Development guidance is not there. It is in the corpus,
and this skill cites it.

A domain that works but appears in none of these is invisible to everyone who
did
not write it.

## The rule is in the corpus, not here

The obligation is stated in
[005-building-a-domain](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md).

| What you need to know | Where |
| --- | --- |
| The permission a new domain needs, and the tables that name it | [adding a permission](https://github.com/osapi-io/specs/blob/main/components/osapi/permissions.md#adding-a-permission) |
| A domain appears everywhere an existing domain appears | [every layer, or not done](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#a-domain-is-in-every-layer-or-it-is-not-done) |

Read this before running anything: the documentation below is
checked by `docs-fmt-check` and `docs-build`, which run in
`just test` and **not** in `just ready`. Writing these pages and then running
only
the build-and-unit commands hands in work that fails continuous integration on
the
pages you just wrote.

## Pages to add

```
docs/docs/sidebar/features/{domain}-management.md
docs/docs/sidebar/sdk/client/{category}/{domain}.md            SDK service page
```

Do not write the CLI pages. `just cli-docs-generate` writes them under
`docs/docs/sidebar/usage/cli/` from the commands: the `Short`, `Long`,
`Example` and flags of each one. Give each command a `Short`, and an `Example`
with one to three command lines and no output. Add a `Long` only when a person
needs it to use the command safely. Put each other fact in its annotation from
`internal/cli/reference`: `reference.Result` for when a write returns
`changed: true` or `false` and when it fails, `reference.Permission` for the
permission, and `reference.Notes` for more detail. `just test` fails when a page does not agree
with its command. The output of each command is on `usage/cli/output.md`, one
time.

The SDK page's title is the `Client` field name, for example `# Power`, not the
Go
struct name. Place it in the category directory its concern belongs to.

## Tables and navigation to update

| File | What to add |
| --- | --- |
| `docs/docusaurus.config.ts` | the feature in the Features dropdown, the service in the SDK dropdown |
| `features/features.md` | a row for the domain |
| `sdk/client/client.md` | the service in its category table |
| `features/authentication.md` | nothing: it points to `osapi token roles`, which prints the permissions from the binary |
| `usage/configuration.md` | the same permission, in the roles table and the YAML reference comments |
| `architecture/system-architecture.md` | endpoints, if they belong in the health or endpoint tables |

A new permission occurs in three places: the spec, the role expansion in code,
and `configuration.md`. If you forget the last one, the YAML reference of the
operator is not complete. `osapi token roles` shows it in all cases.

**There is no API guidelines page to update.** For a new path pattern, the
question is if it obeys the [design
guidelines](https://github.com/osapi-io/specs/blob/main/components/osapi/domains.md#design-guidelines). A table
does not need to list it.

## Writing

- Wrap at 80 characters. `mise exec -- just docs-fmt` is prettier, and
  `just md-fmt` covers markdown outside the site.
- Say what the operation does, then show it. A command with its real output
  answers
  more questions than a paragraph.
- Document the idempotent outcome: what a second run reports.
- Note where an operation is unsupported, and what the caller sees then.
- No development instructions. If it tells a contributor how to build something,
  it belongs in the corpus with a citation from this skill.

## Check

```bash
mise exec -- just docs-build     # broken internal links fail the build
mise exec -- just docs-fmt-check
mise exec -- just md-fmt-check
```

`onBrokenLinks: 'throw'` means that a link to a page that you did not make yet
fails the build. A broken link does not ship.
