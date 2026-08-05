# Documentation

## Reading order

New to the project, in this order:

1. **[COURSE.md](COURSE.md)** — what this repository is an implementation of, and which steps are
   actually done.
2. **[ROADMAP.md](ROADMAP.md)** — the milestones, the effort estimate, and the defects that are
   known and deliberately unfixed.
3. **[ARCHITECTURE.md](ARCHITECTURE.md)** — current layout, target layout, package boundaries,
   testing strategy per layer.
4. **[DECISIONS.md](DECISIONS.md)** — numbered decisions with reasoning, including rejected
   options.
5. **[backup-check.md](backup-check.md)** — the first domain: Longhorn's object graph, the join,
   and what the assertion does and does not claim.

Working on the code:

- **[CODING_GUIDELINES.md](CODING_GUIDELINES.md)** — the rules, as a quick reference.
- **[BEST_PRACTICES.md](BEST_PRACTICES.md)** — the same rules with worked examples.
- **[api.md](api.md)** — HTTP endpoints and CLI commands as they exist now. Milestone 3 changes
  the endpoint set; this file is updated in that milestone, not before.

## What is deliberately not here

This repository is public. Some of what motivated the project is not.

**Not in this repository, ever:**

- Course material. The lecture notes, glossary and reference cards are derived from a paid course
  and stay in a private vault. Documentation here links to the course page instead.
- Anything identifying a specific environment: cluster names, hostnames, storage endpoints, IP
  ranges, namespaces belonging to a tenant, ticket or issue identifiers from a private tracker.
- Measurements taken from a specific environment. A count of volumes in a particular cluster is
  operational state, not a technical fact.
- Incident narrative. No "this failed for us and here is what it looked like".

**In this repository, and worth being precise about the distinction:**

- Upstream behaviour the code must handle. That a `BackupVolume` object name carries a hash suffix,
  so the join uses `spec.volumeName`, is a specification — without it the code looks like a bug.
  Documenting it is not narrative.
- Failure *modes* stated as requirements. "A freshness threshold must account for gaps in the
  backup schedule" is a design constraint. It becomes narrative only when it acquires a date, a
  cluster and a consequence.

The test when adding to these files: could a reader work out **which** environment this came from,
or **when**? If yes, it does not belong here. The reasoning usually survives having those two
things removed.

Traceability is not lost by this — decisions reference the technical constraint that drove them,
and the private record keeps the rest.

## Conventions

- Prose style: no emoji in technical documentation, no numbered headings, no decorative emphasis.
  `CODING_GUIDELINES.md` holds rules and `BEST_PRACTICES.md` holds the examples; keep that split.
- New decisions append to `DECISIONS.md` with the next number and a date. Do not renumber or
  delete an entry — mark it `superseded by NNN`, since the point of the file is that a rejected
  option is not proposed again as if it were new.
- A new domain gets its own file next to `backup-check.md`. Split into a subdirectory once there
  are three of them.
