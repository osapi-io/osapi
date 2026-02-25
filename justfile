# Optional modules: mod? allows `just fetch` to work before .just/remote/ exists.
# Recipes below use `just` subcommands instead of dependency syntax because just
# validates dependencies at parse time, which would fail when modules aren't loaded.
mod? go '.just/remote/go.mod.just'
mod? bats '.just/remote/bats.mod.just'
mod? docs '.just/remote/docs.mod.just'

# --- Fetch ---

# Fetch shared justfiles from osapi-io-justfiles
fetch:
    mkdir -p .just/remote
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-io-justfiles/refs/heads/main/go.mod.just -o .just/remote/go.mod.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-io-justfiles/refs/heads/main/go.just -o .just/remote/go.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-io-justfiles/refs/heads/main/bats.mod.just -o .just/remote/bats.mod.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-io-justfiles/refs/heads/main/bats.just -o .just/remote/bats.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-io-justfiles/refs/heads/main/docs.mod.just -o .just/remote/docs.mod.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-io-justfiles/refs/heads/main/docs.just -o .just/remote/docs.just

# --- Top-level orchestration ---

# Install all dependencies
deps:
    just go::deps
    just go::mod
    just bats::deps
    just docs::deps

# Run all tests
test: linux-tune _bats-clean
    just go::test
    just bats::test

[private]
_bats-clean:
    rm -f database.db

# Generate code
generate:
    just go::generate
    just docs::generate

[linux]
linux-tune:
    sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"

[macos]
linux-tune:

[windows]
linux-tune:
