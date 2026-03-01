# Optional modules: mod? allows `just fetch` to work before .just/remote/ exists.
# Recipes below use `just` subcommands instead of dependency syntax because just
# validates dependencies at parse time, which would fail when modules aren't loaded.
mod? go '.just/remote/go.mod.just'
mod? docs '.just/remote/docs.mod.just'

# --- Fetch ---

# Fetch shared justfiles from osapi-justfiles
fetch:
    mkdir -p .just/remote
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-justfiles/refs/heads/main/go.mod.just -o .just/remote/go.mod.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-justfiles/refs/heads/main/go.just -o .just/remote/go.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-justfiles/refs/heads/main/docs.mod.just -o .just/remote/docs.mod.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-justfiles/refs/heads/main/docs.just -o .just/remote/docs.just

# --- Top-level orchestration ---

# Install all dependencies
deps:
    just go::deps
    just go::mod
    just docs::deps

# Run all tests
test: linux-tune
    just go::test

# Generate code
generate:
    redocly join --prefix-tags-with-info-prop title -o internal/api/gen/api.yaml internal/api/*/gen/api.yaml
    just go::generate
    just docs::generate

[linux]
linux-tune:
    sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"

[macos]
linux-tune:

[windows]
linux-tune:
