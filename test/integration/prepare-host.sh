#!/usr/bin/env bash
# prepare-host.sh makes a Linux host ready for the integration suite's write
# examples. CI runs it on its runner before the suite, and osapi-lab runs it
# on the target VM before the suite. Each step does nothing when it is done.
# Run it as root.
set -euo pipefail

# The interface write examples configure osapi0, a dummy interface, so they
# never change the host's own network.
if ! ip link show osapi0 >/dev/null 2>&1; then
    ip link add osapi0 type dummy
fi

# The ntp examples run chronyc and write to /etc/chrony/sources.d.
if ! command -v chronyc >/dev/null 2>&1; then
    if command -v apt-get >/dev/null 2>&1; then
        apt-get update -qq && apt-get install -y -qq chrony
    elif command -v dnf >/dev/null 2>&1; then
        dnf install -y -q chrony
    elif command -v zypper >/dev/null 2>&1; then
        zypper --non-interactive install -y chrony
    fi
fi

# The user create examples run useradd --create-home, which copies
# /etc/skel. A runner's /etc/skel holds its toolchains for new users, and the
# first copy took more than the job's wait. The examples need a home, not the
# toolchains. Only a CI runner empties it: OSAPI_PREPARE_EMPTY_SKEL=1.
if [ "${OSAPI_PREPARE_EMPTY_SKEL:-}" = "1" ]; then
    find /etc/skel -mindepth 1 -delete
fi
