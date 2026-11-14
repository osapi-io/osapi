set allow-duplicate-variables

# Optional modules: mod? allows `just fetch` to work before .just/remote/ exists.
# Recipes below use `just` subcommands instead of dependency syntax because just
# validates dependencies at parse time, which would fail when modules aren't loaded.
# The React application lives in ui/, not at the repository root, so the react
# module's default of "." is reassigned above the import.

react_dir := "ui"

# The skills `osapi init` installs open with YAML frontmatter, which mdformat
# rewrites into a heading. Claude Code then never loads the skill.

md_extra_excludes := "--exclude 'cmd/resources/skills/**'"

# Minimum total coverage. Below the org-wide 100% because nine statements are
# unreachable guards that cannot execute. Declared again in .github/codecov.yml
# — change both together.

go_coverage_target := "99.9"

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
fetch:
    mkdir -p .just/remote
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-justfiles/refs/heads/main/go/go.just -o .just/remote/go.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-justfiles/refs/heads/main/docusaurus/docusaurus.just -o .just/remote/docusaurus.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-justfiles/refs/heads/main/md/md.just -o .just/remote/md.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-justfiles/refs/heads/main/just/just.just -o .just/remote/just.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-justfiles/refs/heads/main/react/react.just -o .just/remote/react.just

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
    just just-fmt-check
    just docs-fmt-check
    just behavior-gen-check
    just go-generate-check
    just go-pkg-check
    just react-build
    just go-test
    just behavior-cover-check

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
            find internal/controller/api/next docs/docs/sidebar/behaviors -type f \
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

# The behaviors' spec, under next/, stays out of the join until their cutover:
# it describes the same operations as the hand-written specs, so joining both
# fails on duplicate operationIds.

# Generate code
generate:
    bunx @redocly/cli@{{ redocly_version }} join --prefix-tags-with-info-prop title -o internal/controller/api/gen/api.yaml $(ls internal/controller/api/*/gen/api.yaml | grep -v '/next/') internal/controller/api/node/*/gen/api.yaml
    just go-generate
    just docs-generate
    cp internal/controller/api/gen/api.yaml ui/src/sdk/gen/api.yaml
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

# Run every behavior's integration tests, writes included, against a real
# agent on Debian, in a throwaway container from Dockerfile.dev's image. CI's
# agent cannot write /etc, so this is where the write examples run on Linux.
behavior-int-linux:
    #!/usr/bin/env bash
    set -euo pipefail
    image=$(sed -n 's/^FROM //p' Dockerfile.dev | head -1)
    docker run --rm -v "$PWD":/src:ro -w /src \
        -v osapi-int-gomod:/go/pkg/mod -v osapi-int-gocache:/root/.cache/go-build \
        -e OSAPI_INTEGRATION_WRITE_BEHAVIORS=1 -e GOFLAGS=-buildvcs=false \
        "$image" sh -c '
            (apt-get update -qq && apt-get install -y -qq procps) >/dev/null
            head -c16 /dev/urandom | od -An -tx1 | tr -d " \n" > /etc/machine-id
            go test -tags integration -count=1 -v -run TestBehaviorSmokeSuite ./test/integration/'

[linux]
linux-tune:
    sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"

[macos]
linux-tune:

[windows]
linux-tune:
