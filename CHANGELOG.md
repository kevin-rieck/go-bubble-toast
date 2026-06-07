# Changelog

## v0.2.0 - 2026-06-07

- Added README demo GIFs showing the feature showcase, startup Toast, Toast stack, and progress bar behavior.
- Added `examples/progress` for Toast lifetime progress bars.
- Added Toast lifetime helpers and tests.
- Fixed progress bar wrapping and queued Toast timer synchronization.
- Updated dependencies.

## v0.1.0 - 2026-06-06

Initial public release.

- Bubble Tea model for transient, non-blocking terminal toasts.
- Built-in info, success, warning, error, and neutral toast kinds.
- Message-based `Show`, `Dismiss`, and `DismissAll` commands.
- Direct `Model.Push` helper for small apps and tests.
- Queueing, replacement by ID, persistent toasts, and timed expiration.
- Configurable placement, spacing, margins, widths, max height, theme, and custom renderer.
- Runnable examples under `examples/`.
