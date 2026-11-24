set allow-duplicate-variables

# Optional modules: mod? allows `just fetch` to work before .just/remote/ exists.
# Recipes below use `just` subcommands instead of dependency syntax because just
# validates dependencies at parse time, which would fail when modules aren't loaded.
# The React application lives in ui/, not at the repository root, so the react
# module's default of "." is reassigned above the import.

react_dir := "ui"

# The skills, the ones `osapi skill install` installs and the ones in .claude/, are
# written by hand, so md formats them. The shared recipe keeps their YAML
# frontmatter.

md_agent_exclude := ""

import? '.just/remote/go.just'

import? '.just/remote/docusaurus.just'
import? '.just/remote/md.just'
import? '.just/remote/just.just'

import? '.just/remote/react.just'

# --- Fetch ---

# List the recipes
default:
    @just --list

# Fetch shared justfiles from osapi-justfiles
#
# It fetches the commit that main is at, not refs/heads/main: the raw URL of a
# branch is cached for five minutes, so a run just after a merge got the old
# modules and said nothing (osapi-io/osapi-justfiles#108). git ls-remote asks
# the server and has no cache, and the URL of a commit never changes. The
# commit goes to .just/remote/VERSION, so a build records what it used.
fetch:
    #!/usr/bin/env bash
    set -euo pipefail
    sha=$(git ls-remote https://github.com/osapi-io/osapi-justfiles refs/heads/main | cut -f1)
    base="https://raw.githubusercontent.com/osapi-io/osapi-justfiles/$sha"
    mkdir -p .just/remote
    for module in go docusaurus md just react; do
        curl -sSfL "$base/$module/$module.just" -o ".just/remote/$module.just"
    done
    echo "$sha" > .just/remote/VERSION
    echo "osapi-justfiles $sha"

# --- Top-level orchestration ---

# Install all dependencies
deps:
    just go-deps
    just docs-deps
    just react-deps

# Build production binary (includes embedded UI)
build:
    just react-build
    go build -o osapi .

# Run all tests
#
# docs-fmt-check belongs here, not only in CI. A change merged from a
# security advisory's private fork runs no Actions at all, so this recipe is
# the only pre-merge gate those fixes get. Three of them left five pages
# unformatted and broke Docusaurus Lint on main in September 2026.
#
# md-fmt-check is deliberately absent: it needs uv, which the Go workflow does
# not install, and Markdown Lint already runs it with uv available.
test: linux-tune
    just react-build
    just go-test

# The checks of `just test`, which the shared go-test runs and CI reads
# (`just go-ci-checks`): the quick checks share one job, and each slow check is
# a job. A check added here runs on a laptop and in CI, with no change to the
# workflow.
go_quick_checks := "just-fmt-check docs-fmt-check react-fmt-check react-unit react-generate-check go-mod-check go-pkg-check"
go_slow_checks := "behavior-gen-check go-generate-check docs-gen-check cli-docs-gen-check behavior-scaffold-check endpoint-scaffold-check command-scaffold-check behavior-cover-check go-fmt-check go-vet examples-check spec-client-check sdk-gen-check"

# Fail when a behavior that osapi dsl behavior create writes does not pass its tests
#
# The scaffold must keep up with the checks of behaviortest, as each behavior
# must. The recipe writes a behavior in a copy of the checkout, generates it,
# and runs its tests. The checkout itself does not change.
#
# It also writes a behavior that requires managedfile. That behavior passes
# each test but one: its examples fail the rule used until the author writes
# a call to managedfile, and the message of the rule says so.
behavior-scaffold-check:
    #!/usr/bin/env bash
    set -euo pipefail
    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT
    rsync -a --exclude .git --exclude .worktrees --exclude node_modules \
        --exclude /dist --exclude /examples --exclude /docs/build ./ "$tmp/"
    cd "$tmp"
    go run . dsl behavior create --name scaffoldcheck --description "A scaffold check" >/dev/null
    go run . dsl behavior create --name scaffoldchain --description "A scaffold check" \
        --requires managedfile >/dev/null
    go generate ./internal/behaviors/scaffoldcheck/mocks \
        ./internal/behaviors/scaffoldchain/mocks ./internal/behaviors/all
    go test -count=1 ./internal/behaviors/scaffoldcheck/...
    examples='TestBehaviorPublicTestSuite/TestExamples'
    go test -count=1 -skip "$examples" ./internal/behaviors/scaffoldchain/...
    if out=$(go test -count=1 -run "$examples" ./internal/behaviors/scaffoldchain 2>&1); then
        echo "scaffoldchain: its examples pass with no call to managedfile" >&2
        exit 1
    fi
    if ! grep -q "scaffoldchain requires managedfile, and no example calls it" <<<"$out"; then
        echo "$out" >&2
        exit 1
    fi

# Fail when a group that osapi dsl endpoint create writes does not pass
# its tests
#
# As behavior-scaffold-check does for behaviors: the recipe writes a group in
# a copy of the checkout, generates it, and runs its tests and the
# controller's, which mount its routes. The checkout itself does not change.
endpoint-scaffold-check:
    #!/usr/bin/env bash
    set -euo pipefail
    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT
    rsync -a --exclude .git --exclude .worktrees --exclude node_modules \
        --exclude /dist --exclude /examples --exclude /docs/build ./ "$tmp/"
    cd "$tmp"
    go run . dsl endpoint create --name scaffoldcheck --description "A scaffold check" >/dev/null
    go generate ./internal/controller/endpoints/all
    go run . dsl endpoint validate >/dev/null
    go test -count=1 ./internal/controller/endpoints/scaffoldcheck/... ./internal/controller/endpoints/all/ ./internal/controller/

# Fail when a group that osapi dsl command create writes does not pass its tests
#
# As endpoint-scaffold-check does for endpoint groups: the recipe writes a
# local group in a copy of the checkout, generates it, and runs its tests,
# the tests of every local group and the tests of cmd/, which hold each
# command to the CLI's rules. The checkout itself does not change.
command-scaffold-check:
    #!/usr/bin/env bash
    set -euo pipefail
    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT
    rsync -a --exclude .git --exclude .worktrees --exclude node_modules \
        --exclude /dist --exclude /examples --exclude /docs/build ./ "$tmp/"
    cd "$tmp"
    go run . dsl command create --name scaffoldcheck --description "A scaffold check" >/dev/null
    go generate ./internal/cli/commands/all
    go run . dsl command validate >/dev/null
    go test -count=1 ./internal/cli/commands/scaffoldcheck/... ./internal/cli/commands/all/ ./cmd/

# Fail when a behavior's own code is not fully covered
#
# The repository's figure is a total, so one behavior short of 100% can pass
# it while the lines its examples miss go untested. Each behavior package is
# held to 100% on its own; gen/ and mocks/ are generated and left out, as
# .coverignore leaves them out of the total.
behavior-cover-check:
    #!/usr/bin/env bash
    set -euo pipefail
    pkgs=$(go list ./internal/behaviors/... | grep -vE '/(gen|mocks|all)(/|$)')
    short=$(go test -cover $pkgs | grep -v 'coverage: 100.0% of statements' || true)
    if [ -n "$short" ]; then
        echo "$short" >&2
        echo "a behavior is below 100% coverage: add an example, or a case to the test named for the file" >&2
        exit 1
    fi

# Fail when anything under pkg/ reaches internal/
#
# pkg/ is what osapi promises to code outside it, the SDK, and internal/ looks
# to pkg/ for what it needs, never the other way. A pkg/ package that imported
# internal/, its tests included, would tie the SDK to code nothing outside
# this module can see.
go-pkg-check:
    #!/usr/bin/env bash
    set -euo pipefail
    leaks=$(go list -deps -test ./pkg/... | grep '^github.com/osapi-io/osapi/internal/' || true)
    if [ -n "$leaks" ]; then
        echo "pkg/ imports internal/:" >&2
        echo "$leaks" >&2
        exit 1
    fi

# Compile a client that oapi-codegen writes from the full spec with its
# default options, as a person outside osapi would. The SDK does not use
# oapi-codegen: internal/sdk/gen writes it. This holds the spec to a stranger's
# generator. The client builds in a module of its own, outside the checkout,
# with the runtime that go mod tidy picks for it, as a stranger's would.
spec-client-check:
    #!/usr/bin/env bash
    set -euo pipefail
    dir=$(mktemp -d)
    trap 'rm -rf "$dir"' EXIT
    go tool oapi-codegen -generate types,client -package specclient \
        internal/controller/api/gen/api.yaml > "$dir/client.go"
    cd "$dir"
    go mod init specclient > /dev/null 2>&1
    go mod tidy > /dev/null 2>&1
    if ! go build ./...; then
        echo "a client that oapi-codegen generates from the full spec does not compile" >&2
        exit 1
    fi

# Vet each SDK example. The examples build against this checkout's SDK through
# the replace in their go.mod, so a change to the SDK that breaks one fails
# here (#820 broke three, and CI passed). Each file is a separate main, so a
# build of the whole directory fails on the duplicate main: each one is vetted
# alone.
examples-check:
    #!/usr/bin/env bash
    set -euo pipefail
    # A directory with no example, such as one that holds only a built binary
    # that .gitignore ignores, has nothing to check: *.go matches nothing.
    shopt -s nullglob
    broken=0
    for dir in examples/sdk/*/; do
        dir=${dir%/}
        for f in "$dir"/*.go; do
            if ! (cd "$dir" && go vet "$(basename "$f")"); then
                echo "the SDK example $f does not build against this SDK" >&2
                broken=$((broken + 1))
            fi
        done
    done
    if [ "$broken" -gt 0 ]; then
        echo "$broken SDK examples are broken: change each to the SDK's shapes" >&2
        exit 1
    fi

# Write the CLI reference pages under docs/ from the cobra command tree, and
# format them as the site formatter does.
cli-docs-generate:
    just _docs-bun-install
    go run . dsl docs generate
    cd docs && bun run prettier docs/sidebar/usage/cli docs/sidebar/development/dsl docs/sidebar/reference docs/sidebar/features docs/sidebar/usage/configuration.md --write --config prettier.config.js --ignore-path .prettierignore > /dev/null

# The CLI reference pages are generated from the cobra command tree. A command
# without a page, a page without a command, or a flag that changed fails here.
cli-docs-gen-check:
    #!/usr/bin/env bash
    set -euo pipefail
    pages() {
        find docs/docs/sidebar/usage/cli docs/docs/sidebar/development/dsl docs/docs/sidebar/reference docs/docs/sidebar/features docs/docs/sidebar/usage/configuration.md -type f -print | sort | xargs shasum
    }
    before=$(pages)
    just cli-docs-generate
    if [ "$before" != "$(pages)" ]; then
        diff <(echo "$before") <(pages) || true
        echo "CLI reference pages are stale: run just cli-docs-generate and commit the result" >&2
        exit 1
    fi

# The API reference pages under docs/docs/gen are generated from the full
# OpenAPI spec. Nothing compared them with the specs, so /api/next pages went
# stale without a failure. Regenerate them and fail when a page changed.
docs-gen-check:
    #!/usr/bin/env bash
    set -euo pipefail
    pages() {
        find docs/docs/gen -type f -print 2>/dev/null | sort | xargs shasum
    }
    before=$(pages)
    just docs-generate > /dev/null
    if [ "$before" != "$(pages)" ]; then
        diff <(echo "$before") <(pages) || true
        echo "API reference pages are stale: run just docs-generate and commit the result" >&2
        exit 1
    fi

# Fail when what go generate writes is not what is committed
#
# Mocks, OpenAPI server code and behavior output are all generated. An
# interface changed without regenerating its mock compiles until a test calls
# the method the old mock lacks, so every generated file is compared, before
# and after regenerating; uncommitted work elsewhere does not trip it.
go-generate-check:
    #!/usr/bin/env bash
    set -euo pipefail
    generated() {
        find . \( -path ./docs/node_modules -o -path ./ui/node_modules -o -path ./.worktrees -o -path ./.git \) -prune -o \
            -type f \( -name '*.gen.go' -o -path '*/gen/*' \) -print | sort | xargs shasum
    }
    before=$(generated)
    go generate ./...
    if [ "$before" != "$(generated)" ]; then
        diff <(echo "$before") <(generated) || true
        echo "generated code is stale: run go generate ./... and commit the result" >&2
        exit 1
    fi

# Run the unit tests of the UI
#
# The shared react module has no recipe that runs tests. bun runs them itself,
# with no test package: they are the src/**/*.test.ts files of the UI, and
# tsconfig.test.json checks their types.
react-unit:
    cd {{ react_dir }} && bun test

# Fail when the UI client that orval writes is not what is committed
#
# orval writes ui/src/sdk/gen from the full spec. go generate does not run
# orval, so go-generate-check does not find a stale UI client. An orval bump
# that changes the output went to main twice with the old client (#686,
# #1150). The check reads the committed full spec, which behavior-gen-check
# holds fresh. orval takes under a second, so the check is a quick one.
react-generate-check:
    #!/usr/bin/env bash
    set -euo pipefail
    generated() {
        find {{ react_dir }}/src/sdk/gen -type f -print 2>/dev/null | sort | xargs shasum
    }
    before=$(generated)
    just react-generate > /dev/null
    if [ "$before" != "$(generated)" ]; then
        diff <(echo "$before") <(generated) || true
        echo "the UI client is stale: run just react-generate and commit the result" >&2
        exit 1
    fi

# Fail when what is generated from behaviors is not what is committed
#
# The CLI commands, the spec and the docs pages are generated from each
# behavior's definition. A change to a behavior without `go generate`
# leaves them describing the old one, and nothing else would notice. Only the
# generated files are compared, before and after regenerating, so uncommitted
# work elsewhere does not trip it.
behavior-gen-check:
    #!/usr/bin/env bash
    set -euo pipefail
    # The full spec is written in the same step, from the behaviors and the
    # endpoint groups. The SDK, the UI and the API docs are generated from
    # it, so a behavior changed without just generate leaves all three
    # describing its old shape.
    generated() {
        {
            find internal/behaviors -type f -path '*/gen/*'
            echo internal/controller/api/gen/api.yaml
        } 2>/dev/null | sort | xargs shasum
    }
    before=$(generated)
    go generate ./internal/behaviors/all
    if [ "$before" != "$(generated)" ]; then
        diff <(echo "$before") <(generated) || true
        echo "behavior output is stale: run just generate and commit the result" >&2
        exit 1
    fi

# Write the SDK's reference pages under docs/ from the spec and the SDK's
# source, and format them as the site formatter does. go generate writes the
# SDK first, so the pages show its methods.
sdk-docs-generate:
    just _docs-bun-install
    go run ./internal/sdk/gen --docs
    cd docs && bun run prettier docs/sidebar/sdk/client --write --config prettier.config.js --ignore-path .prettierignore > /dev/null

# The SDK's examples and its reference pages are generated from the full spec,
# from the examples of each method's definition, and from the SDK's source. A
# method that changed without just generate leaves an example that calls the
# old shape, or a page that shows it: both went stale by hand before (#1008,
# #1013). Regenerate them and fail when one changed.
sdk-gen-check:
    #!/usr/bin/env bash
    set -euo pipefail
    generated() {
        find examples/sdk/client docs/docs/sidebar/sdk/client -type f -print | sort | xargs shasum
    }
    before=$(generated)
    go generate ./pkg/sdk/client
    just sdk-docs-generate
    if [ "$before" != "$(generated)" ]; then
        diff <(echo "$before") <(generated) || true
        echo "the SDK's examples or pages are stale: run just generate and commit the result" >&2
        exit 1
    fi

# go generate writes the full spec in internal/behaviors/all, which comes
# before pkg/sdk/client, where internal/sdk/gen writes the SDK's typed
# services from it. So one run leaves nothing stale.

# Generate code
generate:
    just go-generate
    just sdk-docs-generate
    just cli-docs-generate
    just docs-generate
    just react-generate

# Format, lint, and generate before committing
ready:
    just generate
    just just-fmt
    just docs-fmt
    just md-fmt
    just go-fmt
    just go-vet
    just react-fmt
    just react-lint
    just react-build

[linux]
linux-tune:
    sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"

[macos]
linux-tune:

[windows]
linux-tune:
