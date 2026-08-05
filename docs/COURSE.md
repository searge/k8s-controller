# Course mapping

This repository is an implementation of the FWDays crash course on Kubernetes controllers. This
file maps the course's implementation steps onto the state of the code, and onto the milestones in
[ROADMAP.md](ROADMAP.md).

No course material is reproduced here. Links point at the course page and the instructors'
reference implementation.

- Course: <https://fwdays.com/event/kubernetes-controllers-course>
- Reference implementation: <https://github.com/den-vasyliev/k8s-controller-tutorial-ref>
- Control plane setup used by the course: <https://github.com/den-vasyliev/k8sdiy-kubernetes-control-plane>

## Two numbering schemes

They are easy to confuse, and earlier notes for this project mixed them.

- **6 lectures** — how the course is delivered.
- **10 implementation steps** — the `feature/stepN-*` branches in the reference repository. This
  is the numbering used throughout this repository's documentation.

## Steps against the code

| Step | Reference branch | Topic | State here |
| --- | --- | --- | --- |
| 1 | `feature/step1-cobra-cli` | Cobra CLI skeleton | Done — `cmd/root.go` |
| 2 | `feature/step2-zerolog-logging` | Structured logging | Done — `pkg/logger` |
| 3 | `feature/step3-pflag-loglevel` | Log level as a flag | Done — `--log-level` on the root command |
| 4 | `feature/step4-fasthttp-server` | HTTP server | Done — `pkg/server`, `cmd/serve.go` |
| 5 | `feature/step5-makefile-docker-ci` | Build, container, CI | Done, and then some — Taskfile instead of Make, plus release automation and provisioning that the course does not ask for |
| 6 | `feature/step6-list-deployments` | client-go, list resources | Done — `pkg/k8s/client.go` |
| 7 | `feature/step7-informer` | Informer and cache | Not started — [milestone 2](ROADMAP.md#milestone-2--step-7-informer) |
| 8 | `feature/step8-api-handler` | Serve from cache | Not started — [milestone 3](ROADMAP.md#milestone-3--step-8-api-from-cache) |
| 9 | `feature/step9-controller-runtime` | Manager, reconciler, CRD | Not started — [milestone 4](ROADMAP.md#milestone-4--step-9-controller-runtime-and-the-backupcheck-resource) |
| 10 | `feature/step10-leader-election` | Leader election, metrics | Not started — [milestone 5](ROADMAP.md#milestone-5--step-10-leader-election-and-metrics) |

Step 5 is where this repository diverged furthest from the course, and decision 013 in
[DECISIONS.md](DECISIONS.md) is a direct response to that.

## Where the milestones diverge from the reference

The reference implementation demonstrates each step against `Deployment`, which is the right
choice for a tutorial: the resource is familiar and the exercise stays focused on the mechanism.

This repository lands the same steps on a different resource set — PVCs and Longhorn's backup
resources — so that finishing step 10 produces something with a reason to keep running. The
mechanisms exercised are the same ones: shared informers, cache-sync gating, work queues, a
custom resource with status conditions, RBAC, leader election, metrics, `envtest`.

One thing that substitution costs, and how it is paid for: a read-only checker never exercises
finalizers, owner references or garbage collection. Milestone 6 exists specifically to cover
them, by creating owned Jobs. See decision 010.

## Lectures against milestones

Approximate, since lecture topics and implementation steps do not line up one-to-one.

| Lecture | Topic | Lands in |
| --- | --- | --- |
| 1 | Kubernetes API, control plane, reconciliation | Background for everything; the provisioning in `ansible/` came from here |
| 2 | Go fundamentals, CLI, HTTP server | Steps 1–4, all done |
| 3 | client-go, informers, serialisation | Step 6 done, steps 7–8 are milestones 2–3 |
| 4 | controller-runtime, manager, reconciler, leader election | Steps 9–10, milestones 4–5 |
| 5 | Custom resources, admission control, testing, observability | Milestones 4–6 |
| 6 | Platform integration, agent interfaces | Out of scope for this repository |

## Personal course notes

Lecture notes, glossary and reference cards live outside this repository and stay there — they are
derived from paid course material. Nothing in this documentation depends on them; the technical
content needed to work on the code is either here or in the upstream documentation linked from
[backup-check.md](backup-check.md) and [ARCHITECTURE.md](ARCHITECTURE.md).
