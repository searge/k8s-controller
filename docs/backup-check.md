# BackupCheck

> [!NOTE]
> **No longer the project's purpose — kept as the design for a future Longhorn plugin.**
> On 2026-08-05 the project settled on being a general cluster interrogation tool with the course
> mechanisms as its spine and every question answered by a plugin
> ([DECISIONS.md](DECISIONS.md) 014). Longhorn is one such plugin, and it is not the first one.
>
> The findings below are worth keeping and were verified against a live cluster: the join key trap,
> the join direction, `status.labels.KubernetesStatus` being history, and the `RecurringJob`
> `default`-group semantics that make label-driven enrolment wrong. Whoever writes the plugin needs
> all of it.
>
> Ignore the milestone references and the custom resource sketch: the resource is now a generic
> `Check` whose `spec` names a plugin (decision 019), not a Longhorn-specific type.

The first domain this controller works in. It answers one question about Longhorn-backed
PersistentVolumeClaims, and the value of the answer depends entirely on how narrowly the
question is stated.

## The assertion

> For every live PVC that Longhorn's recurring jobs cover, the backup target is available and
> the newest completed backup recorded for that volume is younger than the configured policy.

## What it does not assert

Stating these explicitly, because each one is a plausible misreading of the sentence above.

- **Not restorability.** Longhorn's custom resources record that a backup completed and how
  large it was. Proving a backup restores means restoring it somewhere disposable and checking
  the result boots. That is a separate, much larger piece of work.
- **Not application-level dumps.** A `pg_dump` or `mysqldump` written to a file share is a
  different artefact in a different location, and checking it needs filesystem access. Out of
  scope; tracked as a possible `LogicalBackupCheck` in [ROADMAP.md](ROADMAP.md).
- **Not policy correctness.** Whether a 24-hour freshness target is the right target for a given
  volume is a human decision. The controller checks the policy it is given.
- **Not enrolment correctness.** It reports that a volume is uncovered. It does not decide
  whether that volume *should* be covered.

## Object graph

Everything needed lives in the Kubernetes API. No filesystem access, which keeps RBAC small and
deployment simple.

```text
PersistentVolumeClaim          (core/v1)
  spec.volumeName
      |
      v
PersistentVolume               (core/v1)
  spec.csi.driver == "driver.longhorn.io"
  spec.csi.volumeHandle  ------------------.
                                           |  equal strings
Volume            (volumes.longhorn.io)    |
  metadata.name  <-------------------------'
  metadata.labels["recurring-job.longhorn.io/<job>"]
  metadata.labels["recurring-job-group.longhorn.io/<group>"]
  status.lastBackupAt
      |
      |  Volume.name == BackupVolume.spec.volumeName
      v
BackupVolume      (backupvolumes.longhorn.io)
  spec.volumeName
  status.lastBackupAt
  status.lastBackupName
  status.lastSyncedAt

RecurringJob      (recurringjobs.longhorn.io)   -- decides who is enrolled
  spec.task, spec.groups, spec.cron, spec.retain

BackupTarget      (backuptargets.longhorn.io)   -- gates the whole report
  status.available, status.lastSyncedAt
```

### Join key: `spec.volumeName`, never `metadata.name`

Longhorn appends a hash suffix to `BackupVolume` object names. An object named

```text
pvc-11111111-2222-3333-4444-555555555555-abcdef01
```

carries `spec.volumeName: pvc-11111111-2222-3333-4444-555555555555`. Joining on
`metadata.name` therefore matches nothing. The `backup-volume` label carries the same value as
`spec.volumeName` and works as an alternative.

### Join direction: start from the live PVC

Starting from `BackupVolume` and looking for a PVC produces false alarms, because the target
retains records for volumes whose PVC was deleted long ago. Starting from the live PVC and
walking outward keeps those records in their own bucket instead of mixing them into the
freshness verdict.

### `status.labels.KubernetesStatus` is not a join key

`BackupVolume.status.labels.KubernetesStatus` is a JSON *string* holding `pvName`, `namespace`,
`pvcName` and workload references as they were when the backup ran. It is history, not current
state: after a PVC is deleted and recreated, or a volume is restored under a new name, these
fields still describe the old object. Useful for explaining an orphan to a human, never for
deciding whether a PVC is covered.

## Enrolment: derived from RecurringJob groups, not from PVC labels

This is the part most likely to be implemented wrong, because the obvious approach fails.

Longhorn's rule: a `RecurringJob` listing `default` in `spec.groups` applies to every volume that
carries no recurring-job label of its own. Consequences:

1. A volume with **no** `recurring-job.longhorn.io/*` label is still backed up, as long as some
   backup job is in the `default` group. A label-selector-driven design reports such a volume as
   unenrolled, which is wrong.
2. Longhorn writes the resolved job labels back onto the `Volume` custom resource. So the labels
   on `Volume` are a better coverage signal than the labels on the PVC, which may have none.
