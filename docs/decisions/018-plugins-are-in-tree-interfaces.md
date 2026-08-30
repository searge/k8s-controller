# 018 — Plugins are in-tree Go interfaces behind a registry

Decided 2026-08-05. Status: accepted.

A plugin is a Go type satisfying an interface, registered into a registry, declaring the API groups
it requires so 017's detection can decide whether to enable it.

Out-of-process plugins — exec or gRPC, written by other people — are a separate project with
protocol versioning and compatibility guarantees, and they buy nothing at this stage. The in-tree
interface does not preclude them later; it is the boundary they would be built against.
