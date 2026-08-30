# 017 — Assume only core APIs; detect everything else

Decided 2026-08-05. Status: accepted.

kc may assume `core/v1` and `apps/v1` exist. Nothing else. Optional APIs are detected at startup:
find `longhorn.io` and the Longhorn plugins register, find a GitOps controller and its plugins
register, find neither and both are silently absent rather than broken.

The test of this is not a unit test. It is that kc must be fully usable against the single-node
cluster this repository provisions locally, which has no storage provider, no GitOps controller and
no metrics stack. A tool that only works where the author works is the failure mode 014 describes.
