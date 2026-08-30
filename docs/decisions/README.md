# Decisions

One file per decision, numbered in the order they were made. Each records what was decided, the
reasoning, and where it matters what was rejected, so a rejected option does not get proposed
again as if it were new.

Adding one: take the next number, write `NNN-short-slug.md`, and add a row below. Never renumber
and never delete — a decision that stops holding is marked superseded, because the point of the log
is that the old reasoning stays readable.

Status values: **accepted**; **superseded by NNN**; **revisit at course step N** for a decision
that should be re-examined when the work it governs actually starts; and **deferred to the X
plugin** for one that still holds but only inside something not written yet.

Entries 001 to 013 predate the project's purpose being settled. [014](./014-kc-is-a-cluster-interrogation-tool.md)
explains what changed and which of them it overturns; read it before the ones below it.

---

| # | Decision | Status |
| --- | --- | --- |
| [001](./001-assertions-outside-the-api-server.md) | The project targets assertions about state outside the API server | accepted |
| [002](./002-read-only-in-the-cluster.md) | Read-only in the cluster for v1 | accepted |
| [003](./003-never-mutate-gitops-owned-resources.md) | Never mutate resources owned by a GitOps controller | accepted |
| [004](./004-prefer-prometheus-where-it-fits.md) | If a check can be a Prometheus rule, it should be | accepted |
| [005](./005-internal-for-app-code-pkg-for-reusable.md) | Application code lives in `internal/`, `pkg/` is only for reusable code | accepted, revisit at course step 11 |
| [006](./006-unstructured-client-for-longhorn.md) | Unstructured dynamic client for Longhorn resources, not typed imports | deferred to the Longhorn plugin, see 014 |
| [007](./007-join-backups-on-spec-volumename.md) | Join Longhorn backup records on `spec.volumeName` | accepted, applies to the Longhorn plugin only, see 014 |
| [008](./008-enrolment-from-recurringjob-groups.md) | Enrolment comes from RecurringJob groups, not PVC labels | accepted, applies to the Longhorn plugin only, see 014 |
| [009](./009-freshness-policy-respects-schedule-gaps.md) | A freshness policy must account for gaps in the backup schedule | accepted, applies to the Longhorn plugin only, see 014 |
| [010](./010-endpointprobe-as-capstone.md) | `EndpointProbe` is the course capstone, `BackupCheck` ships first | superseded by 019 and 020 |
| [011](./011-first-deliverable-is-a-command.md) | The first deliverable is a command, with no custom resource | superseded by 014 |
| [012](./012-cleanup-belongs-to-its-step.md) | Cleanup belongs to the step that needs it, never its own phase | accepted |
| [013](./013-supporting-tooling-is-frozen.md) | Supporting tooling is frozen until the controller work is done | accepted |
| [014](./014-kc-is-a-cluster-interrogation-tool.md) | kc is a cluster interrogation tool, the course is its spine | accepted |
| [015](./015-core-is-a-go-package-renderers-on-top.md) | The core is a Go package; the CLI, HTTP and MCP are renderers over it | accepted |
| [016](./016-three-deployment-shapes.md) | Three deployment shapes on one core; multi-cluster is a deployment property | accepted |
| [017](./017-assume-only-core-apis.md) | Assume only core APIs; detect everything else | accepted |
| [018](./018-plugins-are-in-tree-interfaces.md) | Plugins are in-tree Go interfaces behind a registry | accepted |
| [019](./019-check-resource-with-status-conditions.md) | `Check` replaces `FrontendPage` and adds status conditions | accepted |
| [020](./020-one-plugin-owns-jobs.md) | One plugin owns Jobs, so garbage collection gets exercised | accepted |
| [021](./021-no-llm-inside-kc.md) | No LLM inside kc, ever | accepted |
| [022](./022-k8sgpt-overlap-is-deliberate.md) | The overlap with k8sgpt is deliberate | accepted |
| [023](./023-mcp-renderer-uses-the-official-go-sdk.md) | The MCP renderer is built on modelcontextprotocol/go-sdk | accepted, revisit at course step 13 |
| [024](./024-agents-verify-tokens-with-tokenreview.md) | Agents authenticate callers with audience-scoped ServiceAccount tokens | accepted, revisit at course step 14 |
