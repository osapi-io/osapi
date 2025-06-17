mod go '.just/remote/go.mod.just'
mod bats '.just/remote/bats.mod.just'
mod docs '.just/remote/docs.mod.just'

# --- Fetch ---

# Fetch shared justfiles from osapi-io-justfiles
fetch:
    mkdir -p .just/remote
    curl -sSL https://raw.githubusercontent.com/osapi-io/osapi-io-justfiles/refs/heads/main/go.just > .just/remote/go.just
    curl -sSL https://raw.githubusercontent.com/osapi-io/osapi-io-justfiles/refs/heads/main/bats.just > .just/remote/bats.just
    curl -sSL https://raw.githubusercontent.com/osapi-io/osapi-io-justfiles/refs/heads/main/docs.just > .just/remote/docs.just

# --- Top-level orchestration ---

# Install all dependencies
deps: go::init bats::deps docs::deps

# Run all tests
test: linux-tune go::test _bats-clean bats::test

[private]
_bats-clean:
    rm -f database.db

# Generate code
generate:
    redocly join --prefix-tags-with-info-prop title -o internal/client/gen/api.yaml internal/api/*/gen/api.yaml
    just go::generate
    just docs::generate

[linux]
linux-tune:
    sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"

[macos]
linux-tune:

[windows]
linux-tune:
