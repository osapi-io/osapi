set allow-duplicate-variables

# Optional modules: mod? allows `just fetch` to work before .just/remote/ exists.
# Recipes below use `just` subcommands instead of dependency syntax because just
# validates dependencies at parse time, which would fail when modules aren't loaded.
# The React application lives in ui/, not at the repository root, so the react
# module's default of "." is reassigned above the import.

react_dir := "ui"

# The skills `osapi init` installs open with YAML frontmatter, which mdformat
# rewrites into a heading. Claude Code then never loads the skill.

md_extra_excludes := "--exclude 'internal/skills/resources/**'"

# Pinned so the combined spec does not change under whoever runs `just
# generate`. Run through bunx rather than from the path: bun is declared in
# .mise.toml, redocly was not declared anywhere.

redocly_version := "2.19.1"

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
go_quick_checks := "just-fmt-check docs-fmt-check react-fmt-check go-mod-check go-pkg-check"
go_slow_checks := "behavior-gen-check go-generate-check docs-gen-check cli-docs-gen-check behavior-scaffold-check endpoint-scaffold-check command-scaffold-check behavior-cover-check go-fmt-check go-vet examples-check spec-client-check"

# Fail when a behavior that osapi behavior create writes does not pass its tests
#
# The scaffold must keep up with the checks of behaviortest, as each behavior
# must. The recipe writes a behavior in a copy of the checkout, generates it,
# and runs its tests. The checkout itself does not change.
behavior-scaffold-check:
    #!/usr/bin/env bash
    set -euo pipefail
    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT
    rsync -a --exclude .git --exclude .worktrees --exclude node_modules \
        --exclude /dist --exclude /examples --exclude /docs/build ./ "$tmp/"
    cd "$tmp"
    go run . behavior create --name scaffoldcheck --description "A scaffold check" >/dev/null
    go generate ./internal/behaviors/scaffoldcheck/mocks ./internal/behaviors/all
    go test -count=1 ./internal/behaviors/scaffoldcheck/...

# Fail when a group that osapi controller endpoint create writes does not pass
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
    go run . controller endpoint create --name scaffoldcheck --description "A scaffold check" >/dev/null
    go generate ./internal/controller/endpoints/all
    go run . controller endpoint validate >/dev/null
    go test -count=1 ./internal/controller/endpoints/scaffoldcheck/... ./internal/controller/endpoints/all/ ./internal/controller/

# Fail when a group that osapi cli command create writes does not pass its tests
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
    go run . cli command create --name scaffoldcheck --description "A scaffold check" >/dev/null
    go generate ./internal/cli/commands/all
    go run . cli command validate >/dev/null
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

# Compile a client that oapi-codegen writes from the joined spec with its
# default options, as a person outside osapi would. The SDK's own client has
# its config; this holds the spec to a stranger's generator. The package goes
# in the module, so it builds with osapi's dependencies, and is removed after.
spec-client-check:
    #!/usr/bin/env bash
    set -euo pipefail
    dir=$(mktemp -d internal/zzspecclientXXXXXX)
    trap 'rm -rf "$dir"' EXIT
    go tool oapi-codegen -generate types,client -package specclient \
        internal/controller/api/gen/api.yaml > "$dir/client.go"
    if ! go build "./$dir"; then
        echo "a client that oapi-codegen generates from the joined spec does not compile" >&2
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
    broken=0
    for dir in examples/sdk/client examples/sdk/platform; do
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
    go run . cli docs generate --dir docs/docs/sidebar/usage/cli
    cd docs && bun run prettier docs/sidebar/usage/cli --write --config prettier.config.js --ignore-path .prettierignore > /dev/null

# The CLI reference pages are generated from the cobra command tree. A command
# without a page, a page without a command, or a flag that changed fails here.
cli-docs-gen-check:
    #!/usr/bin/env bash
    set -euo pipefail
    pages() {
        find docs/docs/sidebar/usage/cli -type f -print | sort | xargs shasum
    }
    before=$(pages)
    just cli-docs-generate
    if [ "$before" != "$(pages)" ]; then
        diff <(echo "$before") <(pages) || true
        echo "CLI reference pages are stale: run just cli-docs-generate and commit the result" >&2
        exit 1
    fi

