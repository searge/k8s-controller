# 019 — `Check` replaces `FrontendPage` and adds status conditions

Decided 2026-08-05. Status: accepted.

`spec` names the plugin, its parameters and how often to re-evaluate. `status.conditions` carries
the result. Reconcile re-runs the check and updates the status. Leader election is genuinely
required rather than decorative: two replicas evaluating the same check write conflicting status.

Writes are limited to the resource's own status, per 002.

Two facts about the reference implementation shaped this. Its `FrontendPage` type has no `status`
field at all, and no custom resource in the course has one, so status conditions are a gap in the
course rather than something being replaced. And its reconciler creates a ConfigMap and a
Deployment, which is where 020 comes in.
