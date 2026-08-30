// Package cmd contains tests for the CLI commands.
// This file tests the serve command definition and the informer-plus-server lifecycle.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// TestRunServeStopsOnCancel is the lifecycle regression test, in four acts:
// the HTTP server answers, readiness follows the informer's sync, the cache
// actually serves data the informer saw, and one cancellation stops everything
// — runServe must not return until it has. The middle acts exist because a
// runServe that silently dropped either component would still "return cleanly
// on cancel".
func TestRunServeStopsOnCancel(t *testing.T) {
	// The fake clientset is seeded before the informer starts, so the initial
	// list carries this deployment into the cache.
	seeded := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "seeded-api"},
	}
	clientset := fake.NewClientset(seeded)
	port := freePort(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- runServe(ctx, clientset, port) }()

	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	client := &http.Client{Timeout: time.Second}
	getStatus := func(path string) (int, string) {
		resp, err := client.Get(base + path)
		if err != nil {
			return 0, ""
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(body)
	}

	// Act one: liveness answers over real TCP.
	waitUntil(t, "the HTTP server to answer /healthz", func() bool {
		status, _ := getStatus("/healthz")
		return status == http.StatusOK
	})

	// Act two: readiness turns 200 once the informer has synced.
	waitUntil(t, "/readyz to report the cache as synced", func() bool {
		status, _ := getStatus("/readyz")
		return status == http.StatusOK
	})

	// Act three: the cache serves what the informer saw, end to end — fake
	// clientset through informer through cache through handler to JSON.
	status, body := getStatus("/deployments")
	if status != http.StatusOK {
		t.Fatalf("GET /deployments = %d, want 200 after sync", status)
	}
	var list struct {
		Count int `json:"count"`
		Items []struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(body), &list); err != nil {
		t.Fatalf("GET /deployments body is not JSON: %v (%q)", err, body)
	}
	if list.Count != 1 || len(list.Items) != 1 || list.Items[0].Name != "seeded-api" {
		t.Errorf("GET /deployments = %+v, want exactly the seeded deployment", list)
	}

	// Act four: one cancel stops both, and runServe waits for both.
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
