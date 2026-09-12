#!/usr/bin/env bash
# tmux-agent-sidebar plugin entry point (run by TPM).

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

tmux bind-key A run-shell "bash '$CURRENT_DIR/scripts/toggle.sh'"

exit 0