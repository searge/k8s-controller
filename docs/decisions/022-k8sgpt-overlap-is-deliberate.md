# 022 — The overlap with k8sgpt is deliberate

Decided 2026-08-05. Status: accepted.

k8sgpt has a CLI, a `serve` mode, a separate in-cluster operator, in-tree analysers and
out-of-process custom analysers. That is the architecture described in 015 through 018, already
built and actively maintained. Decision 001's rule against writing a worse copy of an existing tool
points squarely at this project.

The overlap is accepted anyway, because 014 settles what kc is for: the author is building it to
learn the mechanisms, and there is no user to lose to a better tool.

One real difference holds, and it is 021: k8sgpt calls a model in order to explain what it found.
kc returns data and lets whatever is asking do the explaining.

The failure mode to guard against is not technical. It is telling yourself in six months that this
is a product. If kc ever needs to win on features, the correct response is to contribute to
k8sgpt, not to catch up to it.
