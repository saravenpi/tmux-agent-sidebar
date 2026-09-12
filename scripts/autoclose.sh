#!/usr/bin/env bash
# When the sidebar canvas is the only pane left in a window, close the
# window instead of trapping it open. Runs from the pane-exited hook.

MARKER="canvas.sh"

for session in $(tmux list-sessions -F '#{session_name}' 2>/dev/null); do
    for window in $(tmux list-windows -t "$session" -F '#{window_id}' 2>/dev/null); do
        [ "$(tmux list-panes -t "$window" -F '#{pane_id}' 2>/dev/null | wc -l)" -eq 1 ] || continue
        if tmux list-panes -t "$window" -F '#{pane_start_command}' 2>/dev/null | grep -q "$MARKER"; then
            tmux kill-window -t "$window" 2>/dev/null
        fi
    done
done