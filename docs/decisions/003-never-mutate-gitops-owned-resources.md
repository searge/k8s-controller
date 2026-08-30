# 003 — Never mutate resources owned by a GitOps controller

Decided 2026-08-05. Status: accepted.

A controller that patches objects an ArgoCD or Flux application owns will fight that tool's
self-healing, and lose in a way that is confusing to debug: the patch lands, reverts seconds
later, and the logs of both controllers look normal.

Where a fix belongs in git, the output is a report a human turns into a pull request — not a live
mutation. This holds even where the mutation would be correct.
