# Decisions

Numbered, dated, and kept in one file until the file gets unwieldy. Each entry records what was
decided, the reasoning, and — where it matters — what was rejected, so a rejected option does not
get proposed again as if it were new.

Status values: **accepted**, **superseded by NNN**, **revisit at milestone N**.

---

## 001 — The project targets assertions about state outside the API server

Decided 2026-08-05. Status: accepted.

A controller earns its complexity when the truth it needs cannot be read from the API server
alone. Checking that an object's fields are well-formed is admission control; checking that
resource requests are sane is autoscaling. Both already have mature implementations, and writing
Go to duplicate them produces a worse copy of an installed tool.

What has no off-the-shelf answer is asserting that an **external artefact still exists and is
usable**: a backup on a remote target, the reachability of an endpoint from the specific node a
pod happens to run on. Those need a real read, or a real connection attempt, on a schedule, with
the result recorded as cluster state.

Rejected as project scope: egress policy enforcement, resource right-sizing, WAF or
rate-limiting, anything managing host-level cron. Each is either admission-time work or not
Kubernetes-shaped at all.

## 002 — Read-only in the cluster for v1

Decided 2026-08-05. Status: accepted.

The only writes are to the controller's own custom resource `status` and its own metrics
endpoint. No patching, no deleting, no creating of anything it does not own.

This is a deliberate constraint on a learning project: the failure modes of a reconcile loop are
easiest to learn while the loop cannot damage anything. If remediation is ever added, it arrives
behind an explicit per-object opt-in annotation, and only after the detector has been correct for
an extended period.

Cost of the constraint, named honestly: see 010.

## 003 — Never mutate resources owned by a GitOps controller

Decided 2026-08-05. Status: accepted.

A controller that patches objects an ArgoCD or Flux application owns will fight that tool's
self-healing, and lose in a way that is confusing to debug: the patch lands, reverts seconds
later, and the logs of both controllers look normal.

Where a fix belongs in git, the output is a report a human turns into a pull request — not a live
mutation. This holds even where the mutation would be correct.

## 004 — If a check can be a Prometheus rule, it should be

Decided 2026-08-05. Status: accepted.

A controller that grows into a second alerting system is a liability: two places to look, two
notification paths, two sets of thresholds that drift apart.

The rule: the controller exists to produce assertions Prometheus cannot express — ones needing a
multi-resource join or an external read. Once the assertion is a number, alerting on it is
Prometheus' job, and the deliverable is an example `PrometheusRule`, not a notifier.

## 005 — Application code lives in `internal/`, `pkg/` is only for reusable code

Decided 2026-08-05. Status: accepted, revisit at milestone 4.

The repository grew with everything under `pkg/` and an empty `internal/`, which is backwards for
an application with no external consumers: `pkg/` is a promise of a stable API to importers who
do not exist.

New domain code goes to `internal/`. Existing packages move as the milestone that touches them
comes around, not in a dedicated refactor — see 012.

## 006 — Unstructured dynamic client for Longhorn resources, not typed imports

Decided 2026-08-05. Status: accepted, revisit at milestone 4.

Two options for reading `volumes.longhorn.io` and friends:

- **Typed:** import Longhorn's Go API package. Cleaner code, compile-time field checking, and it
  pins the project to one Longhorn minor version — an upgrade becomes a dependency bump plus a
  build break.
- **Unstructured:** the dynamic client with `unstructured.Unstructured`. More verbose, no
  compile-time checking of field paths, and it survives Longhorn upgrades and works against a
  cluster where the CRDs are a slightly different version.

Chose unstructured, because the alternative couples a learning project's build to a third-party
operator's release cadence, and because the field set actually read here is small enough that the
verbosity is bounded. The accessor calls are wrapped in one place so the field paths appear once,
not scattered.

Revisit if the field count grows past what a handful of accessors can carry.

## 007 — Join Longhorn backup records on `spec.volumeName`

