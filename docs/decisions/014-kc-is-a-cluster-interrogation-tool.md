# 014 — kc is a cluster interrogation tool, the course is its spine

Decided 2026-08-05. Status: accepted.

Earlier decisions treated one specific gap — Longhorn backup freshness — as the reason the project
exists. That was wrong in a way worth recording, because the mistake is easy to repeat: it fused
"learn to build controllers" with "solve the problem in front of me this month" into a single
21-evening plan, and the fused plan is the same shape as the two plans that already stalled.

Two things break the fusion:

- The problem in front of you changes. A tool built around one storage provider in one cluster is
  worth nothing in the next environment, and defending it then means defending a sunk cost.
- The mechanisms the course teaches are general. Informers, caches, custom resources, reconcile,
  leader election and metrics do not care what question they answer.

So the course's mechanisms are the architecture, and every question the tool answers is a plugin
hanging off it. Longhorn becomes one plugin among others, deletable without a trace, rather than
the foundation. `BackupCheck` is no longer the capstone and there is no `backup-report` command.

The reconcile loop is justified because building one is interesting to the author as an
engineering problem. That is the honest reason and it is sufficient. There is no deadline, no
certification and no external consumer of this project; recording that here prevents a later
retelling in which the project had users.
