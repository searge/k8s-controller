// Package server provides HTTP server functionality for the k8s-controller application.
// It implements a FastHTTP-based server with health check endpoints and structured logging.
//
// The server's lifetime is bound to a context: Start blocks until the context is
// cancelled, then shuts down gracefully and returns only when nothing is still
// serving. That is what lets `serve` run the server and the informer side by side
// and stop both with one signal.
package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

// shutdownTimeout bounds how long Start waits for in-flight requests after its
// context is cancelled. Past it, remaining connections are dropped.
const shutdownTimeout = 5 * time.Second

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
			ctx.SetContentType("application/json")
			if _, err := fmt.Fprintf(ctx, `{"status":"ok"}`); err != nil {
				logger.Error().Err(err).Msg("Failed to write health response")
			}
		default:
			ctx.SetContentType("text/plain")
			if _, err := fmt.Fprintf(ctx, "Hello from k8s-controller!"); err != nil {
				logger.Error().Err(err).Msg("Failed to write response")
			}
		}
	}
}

// Server is the application's HTTP server with a context-bound lifetime.
//
// A Server is single-use: after Start has returned, the underlying fasthttp
// server has been shut down and a new Server must be built to serve again.
type Server struct {
	port   int
	logger zerolog.Logger
	srv    *fasthttp.Server

	// mu guards addr, which Start writes from its goroutine and Addr reads
	// from whoever is asking (in practice: tests that bound port 0).
	mu   sync.Mutex
	addr net.Addr
}

// New builds a server for the given port. Port 0 asks the OS for a free port;
// Addr reports which one was granted once Start has bound it.
func New(port int, logger zerolog.Logger) *Server {
	return &Server{
		port:   port,
		logger: logger.With().Str("component", "http-server").Logger(),
		srv:    &fasthttp.Server{Handler: createHandler(logger)},
	}
}

// Start listens on the configured port and serves until ctx is cancelled.
//
// It returns nil after a clean shutdown. A listen failure, a serve failure, or
// a shutdown that exceeds shutdownTimeout each return an error. By the time
// Start returns, the serve goroutine has exited: a caller that saw Start return
// knows nothing is still running, which is what makes a leak a test failure
// rather than a suspicion.
func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	s.mu.Lock()
	s.addr = ln.Addr()
	s.mu.Unlock()

	s.logger.Info().Str("addr", ln.Addr().String()).Msg("Starting HTTP server")

	serveErr := make(chan error, 1)
	go func() { serveErr <- s.srv.Serve(ln) }()

	select {
	case err := <-serveErr:
		return fmt.Errorf("http server failed: %w", err)
	case <-ctx.Done():
	}

	s.logger.Info().Msg("Context cancelled, shutting down HTTP server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := s.srv.ShutdownWithContext(shutdownCtx); err != nil {
		return fmt.Errorf("http server shutdown failed: %w", err)
	}

	// Shutdown closed the listener, so Serve is about to return; wait for it
	// rather than leaving the goroutine to finish after Start has returned.
	if err := <-serveErr; err != nil {
		return fmt.Errorf("http server failed during shutdown: %w", err)
	}

	s.logger.Info().Msg("HTTP server stopped")
	return nil
}

// Addr reports the address the server is bound to, or "" before Start has
// bound the listener. Exists so a test can start on port 0 and learn the port.
func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.addr == nil {
		return ""
	}
	return s.addr.String()
}
