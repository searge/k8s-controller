# API Documentation

## HTTP Endpoints

`serve` runs the HTTP server beside a deployment informer. Handlers read the
informer's cache and never call the Kubernetes API themselves; readiness is
therefore a statement about the cache, not about the process.

Every response carries an `X-Request-ID` header, and the matching value appears
as `request_id` in the request log line.

### Liveness

**Endpoint:** `GET /healthz`

The process is up and serving. Deliberately ignores the cache: a pod that is
alive but not yet ready must fail readiness, not liveness, or it gets restarted
instead of waited for.

| Status | Body | Meaning |
| --- | --- | --- |
| `200` | `{"status":"ok"}` | Always, while the process serves |

### Readiness

**Endpoint:** `GET /readyz`

Ready once the informer cache has synced with the cluster.

| Status | Body | Meaning |
| --- | --- | --- |
| `200` | `{"status":"ok"}` | Cache synced; answers reflect the cluster |
| `503` | `{"error":"cache not synced"}` | Still syncing, or the API server is unreachable |

### Deployments

**Endpoint:** `GET /deployments[?namespace=<name>]`

Deployments from the informer cache. Without `namespace`, all namespaces the
informer watches. Refuses with `503` until the cache has synced, because an
empty answer from an unsynced cache is indistinguishable from a cluster with no
deployments.

```json
{
  "items": [
    {
      "name": "example",
      "namespace": "default",
      "replicas": {"desired": 3, "available": 3, "ready": 3, "updated": 3},
      "age": 86400000000000,
      "images": ["nginx:1.27"],
      "created_at": "2026-08-01T10:00:00Z"
    }
  ],
  "count": 1
}
```

| Status | Meaning |
| --- | --- |
| `200` | List served from the cache |
| `503` | Cache not synced yet |
| `500` | Cache read failed; details go to the log, not the client |

### Anything else

Unknown paths return `404` with `{"error":"not found"}`. Methods other than GET
return `405` with an `Allow: GET` header.

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/readyz
curl -s 'http://localhost:8080/deployments?namespace=default' | jq .
```

## CLI Commands

### Global Flags

- `--log-level string` - Set logging level (debug, info, warn, error, fatal, panic) (default "info")

### serve

Start the HTTP server and the deployment informer under one lifecycle. Both
stop together on SIGINT or SIGTERM. Requires cluster access.

```bash
k8s-controller serve [flags]
```

**Flags:**

- `--port int` - Port to run the server on (default 8080)
- `--kubeconfig string` - Path to kubeconfig (default: `$KUBECONFIG` or `~/.kube/config`)
- `--context string` - Kubeconfig context (default: current context)

### list deployments

List deployments with a live API call (no cache, no informer).

```bash
k8s-controller list deployments [flags]
```

**Flags:**

- `-n, --namespace string` - Namespace (default: all namespaces)
- `-o, --output string` - Output format: table, json, yaml (default "table")
- `-l, --selector string` - Label selector
- `--kubeconfig string`, `--context string` - As in serve
- `--timeout int` - Timeout in seconds (default 30)

### version

Print the version number.

```bash
k8s-controller version
```

## Logging

Structured logging with [zerolog](https://github.com/rs/zerolog), written to
stderr in console format. client-go's own log output (klog) is routed through
the same logger, so reflector and cache-sync failures appear as structured
lines rather than as raw klog text.

HTTP requests are logged after completion with `request_id`, `method`, `path`,
`status` and `duration`.

## Error Handling

- Non-2xx HTTP responses carry `{"error": "..."}`; internal details stay in the log.
- CLI exit code 1 on failure, 0 on success.

## Security Considerations

⚠️ **Warning**: This is a development/learning project. The current implementation:

- Has no authentication or authorization
- Binds to all network interfaces by default
- Does not use HTTPS
- Has no rate limiting

Do not use in production without proper security measures.
