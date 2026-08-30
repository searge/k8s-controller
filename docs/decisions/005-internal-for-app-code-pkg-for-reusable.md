# 005 — Application code lives in `internal/`, `pkg/` is only for reusable code

Decided 2026-08-05. Status: accepted, revisit at course step 11.

The repository grew with everything under `pkg/` and an empty `internal/`, which is backwards for
an application with no external consumers: `pkg/` is a promise of a stable API to importers who
do not exist.

New domain code goes to `internal/`. Existing packages move as the step that touches them
comes around, not in a dedicated refactor — see 012.