Decided 2026-08-05. Status: accepted.

`BackupVolume.metadata.name` carries a hash suffix and is not the volume name; `spec.volumeName`
is. Joining on the object name silently matches nothing, which reads as "no backups exist" rather
than as a bug.

Full object graph and the other two traps — join direction, and
`status.labels.KubernetesStatus` being history rather than current state — in
[backup-check.md](backup-check.md).

## 008 — Enrolment comes from RecurringJob groups, not PVC labels

Decided 2026-08-05. Status: accepted.

The intuitive design — a label selector over PVCs — is wrong for Longhorn. A `RecurringJob` in
the `default` group covers every volume carrying no recurring-job label of its own, so unlabelled
volumes are backed up and a selector-driven check reports them as unenrolled.

Enrolment is therefore computed: resolve the backup jobs, resolve the groups, and read the labels
Longhorn writes back onto the `Volume` resource rather than the labels a human may or may not
have put on the PVC. Resolution order in [backup-check.md](backup-check.md).

Consequence for the custom resource in 011: its selector chooses **what to check**, which is not
the same field as **what is enrolled**.

## 009 — A freshness policy must account for gaps in the backup schedule

Decided 2026-08-05. Status: accepted, revisit at milestone 4.

A flat 24-hour freshness threshold produces a false alarm every time the backup schedule skips a
day. A schedule running Tuesday through Saturday legitimately leaves a two-day-old backup on
Monday.

v1 takes the threshold from the operator and defaults it wide enough to span the schedule's
largest gap, showing the job's cron in the report so the default is explicable. Deriving the
expected interval from the cron expression is correct and deferred: it needs a cron parser and a
timezone decision.

## 010 — `EndpointProbe` is the course capstone, `BackupCheck` ships first

Decided 2026-08-05. Status: accepted.

A read-only verification controller teaches informers, caches, custom resources, status
conditions, RBAC, leader election, metrics and `envtest`. It never teaches the other half —
finalizers, owner references, garbage collection, and conflict handling on objects the controller
owns — because it never creates anything.

`EndpointProbe` creates owned Jobs, one per node when a node selector is set, and therefore
forces exactly those lessons. It is the better *learning* capstone.

It is not the better *first* deliverable: `BackupCheck` is smaller, and its scope is fully
determined by resources that already exist in the cluster. So `BackupCheck` ships first and
`EndpointProbe` closes the course.

## 011 — Milestone 1 is a command, with no custom resource and no controller-runtime

Decided 2026-08-05. Status: accepted.

The custom resource design in [backup-check.md](backup-check.md) only became obvious after
reading real Longhorn objects; the field list would have been guesswork a week earlier.

So the first deliverable is a read-only CLI report doing the join and printing the buckets.
It validates the object graph, produces a usable answer, and defers every architectural decision
that depends on knowing which fields matter. "Custom resource first, informer later" was
considered and rejected on those grounds.

## 012 — Cleanup belongs to the milestone that needs it, never its own phase

Decided 2026-08-05. Status: accepted.

Two successive plans for this project opened with a refactoring-and-documentation phase. Neither
reached the phase after it.

Concretely: the HTTP server gains a context and a `Shutdown` method in the milestone that needs
to run an informer alongside it, not in a preceding cleanup pass. The `pkg/` to `internal/` move
happens per-package as milestones touch them. Known defects with no milestone that needs them
yet stay on a list and stay unfixed, which is the honest state.

## 013 — Supporting tooling is frozen until the controller work is done

Decided 2026-08-05. Status: accepted.

Release automation, changelog tooling, static-analysis configuration, cluster provisioning,
devcontainer customisation, container publishing and micro-benchmarks are all in a working state
and all of them are the least transferable part of the exercise. They are the reason the ratio of
plumbing to Kubernetes learning ended up where it did.

Nothing on that list needs deleting. It needs to stop consuming attention until the controller
milestones are complete. Security updates to dependencies are exempt.
