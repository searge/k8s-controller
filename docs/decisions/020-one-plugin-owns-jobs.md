# 020 — One plugin owns Jobs, so garbage collection gets exercised

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
