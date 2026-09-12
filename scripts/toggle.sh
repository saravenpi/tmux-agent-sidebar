#!/usr/bin/env bash
# Toggle the agent sidebar for the current session.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MARKER="tmux-agent-sidebar"

session="$(tmux display-message -p '#{session_name}')"

if tmux list-panes -t "$session" -F '#{pane_start_command}' 2>/dev/null | grep -q "$MARKER"; then
    for pane in $(tmux list-panes -t "$session" -F '#{pane_id} #{pane_start_command}' | grep "$MARKER" | awk '{print $1}'); do
        tmux kill-pane -t "$pane"
    done
    exit 0
fi

width="$(tmux show-option -gv @tmux-agent-sidebar-width 2>/dev/null || echo 28)"
interval="$(tmux show-option -gv @tmux-agent-sidebar-interval 2>/dev/null || echo 2)"

tmux split-window -h -b -l "$width" -t "$session" \
    "TMUX_AGENT_SIDEBAR_INTERVAL='$interval' bash '$SCRIPT_DIR/sidebar.sh'"