# The API reference pages under docs/docs/gen are generated from the joined
# OpenAPI specs. Nothing compared them with the specs, so /api/next pages went
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
    generated() {
        {
            find internal/behaviors -type f -path '*/gen/*'
            find docs/docs/sidebar/behaviors -type f \
                -not -name behaviors.mdx -not -name _category_.json
        } 2>/dev/null | sort | xargs shasum
    }
    before=$(generated)
    go generate ./internal/behaviors/all
    if [ "$before" != "$(generated)" ]; then
        diff <(echo "$before") <(generated) || true
        echo "behavior output is stale: run go generate ./internal/behaviors/all and commit the result" >&2
        exit 1
    fi
    # A served behavior's spec joins the main one, which the SDK, UI and API
    # docs are generated from, so a behavior changed without just generate
    # leaves all three describing its old shape.
    joined=$(mktemp -d)/api.yaml
    just join-spec "$joined" > /dev/null
    if ! diff -q "$joined" internal/controller/api/gen/api.yaml > /dev/null; then
        diff "$joined" internal/controller/api/gen/api.yaml | head -20 || true
        echo "the joined spec is stale: run just generate and commit the result" >&2
        exit 1
    fi

# The join reads the hand-written specs and each behavior's gen/api.yaml.

# The endpoint groups' and the behaviors' gen/api.yaml are written by go
# generate, and the join reads them. The SDK client is generated, in the same
# go generate pass, from the joined spec. So those specs are written first,
# then joined, then the rest is generated: one run leaves nothing stale.
# Temporary: the first step goes when the generators write the full spec and
# the join is removed (#833).

# Generate code
generate:
    go generate ./internal/controller/endpoints/all ./internal/behaviors/all ./internal/cli/commands/all
    just join-spec internal/controller/api/gen/api.yaml
    just go-generate
    just cli-docs-generate
    just docs-generate
    just react-generate

# redocly join takes the info of the joined spec, its title and description,
# from the first file. The components spec has the API's own, so it is first.
# Join the hand-written specs and each served behavior's gen/api.yaml into out
join-spec out:
    bunx @redocly/cli@{{ redocly_version }} join --prefix-tags-with-info-prop title -o {{ out }} internal/controller/api/components/gen/api.yaml $(find internal/controller/endpoints -path '*/gen/api.yaml' | sort) $(ls internal/controller/api/*/gen/api.yaml | grep -v -e '/components/' -e '^internal/controller/api/gen/') internal/controller/api/node/*/gen/api.yaml $(find internal/behaviors -path '*/gen/api.yaml' | sort)

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

# Run every behavior's integration tests, writes included, against a real
# agent on Debian, in a throwaway container from Dockerfile.dev's image. CI's
# agent cannot write /etc, so this is where the write examples run on Linux.
#
# The sysctl examples write kernel parameters, so the container is privileged:
# without it /proc/sys is read-only. Some of those parameters, such as
# vm.swappiness, are not per container. On macOS they change only Docker's
# VM; on a Linux machine they would change that machine's kernel. Thus the
# recipe runs on macOS only.
[macos]
behavior-int-linux:
    #!/usr/bin/env bash
    set -euo pipefail
    image=$(sed -n 's/^FROM //p' Dockerfile.dev | head -1)
    docker run --rm --privileged -v "$PWD":/src:ro -w /src \
        -v osapi-int-gomod:/go/pkg/mod -v osapi-int-gocache:/root/.cache/go-build \
        -e OSAPI_INTEGRATION_WRITE_BEHAVIORS=1 -e GOFLAGS=-buildvcs=false \
        "$image" sh -c '
            (apt-get update -qq && apt-get install -y -qq procps systemd) >/dev/null
            head -c16 /dev/urandom | od -An -tx1 | tr -d " \n" > /etc/machine-id
            go test -tags integration -count=1 -v -run TestBehaviorSmokeSuite ./test/integration/'

[linux]
linux-tune:
    sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"

[macos]
linux-tune:

[windows]
linux-tune:
