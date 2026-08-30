# 007 — Join Longhorn backup records on `spec.volumeName`

Decided 2026-08-05. Status: accepted, applies to the Longhorn plugin only, see 014.

`BackupVolume.metadata.name` carries a hash suffix and is not the volume name; `spec.volumeName`
is. Joining on the object name silently matches nothing, which reads as "no backups exist" rather
than as a bug.

Full object graph and the other two traps — join direction, and
`status.labels.KubernetesStatus` being history rather than current state — in
[backup-check.md](../backup-check.md).
