# 011 — The first deliverable is a command, with no custom resource

Decided 2026-08-05. Status: superseded by 014.

The custom resource design in [backup-check.md](../backup-check.md) only became obvious after
reading real Longhorn objects; the field list would have been guesswork a week earlier.

So the first deliverable is a read-only CLI report doing the join and printing the buckets.
It validates the object graph, produces a usable answer, and defers every architectural decision
that depends on knowing which fields matter. "Custom resource first, informer later" was
considered and rejected on those grounds.
