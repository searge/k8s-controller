# 002 — Read-only in the cluster for v1

Decided 2026-08-05. Status: accepted.

The only writes are to the controller's own custom resource `status` and its own metrics
endpoint. No patching, no deleting, no creating of anything it does not own.

This is a deliberate constraint on a learning project: the failure modes of a reconcile loop are
easiest to learn while the loop cannot damage anything. If remediation is ever added, it arrives
behind an explicit per-object opt-in annotation, and only after the detector has been correct for
an extended period.

Cost of the constraint, named honestly: see 010.
