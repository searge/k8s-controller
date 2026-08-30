# 008 — Enrolment comes from RecurringJob groups, not PVC labels

Decided 2026-08-05. Status: accepted, applies to the Longhorn plugin only, see 014.

The intuitive design — a label selector over PVCs — is wrong for Longhorn. A `RecurringJob` in
the `default` group covers every volume carrying no recurring-job label of its own, so unlabelled
volumes are backed up and a selector-driven check reports them as unenrolled.

Enrolment is therefore computed: resolve the backup jobs, resolve the groups, and read the labels
Longhorn writes back onto the `Volume` resource rather than the labels a human may or may not
have put on the PVC. Resolution order in [backup-check.md](../backup-check.md).

Consequence for the custom resource in 011: its selector chooses **what to check**, which is not
the same field as **what is enrolled**.
