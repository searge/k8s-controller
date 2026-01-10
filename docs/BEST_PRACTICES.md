# Best Practices

Detailed guide with examples for developing k8s-controller.
See [CODING_GUIDELINES.md](CODING_GUIDELINES.md) for quick reference.

## Table of Contents

- [Code Organization](#code-organization)
- [Naming](#naming)
- [Documentation](#documentation)
- [Error Handling](#error-handling)
- [Logging](#logging)
- [Testing](#testing)
- [Function Design](#function-design)
- [Common Pitfalls](#common-pitfalls)

## Code Organization

### Good Package Structure

```go
// Package k8s provides Kubernetes client functionality.
package k8s

import (
    "context"
    "fmt"

    "k8s.io/client-go/kubernetes"

    "github.com/Searge/k8s-controller/pkg/logger"
)

type Client struct {
    clientset kubernetes.Interface
}
```

### Bad Package Structure

```go
// Missing package documentation
package utils  // Generic name

import (
    "github.com/Searge/k8s-controller/pkg/logger"  // Wrong order
    "context"
    "k8s.io/client-go/kubernetes"
    "fmt"
)

type Helper struct {  // Generic name
    stuff interface{}  // Vague
}
```

## Naming

### Variables

**Good:**

```go
var deploymentName string
var ctx context.Context
var err error

for i, deployment := range deployments {  // Short in short scope
    process(deployment)
}
```

**Bad:**

```go
var n string  // Too short for package scope
var context context.Context  // Shadows package
var error error  // Shadows built-in

for index, dep := range deployments {  // Unnecessarily verbose
    process(dep)
}
```

### Functions

**Good:**

```go
func createK8sClient() (*Client, error)
func validateNamespace(ns string) error
func formatAge(duration time.Duration) string
```

**Bad:**

```go
func CreateKubernetesClientFromConfiguration() (*Client, error)  // Too verbose
func validate(s string) error  // Too vague
func FormatAge(d time.Duration) string  // Unexported should use camelCase
```

### Constants

**Good:**

```go
const (
    defaultTimeout = 30 * time.Second
    maxRetries     = 3

    // Exported constants
    DefaultNamespace = "default"
    MaxNameLength    = 63
)
```

**Bad:**

```go
const TIMEOUT = 30  // Wrong case, magic number
const max_retries = 3  // Snake case
```

### Interfaces

**Good:**

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}
```

**Bad:**

```go
type IReader interface {  // Don't use "I" prefix
    Read(p []byte) (n int, err error)
}

type ReaderInterface interface {  // Redundant suffix
    Read(p []byte) (n int, err error)
}
```

## Documentation

### Package Documentation

**Good:**

```go
// Package logger provides structured logging functionality using zerolog.
// It offers a simple interface for initializing and configuring
// application-wide logging with support for multiple log levels.
package logger
```

**Bad:**

```go
// logger package
package logger

// OR missing entirely
package logger
```

### Function Documentation

**Good:**

```go
// CreateClient creates a new Kubernetes client with the provided configuration.
// It returns a Client instance that wraps the clientset with additional functionality.
//
// The client automatically handles kubeconfig loading from default locations
// or the path specified in config.KubeconfigPath.
func CreateClient(config ClientConfig, logger zerolog.Logger) (*Client, error) {
```

**Bad:**

```go
// Creates client
func CreateClient(config ClientConfig, logger zerolog.Logger) (*Client, error) {

// OR
// This function creates a new client. First it loads the config,
// then it creates the clientset, then it returns the client.
func CreateClient(config ClientConfig, logger zerolog.Logger) (*Client, error) {
```

### Inline Comments

**Good - Explains WHY:**

```go
// Skip logging for version command - it should be clean output
if cmd.Use == "version" {
    return
}

// Use in-cluster config for pods running inside Kubernetes
if inClusterConfig, err := rest.InClusterConfig(); err == nil {
    return inClusterConfig, nil
}
```

**Bad - Explains WHAT:**

```go
// Check if command use equals version
if cmd.Use == "version" {
    return
}

// Try to get in cluster config
if inClusterConfig, err := rest.InClusterConfig(); err == nil {
    return inClusterConfig, nil
}
```

## Error Handling

### Error Wrapping

**Good - Wraps with context:**

```go
if err != nil {
    return fmt.Errorf("failed to create Kubernetes client: %w", err)
}

if err := validateInput(data); err != nil {
    return fmt.Errorf("invalid input: %w", err)
}
```

**Bad - Loses error chain:**

```go
if err != nil {
    return fmt.Errorf("error: %s", err.Error())
}

if err != nil {
    return errors.New("failed to create client")
}
```

### Enhanced Error Messages

**Good - Specific and actionable:**

```go
func enhanceK8sError(err error) error {
    if strings.Contains(err.Error(), "connection refused") {
        return fmt.Errorf("failed to connect to Kubernetes API server - "+
            "is the cluster running and accessible? %w", err)
    }
    if strings.Contains(err.Error(), "forbidden") {
        return fmt.Errorf("insufficient permissions to list deployments - "+
            "check your RBAC configuration: %w", err)
    }
    return fmt.Errorf("failed to list deployments: %w", err)
}
```

**Bad - Generic, not helpful:**

```go
func enhanceK8sError(err error) error {
    return fmt.Errorf("error occurred: %w", err)
}
```

### Early Returns

**Good - Early returns reduce nesting:**

```go
func processDeployment(name string) error {
    if name == "" {
        return fmt.Errorf("name cannot be empty")
    }

    deployment, err := fetchDeployment(name)
    if err != nil {
        return fmt.Errorf("failed to fetch: %w", err)
    }

    if err := validateDeployment(deployment); err != nil {
        return fmt.Errorf("invalid deployment: %w", err)
    }

    return nil
}
```

**Bad - Deep nesting:**

```go
func processDeployment(name string) error {
    if name != "" {
        deployment, err := fetchDeployment(name)
        if err == nil {
            if validateDeployment(deployment) == nil {
                return nil
            } else {
                return fmt.Errorf("invalid deployment")
            }
        } else {
            return fmt.Errorf("failed to fetch: %w", err)
        }
    } else {
        return fmt.Errorf("name cannot be empty")
    }
}
```

## Logging

### Structured Logging

**Good - Structured with context:**

```go
log.Info().
    Str("namespace", namespace).
    Int("count", len(deployments)).
    Str("output", outputFormat).
    Msg("Successfully listed deployments")

log.Error().
    Err(err).
    Str("kubeconfig", kubeconfigPath).
    Msg("Failed to load configuration")
```

**Bad - Unstructured string concatenation:**

```go
log.Info().Msg("Listed " + strconv.Itoa(len(deployments)) +
    " deployments in namespace " + namespace)

log.Error().Msg(fmt.Sprintf("Error: %v", err))
```

### Log Levels

**Good - Appropriate levels:**

```go
log.Debug().Str("path", configPath).Msg("Loading configuration")
log.Info().Msg("Client created successfully")
log.Warn().Err(err).Msg("Failed to close client")
log.Error().Err(err).Msg("Failed to list deployments")
log.Fatal().Err(err).Msg("Cannot continue without database")
```

**Bad - Wrong levels:**

```go
log.Info().Str("path", configPath).Msg("Loading configuration")  // Too verbose
log.Debug().Msg("Client created successfully")  // Should be Info
log.Error().Err(err).Msg("Failed to close client")  // Should be Warn
```

### Logging Context

**Good - Entry/exit of major operations:**

```go
func (c *Client) ListDeployments(
    ctx context.Context,
    opts ListDeploymentsOptions,
) ([]DeploymentInfo, error) {
    c.logger.Debug().
        Str("namespace", opts.Namespace).
        Msg("Listing deployments")

    // ... implementation ...

    c.logger.Info().
        Int("count", len(deployments)).
        Msg("Successfully listed deployments")

    return deployments, nil
}
```

**Bad - No logging context:**

```go
func (c *Client) ListDeployments(
    ctx context.Context,
    opts ListDeploymentsOptions,
) ([]DeploymentInfo, error) {
    log.Info().Msg("Listing")
    // ... implementation ...
    log.Info().Msg("Done")
    return deployments, nil
}
```

## Testing

### Table-Driven Tests

**Good - Comprehensive table-driven test:**

```go
func TestValidateNamespace(t *testing.T) {
    tests := []struct {
        name      string
        namespace string
        wantErr   bool
    }{
        {
            name:      "valid namespace",
            namespace: "default",
            wantErr:   false,
        },
        {
            name:      "empty namespace is valid",
            namespace: "",
            wantErr:   false,
        },
        {
            name:      "namespace too long",
            namespace: strings.Repeat("a", 64),
            wantErr:   true,
        },
        {
            name:      "invalid characters",
            namespace: "My-Namespace",
            wantErr:   true,
        },
        {
            name:      "starts with hyphen",
            namespace: "-invalid",
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateNamespace(tt.namespace)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateNamespace() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Bad - Separate test for each case:**

```go
func TestValidNamespace(t *testing.T) {
    err := validateNamespace("default")
    if err != nil {
        t.Error("expected no error")
    }
}

func TestInvalidNamespace(t *testing.T) {
    err := validateNamespace("My-Namespace")
    if err == nil {
        t.Error("expected error")
    }
}
// ... many more separate tests
```

### Test Constants

**Good - Reusable test constants:**

```go
const (
    testImageNginx       = "nginx:1.21"
    testImageRedis       = "redis:6.2"
    testNamespaceDefault = "default"
    testDeploymentName   = "test-deployment"
)

func TestFormatImages(t *testing.T) {
    images := []string{testImageNginx, testImageRedis}
    result := formatImages(images)
    // ...
}
```

**Bad - Magic strings everywhere:**

```go
func TestFormatImages(t *testing.T) {
    images := []string{"nginx:1.21", "redis:6.2"}
    result := formatImages(images)
    // ...
}

func TestOtherFunction(t *testing.T) {
    deployment := createDeployment("nginx:1.21")  // Duplicated
    // ...
}
```

## Function Design

### Breaking Down Complex Functions

**Good - Main function orchestrates smaller functions:**

```go
func runListDeployments() error {
    if err := validateListParameters(); err != nil {
        return err
    }

    client, err := createK8sClient()
    if err != nil {
        return err
    }
    defer closeClient(client)

    deployments, err := fetchDeployments(client)
    if err != nil {
        return err
    }

    return formatDeploymentOutput(deployments, outputFormat)
}

func validateListParameters() error {
    if err := validateOutputFormat(outputFormat); err != nil {
        return fmt.Errorf("invalid output format: %w", err)
    }
    if err := validateNamespace(namespace); err != nil {
        return fmt.Errorf("invalid namespace: %w", err)
    }
    return nil
}
```

**Bad - Monolithic function:**

```go
func runListDeployments() error {
    // Validate output format
    switch outputFormat {
    case "table", "json":
        // ok
    default:
        return fmt.Errorf("unsupported format")
    }

    // Validate namespace
    if namespace != "" && len(namespace) > 63 {
        return fmt.Errorf("namespace too long")
    }
    for _, r := range namespace {
        if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
            return fmt.Errorf("invalid character")
        }
    }

    // Create client
    config := ClientConfig{
        KubeconfigPath: kubeconfigPath,
        Context:        contextName,
    }
    // ... 50 more lines
}
```

### Pure Functions

**Good - Pure function, no side effects:**

```go
func formatAge(duration time.Duration) string {
    if duration < time.Minute {
        return fmt.Sprintf("%ds", int(duration.Seconds()))
    }
    if duration < time.Hour {
        return fmt.Sprintf("%dm", int(duration.Minutes()))
    }
    if duration < 24*time.Hour {
        return fmt.Sprintf("%dh", int(duration.Hours()))
    }
    days := int(duration.Hours() / 24)
    return fmt.Sprintf("%dd", days)
}
```

**Bad - Has side effects:**

```go
var ageCache = make(map[time.Duration]string)

func formatAge(duration time.Duration) string {
    if cached, ok := ageCache[duration]; ok {
        return cached
    }
    result := fmt.Sprintf("%ds", int(duration.Seconds()))
    ageCache[duration] = result  // Mutates global state
    return result
}
```

### Using Config Structs

**Good - Config struct for many parameters:**

```go
type ClientConfig struct {
    KubeconfigPath string
    Context        string
    Timeout        time.Duration
    Retries        int
}

func CreateClient(config ClientConfig, logger zerolog.Logger) (*Client, error) {
    // ...
}

// Usage
client, err := CreateClient(ClientConfig{
    KubeconfigPath: "/path/to/config",
    Context:        "prod-cluster",
    Timeout:        30 * time.Second,
}, logger)
```

**Bad - Too many parameters:**

```go
func CreateClient(
    kubeconfigPath string,
    context string,
    timeout time.Duration,
    retries int,
    logger zerolog.Logger,
) (*Client, error) {
    // ...
}
```

## Common Pitfalls

### Ignoring Errors

**Good - Handle or explicitly ignore:**

```go
if err := client.Close(); err != nil {
    log.Warn().Err(err).Msg("Failed to close client")
}

// Intentionally ignored with comment
_ = writer.Flush()  // Ignore flush errors, data already processed
```

**Bad - Silently ignoring:**

```go
client.Close()
writer.Flush()
```

### Not Using Context

**Good - Accept and use context:**

```go
func (c *Client) ListDeployments(
    ctx context.Context,
    opts ListDeploymentsOptions,
) ([]DeploymentInfo, error) {
    deploymentList, err := c.clientset.AppsV1().
        Deployments(opts.Namespace).
        List(ctx, listOpts)
    if err != nil {
        return nil, err
    }
    return deployments, nil
}
```

**Bad - No context, can't be canceled:**

```go
func (c *Client) ListDeployments(
    opts ListDeploymentsOptions,
) ([]DeploymentInfo, error) {
    deploymentList, err := c.clientset.AppsV1().
        Deployments(opts.Namespace).
        List(context.Background(), listOpts)
    // ...
}
```

### Magic Values

**Good - Named constants:**

```go
const (
    maxNamespaceLength = 63
    minPasswordLength  = 8
    defaultTimeout     = 30 * time.Second
)

if len(namespace) > maxNamespaceLength {
    return fmt.Errorf("namespace too long (max %d)", maxNamespaceLength)
}
```

**Bad - Magic numbers:**

```go
if len(namespace) > 63 {
    return fmt.Errorf("namespace too long")
}

if len(password) < 8 {
    return fmt.Errorf("password too short")
}
```

### Global Mutable State

**Good - Encapsulated state:**

```go
type Cache struct {
    mu    sync.RWMutex
    items map[string]interface{}
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.items[key]
    return val, ok
}
```

**Bad - Global mutable variable:**

```go
var globalCache = make(map[string]interface{})

func GetFromCache(key string) interface{} {
    return globalCache[key]  // Race condition!
}
```

### Deep Nesting

**Good - Flat structure with early returns:**

```go
func processItem(item Item) error {
    if !item.Valid() {
        return fmt.Errorf("invalid item")
    }

    data, err := item.GetData()
    if err != nil {
        return fmt.Errorf("get data: %w", err)
    }

    if err := validateData(data); err != nil {
        return fmt.Errorf("validate: %w", err)
    }

    return saveData(data)
}
```

**Bad - Deeply nested:**

```go
func processItem(item Item) error {
    if item.Valid() {
        data, err := item.GetData()
        if err == nil {
            if validateData(data) == nil {
                if saveData(data) == nil {
                    return nil
                } else {
                    return fmt.Errorf("save failed")
                }
            } else {
                return fmt.Errorf("validation failed")
            }
        } else {
            return fmt.Errorf("get data: %w", err)
        }
    } else {
        return fmt.Errorf("invalid item")
    }
}
```

## Additional Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [100 Go Mistakes and How to Avoid Them](https://github.com/teivah/100-go-mistakes)
