# 010 — `EndpointProbe` is the course capstone, `BackupCheck` ships first

Decided 2026-08-05. Status: superseded by 019 and 020.

A read-only verification controller teaches informers, caches, custom resources, status
conditions, RBAC, leader election, metrics and `envtest`. It never teaches the other half —
finalizers, owner references, garbage collection, and conflict handling on objects the controller
owns — because it never creates anything.

`EndpointProbe` creates owned Jobs, one per node when a node selector is set, and therefore
forces exactly those lessons. It is the better *learning* capstone.

It is not the better *first* deliverable: `BackupCheck` is smaller, and its scope is fully
determined by resources that already exist in the cluster. So `BackupCheck` ships first and
`EndpointProbe` closes the course.
