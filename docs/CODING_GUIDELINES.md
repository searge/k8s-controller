# Coding Guidelines

Quick reference for code standards in the k8s-controller project. Based on
[Google Go Style Guide](https://google.github.io/styleguide/go/) and
[Effective Go](https://go.dev/doc/effective_go).

For detailed examples and patterns, see [BEST_PRACTICES.md](BEST_PRACTICES.md).

## Code Organization

### Package Structure

- One purpose per package
- Short, lowercase, single-word names
- No generic names (`util`, `common`, `helpers`)
- Required package documentation

### Import Order

1. Standard library
2. Third-party packages
3. Project packages

Separate groups with blank lines.

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

- **Package**: Required for all packages
- **Exported elements**: All functions, types, constants
- **Comments**: Complete sentences, start with element name
- **Explain why**, not what (code shows what)

## Error Handling

- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Check immediately after occurrence
- Use early returns to reduce nesting
- Error messages: lowercase, specific, actionable

## Logging

### Structured Logging with Zerolog

```go
log.Info().
    Str("namespace", namespace).
    Int("count", count).
    Msg("Operation completed")
```

### Log Levels

- **Debug**: Diagnostic details
- **Info**: Normal operations (default)
- **Warn**: Potentially harmful situations
- **Error**: Failures, execution continues
- **Fatal**: Critical errors, app terminates

## Testing

- **Table-driven tests** for comprehensive coverage
- **Constants** for test data to avoid duplication
- **Coverage**: Aim for >80%
- **Tests**: Fast, independent, deterministic
- **Mocks**: Use interfaces for external dependencies

## Function Design

| Guideline             | Standard                         |
| --------------------- | -------------------------------- |
| Lines per function    | ~50 max                          |
| Cyclomatic complexity | 10-15 max                        |
| Parameters            | 3-4 max, use structs for more    |
| Purpose               | Single, clear responsibility     |
| Return values         | Error always last                |

## Code Quality Standards

### Required Practices

- Use `defer` for cleanup operations
- Context as first parameter in functions
- Set appropriate timeouts
- Validate input early
- Use constants over magic values
- Format with `gofmt` before commit
- Run `golangci-lint` before commit

### Preferred Patterns

- Pure functions (no side effects)
- Immutable data structures
- Zero-value usability
- Interface-based design
- Small, focused functions

### Avoid

- Global mutable state
- Magic numbers/strings
- Deep nesting (>3 levels)
- Long functions (>50 lines)
- Generic package names
- Ignoring errors
- Logging sensitive data

## Code Review Checklist

- [ ] All tests pass
- [ ] Code formatted with `gofmt`
- [ ] Linter passes (`golangci-lint`)
- [ ] Documentation added/updated
- [ ] Error handling appropriate
- [ ] Logging structured and appropriate
- [ ] No sensitive data in logs/comments
- [ ] Function complexity within limits
- [ ] Test coverage maintained/improved

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Best Practices](BEST_PRACTICES.md) - Detailed examples and patterns
