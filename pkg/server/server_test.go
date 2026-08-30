// Package server contains tests for the HTTP server functionality.
// This file tests the context-bound server lifecycle; handlers_test.go covers
// the endpoints themselves.
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

// startTimeout bounds every wait in these tests: long enough for a slow machine,
// short enough that a hang fails the run instead of stalling it.
const startTimeout = 5 * time.Second

// startServer runs a Server on an OS-assigned port and waits until it is bound.
// It returns the base URL and a stop function that cancels the server and
// asserts Start returned cleanly — which is the leak check every test gets for
// free by using this helper.
func startServer(t *testing.T) (string, func()) {
	t.Helper()

	srv := New(0, &fakeSource{synced: true}, zerolog.New(&bytes.Buffer{}))

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

	status, body, err := get(t, url+"/healthz")
	if err != nil {
		t.Fatalf("GET /healthz against running server: %v", err)
	}
	if status != 200 || !strings.Contains(body, `"ok"`) {
		t.Errorf("GET /healthz = %d %q, want 200 with ok body", status, body)
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

	if _, _, err := get(t, url+"/healthz"); err == nil {
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

	second := New(port, &fakeSource{synced: true}, zerolog.New(&bytes.Buffer{}))
	if err := second.Start(context.Background()); err == nil {
		t.Error("Start() on an occupied port returned nil, want an error")
	}
}

// TestAddrBeforeStart pins the "not started yet" contract.
func TestAddrBeforeStart(t *testing.T) {
	srv := New(0, &fakeSource{synced: true}, zerolog.New(&bytes.Buffer{}))
	if got := srv.Addr(); got != "" {
		t.Errorf(`Addr() before Start = %q, want ""`, got)
	}
}
