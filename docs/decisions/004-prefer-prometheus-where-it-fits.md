# 004 — If a check can be a Prometheus rule, it should be

Decided 2026-08-05. Status: accepted.

A controller that grows into a second alerting system is a liability: two places to look, two
notification paths, two sets of thresholds that drift apart.

The rule: the controller exists to produce assertions Prometheus cannot express — ones needing a
multi-resource join or an external read. Once the assertion is a number, alerting on it is
Prometheus' job, and the deliverable is an example `PrometheusRule`, not a notifier.
