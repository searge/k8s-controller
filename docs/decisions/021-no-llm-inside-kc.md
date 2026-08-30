# 021 — No LLM inside kc, ever

Decided 2026-08-05. Status: accepted.

kc never calls a model. The MCP renderer is the only point of contact, and it points outward: a
model uses kc as a tool, kc does not use a model.

A tool built well enough for a person to use is already good enough for a model to use, so there is
nothing to gain by building a second, model-shaped interface. Models change quickly; a tool that
does a defined job and returns quickly can go a long time without updates. Chasing what broke in
this month's agent framework is not work.

What the constraint buys, concretely: no API keys, no rate limits, no per-call cost, deterministic
output, and tests that do not need a model to pass. It also rules out the tempting middle ground of
"just one plugin with an LLM in it", which would drag all of the above into the whole project.
