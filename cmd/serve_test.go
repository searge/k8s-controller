// Package cmd contains tests for the CLI commands.
// This file tests the serve command definition and the informer-plus-server lifecycle.
package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/fake"
)

// TestServeCommandDefined verifies that the serve command is properly defined
// and configured with the expected flags and properties.
func TestServeCommandDefined(t *testing.T) {
	if serveCmd == nil {
		t.Fatal("serveCmd should be defined")
	}

	if serveCmd.Use != "serve" {
		t.Errorf("expected command use 'serve', got %s", serveCmd.Use)
	}

	for _, name := range []string{"port", "kubeconfig", "context"} {
		if serveCmd.Flags().Lookup(name) == nil {
			t.Errorf("expected %q flag to be defined", name)
		}
	}
}

// serveTestTimeout bounds every wait in the lifecycle test.
const serveTestTimeout = 5 * time.Second

// freePort asks the OS for a port and releases it, so runServe can bind it.
// The gap between releasing and rebinding is a race in principle; in practice
// this is the standard technique and collisions do not happen in test runs.
func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find a free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	if err := ln.Close(); err != nil {
		t.Fatalf("failed to release the probe listener: %v", err)
	}
	return port
}

// waitUntil polls cond until it holds or the timeout fails the test.
func waitUntil(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(serveTestTimeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// TestRunServeStopsOnCancel is the milestone's regression test, in three acts:
// the HTTP server actually answers, the informer actually watches, and one
// cancellation stops them both — runServe must not return until it has. The
// first two assertions exist because a runServe that silently dropped either
// component would still "return cleanly on cancel".
func TestRunServeStopsOnCancel(t *testing.T) {
	clientset := fake.NewClientset()
	port := freePort(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- runServe(ctx, clientset, port) }()

	// Act one: the server answers over real TCP.
	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	client := &http.Client{Timeout: time.Second}
	waitUntil(t, "the HTTP server to answer /health", func() bool {
		resp, err := client.Get(url)
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode == http.StatusOK
	})

	// Act two: the informer is watching. The fake clientset records every call,
	// so the informer's list-and-watch shows up in the action log.
	waitUntil(t, "the informer to start watching deployments", func() bool {
		for _, action := range clientset.Actions() {
			if action.GetVerb() == "watch" && action.GetResource().Resource == "deployments" {
				return true
			}
		}
		return false
	})

	// Act three: one cancel stops both, and runServe waits for both.
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("runServe() after cancellation = %v, want nil", err)
		}
	case <-time.After(serveTestTimeout):
		t.Error("runServe() did not return after cancellation: informer or server leaked")
	}
}

// TestValidatePort covers the port range check.
func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{name: "valid common port", port: 8080, wantErr: false},
		{name: "lowest valid port", port: 1, wantErr: false},
		{name: "highest valid port", port: 65535, wantErr: false},
		{name: "zero is rejected", port: 0, wantErr: true},
		{name: "negative is rejected", port: -1, wantErr: true},
		{name: "above range is rejected", port: 65536, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePort(%d) error = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
		})
	}
}
