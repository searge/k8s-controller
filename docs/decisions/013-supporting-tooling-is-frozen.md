# 013 — Supporting tooling is frozen until the controller work is done

Decided 2026-08-05. Status: accepted.

Release automation, changelog tooling, static-analysis configuration, cluster provisioning,
devcontainer customisation, container publishing and micro-benchmarks are all in a working state
and all of them are the least transferable part of the exercise. They are the reason the ratio of
plumbing to Kubernetes learning ended up where it did.

Nothing on that list needs deleting. It needs to stop consuming attention until the controller
course steps are complete. Security updates to dependencies are exempt.
