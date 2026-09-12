#!/usr/bin/env bash
# Sidebar renderer: one line per pane that has agent state.
# Run by tmux-agent-sidebar inside a narrow left pane; loops until killed.

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STATE_DIR="${TMUX_AGENT_SIDEBAR_DIR:-$HOME/.cache/tmux-agent-sidebar}"
INTERVAL="${TMUX_AGENT_SIDEBAR_INTERVAL:-2}"

color_for() {
    case "$1" in
        working) printf '\033[32m' ;;
        waiting) printf '\033[33m' ;;
        done)    printf '\033[36m' ;;
        idle)    printf '\033[90m' ;;
        *)       printf '\033[0m'  ;;
    esac
}

render() {
    printf '\033[H\033[J'
    printf '\033[1m agents\033[0m\r\n'
    printf -- '────────────\r\n'
    found=0
    for f in "$STATE_DIR"/*.json; do
        [ -f "$f" ] || continue
        found=1
        pane="$(sed -n 's/.*"pane"[: ]*"\([^"]*\)".*/\1/p' "$f")"
        status="$(sed -n 's/.*"status"[: ]*"\([^"]*\)".*/\1/p' "$f")"
        summary="$(sed -n 's/.*"summary"[: ]*"\([^"]*\)".*/\1/p' "$f")"
        window="$(sed -n 's/.*"window"[: ]*"\([^"]*\)".*/\1/p' "$f")"
        [ -n "$window" ] || window="$(tmux display-message -p -t "$pane" '#{window_index}:#{window_name}' 2>/dev/null || echo '?')"
        icon="$(color_for "$status")"
        case "$status" in
            working) sym='▸' ;;
            waiting) sym='?' ;;
            done)    sym='✓' ;;
            *)       sym='·' ;;
        esac
        if [ -n "$summary" ]; then
            label="$summary"
        else
            label="$window"
        fi
        printf '%s%s\033[0m %s %s\r\n' "$icon" "$sym" "$window" "$label"
    done
    if [ "$found" -eq 0 ]; then
        printf '\033[90mno agents\033[0m\r\n'
    fi
}

mkdir -p "$STATE_DIR"
while true; do
    render
    sleep "$INTERVAL"
done