3. Only `spec.task == "backup"` enrols a volume for *backup*. The snapshot tasks
   (`snapshot`, `snapshot-cleanup`, `snapshot-delete`) attach identical-looking labels and
   produce no backups at all. Matching on the label prefix without checking the task
   over-reports coverage.

Resolution order for "is this volume enrolled for backup":

```text
1. Volume has recurring-job.longhorn.io/<job> for a job with task == "backup"   -> enrolled
2. Volume has recurring-job-group.longhorn.io/<group> containing a backup job   -> enrolled
3. Volume has no recurring-job or group labels at all,
   and some backup job lists "default" in spec.groups                           -> enrolled
4. otherwise                                                                    -> not enrolled
```

## Buckets

| Bucket | Condition | Meaning |
| --- | --- | --- |
| `fresh` | enrolled, `BackupVolume` present, newest backup within policy | Working as intended |
| `stale` | enrolled, `BackupVolume` present, newest backup older than policy | Backups were running and stopped |
| `uncovered` | enrolled, **no `BackupVolume` at all** | Enrolled and never backed up, or the record vanished |
| `unenrolled` | live PVC, no backup job applies | Reported for visibility; not necessarily a fault |
| `orphan` | `BackupVolume` with no live PVC | Target retains history for a deleted volume |

`uncovered` is the most serious state and the easiest to miss, because a bucket scheme built
only from backup ages cannot express it: there is no age to compare. It is the state a volume
reaches when its enrolment disappears — the record of past backups may still exist under a
different volume name, or not at all.

`orphan` is not a fault. It is target-side history plus, over time, consumed space on the backup
target. Its value is that separating orphans from `stale` is exactly what makes the `stale`
count trustworthy; without the join, every deleted volume's leftover record looks like a
neglected backup.

## Freshness policy and schedule gaps

A naive `--max-age 24h` produces a false alarm on any day following a day the backup job does
not run. A cron of `20 4 * * 2-6` runs Tuesday through Saturday, so on Sunday and Monday the
newest backup is legitimately more than 24 hours old.

Two ways to handle this:

- **v1:** an operator-supplied `--max-age` with a default wide enough to span the schedule's
  largest gap, and the job's cron shown in the report so the number is explicable.
- **later:** derive the expected interval from `RecurringJob.spec.cron` and flag a volume only
  once it misses a scheduled run. Correct, and it needs a cron parser plus a decision about
  timezones.

`RecurringJob.spec.retain` is the natural source for a `minCompletedBackups` policy: a job
retaining 3 backups cannot satisfy a policy demanding 10.

## Target availability gates everything

If `BackupTarget.status.available` is false, the freshness numbers below it are **unverifiable
rather than false** — Longhorn cannot sync the target, so `status.lastBackupAt` on every
`BackupVolume` is as stale as `status.lastSyncedAt`. Reporting those volumes as `stale` conflates
two different failures. The report states target availability first and marks freshness verdicts
as unknown when the target is down.

The same applies at a smaller scale to a single `BackupVolume` whose `status.lastSyncedAt` is old.

## Known limits

- **Backup history is keyed to the Longhorn volume name.** A volume restored under a new name
  starts with no history, so it appears `uncovered` even though its data was recovered from a
  backup. Needs an explicit test case, and probably an annotation-based grace period.
- **A `Volume.status.lastBackupAt` may come from a manual backup**, not from the recurring job.
  It proves a backup happened; it does not prove the schedule is working.
- **Per-PVC metric series** are fine at the scale of tens of volumes. At hundreds of namespaces
  the cardinality needs review before this is deployed widely.

## Custom resource sketch

Introduced at milestone 4, not before — the field list below only became obvious after reading
real objects.

The API group below is a placeholder. Picking the real one is part of milestone 4, and it needs
a domain that will not collide with anything upstream.

```yaml
apiVersion: PLACEHOLDER/v1alpha1
kind: BackupCheck
spec:
  # Namespaces and PVCs in scope. Note this selects *what to check*, which is not
  # the same as *what is enrolled* -- enrolment is derived from RecurringJob groups.
  namespaceSelector: {matchLabels: {}}
  pvcSelector: {matchLabels: {}}
  maxBackupAge: 72h
  minCompletedBackups: 3
  backupTargetName: default
status:
  conditions:
    - type: TargetAvailable      # the target synced recently
    - type: MetadataFresh        # every covered volume is within maxBackupAge
    - type: CoverageComplete     # no enrolled volume is in the uncovered bucket
  coveredPVCs: 0
  staleVolumes: []
  uncoveredVolumes: []
  lastCheckTime: null
```

Three conditions rather than one, because the three failures need different responses: an
unavailable target is an infrastructure problem, a stale volume is a scheduling problem, and an
uncovered volume is an enrolment problem.

## Reading list

- [Longhorn: recurring snapshots and backups](https://longhorn.io/docs/latest/snapshots-and-backups/scheduling-backups-and-snapshots/)
- [Longhorn: backup target setup](https://longhorn.io/docs/latest/snapshots-and-backups/backup-and-restore/set-backup-target/)
- [Longhorn API reference](https://longhorn.io/docs/latest/references/api/)
