// This file holds the HTTP routing and endpoint handlers; server.go owns the
// listener lifecycle.
package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/Searge/k8s-controller/pkg/k8s"
)

// contentTypeJSON is the Content-Type every endpoint responds with.
const contentTypeJSON = "application/json"

// DeploymentSource is what the handlers read deployments from. In production it
// is the informer's cache; in tests it is a fake. It is deliberately the only
// path handlers have to cluster data -- there is no clientset to reach for, so
// "serves from cache, never from the API" holds by construction.
type DeploymentSource interface {
	// List returns deployments from the cache. Empty namespace means all.
	List(namespace string, selector labels.Selector) ([]*appsv1.Deployment, error)
	// HasSynced reports whether the cache holds the cluster's current state.
	HasSynced() bool
}

// deploymentList is the JSON shape of GET /deployments.
type deploymentList struct {
	Items []k8s.DeploymentInfo `json:"items"`
	Count int                  `json:"count"`
}

// errorBody is the JSON shape of every non-2xx response.
type errorBody struct {
	Error string `json:"error"`
}

// newHandler builds the routing handler with request logging around it.
//
// Every request gets an id, returned in X-Request-ID and attached to the log
// line, so one slow or failing request can be followed from client to log
// without grepping by timestamp.
func newHandler(source DeploymentSource, logger zerolog.Logger) fasthttp.RequestHandler {
	routes := func(ctx *fasthttp.RequestCtx) {
		// Every endpoint is a read; anything but GET is refused before dispatch,
		// with the Allow header RFC 9110 requires alongside a 405.
		if !ctx.IsGet() {
			ctx.Response.Header.Set("Allow", "GET")
			writeJSON(ctx, fasthttp.StatusMethodNotAllowed, errorBody{Error: "method not allowed"}, logger)
			return
		}

		switch string(ctx.Path()) {
		case "/healthz":
			handleHealthz(ctx)
		case "/readyz":
			handleReadyz(ctx, source, logger)
		case "/deployments":
			handleDeployments(ctx, source, logger, time.Now())
		default:
			writeJSON(ctx, fasthttp.StatusNotFound, errorBody{Error: "not found"}, logger)
		}
	}

	return func(ctx *fasthttp.RequestCtx) {
		requestID := newRequestID()
		ctx.Response.Header.Set("X-Request-ID", requestID)

		start := time.Now()
		routes(ctx)

		logger.Info().
			Str("request_id", requestID).
			Str("method", string(ctx.Method())).
			Str("path", string(ctx.Path())).
			Int("status", ctx.Response.StatusCode()).
			Dur("duration", time.Since(start)).
			Msg("Request handled")
	}
}

// newRequestID returns 8 random bytes as hex: unique enough to correlate a
// request with a log line, with no dependency for the privilege.
func newRequestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand failing means the platform is broken in ways a request
		// id cannot fix; degrade to a constant rather than refuse requests.
		return "0000000000000000"
	}
	return hex.EncodeToString(b[:])
}

// handleHealthz is liveness: the process is up and serving. Deliberately knows
// nothing about the cache -- a pod that is alive but not yet ready must fail
// readiness, not liveness, or it gets restarted instead of waited for.
func handleHealthz(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType(contentTypeJSON)
	ctx.SetBodyString(`{"status":"ok"}`)
}

// handleReadyz is readiness: 200 only once the informer cache has synced.
// Before that the process is healthy but its answers would be arbitrary.
// The 503 carries the same {"error": ...} shape as every other non-2xx
// response, so one parser covers the whole API.
func handleReadyz(ctx *fasthttp.RequestCtx, source DeploymentSource, logger zerolog.Logger) {
	if !source.HasSynced() {
		writeJSON(ctx, fasthttp.StatusServiceUnavailable, errorBody{Error: "cache not synced"}, logger)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType(contentTypeJSON)
	ctx.SetBodyString(`{"status":"ok"}`)
}

// handleDeployments serves the deployment list from the cache.
//
// It refuses with 503 until the cache is synced: an empty answer from an
// unsynced cache is indistinguishable from "the cluster has no deployments",
// and of the two failure modes, a client that retries beats a client that
// trusts a wrong answer.
func handleDeployments(ctx *fasthttp.RequestCtx, source DeploymentSource, logger zerolog.Logger, now time.Time) {
	if !source.HasSynced() {
		writeJSON(ctx, fasthttp.StatusServiceUnavailable, errorBody{Error: "cache not synced yet"}, logger)
		return
	}

	namespace := string(ctx.QueryArgs().Peek("namespace"))

	deployments, err := source.List(namespace, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list deployments from cache")
		writeJSON(ctx, fasthttp.StatusInternalServerError, errorBody{Error: "failed to list deployments"}, logger)
		return
	}

	items := make([]k8s.DeploymentInfo, 0, len(deployments))
	for _, d := range deployments {
		items = append(items, k8s.NewDeploymentInfo(d, now))
	}

	writeJSON(ctx, fasthttp.StatusOK, deploymentList{Items: items, Count: len(items)}, logger)
}

// writeJSON marshals body into the response with the given status.
func writeJSON(ctx *fasthttp.RequestCtx, status int, body any, logger zerolog.Logger) {
	ctx.SetStatusCode(status)
	ctx.SetContentType(contentTypeJSON)
	if err := json.NewEncoder(ctx).Encode(body); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
	}
}
