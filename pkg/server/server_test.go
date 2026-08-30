// Package server contains tests for the HTTP server functionality.
// This file tests the HTTP handlers and the context-bound server lifecycle.
package server

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

// HelloMessage is the default message returned by the server.
const HelloMessage = "Hello from k8s-controller!"

// startTimeout bounds every wait in these tests: long enough for a slow machine,
// short enough that a hang fails the run instead of stalling it.
const startTimeout = 5 * time.Second

// TestCreateHandler tests the HTTP request routing and response generation
// for all supported endpoints. It directly tests the handler function
// without network dependencies.
func TestCreateHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "health endpoint GET",
			path:           "/health",
			method:         "GET",
			expectedStatus: 200,
			expectedBody:   `{"status":"ok"}`,
		},
		{
			name:           "root endpoint",
			path:           "/",
			method:         "GET",
			expectedStatus: 200,
			expectedBody:   HelloMessage,
		},
		{
			name:           "unknown endpoint",
			path:           "/unknown",
			method:         "GET",
			expectedStatus: 200,
			expectedBody:   HelloMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture log output
			var logBuf bytes.Buffer
			logger := zerolog.New(&logBuf).With().Timestamp().Logger()

			// Create handler
			handler := createHandler(logger)

			// Create fasthttp context
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetRequestURI(tt.path)
			ctx.Request.Header.SetMethod(tt.method)

			// Call handler
			handler(ctx)

			// Verify response status code
			if ctx.Response.StatusCode() != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, ctx.Response.StatusCode())
			}

			// Verify response body content
			body := string(ctx.Response.Body())
			if body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}

			// Verify that request was logged
			logOutput := logBuf.String()
			expectedLogContent := fmt.Sprintf("%s %s", tt.method, tt.path)
			if !strings.Contains(logOutput, expectedLogContent) {
				t.Errorf("Expected log to contain %q, got %q", expectedLogContent, logOutput)
			}
		})
	}
}

// startServer runs a Server on an OS-assigned port and waits until it is bound.
// It returns the base URL and a stop function that cancels the server and
// asserts Start returned cleanly — which is the leak check every test gets for
// free by using this helper.
func startServer(t *testing.T) (string, func()) {
	t.Helper()

	srv := New(0, zerolog.New(&bytes.Buffer{}))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Start(ctx) }()

	deadline := time.Now().Add(startTimeout)
	for srv.Addr() == "" {
		if time.Now().After(deadline) {
			cancel()
			t.Fatal("server did not bind a port within the timeout")
		}
		time.Sleep(2 * time.Millisecond)
	}

	// The server binds every interface, so Addr() reports "[::]:PORT" — not an
	// address a client can dial. Take the port and aim at loopback instead.
	_, port, err := net.SplitHostPort(srv.Addr())
	if err != nil {
		cancel()
		t.Fatalf("unexpected listen address %q: %v", srv.Addr(), err)
	}

	stop := func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Start() returned error after cancellation: %v", err)
			}
		case <-time.After(startTimeout):
			t.Error("Start() did not return after cancellation: serve goroutine leaked")
		}
	}

	return "http://127.0.0.1:" + port, stop
}

// get performs one HTTP GET against a running server.
func get(t *testing.T, url string) (int, string, error) {
	t.Helper()

	client := &fasthttp.Client{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")

	if err := client.Do(req, resp); err != nil {
		return 0, "", err
	}
	return resp.StatusCode(), string(resp.Body()), nil
}

// TestServerServesUntilCancelled covers the whole lifecycle: bind, serve a real
// request over TCP, shut down on cancel, and release the port.
func TestServerServesUntilCancelled(t *testing.T) {
	url, stop := startServer(t)

	status, body, err := get(t, url+"/health")
	if err != nil {
		t.Fatalf("GET /health against running server: %v", err)
	}
	if status != 200 || body != `{"status":"ok"}` {
		t.Errorf("GET /health = %d %q, want 200 with ok body", status, body)
	}

	stop()

	// The port must be free again: shutdown that leaves the listener bound is
	// exactly the defect the old ListenAndServe path had.
	addr := strings.TrimPrefix(url, "http://")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("port still bound after shutdown: %v", err)
	}
	if err := ln.Close(); err != nil {
		t.Errorf("failed to close probe listener: %v", err)
	}

	if _, _, err := get(t, url+"/health"); err == nil {
		t.Error("server still answering after shutdown")
	}
}

// TestServerListenFailure makes sure a port that cannot be bound surfaces as an
// error from Start rather than as a log line and a hung process.
func TestServerListenFailure(t *testing.T) {
	url, stop := startServer(t)
	defer stop()

	// A second server on the same port must fail to listen.
	addr := strings.TrimPrefix(url, "http://")
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("unexpected addr %q: %v", addr, err)
	}

	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		t.Fatalf("unexpected port %q: %v", portStr, err)
	}

	second := New(port, zerolog.New(&bytes.Buffer{}))
	if err := second.Start(context.Background()); err == nil {
		t.Error("Start() on an occupied port returned nil, want an error")
	}
}

// TestAddrBeforeStart pins the "not started yet" contract.
func TestAddrBeforeStart(t *testing.T) {
	srv := New(0, zerolog.New(&bytes.Buffer{}))
	if got := srv.Addr(); got != "" {
		t.Errorf(`Addr() before Start = %q, want ""`, got)
	}
}
