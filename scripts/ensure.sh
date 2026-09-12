#!/usr/bin/env bash
# Open a sidebar in every window of the given session that lacks one.
# Called by after-new-window when the session's sidebar is on.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
session="${1:-$(tmux display-message -p '#{session_name}' 2>/dev/null)}"
[ -n "$session" ] || exit 0

[ "$(tmux show-option -gv "@agent-sidebar-on-$session" 2>/dev/null)" = "1" ] || exit 0

width="$(tmux show-option -gv @tmux-agent-sidebar-width 2>/dev/null || echo 28)"
interval="$(tmux show-option -gv @tmux-agent-sidebar-interval 2>/dev/null || echo 2)"

for window in $(tmux list-windows -t "$session" -F '#{window_id}' 2>/dev/null); do
    tmux list-panes -t "$window" -F '#{pane_start_command}' 2>/dev/null | grep -q sidebar.sh && continue
    tmux split-window -h -b -l "$width" -t "$window" \
        "TMUX_AGENT_SIDEBAR_INTERVAL='$interval' bash '$SCRIPT_DIR/sidebar.sh'"
done