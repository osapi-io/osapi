set allow-duplicate-variables

# Optional modules: mod? allows `just fetch` to work before .just/remote/ exists.
# Recipes below use `just` subcommands instead of dependency syntax because just
# validates dependencies at parse time, which would fail when modules aren't loaded.
# The React application lives in ui/, not at the repository root, so the react
# module's default of "." is reassigned above the import.

react_dir := "ui"

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
    just go-mod
    just docusaurus-deps
    just react-deps

# Build production binary (includes embedded UI)
build:
    just react-build
    go build -o osapi .

# Run all tests
test: linux-tune
    just just-fmt-check
    just react-build
    just go-test

# Generate code
generate:
    bunx @redocly/cli@{{ redocly_version }} join --prefix-tags-with-info-prop title -o internal/controller/api/gen/api.yaml internal/controller/api/*/gen/api.yaml internal/controller/api/node/*/gen/api.yaml
    just go-generate
    just docusaurus-generate
    cp internal/controller/api/gen/api.yaml ui/src/sdk/gen/api.yaml
    just react-generate

# Format, lint, and generate before committing
ready:
    just generate
    just just-fmt
    just docusaurus-fmt
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
