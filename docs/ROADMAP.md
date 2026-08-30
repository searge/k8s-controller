# Roadmap

> [!WARNING]
> **Being rewritten.** The milestones below are built around a Longhorn backup checker as the
> project's purpose. That premise was overturned on 2026-08-05: see
> [DECISIONS.md](DECISIONS.md) 014 through 022 for what kc is actually for. Longhorn is now one
> optional plugin, not the foundation, and there is no `backup-report` command.
>
> What survives from below: the informer, cache-sync gating and lifecycle work in milestone 2, and
> the rule that cleanup belongs to the milestone needing it. What does not: the domain, the
> ordering, the effort estimates, and the definition of done.
>
> Rewritten once the informer work lands, so the new plan is written against code that exists.

Supersedes every earlier plan and progress file for this project. Those described a schedule from
mid-2025 and a completion percentage that stopped meaning anything.

**Where the code is:** step 6 of 14 — a Cobra CLI, structured logging, an HTTP server, and
`list deployments` through client-go. No informers, no reconciliation, no controller-runtime. See
[COURSE.md](COURSE.md) for the step-by-step mapping.

**Two goals, deliberately separated:**

- *Learning* — finish steps 7 through 10 properly, including the lessons a read-only controller
  cannot teach: owned resources, owner references, garbage collection. That makes
  `EndpointProbe` the course capstone.
- *Useful output* — ship a check for a class of failure that neither a policy engine nor a metrics
  rule can express. That is `BackupCheck`, and it comes first because it is smaller and its scope
  is bounded by resources that already exist.

Reasoning behind both, and the options rejected along the way, in [DECISIONS.md](DECISIONS.md).

## Milestone 1 — a report, no custom resource

Effort: 1–2 evenings.

The smallest thing that produces a usable answer. Exists so the project ships something before it
ships architecture.

- `kc backup-report [--max-age 72h]`.
- Joins live PVCs through PV to the Longhorn `Volume` and its `BackupVolume`, on
  `spec.volumeName`.
- Derives enrolment from `RecurringJob` group semantics rather than PVC labels.
- Prints the buckets: `fresh`, `stale`, `uncovered`, `unenrolled`, `orphan`.
- Read-only. Unstructured dynamic client, no Longhorn type dependency.

Design in [backup-check.md](backup-check.md).

**Done when:** the join logic is covered by table-driven tests with no cluster, and a run against
a real cluster puts every volume in a bucket a human agrees with — in particular, an `orphan` is
never reported as `stale`.

## Milestone 2 — step 7, informer

Effort: 1–2 evenings.

- `internal/informer`: shared informer factory, PVC informer, resync interval, structured event
  logging.
- Cache-sync gating: nothing serves before `WaitForCacheSync` returns.
- Alongside, because this milestone is the first that needs it: a root signal context in `serve`,
  and `Server.Start(ctx)` / `Server.Shutdown()` replacing the current
  `ListenAndServe`-then-`os.Exit` path.

**Done when:** a fake watch drives add, update and delete through the cache and the assertions
read cache contents; cancelling the root context shuts down both the informer and the HTTP
server, and a test fails if either leaks.

## Milestone 3 — step 8, API from cache

Effort: 1 evening.

- `GET /pvcs` and `GET /backup-report` served from the informer cache, with a test that fails if
  a handler reaches for the clientset.
- `/healthz` for process liveness, `/readyz` gated on cache sync. Replaces the current `/health`,
  which returns a constant and will be meaningless once a cache exists.
- Structured request logging with a request id, replacing the current formatted-string log line.

[api.md](api.md) is updated in this milestone, not before.

## Milestone 4 — step 9, controller-runtime and the `BackupCheck` resource

Effort: 3–6 evenings.

The largest step, and where an overrun would come from: manager wiring, API types, CRD manifests,
`envtest` fixtures.

- Manager and scheme; health and readiness probes wired through the manager.
- The `BackupCheck` custom resource — field sketch in [backup-check.md](backup-check.md).
- Three status conditions rather than one: an unavailable target, a stale volume and an uncovered
  volume are three different problems needing three different responses.
- RBAC generated for exactly the resources it reads.
- `envtest` cases covering the awkward paths: a volume restored under a new name, an orphan
  record, an enrolled PVC with no `BackupVolume` at all, and an unavailable backup target.

Decisions due here, currently leaning as recorded in [DECISIONS.md](DECISIONS.md): the API group
name, whether the unstructured client still pays for itself (006), and whether the freshness
threshold becomes cron-derived (009).

## Milestone 5 — step 10, leader election and metrics

Effort: 1–2 evenings.

