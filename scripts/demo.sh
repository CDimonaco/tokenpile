#!/usr/bin/env bash
# Demo script for asciinema recording.
# Usage: asciinema rec demo.cast --command ./scripts/demo.sh
#
# Records a clean session using an isolated temporary DB.
# Skips the interactive TUI (record that separately as a GIF).

set -euo pipefail

DEMO_DIR="$(mktemp -d)"
export XDG_DATA_HOME="$DEMO_DIR"
export XDG_CONFIG_HOME="$DEMO_DIR/config"
DEMO_REPO="cdimonaco/tokenpile"

TOKENPILE="${TOKENPILE_BIN:-$(cd "$(dirname "$0")/.." && pwd)/tokenpile}"

# --- helpers ---

type_cmd() {
    printf "\033[1;32m\$\033[0m "
    local cmd="$1"
    local i
    for (( i=0; i<${#cmd}; i++ )); do
        printf "%s" "${cmd:$i:1}"
        sleep 0.04
    done
    echo
}

run() {
    type_cmd "$*"
    sleep 0.3
    "$@" || true
    echo
    sleep 0.8
}

pause() {
    sleep "${1:-1.2}"
}

section() {
    echo
    printf "\033[1;34m# %s\033[0m\n" "$1"
    echo
    sleep 0.6
}

cleanup() {
    rm -rf "$DEMO_DIR"
}
trap cleanup EXIT

# --- demo ---

clear
sleep 0.5

section "check version"
run "$TOKENPILE" --version

section "list supported agent integrations"
run "$TOKENPILE" skill list

section "install tokenpile skill for claude-code"
run "$TOKENPILE" skill install --agent claude-code

section "log usage: first session on issue #1"
run "$TOKENPILE" log \
    --repo "$DEMO_REPO" \
    --issue 1 \
    --agent claude-code \
    --model claude-opus-4-8 \
    --tokens-in 3200 \
    --tokens-out 890

run "$TOKENPILE" log \
    --repo "$DEMO_REPO" \
    --issue 1 \
    --agent claude-code \
    --model claude-opus-4-8 \
    --tokens-in 1800 \
    --tokens-out 420

section "log usage: same issue, different agent and model"
run "$TOKENPILE" log \
    --repo "$DEMO_REPO" \
    --issue 1 \
    --agent opencode \
    --model claude-sonnet-5 \
    --tokens-in 950 \
    --tokens-out 310

section "text report for issue #1"
run "$TOKENPILE" report \
    --repo "$DEMO_REPO" \
    --issue 1

section "export signed JSON for issue #1"
run "$TOKENPILE" export \
    --repo "$DEMO_REPO" \
    --issue 1 \
    --output /tmp/tokenpile-issue1.json

section "verify the signed export"
run "$TOKENPILE" export verify --file /tmp/tokenpile-issue1.json

section "done — see the interactive TUI demo in the GIF"
sleep 1.5
