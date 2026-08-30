# Decisions

Numbered, dated, and kept in one file until the file gets unwieldy. Each entry records what was
decided, the reasoning, and — where it matters — what was rejected, so a rejected option does not
get proposed again as if it were new.

Status values: **accepted**, **superseded by NNN**, **revisit at milestone N**, and
**deferred to the X plugin** for a decision that still holds but only inside a plugin that is not
written yet.

Entries 001 through 013 were written before the project's purpose was settled; 014 explains what
changed and which of them it overturns. Read 014 first.

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

Decided 2026-08-05. Status: deferred to the Longhorn plugin, see 014.

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

Decided 2026-08-05. Status: accepted, applies to the Longhorn plugin only, see 014.

`BackupVolume.metadata.name` carries a hash suffix and is not the volume name; `spec.volumeName`
is. Joining on the object name silently matches nothing, which reads as "no backups exist" rather
than as a bug.

Full object graph and the other two traps — join direction, and
`status.labels.KubernetesStatus` being history rather than current state — in
[backup-check.md](backup-check.md).

## 008 — Enrolment comes from RecurringJob groups, not PVC labels

Decided 2026-08-05. Status: accepted, applies to the Longhorn plugin only, see 014.

The intuitive design — a label selector over PVCs — is wrong for Longhorn. A `RecurringJob` in
the `default` group covers every volume carrying no recurring-job label of its own, so unlabelled
volumes are backed up and a selector-driven check reports them as unenrolled.

Enrolment is therefore computed: resolve the backup jobs, resolve the groups, and read the labels
Longhorn writes back onto the `Volume` resource rather than the labels a human may or may not
have put on the PVC. Resolution order in [backup-check.md](backup-check.md).

Consequence for the custom resource in 011: its selector chooses **what to check**, which is not
the same field as **what is enrolled**.

## 009 — A freshness policy must account for gaps in the backup schedule

Decided 2026-08-05. Status: accepted, applies to the Longhorn plugin only, see 014.

A flat 24-hour freshness threshold produces a false alarm every time the backup schedule skips a
day. A schedule running Tuesday through Saturday legitimately leaves a two-day-old backup on
Monday.

v1 takes the threshold from the operator and defaults it wide enough to span the schedule's
largest gap, showing the job's cron in the report so the default is explicable. Deriving the
expected interval from the cron expression is correct and deferred: it needs a cron parser and a
timezone decision.

## 010 — `EndpointProbe` is the course capstone, `BackupCheck` ships first

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

## 011 — Milestone 1 is a command, with no custom resource and no controller-runtime

Decided 2026-08-05. Status: superseded by 014.

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

## 014 — kc is a cluster interrogation tool, the course is its spine

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

## 015 — The core is a Go package; the CLI, HTTP and MCP are renderers over it

Decided 2026-08-05. Status: accepted.

Every capability is a named, parameterised query returning structured data. The CLI is the first
renderer over that data. HTTP and MCP are later renderers over the same package.

The reference implementation makes an HTTP API the core, and its MCP handlers call that API. That
is right for a platform API and wrong here: a CLI that has to start a web server before it can
answer a question is complexity with no cause. The lesson of building an API layer is still
learned, it just lands as a package rather than as a network service.

Deciding this now is nearly free. Retrofitting it after the CLI has grown its own formatting and
business logic is not.

## 016 — Three deployment shapes on one core; multi-cluster is a deployment property

Decided 2026-08-05. Status: accepted.

The same core runs in three shapes:

- **local** — the CLI reads a kubeconfig and answers about that cluster.
- **agent** — kc runs inside a cluster and exposes the HTTP renderer.
- **client** — the CLI talks to a remote agent instead of to a cluster.

"Agent" is meant in the sense of a Wazuh agent or a Salt minion, not in the sense of an LLM agent.
A fleet of them, one per cluster, is queried by pointing the CLI at whichever endpoint is wanted.

This is why nothing in the query logic knows about more than one cluster: the fan-out lives in
where the agents are deployed and in which endpoint the caller picks, not in the code. It is also
why 021's decision about authentication is load-bearing rather than theoretical, since an agent is
reachable over the network.

The boundary is drawn now. The server itself is built when the course reaches its API step.

