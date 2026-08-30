# 016 — Three deployment shapes on one core; multi-cluster is a deployment property

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
