# 015 — The core is a Go package; the CLI, HTTP and MCP are renderers over it

Decided 2026-08-05. Status: accepted.

Every capability is a named, parameterised query returning structured data. The CLI is the first
renderer over that data. HTTP and MCP are later renderers over the same package.

The reference implementation makes an HTTP API the core, and its MCP handlers call that API. That
is right for a platform API and wrong here: a CLI that has to start a web server before it can
answer a question is complexity with no cause. The lesson of building an API layer is still
learned, it just lands as a package rather than as a network service.

Deciding this now is nearly free. Retrofitting it after the CLI has grown its own formatting and
business logic is not.