## 017 — Assume only core APIs; detect everything else

Decided 2026-08-05. Status: accepted.

kc may assume `core/v1` and `apps/v1` exist. Nothing else. Optional APIs are detected at startup:
find `longhorn.io` and the Longhorn plugins register, find a GitOps controller and its plugins
register, find neither and both are silently absent rather than broken.

The test of this is not a unit test. It is that kc must be fully usable against the single-node
cluster this repository provisions locally, which has no storage provider, no GitOps controller and
no metrics stack. A tool that only works where the author works is the failure mode 014 describes.

## 018 — Plugins are in-tree Go interfaces behind a registry

Decided 2026-08-05. Status: accepted.

A plugin is a Go type satisfying an interface, registered into a registry, declaring the API groups
it requires so 017's detection can decide whether to enable it.

Out-of-process plugins — exec or gRPC, written by other people — are a separate project with
protocol versioning and compatibility guarantees, and they buy nothing at this stage. The in-tree
interface does not preclude them later; it is the boundary they would be built against.

## 019 — `Check` replaces `FrontendPage` and adds status conditions

Decided 2026-08-05. Status: accepted.

`spec` names the plugin, its parameters and how often to re-evaluate. `status.conditions` carries
the result. Reconcile re-runs the check and updates the status. Leader election is genuinely
required rather than decorative: two replicas evaluating the same check write conflicting status.

Writes are limited to the resource's own status, per 002.

Two facts about the reference implementation shaped this. Its `FrontendPage` type has no `status`
field at all, and no custom resource in the course has one, so status conditions are a gap in the
course rather than something being replaced. And its reconciler creates a ConfigMap and a
Deployment, which is where 020 comes in.

## 020 — One plugin owns Jobs, so garbage collection gets exercised

Decided 2026-08-05. Status: accepted.

A status-only reconciler never touches owner references, garbage collection or the conflict
handling that comes with owning objects. In the reference implementation `SetControllerReference`
and `Owns()` appear in exactly one file, the `FrontendPage` controller, so replacing that resource
with a status-only `Check` would remove the only place those are taught.

The fix is a plugin whose question genuinely cannot be answered from inside the API server:
reachability from a particular node, DNS resolution as a pod sees it, the expiry of a certificate
on an endpoint. Such a plugin creates a Job, owns it, aggregates its result into the `Check`
status, and lets garbage collection clean up.

This is the idea previously scoped as an `EndpointProbe` capstone, demoted to one plugin. The
demotion is what makes it honest: the Job exists because the check needs it, not because the
curriculum needs a Job.

Finalizers stay untaught. The course never covers them either, and a read-only checker has no
legitimate use for one. Adding a finalizer to have used a finalizer is exactly the decorative work
this decision avoids.

## 021 — No LLM inside kc, ever

Decided 2026-08-05. Status: accepted.

kc never calls a model. The MCP renderer is the only point of contact, and it points outward: a
model uses kc as a tool, kc does not use a model.

A tool built well enough for a person to use is already good enough for a model to use, so there is
nothing to gain by building a second, model-shaped interface. Models change quickly; a tool that
does a defined job and returns quickly can go a long time without updates. Chasing what broke in
this month's agent framework is not work.

What the constraint buys, concretely: no API keys, no rate limits, no per-call cost, deterministic
output, and tests that do not need a model to pass. It also rules out the tempting middle ground of
"just one plugin with an LLM in it", which would drag all of the above into the whole project.

## 022 — The overlap with k8sgpt is deliberate

Decided 2026-08-05. Status: accepted.

k8sgpt has a CLI, a `serve` mode, a separate in-cluster operator, in-tree analysers and
out-of-process custom analysers. That is the architecture described in 015 through 018, already
built and actively maintained. Decision 001's rule against writing a worse copy of an existing tool
points squarely at this project.

The overlap is accepted anyway, because 014 settles what kc is for: the author is building it to
learn the mechanisms, and there is no user to lose to a better tool.

One real difference holds, and it is 021: k8sgpt calls a model in order to explain what it found.
kc returns data and lets whatever is asking do the explaining.

The failure mode to guard against is not technical. It is telling yourself in six months that this
is a product. If kc ever needs to win on features, the correct response is to contribute to
k8sgpt, not to catch up to it.
