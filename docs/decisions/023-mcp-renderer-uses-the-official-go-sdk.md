# 023 — The MCP renderer is built on modelcontextprotocol/go-sdk

Decided 2026-08-30. Status: accepted, revisit at course step 13.

The course reference pins `mark3labs/mcp-go v0.31.0`. kc uses the official
`modelcontextprotocol/go-sdk` instead. Four reasons, in order of weight:

- **A compatibility promise that exists and has held.** go-sdk's v1.0.0 release
  states plainly that "going forward we won't make breaking API changes", with a
  deprecation policy in CONTRIBUTING.md. Eleven months and seven minor releases
  later, including a disruptive spec revision, it has held. mcp-go spent its
  whole life on 0.x and tagged its first v1 beta two weeks before this decision;
  no written policy accompanies it.
- **Spec currency in the stable line.** go-sdk v1.7.0 supports the current
  revision down to the oldest. mcp-go's last stable tag tops out three revisions
  behind; reaching the current one means running a beta.
- **Authentication is already the right shape.** `auth.RequireBearerToken`
  returns ordinary `func(http.Handler) http.Handler` middleware driven by a
  `TokenVerifier`. The JWT work in step 14 becomes writing a verifier and
  wrapping the transport, rather than building the verification layer first.
- **Maintenance base.** go-sdk is maintained by the MCP organisation together
  with Google's Go team, with a governance process. mcp-go is effectively one
  person: its top contributor has an order of magnitude more commits than the
  next.

Cost of diverging from the reference: none worth counting. Decision 015 already
commits kc to different wiring than the course -- a renderer over the core Go
package rather than over an HTTP API -- so line-by-line fidelity was never
available, and the reference's step 13 handlers are stubs in any case.

Rejected: following the course's pin for fidelity. It is a May-2025 snapshot
with no working handlers to inherit.

Findings, with sources and the facts to re-check, in the working note
`tmp/roadmap/06-mcp-library.md` (not tracked). Releases move; the recommendation
should be re-verified when step 13 actually starts.
