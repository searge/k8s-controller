# 024 — Agents authenticate callers with audience-scoped ServiceAccount tokens

Decided 2026-08-30. Status: accepted, revisit at course step 14.

A caller presents a Kubernetes ServiceAccount token minted for the agent's own
audience; the agent verifies it by posting a TokenReview to
`authentication.k8s.io/v1` and requires both `status.authenticated` and its
audience appearing in `status.audiences`. The agent's entire RBAC for this is
`create` on `tokenreviews`, half of the built-in `system:auth-delegator`.

The decisive property is that audience binding is enforced by the API server
rather than by convention, and it was observed working rather than read about: a
token minted for one audience reviews as authenticated for that audience and is
rejected for any other, with the API server reporting the mismatch. So a token
issued for an agent cannot be replayed against the API server or against a
different agent, and a ServiceAccount token stolen from elsewhere in the cluster
does not authenticate to kc. Nothing else on the shortlist provides that without
new infrastructure.

The other candidates fail on the constraints already recorded:

- A shared-secret JWT, which is what the course does, makes one leaked key a
  compromise of every agent in the fleet, and puts key distribution and rotation
  on a single part-time maintainer. The reference implementation is worth reading
  as a warning: its token endpoint sits outside its own middleware and issues a
  valid one-hour token for a fixed subject to anyone who can reach the port, its
  signing secret defaults to empty while the help text calls it required, and its
  data endpoints have no middleware at all. The transferable lesson there is the
  shape of a bearer-token middleware, not the token model.
- OIDC against the caller's identity provider is an external service dependency,
  which 021 excludes, and an assumption about infrastructure, which 017 excludes.
- mTLS cannot be built from what a cluster hands out: the ServiceAccount CA
  bundle is only guaranteed to verify connections to the API server,
  `kube-apiserver-client` certificate requests are never auto-approved, and no
  built-in signer issues serving certificates for an arbitrary service.

Identity arrives with the verification: TokenReview returns the username and
groups, which is enough for an optional subject allowlist without kc maintaining
a user list of its own. Per-namespace authorization would need SubjectAccessReview
on top, at the cost of an extra API call per distinct namespace and resource, and
is deferred rather than designed away -- its appeal is that kc's authorization
would then be the cluster's authorization.

One verifier serves both renderers. go-sdk's `auth.RequireBearerToken` is
ordinary `net/http` middleware; the HTTP renderer is fasthttp and needs a thin
wrap of the same verifier. Two adapters, one auth model.

Deliberately not covered, and to be stated in `docs/api.md` rather than implied:
there is no transport security yet, so a bearer token on plain HTTP is exposed on
the wire; callers authenticate as workloads, not people, so audit trails name a
ServiceAccount; individual tokens cannot be revoked, only outlived; and a single
credential does not work across the fleet, because TokenReview validates against
one cluster by construction. Under the MCP specification authorization is
optional for HTTP transports, so accepting bearer tokens without being an
OAuth 2.1 resource server is legal -- but it is a deviation worth naming.

What would change this: kc needing to identify a person rather than a workload,
or a single credential having to work across clusters. Both point at OIDC.

Findings, sources and the live round trip in the working note
`tmp/roadmap/08-agent-auth.md` (not tracked).
