# Changelog

## 0.2.2

### Fixed

- Sidebar pane now opens at the far left edge of the window in every layout, and focus returns to the leftmost non-canvas pane

## 0.2.1

### Changed

- Extra blank line above the " agents" title and a leading space before spinner/status icons

## 0.2.0

### Changed

- Auto-close is scoped per window via a `pane-died` hook set when the canvas pane is created, instead of global `pane-died` / `after-kill-pane` hooks that scanned every session
- Auto-close never kills the last window of a session
- Toggle key moved from `prefix + A` to `prefix + b`
- `ensure.sh` opens the Go canvas (was still launching the bash renderer)

### Fixed

- Killing a window whose last pane was the canvas could kill the whole session through the global hook chain
