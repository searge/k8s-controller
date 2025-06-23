# API Documentation

## HTTP Endpoints

The k8s-controller provides a simple HTTP API for health checking and basic operations.

### Health Check

**Endpoint:** `GET /health`

**Description:** Returns the health status of the application.

**Response:**

```json
{
  "status": "ok"
}
```

**Status Codes:**

- `200 OK` - Service is healthy

**Example:**

```bash
curl http://localhost:8080/health
```

### Default Endpoint

**Endpoint:** `GET /*` (all other paths)

**Description:** Returns a default greeting message.

**Response:**

```text
Hello from k8s-controller!
```

**Status Codes:**

- `200 OK` - Request processed successfully

**Example:**

```bash
curl http://localhost:8080/
curl http://localhost:8080/any-path
```

## CLI Commands

### Global Flags

- `--log-level string` - Set logging level (debug, info, warn, error, fatal, panic) (default "info")

### Commands

#### serve

Start the HTTP server.

```bash
k8s-controller serve [flags]
```

**Flags:**

- `--port int` - Port to run the server on (default 8080)

**Examples:**

```bash
# Start server on default port 8080
k8s-controller serve

# Start server on custom port with debug logging
k8s-controller serve --port=9090 --log-level=debug
```

#### version

Print the version number of k8s-controller.

```bash
k8s-controller version
```

**Example output:**

```bash
k8s-controller version v0.1.0
```

## Configuration

### Environment Variables

Currently, the application doesn't use environment variables for configuration. All configuration is done via CLI flags.

### Logging

The application uses structured logging with [zerolog](https://github.com/rs/zerolog). Log levels can be configured using the `--log-level` flag.

Available log levels:

- `debug` - Detailed debug information
- `info` - General information (default)
- `warn` - Warning messages
- `error` - Error messages
- `fatal` - Fatal errors (application exits)
- `panic` - Panic-level errors (application panics)

### Server Configuration

- **Port**: Configurable via `--port` flag (default: 8080)
- **Bind Address**: Currently binds to all interfaces (0.0.0.0)
- **Protocol**: HTTP (HTTPS not yet implemented)

## Error Handling

### HTTP Errors

Currently, the server doesn't return specific HTTP error codes for client errors. All endpoints return 200 OK for valid requests.

### CLI Errors

- Exit code 1 - Command execution failed or server startup failed
- Exit code 0 - Successful execution

## Security Considerations

⚠️ **Warning**: This is a development/learning project. The current implementation:

- Has no authentication or authorization
- Binds to all network interfaces by default
- Does not use HTTPS
- Has no rate limiting

Do not use in production without proper security measures.

## Future Enhancements

This documentation will be updated as new features are added:

- Kubernetes client integration
- Authentication and authorization
- HTTPS support
- Metrics endpoints
- Custom resource management
