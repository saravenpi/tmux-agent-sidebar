# Changelog

## 0.2.0

### Changed

- Auto-close is scoped per window via a `pane-died` hook set when the canvas pane is created, instead of global `pane-died` / `after-kill-pane` hooks that scanned every session
- Auto-close never kills the last window of a session
- Toggle key moved from `prefix + A` to `prefix + b`
- `ensure.sh` opens the Go canvas (was still launching the bash renderer)

### Fixed

- Killing a window whose last pane was the canvas could kill the whole session through the global hook chain
