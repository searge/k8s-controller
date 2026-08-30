# 009 — A freshness policy must account for gaps in the backup schedule

Decided 2026-08-05. Status: accepted, applies to the Longhorn plugin only, see 014.

A flat 24-hour freshness threshold produces a false alarm every time the backup schedule skips a
day. A schedule running Tuesday through Saturday legitimately leaves a two-day-old backup on
Monday.

v1 takes the threshold from the operator and defaults it wide enough to span the schedule's
largest gap, showing the job's cron in the report so the default is explicable. Deriving the
expected interval from the cron expression is correct and deferred: it needs a cron parser and a
timezone decision.
