// Package server contains tests for the HTTP server functionality.
// This file tests the server's HTTP handlers and routing logic.
package server

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

// TestServerHandlers tests the HTTP request routing and response generation
// for all supported endpoints. It uses an in-memory listener to avoid
// binding to real network ports during testing.
func TestServerHandlers(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "health endpoint",
			path:           "/health",
			expectedStatus: 200,
			expectedBody:   `{"status":"ok"}`,
		},
		{
			name:           "default endpoint",
			path:           "/",
			expectedStatus: 200,
			expectedBody:   "Hello from k8s-controller!",
		},
		{
			name:           "unknown endpoint",
			path:           "/unknown",
			expectedStatus: 200,
			expectedBody:   "Hello from k8s-controller!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create in-memory listener for testing
			ln := fasthttputil.NewInmemoryListener()
			defer func() {
				if err := ln.Close(); err != nil {
					t.Errorf("Failed to close listener: %v", err)
				}
			}()

			// Start server using our actual handler logic
			go func() {
				// Create a test logger that writes to stderr
				logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
				handler := createHandler(logger)
				if err := fasthttp.Serve(ln, handler); err != nil {
					t.Errorf("Failed to serve: %v", err)
				}
			}()

			// Give server time to start
			time.Sleep(10 * time.Millisecond)

			// Create client with custom dialer for in-memory connection
			client := &fasthttp.Client{
				Dial: func(_ string) (net.Conn, error) {
					return ln.Dial()
				},
			}

			// Prepare HTTP request
			req := fasthttp.AcquireRequest()
			resp := fasthttp.AcquireResponse()
			defer fasthttp.ReleaseRequest(req)
			defer fasthttp.ReleaseResponse(resp)

			req.SetRequestURI(tt.path)
			req.Header.SetMethod("GET")
			req.Header.SetHost("localhost") // FastHTTP requires Host header

			// Execute the request
			err := client.Do(req, resp)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}

			// Verify response status code
			if resp.StatusCode() != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode())
			}

			// Verify response body content
			body := string(resp.Body())
			if body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

// ExampleStart demonstrates how to start the HTTP server.
// This example shows the basic usage of the Start function with
// a logger and port configuration.
func ExampleStart() {
	// This example shows how to start the server
	// Note: In real usage, this would block until the server stops

	// Start server on port 8080
	// err := Start(8080, logger)
	// if err != nil {
	//     log.Fatal(err)
	// }

	fmt.Println("Server would start on :8080")
	// Output: Server would start on :8080
}
