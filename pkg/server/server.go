// Package server provides HTTP server functionality for the k8s-controller application.
// It implements a FastHTTP-based server with health check endpoints and structured logging.
package server

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

// createHandler creates an HTTP handler function with the application's routing logic.
// It accepts a zerolog.Logger for structured logging of HTTP requests and errors.
// The handler supports the following endpoints:
//   - GET /health: Returns a JSON health status response
//   - GET /*: Returns a default greeting message for all other paths
func createHandler(logger zerolog.Logger) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())

		logger.Info().Msgf("Request: %s %s", ctx.Method(), path)

		switch path {
		case "/health":
			ctx.SetStatusCode(200)
			if _, err := fmt.Fprintf(ctx, `{"status":"ok"}`); err != nil {
				logger.Error().Err(err).Msg("Failed to write health response")
			}
		default:
			if _, err := fmt.Fprintf(ctx, "Hello from k8s-controller!"); err != nil {
				logger.Error().Err(err).Msg("Failed to write response")
			}
		}
	}
}

// Start starts the HTTP server on the specified port.
// It creates a FastHTTP server with the application's handler and begins listening
// for incoming requests. The function blocks until the server encounters an error.
//
// Parameters:
//   - port: The TCP port number to bind the server to
//   - logger: A zerolog.Logger instance for structured logging
//
// Returns an error if the server fails to start or encounters a runtime error.
func Start(port int, logger zerolog.Logger) error {
	addr := fmt.Sprintf(":%d", port)

	logger.Info().Msgf("Starting HTTP server on %s", addr)

	handler := createHandler(logger)

	return fasthttp.ListenAndServe(addr, handler)
}
