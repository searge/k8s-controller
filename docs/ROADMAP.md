# Roadmap

**Where the code is:** step 8 of 14. The CLI, structured logging, `list deployments` through
client-go, a cached Deployment informer, and an HTTP server that answers from that cache under a
signal-bound lifecycle. No controller-runtime, no custom resource, no plugins yet.

What kc is and why it is shaped this way: [the decision log](decisions/), starting at
[014](decisions/014-kc-is-a-cluster-interrogation-tool.md). Which
course step maps to what: [COURSE.md](COURSE.md).

Two things drive the order below. The course supplies the mechanisms, in its own sequence. The
plugins supply the reasons, and they are deliberately allowed to arrive before the machinery that
will eventually host them, so the project produces usable answers while the architecture is still
being built.

## Next — the plugin interface, proven by one plugin

Effort: 2–3 evenings.

The plugin boundary (decision 018) and capability detection (017) are the two decisions with
nothing behind them yet. Both are cheap to get wrong and expensive to change later, so they get
built against a real plugin rather than designed in the abstract.

- `internal/plugin`: the interface, a registry, and declared API-group requirements per plugin.
- Capability detection at startup: discover which API groups the cluster serves, register the
  plugins whose requirements are met, and skip the rest without complaint.
- The first plugin: **hand-edit detection over `managedFields`**. It reports objects where a
  `kubectl`-class field manager overlaps a GitOps-class one, with the manager, the timestamp and
  the field paths it owns. Core APIs only, so it runs anywhere.
- A CLI capability, not a custom resource. No controller-runtime is involved.

Feasibility is already established rather than assumed: a probe against a live cluster found six
such objects out of two hundred, including a January probe tweak that never reached git. Findings
and the honest limits — the manager is a tool rather than a person, and surviving traces are not a
history — are in the working note `tmp/roadmap/07-managedfields-plugin.md`.

**Done when:** the plugin runs against a cluster with no GitOps controller and reports nothing
rather than failing, and a second plugin could be added without touching the first.

## Step 9 — controller-runtime and the manager

Effort: 2–4 evenings.

- A manager owning the caches, the health and readiness probes, and the client.
- The existing informer moves under it, or is replaced by the manager's cache. Which one is a real
  decision to make with the code in front of us, not now.
- `envtest` set up, since everything after this step needs it.

Cost to watch: this is where the estimate is least reliable. Manager wiring and `envtest` fixtures
are the parts of the course most likely to consume an evening on setup alone.

## Step 10 — leader election and metrics

Effort: 1–2 evenings.

- Leader election through the manager. Required rather than optional once a reconciler writes
  status: two replicas evaluating the same check produce conflicting writes (decision 019).
- Metrics: the informer's event counters already exist and are the natural first series, plus
  reconcile counts and errors. Watch cardinality before adding anything per-object.
- One example `PrometheusRule`, because alerting belongs to Prometheus (decision 004).

## Step 11 — the `Check` custom resource

Effort: 3–5 evenings.

The course's `FrontendPage` is replaced by `Check` (decision 019): `spec` names a plugin, its
parameters and how often to re-evaluate; `status.conditions` carries the result. Writes are limited
to the resource's own status.

- API types, deepcopy generation, CRD manifests, RBAC generated for what is actually read.
- A reconciler that runs the named plugin and records the outcome as conditions.
- Decision due here and currently open: the API group name. It needs a domain that will not
  collide upstream.
- `envtest` cases for the awkward paths: a plugin whose API group is absent, a check whose plugin
  name does not exist, and a status update that loses a conflict.

Note what this step does *not* teach, and where that is repaid: a status-only reconciler never
touches owner references or garbage collection. Decision 020 puts those behind a later plugin whose
question genuinely needs an owned Job — reachability from a particular node, or the expiry of a
certificate on an endpoint.

## Step 12 — the core package boundary

Effort: 1–2 evenings.

Decision 015 already commits to it and the current code half-implements it: `pkg/server` reads
through a `DeploymentSource` interface and `k8s.NewDeploymentInfo` is a shared pure projection.
This step makes it explicit — every capability is a named, parameterised query returning structured
data, and the CLI is one renderer over that.

The course builds an HTTP API here. kc builds a Go package; HTTP is a renderer over it, not the
core.

## Step 13 — the MCP renderer

Effort: 2–3 evenings.

- An MCP server over the core package, built on `modelcontextprotocol/go-sdk` (decision 023).
- One tool per capability, with typed input and output structs.
- Transport decision waiting here: go-sdk's middleware is `net/http` while `pkg/server` is
  fasthttp, and adapting between them buffers responses in a way that breaks streaming. Serving MCP
  on `net/http` directly is the likely answer.
- Re-verify the library choice first. Both projects release often and the note lists what to check.

## Step 14 — authentication on the agent

Effort: 2–3 evenings.

- Audience-scoped ServiceAccount tokens verified by TokenReview (decision 024).
- One verifier, two thin adapters: go-sdk's `auth.RequireBearerToken` for MCP, a fasthttp wrap for
  the HTTP renderer.
- `docs/api.md` states plainly what is not covered: no transport security, workloads rather than
  people in audit trails, no individual revocation, and one credential per cluster.
- Deferred deliberately: per-namespace authorization via SubjectAccessReview, designed in the
  working note and not built.

## After the course

Not scheduled, recorded so they are not mistaken for oversights.

- **More plugins.** The rule from the grilling session stands: the next plugin is whichever
  question costs more than a minute of `kubectl` in real work.
- **A plugin with owned Jobs**, which is where owner references and garbage collection finally get
  exercised (decision 020).
- **The agent deployment shape.** A manifest, TLS, and the fleet story from decision 016.
- **A Longhorn plugin**, if it is still wanted by then. The object graph, the join key trap and the
  RecurringJob group semantics are all worked out in [backup-check.md](backup-check.md).

## Rules that hold throughout

1. **Read-only in the cluster.** The only writes are the controller's own status and its metrics.
2. **Never mutate what a GitOps controller owns.**
3. **If Prometheus can express it, use Prometheus.**
4. **Every step ends with a test that would fail if the behaviour regressed** — not with a coverage
   number. In practice this has meant checking new tests by mutation: break the behaviour on
   purpose and confirm the suite notices.
5. **Cleanup belongs to the step that needs it**, never to a phase of its own.
6. **Assume only core APIs.** Everything else is detected, and its absence is not an error.

## Known defects

Recorded rather than fixed, per decision 012, and re-checked against the code on 2026-08-30. Each
will be picked up by whichever step first needs the code it touches.

| Where | Problem |
| --- | --- |
| `cmd/flags.go`, `cmd/list.go` | Mutable package-level globals hold flag values; commands cannot be constructed independently or tested in isolation |
| `pkg/k8s/client.go:77` | `rest.InClusterConfig()` is tried before explicit flags, so `--kubeconfig` and `--context` are ignored when running inside a cluster |
| `pkg/k8s/client.go` | Client is built with no `UserAgent`, `QPS` or `Burst` |
| `cmd/list.go:163` | API errors classified by substring matching on the error text instead of `apierrors.IsForbidden` and friends |
| `cmd/list.go:145`, `cmd/connection.go:35` | `context.WithTimeout` built from an unvalidated timeout; zero means immediate cancellation |
| `Dockerfile:40` | Image label still claims a production-grade controller |
| `internal/informer` | `Config.Namespace` exists but no command exposes it, so the informer always watches every namespace |

Resolved since the last revision: the server tests that leaked a goroutine and a port per run, by
the lifecycle work in step 7.
