# 006 — Unstructured dynamic client for Longhorn resources, not typed imports

Decided 2026-08-05. Status: deferred to the Longhorn plugin, see 014.

Two options for reading `volumes.longhorn.io` and friends:

- **Typed:** import Longhorn's Go API package. Cleaner code, compile-time field checking, and it
  pins the project to one Longhorn minor version — an upgrade becomes a dependency bump plus a
  build break.
- **Unstructured:** the dynamic client with `unstructured.Unstructured`. More verbose, no
  compile-time checking of field paths, and it survives Longhorn upgrades and works against a
  cluster where the CRDs are a slightly different version.

Chose unstructured, because the alternative couples a learning project's build to a third-party
operator's release cadence, and because the field set actually read here is small enough that the
verbosity is bounded. The accessor calls are wrapped in one place so the field paths appear once,
not scattered.

Revisit if the field count grows past what a handful of accessors can carry.