- Leader election through the manager. Required rather than optional: two replicas asserting the
  same external state duplicate the work and can publish contradictory status.
- Metrics: backup age per volume, target availability, coverage counts, reconcile errors. Watch
  series cardinality — per-volume series are fine at this scale and need review before wide
  deployment.
- One example `PrometheusRule`, because alerting belongs to Prometheus (004).

**Done when:** two replicas run and only the leader reconciles; the metrics answer "which volume
has no recent backup" without anyone reading a log.

## Milestone 6 — capstone, `EndpointProbe`

Effort: 4–8 evenings.

The part that makes this a controller rather than a reporter: it creates and owns objects.

- `EndpointProbe` custom resource: endpoint, protocol, node selector, schedule, timeout.
- The controller creates **owned Jobs** — one per matching node when a node selector is set,
  because reachability can differ per node and a single central probe hides that entirely.
- Owner references and garbage collection, a finalizer for probe cleanup, status aggregated
  across child Jobs, and conflict handling when a Job changes under the reconciler.
- Conditions for reachability, certificate validity and per-node failures. Metrics for probe
  success, latency and certificate expiry.

## Not in v1

Named so they are not silently assumed to be included.

- **Restorability.** Proving a backup restores means restoring it somewhere disposable and
  checking the result boots. Real value, much larger project.
- **Application-level dump verification.** A different artefact in a different location, needing
  filesystem access. Its own increment if it ever happens.
- **Any mutation of cluster state**, including remediating what the detectors find (002).

## Stretch — stale `VolumeAttachment` detector

A CSI node plugin that dies under node memory pressure leaves a `VolumeAttachment` marked
attached with nothing alive to detach it, which blocks any other node from attaching that volume.
The pod sits in `ContainerCreating` with very little event signal, and diagnosing it means
correlating `VolumeAttachment`, Pod, PVC and Node — a multi-resource join, which is the shape this
project is already built for.

Detector only: expose a count and a condition, delete nothing. Deletion here is destructive and
the failure is rare enough that a human should stay in the loop.

## Rules that hold for every milestone

1. **Read-only in the cluster.** The only writes are the controller's own status and metrics.
2. **Never mutate what a GitOps controller owns.**
3. **Non-production clusters only** until the detector has been correct for an extended period.
4. **If Prometheus can express it, use Prometheus.**
5. **Every milestone ends with a test that would fail if the behaviour regressed** — not with a
   coverage number.
6. **Cleanup belongs to the milestone that needs it**, never to a phase of its own.

## Effort

Evenings, not working days. The distinction matters: an earlier version of this plan quoted
"four focused days", which was a working-day estimate presented as a schedule.

| Milestone | Evenings | Cumulative |
| --- | --- | --- |
| 1 — backup report | 1–2 | 2 |
| 2 — informer (step 7) | 1–2 | 4 |
| 3 — API from cache (step 8) | 1 | 5 |
| 4 — controller-runtime and CRD (step 9) | 3–6 | 11 |
| 5 — leader election and metrics (step 10) | 1–2 | 13 |
| 6 — `EndpointProbe` capstone | 4–8 | 21 |

Roughly two to three weeks of evenings to finish steps 7 through 10, and about a month to the
capstone.

## Definition of done

> Running on a non-production cluster: a `BackupCheck` reports honest conditions across the
> Longhorn volumes in scope, separating volumes whose backups stopped from records left behind by
> deleted volumes, and a Prometheus alert fires when a covered volume's newest backup ages past
> policy. Then `EndpointProbe` runs per-node probes through owned Jobs, with owner references and
> garbage collection actually exercised.

## Known defects not yet assigned to a milestone

Recorded rather than fixed, per decision 012. Each will be picked up by whichever milestone first
needs the code it touches.

| Where | Problem |
| --- | --- |
| `cmd/flags.go`, `cmd/list.go` | Mutable package-level globals hold flag values; commands cannot be constructed independently or tested in isolation |
| `pkg/k8s/client.go` | `rest.InClusterConfig()` is tried before explicit flags, so `--kubeconfig` and `--context` are ignored when running inside a cluster |
| `pkg/k8s/client.go` | Client is built with no `UserAgent`, `QPS` or `Burst` |
| `cmd/list.go` | API errors classified by substring matching on the error text instead of `apierrors.IsForbidden` and friends |
| `cmd/list.go`, `cmd/connection.go` | `context.WithTimeout` built from an unvalidated timeout; zero means immediate cancellation |
| `Dockerfile` | Image label claims a production-grade controller; there is no controller yet |
| `pkg/server/server_test.go` | Tests start the server in a goroutine with no shutdown path, leaking goroutines and ports — resolved by milestone 2 |
