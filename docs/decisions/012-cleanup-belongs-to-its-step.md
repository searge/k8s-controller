# 012 — Cleanup belongs to the step that needs it, never its own phase

Decided 2026-08-05. Status: accepted.

Two successive plans for this project opened with a refactoring-and-documentation phase. Neither
reached the phase after it.

Concretely: the HTTP server gains a context and a `Shutdown` method in the step that needs
to run an informer alongside it, not in a preceding cleanup pass. The `pkg/` to `internal/` move
happens per-package as steps touch them. Known defects with no step that needs them
yet stay on a list and stay unfixed, which is the honest state.
