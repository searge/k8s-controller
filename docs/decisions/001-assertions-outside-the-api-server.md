# 001 — The project targets assertions about state outside the API server

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
