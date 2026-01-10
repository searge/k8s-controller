# Coding Guidelines

Quick reference for code standards in the k8s-controller project. Based on
[Google Go Style Guide](https://google.github.io/styleguide/go/) and
[Effective Go](https://go.dev/doc/effective_go).

We follow functional programming principles: pure functions, immutability,
and composition over inheritance.

For detailed examples, see [BEST_PRACTICES.md](BEST_PRACTICES.md).

## Code Organization

### File Size Limits

Keep files and functions focused and maintainable.

| Element  | Target    | Maximum | Action if exceeded                 |
| -------- | --------- | ------- | ---------------------------------- |
| File     | 150-250   | 500     | Split into multiple files/packages |
| Function | 20-30     | 50      | Extract helper functions           |
| Test     | 20-30     | 50      | Use table-driven test structure    |

Maximum limits enforced by Codacy. Target ranges represent optimal readability.

### Package Structure

- One clear purpose per package
- Short, lowercase, single-word names
- Avoid generic names: `util`, `common`, `helpers`
- All packages require documentation

### Import Order

Groups separated by blank lines:

1. Standard library
2. Third-party packages
3. Project packages

## Naming Conventions

| Element              | Convention   | Example              |
| -------------------- | ------------ | -------------------- |
| Exported functions   | PascalCase   | `CreateClient`       |
| Unexported functions | camelCase    | `createTableWriter`  |
| Exported constants   | PascalCase   | `DefaultTimeout`     |
| Unexported constants | camelCase    | `maxRetries`         |
| Local variables      | camelCase    | `namespace`          |
| Short scope          | 1-2 chars    | `i`, `ctx`, `err`    |
| Interfaces           | `-er` suffix | `Reader`, `Writer`   |

## Documentation

All exported elements require documentation comments.

- Package: Required for all packages
- Functions: Start with function name, explain purpose and context
- Complex logic: Explain why, not what
- Comments: Complete sentences

## Error Handling

- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Check errors immediately after occurrence
- Use early returns to reduce nesting
- Error messages: lowercase, specific, actionable

## Logging

### Structured Logging with Zerolog

All logging uses structured format with relevant context fields.

### Log Levels

- Debug: Diagnostic details for development
- Info: Normal operations, significant events
- Warn: Potentially harmful situations, degraded state
- Error: Failures requiring attention, execution continues
- Fatal: Critical errors, application terminates

## Testing

- Table-driven tests for comprehensive coverage
- Constants for reusable test data
- Target coverage: >80%
- Tests must be: fast, independent, deterministic
- Use interfaces for mocking external dependencies

## Function Design

| Guideline             | Target  | Maximum | Notes                           |
| --------------------- | ------- | ------- | ------------------------------- |
| Lines per function    | 20-30   | 50      | Enforced by Codacy              |
| Cyclomatic complexity | 5-10    | 15      | Break complex logic into steps  |
| Parameters            | 1-3     | 4       | Use config structs beyond limit |
| Purpose               | Single  | -       | One clear responsibility        |
| Return values         | -       | -       | Error always last               |

### Single Responsibility Principle

Each function performs one well-defined task. Complex operations compose smaller functions rather than implementing everything inline.

## Functional Programming Principles

### Pure Functions

Functions that are deterministic and free of side effects:

- Same input always produces same output
- No I/O operations (file, network, console)
- No modification of global state
- No mutation of parameters

Benefits: easier testing, reasoning, caching, and concurrent execution.

### Immutability

Create new data structures instead of modifying existing ones. Particularly important for slices, maps, and structs passed as parameters.

### Composition Over Complex Logic

Build complex behavior from small, focused, reusable functions. Main functions orchestrate helper functions rather than implementing all logic inline.

### Isolating Side Effects

Separate pure business logic from impure I/O operations:

- Pure core: calculations, transformations, validations
- Impure shell: API calls, database operations, console output

Push I/O to program edges, keep core logic pure.

## Code Quality Standards

### Required Practices

- Use `defer` for cleanup operations
- Context as first parameter in functions
- Set appropriate timeouts for operations
- Validate input early in functions
- Use named constants instead of magic values
- Format with `gofmt` before commit
- Run linters before commit

### Preferred Patterns

- Pure functions: deterministic, no side effects
- Immutability: create new data, don't modify existing
- Composition: build complexity from simple functions
- Side effects at edges: isolate I/O at boundaries
- Zero-value usability: structs work without initialization
- Interface-based design: depend on interfaces, not implementations
- Small focused functions: single responsibility

### Avoid

- Global mutable state
- Mutating function parameters
- Side effects in business logic
- Magic numbers and strings
- Deep nesting (>3 levels)
- Long functions (>50 lines)
- Long files (>500 lines)
- Generic package names
- Ignoring errors
- Logging sensitive data

## Code Review Checklist

### Code Quality

- All tests pass
- Code formatted with `gofmt`
- `golangci-lint` passes
- No Codacy violations

### Size Limits

- Functions ≤50 lines (including tests)
- Files ≤500 lines
- Cyclomatic complexity ≤15

### Functional Programming

- Business logic uses pure functions where possible
- Side effects isolated at program edges
- No mutation of function parameters
- Immutable data structures preferred

### General Best Practices

- Documentation added/updated for exported elements
- Error handling appropriate, wrapped with context
- Logging structured with appropriate level
- No sensitive data in logs/comments
- Named constants instead of magic values
- Test coverage maintained/improved (>80%)

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Best Practices](BEST_PRACTICES.md) - Detailed examples and patterns
